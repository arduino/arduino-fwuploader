package utils

import (
	"github.com/facchinm/go-serial"
)

func OpenSerial(portName string) (serial.Port, error) {
	mode := &serial.Mode{
		// This bound rate works on osx 10.14
		BaudRate: 115200,
		Vtimeout: 255,
		Vmin:     0,
	}

	return serial.Open(portName, mode)
}
