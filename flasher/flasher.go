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
	"fmt"
	"io"
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
	FlashFirmware(firmwareFile *paths.Path, flasherOut io.Writer) error
	FlashCertificates(certificatePaths *paths.PathList, URLs []string, flasherOut io.Writer) error
	Close() error
	SetProgressCallback(func(progress int))

	hello() error
	write(address uint32, buffer []byte) error
	flashChunk(offset int, buffer []byte) error
	getMaximumPayloadSize() (uint16, error)
	serialFillBuffer(buffer []byte) error
	sendCommand(data CommandData) error
}

// OpenSerial opens a new serial connection with the specified portAddress
func OpenSerial(portAddress string, baudRate int, readTimeout int) (serial.Port, error) {

	port, err := serial.Open(portAddress, &serial.Mode{BaudRate: baudRate})
	if err != nil {
		return nil, err
	}
	logrus.Infof("Opened port %s at %d", portAddress, baudRate)

	if err := port.SetReadTimeout(time.Duration(readTimeout) * time.Second); err != nil {
		err = fmt.Errorf("could not set timeout on serial port: %s", err)
		logrus.Error(err)
		return nil, err
	}
	return port, nil
}

// FlashResult contains the result of the flashing procedure
type FlashResult struct {
	Programmer *ExecOutput `json:"programmer"`
	Flasher    *ExecOutput `json:"flasher,omitempty"`
	Version    string      `json:"version,omitempty"`
}

// ExecOutput contais the stdout and stderr output, they are used to store the output of the flashing and upload
type ExecOutput struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

func (r *FlashResult) Data() interface{} {
	return r
}

func (r *FlashResult) String() string {
	// The output is already printed via os.Stdout/os.Stdin
	return ""
}
