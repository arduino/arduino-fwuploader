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
	"encoding/binary"
	"fmt"
	"time"

	"github.com/arduino/go-paths-helper"
	"go.bug.st/serial"
)

// NewNinaFlasher creates an new instance of NinaFlasher
func NewNinaFlasher(portAddress string) (*NinaFlasher, error) {
	port, err := openSerial(portAddress)
	if err != nil {
		return nil, err
	}
	f := &NinaFlasher{port: port}
	payloadSize, err := f.getMaximumPayloadSize()
	if err != nil {
		return nil, err
	}
	if payloadSize < 1024 {
		return nil, fmt.Errorf("programmer reports %d as maximum payload size (1024 is needed)", payloadSize)
	}
	f.payloadSize = int(payloadSize)
	return f, nil
}

type NinaFlasher struct {
	port        serial.Port
	payloadSize int
}

// FlashFirmware in board connected to port using data from firmwareFile
func (f *NinaFlasher) FlashFirmware(firmwareFile *paths.Path) error {
	if err := f.hello(); err != nil {
		return err
	}

	data, err := firmwareFile.ReadFile()
	if err != nil {
		return err
	}

	firmwareOffset := 0x0000
	if err := f.flashChunk(firmwareOffset, data); err != nil {
		return err
	}

	return f.md5sum(data)
}

func (f *NinaFlasher) FlashCertificates(certificatePaths *paths.PathList) error {
	// TODO
	return nil
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
		return err
	}

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Receive response
	res := make([]byte, 65535)
	n, err := f.port.Read(res)
	if err != nil {
		return err
	}
	// flush eventual leftover from the rx buffer
	if n >= 6 {
		res = res[n-6 : n]
	}

	if res[0] != 'v' {
		return FlasherError{err: "Programmer is not responding"}
	}
	if string(res) != "v10000" {
		// TODO: Do we really need this check? What is it trying to verify?
		return FlasherError{err: fmt.Sprintf("Programmer version mismatch, v10000 needed: %s", res)}
	}
	return nil
}

// flashChunk flashes a chunk of data
func (f *NinaFlasher) flashChunk(offset int, buffer []byte) error {
	chunkSize := int(f.payloadSize)
	bufferLength := len(buffer)

	if err := f.erase(uint32(offset), uint32(bufferLength)); err != nil {
		return err
	}

	for i := 0; i < bufferLength; i += chunkSize {
		// fmt.Printf("\rFlashing: " + strconv.Itoa((i*100)/bufferLength) + "%%")
		start := i
		end := i + chunkSize
		if end > bufferLength {
			end = bufferLength
		}
		if err := f.write(uint32(offset+i), buffer[start:end]); err != nil {
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
		return 0, err
	}

	// Receive response
	res := make([]byte, 2)
	if err := f.serialFillBuffer(res); err != nil {
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
			return err
		}
		if n == 0 {
			return FlasherError{err: "Serial port closed unexpectedly"}
		}
		read += n
	}
	return nil
}

// sendCommand sends the data over serial port to connected board
func (f *NinaFlasher) sendCommand(data CommandData) error {
	buff := new(bytes.Buffer)
	if err := binary.Write(buff, binary.BigEndian, data.Command); err != nil {
		return err
	}
	if err := binary.Write(buff, binary.BigEndian, data.Address); err != nil {
		return err
	}
	if err := binary.Write(buff, binary.BigEndian, data.Value); err != nil {
		return err
	}
	var length uint16
	if data.Payload == nil {
		length = 0
	} else {
		length = uint16(len(data.Payload))
	}
	if err := binary.Write(buff, binary.BigEndian, length); err != nil {
		return err
	}
	if data.Payload != nil {
		buff.Write(data.Payload)
	}
	bufferData := buff.Bytes()
	for {
		sent, err := f.port.Write(bufferData)
		if err != nil {
			return err
		}
		if sent == len(bufferData) {
			break
		}
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
		return nil, err
	}

	// Receive response
	result := make([]byte, length)
	if err := f.serialFillBuffer(result); err != nil {
		return nil, err
	}
	ack := make([]byte, 2)
	if err := f.serialFillBuffer(ack); err != nil {
		return nil, err
	}
	if string(ack) != "OK" {
		return nil, FlasherError{err: fmt.Sprintf("Missing ack on read: %s, result: %s", ack, result)}
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
		return err
	}

	// wait acknowledge
	ack := make([]byte, 2)
	if err := f.serialFillBuffer(ack); err != nil {
		return err
	}
	if string(ack) != "OK" {
		return FlasherError{err: fmt.Sprintf("Missing ack on write: %s", ack)}
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
		return err
	}

	// wait acknowledge
	ack := make([]byte, 2)
	if err := f.serialFillBuffer(ack); err != nil {
		return err
	}
	if string(ack) != "OK" {
		return FlasherError{err: fmt.Sprintf("Missing ack on erase: %s", ack)}
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
		return err
	}

	// Wait acknowledge
	ack := make([]byte, 2)
	if err := f.serialFillBuffer(ack); err != nil {
		return err
	}
	if string(ack) != "OK" {
		return FlasherError{err: fmt.Sprintf("Missing ack on md5sum: %s", ack)}
	}

	// Wait md5
	md5sum := make([]byte, 16)
	if err := f.serialFillBuffer(md5sum); err != nil {
		return err
	}

	md5sumfromdevice := hasher.Sum(nil)

	for i := 0; i < 16; i++ {
		if md5sumfromdevice[i] != md5sum[i] {
			return &FlasherError{err: "MD5sum failed"}
		}
	}

	// log.Println("Integrity ok")

	return nil
}
