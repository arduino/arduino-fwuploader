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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/arduino/go-paths-helper"
	"go.bug.st/serial"
)

func NewSaraFlasher(portAddress string) (*SaraFlasher, error) {
	port, err := openSerial(portAddress)
	if err != nil {
		return nil, err
	}
	// Magic numbers ¯\_(ツ)_/¯
	return &SaraFlasher{port: port, payloadSize: 128}, nil
}

type SaraFlasher struct {
	port        serial.Port
	payloadSize int
}

func (f *SaraFlasher) FlashFirmware(firmwareFile *paths.Path) error {
	data, err := firmwareFile.ReadFile()
	if err != nil {
		return err
	}

	_, err = f.expectMinBytes("AT+ULSTFILE", "+ULSTFILE:", 1000, 0)
	if err != nil {
		return err
	}

	_, err = f.expectMinBytes("AT+UDWNFILE=\"UPDATE.BIN\","+strconv.Itoa(len(data))+",\"FOAT\"", ">", 20000, 0)
	if err != nil {
		return err
	}

	firmwareOffset := 0x0000
	err = f.flashChunk(firmwareOffset, data)
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	_, err = f.expectMinBytes("", "OK", 1000, 0)
	if err != nil {
		return err
	}

	_, err = f.expectMinBytes("AT+UFWINSTALL", "OK", 60000, 0)
	if err != nil {
		return err
	}

	time.Sleep(10 * time.Second)

	// wait up to 20 minutes trying to ping the module. After 20 minutes signal the error
	start := time.Now()
	for time.Since(start) < time.Minute*20 {
		err = f.hello()
		if err == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	return err
}

func (f *SaraFlasher) FlashCertificates(certificatePaths *paths.PathList) error {
	// TODO
	return nil
}

func (f *SaraFlasher) Close() error {
	return f.port.Close()
}

func (f *SaraFlasher) hello() error {
	f.expectMinBytes("ATE0", "OK", 100, 0)
	f.expectMinBytes("ATE0", "OK", 100, 0)
	f.expectMinBytes("ATE0", "OK", 100, 0)
	_, err := f.expectMinBytes("AT", "OK", 100, 0)
	return err
}

func (f *SaraFlasher) write(address uint32, buffer []byte) error {
	return f.sendCommand(CommandData{
		Payload: buffer,
	})
}

func (f *SaraFlasher) flashChunk(offset int, buffer []byte) error {
	bufferLength := len(buffer)

	for i := 0; i < bufferLength; i += f.payloadSize {
		fmt.Printf("\rFlashing: " + strconv.Itoa((i*100)/bufferLength) + "%%")
		start := i
		end := i + f.payloadSize
		if end > bufferLength {
			end = bufferLength
		}
		if err := f.write(uint32(offset+i), buffer[start:end]); err != nil {
			return err
		}
		//time.Sleep(1 * time.Millisecond)
	}

	return nil
}

func (f *SaraFlasher) getMaximumPayloadSize() (uint16, error) {
	return 0, fmt.Errorf("Not supported by SaraFlasher")
}

func (f *SaraFlasher) serialFillBuffer(buffer []byte) error {
	read := 0
	for read < len(buffer) {
		n, err := f.port.Read(buffer[read:])
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

func (f *SaraFlasher) sendCommand(data CommandData) error {
	if data.Payload != nil {
		for {
			if sent, err := f.port.Write(data.Payload); err != nil {
				return err
			} else if sent < len(data.Payload) {
				data.Payload = data.Payload[sent:]
			} else {
				break
			}
		}
	}
	return nil
}

func (f *SaraFlasher) expectMinBytes(buffer string, response string, timeout int, min_bytes int) (string, error) {
	err := f.sendCommand(CommandData{
		Payload: []byte(buffer + "\r\n"),
	})
	if err != nil {
		return "", err
	}

	// log.Println("Sending " + buffer)

	// Receive response
	var res []byte
	n := 0

	start := time.Now()
	for (time.Since(start) < time.Duration(timeout)*time.Millisecond && !strings.Contains(string(res), response)) || (len(res) < min_bytes) {
		data := 0
		partial := make([]byte, 65535)
		data, err = f.port.Read(partial)
		res = append(res, partial[:data]...)
		n += data
		if err != nil {
			return "", err
		}
	}

	// log.Println(string(res))

	if !strings.Contains(string(res), response) {
		return string(res), FlasherError{err: fmt.Sprintf("Expected %s, got %s", response, res)}
	}
	return string(res), nil
}
func (f *SaraFlasher) getFirmwareVersion() (string, error) {
	return f.expectMinBytes("ATI9", "05.06,A.02.", 100, 25)
}
