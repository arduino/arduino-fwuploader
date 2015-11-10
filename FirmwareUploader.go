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
	"flasher"
	"log"
	"os"
)

var portName string

func init() {
	flag.StringVar(&portName, "port", "", "serial port to use for flashing")
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

	//if err := f.Write(1024, make([]byte, payloadSize)); err != nil {
	//	log.Fatal(err)
	//}

	for i := 0; i < 256; i++ {
		data, err := f.Read(uint32(i*1024), 1024)
		if err != nil {
			log.Fatal(err.Error())
		}
		os.Stdout.Write(data)
	}

	f.Close()
}
