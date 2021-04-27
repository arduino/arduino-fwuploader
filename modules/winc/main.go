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

package winc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/arduino/FirmwareUploader/programmers/bossac"
	"github.com/arduino/FirmwareUploader/utils/context"
	"github.com/pkg/errors"
)

var f *Flasher
var payloadSize uint16

func Run(ctx *context.Context) error {

	programmer := bossac.NewBossac(ctx)

	if ctx.FWUploaderBinary != "" {
		log.Println("Flashing firmware uploader winc")
		if err := programmer.Flash(ctx.FWUploaderBinary, nil); err != nil {
			return err
		}
	}

	log.Println("Connecting to programmer")
	if _f, err := OpenFlasher(ctx.PortName); err != nil {
		return err
	} else {
		f = _f
	}
	defer f.Close()

	// Synchronize with programmer
	log.Println("Sync with programmer")
	if err := f.Hello(); err != nil {
		return err
	}

	// Check maximum supported payload size
	log.Println("Reading max payload size")
	_payloadSize, err := f.GetMaximumPayloadSize()
	if err != nil {
		return err
	} else {
		payloadSize = _payloadSize
	}
	if payloadSize < 1024 {
		return errors.Errorf("Programmer reports %d as maximum payload size (1024 is needed)", payloadSize)
	}

	if ctx.FirmwareFile != "" {
		if err := flashFirmware(ctx); err != nil {
			return err
		}
	}

	if ctx.RootCertDir != "" || len(ctx.Addresses) != 0 {
		if err := flashCerts(ctx); err != nil {
			return err
		}
	}

	if ctx.ReadAll {
		log.Println("Reading all flash")
		if err := readAllFlash(); err != nil {
			return err
		}
	}

	f.Close()

	if ctx.BinaryToRestore != "" {
		log.Println("Restoring previous sketch")

		if err := programmer.Flash(ctx.BinaryToRestore, nil); err != nil {
			return err
		}
	}
	return nil
}

func readAllFlash() error {
	for i := 0; i < 256; i++ {
		if data, err := f.Read(uint32(i*1024), 1024); err != nil {
			return err
		} else {
			os.Stdout.Write(data)
		}
	}
	return nil
}

func flashCerts(ctx *context.Context) error {
	CertificatesOffset := 0x4000

	if ctx.RootCertDir != "" {
		log.Printf("Converting and flashing certificates from '%v'", ctx.RootCertDir)
	}

	certificatesData, err := ConvertCertificates(ctx.RootCertDir, ctx.Addresses)
	if err != nil {
		return err
	}

	return flashChunk(CertificatesOffset, certificatesData)
}

func flashFirmware(ctx *context.Context) error {
	FirmwareOffset := 0x0000

	log.Printf("Flashing firmware from '%v'", ctx.FirmwareFile)

	fwData, err := ioutil.ReadFile(ctx.FirmwareFile)
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
		fmt.Printf("\rFlashing: " + strconv.Itoa((i*100)/bufferLength) + "%%")
		start := i
		end := i + chunkSize
		if end > bufferLength {
			end = bufferLength
		}
		if err := f.Write(uint32(offset+i), buffer[start:end]); err != nil {
			return err
		}
	}

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

	return nil
}
