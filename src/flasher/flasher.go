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

package flasher

import (
	"encoding/binary"
	"go.bug.st/serial"
	"time"
)

type FlasherError struct {
	err string
}

func (e FlasherError) Error() string {
	return e.err
}

type Flasher struct {
	port *serial.SerialPort
}

// Ping the programmer to see if it is alive.
// Also check if the version of the programmer protocol match the uploader
func (flasher *Flasher) Hello() error {
	// "HELLO" command
	flasher.sendCommand(0x99, 0x11223344, 0x55667788, 0)

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Receive response
	res := make([]byte, 65535)
	n, err := flasher.port.Read(res)
	if err != nil {
		return err
	}
	// flush eventual leftover from the rx buffer
	res = res[n-6 : n]

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
	flasher.sendCommand(0x50, 0, 0, 0)

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
	flasher.sendCommand(0x01, address, length, 0)

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
	flasher.sendCommand(0x02, address, 0, uint16(len(buffer)))

	// send payload
	if _, err := flasher.port.Write(buffer); err != nil {
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
	flasher.sendCommand(0x03, address, length, 0)

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

func (flasher *Flasher) sendCommand(command byte, address uint32, val uint32, length uint16) error {
	if err := binary.Write(flasher.port, binary.BigEndian, command); err != nil {
		return err
	}
	if err := binary.Write(flasher.port, binary.BigEndian, address); err != nil {
		return err
	}
	if err := binary.Write(flasher.port, binary.BigEndian, val); err != nil {
		return err
	}
	if err := binary.Write(flasher.port, binary.BigEndian, length); err != nil {
		return err
	}
	return nil
}

func Open(portName string) (*Flasher, error) {
	mode := &serial.Mode{
		BaudRate: 115200,
	}

	port, err := serial.OpenPort(portName, mode)
	if err != nil {
		return nil, err
	}

	flasher := &Flasher{
		port: port,
	}

	// Wait for the complete reset of the board
	time.Sleep(2500 * time.Millisecond)

	return flasher, err
}
