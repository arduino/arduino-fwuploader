/*
  FirmwareUploader
  Copyright (c) 2021 Arduino LLC.  All right reserved.

  This library is free software; you can redistribute it and/or
  modify it under the terms of the GNU Lesser General Public
  License as published by the Free Software Foundation; either
  version 2.1 of the License, or (at your option) any later version.

  This library is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
  Lesser General Public License for more details.

  You should have received a copy of the GNU Lesser General Public
  License along with this library; if not, write to the Free Software
  Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
*/

package flasher

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"go.bug.st/serial"
)

// NewNinaFlasher creates an new instance of NinaFlasher
func NewNinaFlasher(portAddress string) (*NinaFlasher, error) {
	port, err := openSerial(portAddress)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	time.Sleep(2 * time.Second)
	f := &NinaFlasher{port: port}
	payloadSize, err := f.getMaximumPayloadSize()
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	if payloadSize < 1024 {
		err = fmt.Errorf("programmer reports %d as maximum payload size (1024 is needed)", payloadSize)
		logrus.Error(err)
		return nil, err
	}
	f.payloadSize = int(payloadSize)
	return f, nil
}

type NinaFlasher struct {
	port        serial.Port
	payloadSize int
}

// FlashFirmware in board connected to port using data from firmwareFile
func (f *NinaFlasher) FlashFirmware(firmwareFile *paths.Path, flasherOut io.Writer) error {
	logrus.Infof("Flashing firmware %s", firmwareFile)
	flasherOut.Write([]byte(fmt.Sprintf("Flashing firmware %s", firmwareFile)))
	flasherOut.Write([]byte(fmt.Sprintln()))
	if err := f.hello(); err != nil {
		logrus.Error(err)
		return err
	}

	logrus.Debugf("Reading file %s", firmwareFile)
	data, err := firmwareFile.ReadFile()
	if err != nil {
		logrus.Error(err)
		return err
	}

	logrus.Debugf("Flashing chunks")
	firmwareOffset := 0x0000
	if err := f.flashChunk(firmwareOffset, data); err != nil {
		logrus.Error(err)
		return err
	}

	logrus.Debugf("Checking md5")
	if err := f.md5sum(data); err != nil {
		logrus.Error(err)
		return err
	}
	logrus.Infof("Flashed all the things")
	flasherOut.Write([]byte("Flashed all the things"))
	flasherOut.Write([]byte(fmt.Sprintln()))
	return err //should be nil
}

func (f *NinaFlasher) FlashCertificates(certificatePaths *paths.PathList, URLs []string) error {
	var certificatesData []byte
	for _, certPath := range *certificatePaths {
		logrus.Infof("Converting and flashing certificate %s", certPath)

		data, err := f.certificateFromFile(certPath)
		if err != nil {
			return err
		}
		certificatesData = append(certificatesData, data...)
	}

	for _, URL := range URLs {
		logrus.Infof("Converting and flashing certificate from %s", URL)
		data, err := f.certificateFromURL(URL)
		if err != nil {
			return err
		}
		certificatesData = append(certificatesData, data...)
	}

	certificatesDataLimit := 0x20000
	if len(certificatesData) > certificatesDataLimit {
		err := fmt.Errorf("certificates data %d exceeds limit of %d bytes", len(certificatesData), certificatesDataLimit)
		logrus.Error(err)
		return err
	}

	// Pad certificatesData to flash page
	for len(certificatesData)%int(f.payloadSize) != 0 {
		certificatesData = append(certificatesData, 0)
	}

	certificatesOffset := 0x10000
	return f.flashChunk(certificatesOffset, certificatesData)
}

func (f *NinaFlasher) certificateFromFile(certificateFile *paths.Path) ([]byte, error) {
	data, err := certificateFile.ReadFile()
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	cert, err := x509.ParseCertificate(data)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}), nil
}

func (f *NinaFlasher) certificateFromURL(URL string) ([]byte, error) {
	config := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", URL, config)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	defer conn.Close()

	if err := conn.Handshake(); err != nil {
		logrus.Error(err)
		return nil, err
	}

	peerCertificates := conn.ConnectionState().PeerCertificates
	if len(peerCertificates) == 0 {
		err = fmt.Errorf("no peer certificates found at %s", URL)
		logrus.Error(err)
		return nil, err
	}

	rootCertificate := peerCertificates[len(peerCertificates)-1]
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rootCertificate.Raw}), nil
}

