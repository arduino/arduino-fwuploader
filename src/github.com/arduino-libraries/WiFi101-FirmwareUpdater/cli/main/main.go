package main

import (
	"flag"
	"log"
)

type addressFlags []string

func (af *addressFlags) String() string {
	return ""
}

func (af *addressFlags) Set(value string) error {
	*af = append(*af, value)
	return nil
}

var portName string
var rootCertDir string
var addresses addressFlags
var firmwareFile string
var readAll bool

func init() {
	flag.StringVar(&portName, "port", "", "serial port to use for flashing")
	flag.StringVar(&rootCertDir, "certs", "", "root certificate directory")
	flag.Var(&addresses, "address", "address (host:port) to fetch and flash root certificate for, multiple values allowed")
	flag.StringVar(&firmwareFile, "firmware", "", "firmware file to flash")
	flag.BoolVar(&readAll, "read", false, "read all firmware and output to stdout")
}

func main() {
	flag.Parse()

	if portName == "" {
		log.Fatal("Please specify a serial port")
	}

	winc_flasher()
}
