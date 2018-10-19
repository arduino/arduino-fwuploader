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

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/arduino-libraries/WiFi101-FirmwareUpdater/certificates"
	"github.com/arduino-libraries/WiFi101-FirmwareUpdater/flasher"
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

var f *flasher.Flasher
var payloadSize uint16

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

	log.Println("Connecting to programmer")
	if _f, err := flasher.Open(portName); err != nil {
		log.Fatal(err)
	} else {
		f = _f
	}
	defer f.Close()

	// Synchronize with programmer
	log.Println("Synch with programmer")
	if err := f.Hello(); err != nil {
		log.Fatal(err)
	}

	// Check maximum supported payload size
	log.Println("Reading max payload size")
	_payloadSize, err := f.GetMaximumPayloadSize()
	if err != nil {
		log.Fatal(err)
	} else {
		payloadSize = _payloadSize
	}
	if payloadSize < 1024 {
		log.Fatalf("Programmer reports %d as maximum payload size (1024 is needed)", payloadSize)
	}

	if firmwareFile != "" {
		if err := flashFirmware(); err != nil {
			log.Fatal(err)
		}
	}

	if rootCertDir != "" || len(addresses) != 0 {
		if err := flashCerts(); err != nil {
			log.Fatal(err)
		}
	}

	if readAll {
		log.Println("Reading all flash")
		if err := readAllFlash(); err != nil {
			log.Fatal(err)
		}
	}
}

func readAllFlash() error {
	for i := 0; i < 256; i++ {
		if data, err := f.Read(uint32(i*1024), 1024); err != nil {
			log.Fatal(err)
		} else {
			os.Stdout.Write(data)
		}
	}
	return nil
}

func flashCerts() error {
	CertificatesOffset := 0x4000

	if rootCertDir != "" {
		log.Printf("Converting and flashing certificates from '%v'", rootCertDir)
	}

	certificatesData, err := certificates.Convert(rootCertDir, addresses)
	if err != nil {
		return err
	}

	return flashChunk(CertificatesOffset, certificatesData)
}

func flashFirmware() error {
	FirmwareOffset := 0x0000

	log.Printf("Flashing firmware from '%v'", firmwareFile)

	fwData, err := ioutil.ReadFile(firmwareFile)
	if err != nil {
		return err
	}

	return flashChunk(FirmwareOffset, fwData)
}

func flashChunk(offset int, buffer []byte) error {
	chunkSize := int(payloadSize)
	bufferLength := len(buffer)

	if err := f.Erase(uint32(offset), uint32(bufferLength)); err != nil {
		return err
	}

	for i := 0; i < bufferLength; i += chunkSize {
		fmt.Print(i * 100 / bufferLength)
		fmt.Println("%")
		start := i
		end := i + chunkSize
		if end > bufferLength {
			end = bufferLength
		}
		if err := f.Write(uint32(offset+i), buffer[start:end]); err != nil {
			return err
		}
	}

	return f.Md5sum(buffer)

	/*
		fmt.Println("read")

		var flashData []byte
		for i := 0; i < bufferLength; i += chunkSize {
			readLength := chunkSize
			if (i + chunkSize) > bufferLength {
				readLength = bufferLength % chunkSize
			}

			data, err := f.Read(uint32(offset+i), uint32(readLength))
			if err != nil {
				return err
			}

			flashData = append(flashData, data...)
		}

		if !bytes.Equal(buffer, flashData) {
			return errors.New("Flash data does not match written!")
		}
	*/
	return nil
}
