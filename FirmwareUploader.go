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
	"bytes"
	"certificates"
	"errors"
	"flag"
	"flasher"
	"log"
)

var portName string
var rootCertDir string

func init() {
	flag.StringVar(&portName, "port", "", "serial port to use for flashing")
	flag.StringVar(&rootCertDir, "certs", "", "root certificate directory")
}

func main() {
	flag.Parse()

	f, err := flasher.Open(portName)
	if err != nil {
		log.Fatal(err)
	}

	// Synchronize with programmer
	if err := f.Hello(); err != nil {
		log.Fatal(err)
	}

	// Check maximum supported payload size
	payloadSize, err := f.GetMaximumPayloadSize()
	if err != nil {
		log.Fatal(err)
	}
	if payloadSize < 1024 {
		log.Fatalf("Programmer reports %d as maximum payload size (1024 is needed)", payloadSize)
	}

	if rootCertDir != "" {
		log.Printf("Converting and flashing certificates from '%v'", rootCertDir)
		if err := flashCerts(f, int(payloadSize)); err != nil {
			log.Fatal(err)
		}
	}

	f.Close()
}

func flashCerts(f *flasher.Flasher, payloadSize int) error {
	CERTIFICATES_OFFSET := 0x4000
	CERTIFICATES_LENGTH := 4096

	certificatesData, err := certificates.Convert(rootCertDir)
	if err != nil {
		return err
	}

	if err := f.Erase(uint32(CERTIFICATES_OFFSET), uint32(CERTIFICATES_LENGTH)); err != nil {
		return err
	}

	for i := 0; i < CERTIFICATES_LENGTH; i += payloadSize {
		if err := f.Write(uint32(CERTIFICATES_OFFSET+i), certificatesData[i:i+payloadSize]); err != nil {
			return err
		}
	}

	var flashData []byte

	for i := 0; i < CERTIFICATES_LENGTH; i += payloadSize {
		data, err := f.Read(uint32(CERTIFICATES_OFFSET+i), uint32(payloadSize))
		if err != nil {
			return err
		}

		flashData = append(flashData, data...)
	}

	if !bytes.Equal(certificatesData, flashData) {
		return errors.New("Flash data does not match written!")
	}

	return nil
}
