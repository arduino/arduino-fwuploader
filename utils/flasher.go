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

package utils

import (
	"log"
	"time"

	"github.com/arduino/arduino-cli/arduino/serialutils"
	"go.bug.st/serial"
)

// http://www.ni.com/product-documentation/54548/en/
var baudRates = []int{
	// Standard baud rates supported by most serial ports
	115200,
	57600,
	56000,
	38400,
}

type Programmer interface {
	Flash(filename string, cb *serialutils.ResetProgressCallbacks) error
}

func OpenSerial(portName string) (serial.Port, error) {
	var lastError error

	for _, baudRate := range baudRates {
		port, err := serial.Open(portName, &serial.Mode{BaudRate: baudRate})
		if err != nil {
			lastError = err
			// try another baudrate
			continue
		}
		log.Printf("Opened the serial port with baud rate %d", baudRate)

		if err := port.SetReadTimeout(30 * time.Second); err != nil {
			log.Fatalf("Could not set timeout on serial port: %s", err)
			return nil, err
		}

		return port, nil
	}

	return nil, lastError
}