// Close the port used by this flasher
func (f *NinaFlasher) Close() error {
	return f.port.Close()
}

// Ping the programmer to see if it is alive.
// Also check if the version of the programmer protocol match the uploader
func (f *NinaFlasher) hello() error {
	// "HELLO" command
	err := f.sendCommand(CommandData{
		Command: 0x99,
		Address: 0x11223344,
		Value:   0x55667788,
		Payload: nil,
	})
	if err != nil {
		logrus.Error(err)
		return err
	}

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Receive response
	res := make([]byte, 65535)
	n, err := f.port.Read(res)
	if err != nil {
		logrus.Error(err)
		return err
	}
	// flush eventual leftover from the rx buffer
	if n >= 6 {
		res = res[n-6 : n]
	}

	if res[0] != 'v' {
		err = FlasherError{err: "Programmer is not responding"}
		logrus.Error(err)
		return err
	}
	if string(res) != "v10000" {
		// TODO: Do we really need this check? What is it trying to verify?
		err = FlasherError{err: fmt.Sprintf("Programmer version mismatch, v10000 needed: %s", res)}
		logrus.Error(err)
		return err
	}
	return nil
}

// flashChunk flashes a chunk of data
func (f *NinaFlasher) flashChunk(offset int, buffer []byte) error {
	chunkSize := int(f.payloadSize)
	bufferLength := len(buffer)

	if err := f.erase(uint32(offset), uint32(bufferLength)); err != nil {
		logrus.Error(err)
		return err
	}

	for i := 0; i < bufferLength; i += chunkSize {
		logrus.Debugf("Flashing chunk: %s%%", strconv.Itoa((i*100)/bufferLength))
		start := i
		end := i + chunkSize
		if end > bufferLength {
			end = bufferLength
		}
		if err := f.write(uint32(offset+i), buffer[start:end]); err != nil {
			logrus.Error(err)
			return err
		}
	}

	return nil
}

// getMaximumPayloadSize asks the board the maximum payload size
func (f *NinaFlasher) getMaximumPayloadSize() (uint16, error) {
	// "MAX_PAYLOAD_SIZE" command
	err := f.sendCommand(CommandData{
		Command: 0x50,
		Address: 0,
		Value:   0,
		Payload: nil,
	})
	if err != nil {
		logrus.Error(err)
		return 0, err
	}

	// Receive response
	res := make([]byte, 2)
	if err := f.serialFillBuffer(res); err != nil {
		logrus.Error(err)
		return 0, err
	}
	return (uint16(res[0]) << 8) + uint16(res[1]), nil
}

// serialFillBuffer fills buffer with data read from the serial port
func (f *NinaFlasher) serialFillBuffer(buffer []byte) error {
	read := 0
	for read < len(buffer) {
		n, err := f.port.Read(buffer[read:])
		if err != nil {
			logrus.Error(err)
			return err
		}
		if n == 0 {
			err = FlasherError{err: "Serial port closed unexpectedly"}
			logrus.Error(err)
			return err
		}
		read += n
	}
	return nil
}

// sendCommand sends the data over serial port to connected board
func (f *NinaFlasher) sendCommand(data CommandData) error {
	logrus.Debugf("sending command data %s", data)
	buff := new(bytes.Buffer)
	if err := binary.Write(buff, binary.BigEndian, data.Command); err != nil {
		err = fmt.Errorf("writing command: %s", err)
		logrus.Error(err)
		return err
	}
	if err := binary.Write(buff, binary.BigEndian, data.Address); err != nil {
		err = fmt.Errorf("writing address: %s", err)
		logrus.Error(err)
		return err
	}
	if err := binary.Write(buff, binary.BigEndian, data.Value); err != nil {
		err = fmt.Errorf("writing value: %s", err)
		logrus.Error(err)
		return err
	}
	var length uint16
	if data.Payload == nil {
		length = 0
	} else {
		length = uint16(len(data.Payload))
	}
	if err := binary.Write(buff, binary.BigEndian, length); err != nil {
		err = fmt.Errorf("writing payload length: %s", err)
		logrus.Error(err)
		return err
	}
	if data.Payload != nil {
		buff.Write(data.Payload)
	}
	bufferData := buff.Bytes()
	for {
		sent, err := f.port.Write(bufferData)
		if err != nil {
			err = fmt.Errorf("writing data: %s", err)
			logrus.Error(err)
			return err
		}
		if sent == len(bufferData) {
			break
		}
		logrus.Debugf("Sent %d bytes out of %d", sent, len(bufferData))
		bufferData = bufferData[sent:]
	}
	return nil
}

