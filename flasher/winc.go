/*
	arduino-fwuploader
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
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"go.bug.st/serial"
)

func NewWincFlasher(portAddress string, baudRate int) (*WincFlasher, error) {
	port, err := OpenSerial(portAddress, baudRate)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	f := &WincFlasher{port: port}
	payloadSize, err := f.getMaximumPayloadSize()
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	if payloadSize < 1024 {
		return nil, fmt.Errorf("programmer reports %d as maximum payload size (1024 is needed)", payloadSize)
	}
	f.payloadSize = int(payloadSize)
	return f, nil
}

type WincFlasher struct {
	port             serial.Port
	payloadSize      int
	progressCallback func(int)
}

func (f *WincFlasher) FlashFirmware(firmwareFile *paths.Path, flasherOut io.Writer) error {
	logrus.Infof("Flashing firmware %s", firmwareFile)
	flasherOut.Write([]byte(fmt.Sprintf("Flashing firmware %s\n", firmwareFile)))
	data, err := firmwareFile.ReadFile()
	if err != nil {
		logrus.Error(err)
		return err
	}
	firmwareOffset := 0x0000
	if err = f.flashChunk(firmwareOffset, data); err != nil {
		logrus.Error(err)
		return err
	}
	logrus.Infof("Flashed all the things")
	flasherOut.Write([]byte("Flashing progress: 100%\n"))
	return nil
}

func (f *WincFlasher) FlashCertificates(certificatePaths *paths.PathList, URLs []string, flasherOut io.Writer) error {
	var certificatesData []byte
	certificatesNumber := 0
	for _, certPath := range *certificatePaths {
		logrus.Infof("Converting and flashing certificate %s", certPath)
		flasherOut.Write([]byte(fmt.Sprintf("Converting and flashing certificate %s\n", certPath)))

		data, err := f.certificateFromFile(certPath)
		if err != nil {
			return err
		}
		certificatesData = append(certificatesData, data...)
		certificatesNumber++
	}

	for _, URL := range URLs {
		logrus.Infof("Converting and flashing certificate from %s", URL)
		flasherOut.Write([]byte(fmt.Sprintf("Converting and flashing certificate from %s\n", URL)))
		data, err := f.certificateFromURL(URL)
		if err != nil {
			return err
		}
		certificatesData = append(certificatesData, data...)
		certificatesNumber++
	}

	certificatesOffset := 0x4000
	if err := f.flashChunk(certificatesOffset, certificatesData); err != nil {
		logrus.Error(err)
		return err
	}
	logrus.Infof("Flashed all the things")
	flasherOut.Write([]byte("Flashed all the things\n"))
	return nil
}

func (f *WincFlasher) certificateFromFile(certificateFile *paths.Path) ([]byte, error) {
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

	return f.getCertificateData(cert)
}

func (f *WincFlasher) certificateFromURL(URL string) ([]byte, error) {
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
	return f.getCertificateData(rootCertificate)
}

func (f *WincFlasher) getCertificateData(cert *x509.Certificate) ([]byte, error) {
	b := []byte{}
	nameSHA1Bytes, err := calculateNameSha1(cert)
	if err != nil {
		return nil, err
	}

	notBeforeBytes, err := convertTime(cert.NotBefore)
	if err != nil {
		return nil, err
	}

	notAfterBytes, err := convertTime(cert.NotAfter)
	if err != nil {
		return nil, err
	}

	rsaPublicKey := *cert.PublicKey.(*rsa.PublicKey)

	rsaModulusNBytes := modulusN(rsaPublicKey)
	rsaPublicExponentBytes := publicExponent(rsaPublicKey)

	rsaModulusNLenBytes := uint16ToBytes(len(rsaModulusNBytes))
	rsaPublicExponentLenBytes := uint16ToBytes(len(rsaPublicExponentBytes))

	b = append(b, nameSHA1Bytes...)
	b = append(b, rsaModulusNLenBytes...)
	b = append(b, rsaPublicExponentLenBytes...)
	b = append(b, notBeforeBytes...)
	b = append(b, notAfterBytes...)
	b = append(b, rsaModulusNBytes...)
	b = append(b, rsaPublicExponentBytes...)
	for (len(b) & 3) != 0 {
		b = append(b, 0xff) // padding
	}
	return b, nil
}

func (f *WincFlasher) Close() error {
	return f.port.Close()
}

func (f *WincFlasher) hello() error {
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

func (f *WincFlasher) write(address uint32, buffer []byte) error {
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

func (f *WincFlasher) flashChunk(offset int, buffer []byte) error {
	bufferLength := len(buffer)

	if err := f.erase(uint32(offset), uint32(bufferLength)); err != nil {
		logrus.Error(err)
		return err
	}

	for i := 0; i < bufferLength; i += f.payloadSize {
		progress := ((i * 100) / bufferLength)
		logrus.Debugf("Flashing chunk: %d%%", progress)
		if f.progressCallback != nil {
			f.progressCallback(progress)
		}
		start := i
		end := i + f.payloadSize
		if end > bufferLength {
			end = bufferLength
		}
		if err := f.write(uint32(offset+i), buffer[start:end]); err != nil {
			logrus.Error(err)
			return err
		}
	}

	var flashData []byte
	for i := 0; i < bufferLength; i += f.payloadSize {
		readLength := f.payloadSize
		if (i + f.payloadSize) > bufferLength {
			readLength = bufferLength % f.payloadSize
		}

		data, err := f.read(uint32(offset+i), uint32(readLength))
		if err != nil {
			logrus.Error(err)
			return err
		}

		flashData = append(flashData, data...)
	}

	if !bytes.Equal(buffer, flashData) {
		err := errors.New("flash data does not match written")
		logrus.Error(err)
		return err
	}

	return nil
}

func (f *WincFlasher) getMaximumPayloadSize() (uint16, error) {
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

func (f *WincFlasher) serialFillBuffer(buffer []byte) error {
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

func (f *WincFlasher) sendCommand(data CommandData) error {
	logrus.Debugf("sending command data %s", data)
	buff := new(bytes.Buffer)
	if err := binary.Write(buff, binary.BigEndian, data.Command); err != nil {
		logrus.Error(err)
		return err
	}
	if err := binary.Write(buff, binary.BigEndian, data.Address); err != nil {
		logrus.Error(err)
		return err
	}
	if err := binary.Write(buff, binary.BigEndian, data.Value); err != nil {
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

// Read a block of flash memory
func (f *WincFlasher) read(address uint32, length uint32) ([]byte, error) {
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
		err = FlasherError{err: fmt.Sprintf("Missing ack on read: %s", ack)}
		logrus.Error(err)
		return nil, err
	}
	return result, nil
}

// Erase a block of flash memory
func (f *WincFlasher) erase(address uint32, length uint32) error {
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

	logrus.Debugf("Erasing %d bytes from address 0x%X\n", length, address)

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

func (f *WincFlasher) SetProgressCallback(callback func(progress int)) {
	f.progressCallback = callback
}
