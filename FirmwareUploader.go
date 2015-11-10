/*
  FirmwareUploader.go - A firmware uploader for the WiFi101 module.
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
package main

import "go.bug.st/serial"
import _ "fmt"
import "time"
import "log"
import "encoding/binary"
import "os"
import "flag"

type UpdaterError struct {
	err string
}

func (e UpdaterError) Error() string {
	return e.err
}

// Fill buffer with data coming from serial port.
// Blocks until the buffer is full.
func serialFillBuffer(port *serial.SerialPort, buffer []byte) error {
	read := 0
	for read < len(buffer) {
		n, err := port.Read(buffer[read:])
		if err != nil {
			return err
		}
		if n == 0 {
			return &UpdaterError{err: "Serial port closed unexpectedly"}
		}
		read += n
	}
	return nil
}

func programmerSendCommand(port *serial.SerialPort, command byte, address uint32, val uint32, length uint16) error {
	if err := binary.Write(port, binary.BigEndian, command); err != nil {
		return err
	}
	if err := binary.Write(port, binary.BigEndian, address); err != nil {
		return err
	}
	if err := binary.Write(port, binary.BigEndian, val); err != nil {
		return err
	}
	if err := binary.Write(port, binary.BigEndian, length); err != nil {
		return err
	}
	return nil
}

// Ping the programmer to see if it is alive.
// Also check if the version of the programmer protocol match the uploader
func Hello(port *serial.SerialPort) error {
	// "HELLO" command
	programmerSendCommand(port, 0x99, 0x11223344, 0x55667788, 0)

	// Receive response
	res := make([]byte, 65535)
	n, err := port.Read(res)
	if err != nil {
		return err
	}

	// flush eventual leftover from the rx buffer
	res = res[n-6 : n]

	if res[0] != 'v' {
		return &UpdaterError{err: "Programmer is not responding"}
	}
	if string(res) != "v10000" {
		return &UpdaterError{err: "Programmer version mismatch: " + string(res) + " (needed v10000)"}
	}
	return nil
}

// Get maximum payload size for upload.
func GetMaximumPayloadSize(port *serial.SerialPort) (uint16, error) {
	// "MAX_PAYLOAD_SIZE" command
	programmerSendCommand(port, 0x50, 0, 0, 0)

	// Receive response
	res := make([]byte, 2)
	if err := serialFillBuffer(port, res); err != nil {
		return 0, err
	}
	return (uint16(res[0]) << 8) + uint16(res[1]), nil
}

// Read a block of flash memory
func FlashRead(port *serial.SerialPort, address uint32, length uint32) ([]byte, error) {
	// "FLASH_READ" command
	programmerSendCommand(port, 0x01, address, length, 0)

	// Receive response
	result := make([]byte, length)
	if err := serialFillBuffer(port, result); err != nil {
		return nil, err
	}
	ack := make([]byte, 2)
	if err := serialFillBuffer(port, ack); err != nil {
		return nil, err
	}
	if string(ack) != "OK" {
		return nil, &UpdaterError{err: "Error during FlashRead()"}
	}
	return result, nil
}

// Write a block of flash memory
func FlashWrite(port *serial.SerialPort, address uint32, buffer []byte) error {
	// "FLASH_WRITE" command
	programmerSendCommand(port, 0x02, address, 0, uint16(len(buffer)))

	// send payload
	if _, err := port.Write(buffer); err != nil {
		return err
	}

	// wait acknowledge
	ack := make([]byte, 2)
	if err := serialFillBuffer(port, ack); err != nil {
		return err
	}
	if string(ack) != "OK" {
		return &UpdaterError{err: "Error during FlashWrite()"}
	}
	return nil
}

// Erase a block of flash memory
func FlashErase(port *serial.SerialPort, address uint32, length uint32) error {
	// "FLASH_ERASE" command
	programmerSendCommand(port, 0x03, address, length, 0)

	// wait acknowledge
	ack := make([]byte, 2)
	if err := serialFillBuffer(port, ack); err != nil {
		return err
	}
	if string(ack) != "OK" {
		return &UpdaterError{err: "Error during FlashErase()"}
	}
	return nil
}

var portName string

func init() {
	flag.StringVar(&portName, "port", "", "serial port to use for flashing")
}

func main() {
	flag.Parse()

	mode := &serial.Mode{
		BaudRate: 115200,
	}
	port, err := serial.OpenPort(portName, mode)
	if err != nil {
		log.Fatal(err)
	}

	// Wait for the complete reset of the board
	time.Sleep(2500 * time.Millisecond)

	// Synchronize with programmer
	if err := Hello(port); err != nil {
		log.Fatal(err)
	}

	// Check maximum supported payload size
	payloadSize, err := GetMaximumPayloadSize(port)
	if err != nil {
		log.Fatal(err)
	}
	if payloadSize < 1024 {
		log.Fatalf("Programmer reports %d as maximum payload size (1024 is needed)", payloadSize)
	}

	//if err := FlashWrite(port, 1024, make([]byte, payloadSize)); err != nil {
	//	log.Fatal(err)
	//}

	for i := 0; i < 256; i++ {
		data, err := FlashRead(port, uint32(i*1024), 1024)
		if err != nil {
			log.Fatal(err.Error())
		}
		os.Stdout.Write(data)
	}

	port.Close()
}