// read a block of flash memory
func (f *NinaFlasher) read(address uint32, length uint32) ([]byte, error) {
	// "FLASH_READ" command
	err := f.sendCommand(CommandData{
		Command: 0x01,
		Address: address,
		Value:   length,
		Payload: nil,
	})
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	// Receive response
	result := make([]byte, length)
	if err := f.serialFillBuffer(result); err != nil {
		logrus.Error(err)
		return nil, err
	}
	ack := make([]byte, 2)
	if err := f.serialFillBuffer(ack); err != nil {
		logrus.Error(err)
		return nil, err
	}
	if string(ack) != "OK" {
		err = FlasherError{err: fmt.Sprintf("Missing ack on read: %s, result: %s", ack, result)}
		logrus.Error(err)
		return nil, err
	}
	return result, nil
}

// write a block of flash memory
func (f *NinaFlasher) write(address uint32, buffer []byte) error {
	// "FLASH_WRITE" command
	err := f.sendCommand(CommandData{
		Command: 0x02,
		Address: address,
		Value:   0,
		Payload: buffer,
	})
	if err != nil {
		logrus.Error(err)
		return err
	}

	// wait acknowledge
	ack := make([]byte, 2)
	if err := f.serialFillBuffer(ack); err != nil {
		logrus.Error(err)
		return err
	}
	if string(ack) != "OK" {
		err = FlasherError{err: fmt.Sprintf("Missing ack on write: %s", ack)}
		logrus.Error(err)
		return err
	}
	return nil
}

// erase a block of flash memory
func (f *NinaFlasher) erase(address uint32, length uint32) error {
	// "FLASH_ERASE" command
	err := f.sendCommand(CommandData{
		Command: 0x03,
		Address: address,
		Value:   length,
		Payload: nil,
	})
	if err != nil {
		logrus.Error(err)
		return err
	}

	// wait acknowledge
	ack := make([]byte, 2)
	if err := f.serialFillBuffer(ack); err != nil {
		logrus.Error(err)
		return err
	}
	if string(ack) != "OK" {
		err = FlasherError{err: fmt.Sprintf("Missing ack on erase: %s", ack)}
		logrus.Error(err)
		return err
	}
	return nil
}

func (f *NinaFlasher) md5sum(data []byte) error {
	hasher := md5.New()
	hasher.Write(data)

	// Get md5sum
	err := f.sendCommand(CommandData{
		Command: 0x04,
		Address: 0,
		Value:   uint32(len(data)),
		Payload: nil,
	})
	if err != nil {
		logrus.Error(err)
		return err
	}

	// Wait acknowledge
	ack := make([]byte, 2)
	if err := f.serialFillBuffer(ack); err != nil {
		logrus.Error(err)
		return err
	}
	if string(ack) != "OK" {
		err := FlasherError{err: fmt.Sprintf("Missing ack on md5sum: %s", ack)}
		logrus.Error(err)
		return err
	}

	// Wait md5
	md5sumfromdevice := make([]byte, 16)
	if err := f.serialFillBuffer(md5sumfromdevice); err != nil {
		return err
	}

	md5sum := hasher.Sum(nil)
	logrus.Debugf("md5 read from device %s", md5sumfromdevice)
	logrus.Debugf("md5 of data %s", md5sum)

	for i := 0; i < 16; i++ {
		if md5sum[i] != md5sumfromdevice[i] {
			err := FlasherError{err: "MD5sum failed"}
			logrus.Error(err)
			return err
		}
	}

	return nil
}
