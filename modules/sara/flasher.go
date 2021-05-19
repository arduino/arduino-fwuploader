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

package sara

import (
	"log"
	"strings"
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

func (flasher *Flasher) Hello() error {
	flasher.Expect("ATE0", "OK", 100)
	flasher.Expect("ATE0", "OK", 100)
	flasher.Expect("ATE0", "OK", 100)
	_, err := flasher.Expect("AT", "OK", 100)
	return err
}

func (flasher *Flasher) ExpectMinBytes(buffer string, response string, timeout int, min_bytes int) (string, error) {
	err := flasher.sendCommand([]byte(buffer + "\r\n"))
	if err != nil {
		return "", err
	}

	log.Println("Sending " + buffer)

	// Wait a bit
	// time.Sleep(time.Duration(timeout) * time.Millisecond)

	// Receive response
	var res []byte
	n := 0

	start := time.Now()

	for (time.Since(start) < time.Duration(timeout)*time.Millisecond && !strings.Contains(string(res), response)) || (len(res) < min_bytes) {
		data := 0
		partial := make([]byte, 65535)
		data, err = flasher.port.Read(partial)
		res = append(res, partial[:data]...)
		n += data
		if err != nil {
			return "", err
		}
	}

	log.Println(string(res))

	if !strings.Contains(string(res), response) {
		return string(res), &FlasherError{err: "Expected " + response + ", got " + string(res)}
	}
	return string(res), nil
}

func (flasher *Flasher) Expect(buffer string, response string, timeout int) (string, error) {
	return flasher.ExpectMinBytes(buffer, response, timeout, 0)
}

func (flasher *Flasher) Close() error {
	return flasher.port.Close()
}

func (flasher *Flasher) GetFwVersion() (string, error) {
	return flasher.ExpectMinBytes("ATI9", "05.06,A.02.", 100, 25)
}

// Write a block of flash memory
func (flasher *Flasher) Write(address uint32, buffer []byte) error {
	if err := flasher.sendCommand(buffer); err != nil {
		return err
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

func (flasher *Flasher) sendCommand(payload []byte) error {
	if payload != nil {
		for {
			if sent, err := flasher.port.Write(payload); err != nil {
				return err
			} else if sent < len(payload) {
				payload = payload[sent:]
			} else {
				break
			}
		}
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
