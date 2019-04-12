package utils

import (
	"github.com/facchinm/go-serial"
	"log"
)

// http://www.ni.com/product-documentation/54548/en/
var baudRates = []int{
	// Standard baud rates supported by most serial ports
	115200,
	57600,
	56000,
	38400,
}

func OpenSerial(portName string) (serial.Port, error) {
	var port serial.Port
	var err error
	for _, baudRate := range baudRates {
		mode := &serial.Mode{
			BaudRate: baudRate,
			Vtimeout: 255,
			Vmin:     0,
		}
		port, err := serial.Open(portName, mode)
		if err == nil {
			log.Printf("Open the serial port with baud rate %d", baudRate)
			return port, nil
		}
	}
	return port, err

}
