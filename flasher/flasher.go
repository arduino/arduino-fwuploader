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
	"time"

	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"go.bug.st/serial"
)

type CommandData struct {
	Command byte
	Address uint32
	Value   uint32
	Payload []byte
}

func (d CommandData) String() string {
	return fmt.Sprintf("%+v, %+v, %+v, %+v", d.Command, d.Address, d.Value, d.Payload)
}

type FlasherError struct {
	err string
}

func (e FlasherError) Error() string {
	return e.err
}

type Flasher interface {
	FlashFirmware(firmwareFile *paths.Path) error
	FlashCertificates(certificatePaths *paths.PathList, URLs []string) error
	Close() error

	hello() error
	write(address uint32, buffer []byte) error
	flashChunk(offset int, buffer []byte) error
	getMaximumPayloadSize() (uint16, error)
	serialFillBuffer(buffer []byte) error
	sendCommand(data CommandData) error
}

// http://www.ni.com/product-documentation/54548/en/
// Standard baud rates supported by most serial ports
var baudRates = []int{
	115200,
	57600,
	56000,
	38400,
}

func openSerial(portAddress string) (serial.Port, error) {
	var lastError error

	for _, baudRate := range baudRates {
		port, err := serial.Open(portAddress, &serial.Mode{BaudRate: baudRate})
		if err != nil {
			lastError = err
			// Try another baudrate
			continue
		}
		logrus.Infof("Opened port %s at %d", portAddress, baudRate)

		if err := port.SetReadTimeout(30 * time.Second); err != nil {
			err = fmt.Errorf("could not set timeout on serial port: %s", err)
			logrus.Error(err)
			return nil, err
		}

		return port, nil
	}

	return nil, lastError
}
