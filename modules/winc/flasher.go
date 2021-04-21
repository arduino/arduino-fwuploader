/*
  flasher.go - A firmware uploader for the WiFi101 module.
  Copyright (c) 2015 Arduino LLC.  All right reserved.

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

package winc

import (
	"bytes"
	"encoding/binary"
	"log"
	"time"

	"github.com/arduino/FirmwareUploader/utils"
	"go.bug.st/serial"
)

type FlasherError struct {
	err string
}

func (e FlasherError) Error() string {
	return e.err
}

type Flasher struct {
	port serial.Port
}

// Ping the programmer to see if it is alive.
// Also check if the version of the programmer protocol match the uploader
func (flasher *Flasher) Hello() error {
	// "HELLO" command
	if err := flasher.sendCommand(0x99, 0x11223344, 0x55667788, nil); err != nil {
		return err
	}

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Receive response
	res := make([]byte, 65535)
	n, err := flasher.port.Read(res)
	if err != nil {
		return err
	}
	// flush eventual leftover from the rx buffer
	if n >= 6 {
		res = res[n-6 : n]
	}

	if res[0] != 'v' {
		return &FlasherError{err: "Programmer is not responding"}
	}
	if string(res) != "v10000" {
		return &FlasherError{err: "Programmer version mismatch: " + string(res) + " (needed v10000)"}
	}
	return nil
}

func (flasher *Flasher) Close() error {
	return flasher.port.Close()
}

// Get maximum payload size for upload.
func (flasher *Flasher) GetMaximumPayloadSize() (uint16, error) {
	// "MAX_PAYLOAD_SIZE" command
	if err := flasher.sendCommand(0x50, 0, 0, nil); err != nil {
		return 0, err
	}

	// Receive response
	res := make([]byte, 2)
	if err := flasher.serialFillBuffer(res); err != nil {
		return 0, err
	}
	return (uint16(res[0]) << 8) + uint16(res[1]), nil
}

// Read a block of flash memory
func (flasher *Flasher) Read(address uint32, length uint32) ([]byte, error) {
	// "FLASH_READ" command
	if err := flasher.sendCommand(0x01, address, length, nil); err != nil {
		return nil, err
	}

	// Receive response
	result := make([]byte, length)
	if err := flasher.serialFillBuffer(result); err != nil {
		return nil, err
	}
	ack := make([]byte, 2)
	if err := flasher.serialFillBuffer(ack); err != nil {
		return nil, err
	}
	if string(ack) != "OK" {
		return nil, &FlasherError{err: "Error during FlashRead()"}
	}
	return result, nil
}

// Write a block of flash memory
func (flasher *Flasher) Write(address uint32, buffer []byte) error {
	// "FLASH_WRITE" command
	if err := flasher.sendCommand(0x02, address, 0, buffer); err != nil {
		return err
	}

	// wait acknowledge
	ack := make([]byte, 2)
	if err := flasher.serialFillBuffer(ack); err != nil {
		return err
	}
	if string(ack) != "OK" {
		return &FlasherError{err: "Error during FlashWrite()"}
	}
	return nil
}

// Erase a block of flash memory
func (flasher *Flasher) Erase(address uint32, length uint32) error {
	// "FLASH_ERASE" command
	if err := flasher.sendCommand(0x03, address, length, nil); err != nil {
		return err
	}

	log.Printf("Erasing %d bytes from address 0x%X\n", length, address)

	// wait acknowledge
	ack := make([]byte, 2)
	if err := flasher.serialFillBuffer(ack); err != nil {
		return err
	}
	if string(ack) != "OK" {
		return &FlasherError{err: "Error during FlashErase()"}
	}
	return nil
}

// Fill buffer with data coming from serial port.
// Blocks until the buffer is full.
func (flasher *Flasher) serialFillBuffer(buffer []byte) error {
	read := 0
	for read < len(buffer) {
		n, err := flasher.port.Read(buffer[read:])
		if err != nil {
			return err
		}
		if n == 0 {
			return &FlasherError{err: "Serial port closed unexpectedly"}
		}
		read += n
	}
	return nil
}

func (flasher *Flasher) sendCommand(command byte, address uint32, val uint32, payload []byte) error {
	buff := new(bytes.Buffer)
	if err := binary.Write(buff, binary.BigEndian, command); err != nil {
		return err
	}
	if err := binary.Write(buff, binary.BigEndian, address); err != nil {
		return err
	}
	if err := binary.Write(buff, binary.BigEndian, val); err != nil {
		return err
	}
	var length uint16
	if payload == nil {
		length = 0
	} else {
		length = uint16(len(payload))
	}
	if err := binary.Write(buff, binary.BigEndian, length); err != nil {
		return err
	}
	if payload != nil {
		buff.Write(payload)
	}

	data := buff.Bytes()
	for {
		sent, err := flasher.port.Write(data)
		if err != nil {
			return err
		}
		if sent == len(data) {
			break
		}
		// fmt.Println("HEY! sent", sent, "out of", len(data))
		data = data[sent:]
	}
	return nil
}

func OpenFlasher(portName string) (*Flasher, error) {

	port, err := utils.OpenSerial(portName)
	if err != nil {
		return nil, &FlasherError{err: "Error opening serial port. " + err.Error()}
	}

	flasher := &Flasher{
		port: port,
	}

	return flasher, err
}
