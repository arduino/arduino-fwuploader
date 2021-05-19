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

package nina

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/arduino/FirmwareUploader/programmers/avrdude"
	"github.com/arduino/FirmwareUploader/programmers/bossac"
	"github.com/arduino/FirmwareUploader/programmers/rp2040load"
	"github.com/arduino/FirmwareUploader/utils"
	"github.com/arduino/FirmwareUploader/utils/context"
	"github.com/pkg/errors"
)

var flasher *Flasher
var payloadSize uint16

func Run(ctx *context.Context) error {
	var programmer utils.Programmer

	if ctx.ProgrammerPath != "" {
		if strings.Contains(filepath.Base(ctx.ProgrammerPath), "bossac") {
			programmer = bossac.NewBossac(ctx)
		} else if strings.Contains(filepath.Base(ctx.ProgrammerPath), "avrdude") {
			programmer = avrdude.NewAvrdude(ctx)
		} else if strings.Contains(filepath.Base(ctx.ProgrammerPath), "rp2040load") {
			programmer = rp2040load.NewRP2040Load(ctx)
		} else {
			return errors.New("Programmer path not specified correctly, programmer path set to: " + ctx.ProgrammerPath)
		}
	}

	if ctx.FWUploaderBinary != "" {
		log.Println("Flashing firmware uploader nina")
		if programmer == nil {
			return errors.New("ERROR: You must specify a programmer!")
		}
		if err := programmer.Flash(ctx.FWUploaderBinary, nil); err != nil {
			return err
		}
	}

	log.Println("Connecting to programmer")
	if f, err := OpenFlasher(ctx.PortName); err != nil {
		return err
	} else {
		flasher = f
	}
	defer flasher.Close()

	// Synchronize with programmer
	log.Println("Sync with programmer")
	if err := flasher.Hello(); err != nil {
		return err
	}

	// Check maximum supported payload size
	log.Println("Reading max payload size")
	if _payloadSize, err := flasher.GetMaximumPayloadSize(); err != nil {
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

	flasher.Close()

	if ctx.BinaryToRestore != "" {
		log.Println("Restoring binary")
		if programmer == nil {
			errors.New("ERROR: You must specify a programmer!")
		}
		if err := programmer.Flash(ctx.BinaryToRestore, nil); err != nil {
			return err
		}
	}
	return nil
}

func readAllFlash() error {
	for i := 0; i < 256; i++ {
		if data, err := flasher.Read(uint32(i*1024), 1024); err != nil {
			return err
		} else {
			os.Stdout.Write(data)
		}
	}
	return nil
}

func flashCerts(ctx *context.Context) error {
	CertificatesOffset := 0x10000

	if ctx.RootCertDir != "" {
		log.Printf("Converting and flashing certificates from '%v'", ctx.RootCertDir)
	}

	certificatesData, err := ConvertCertificates(ctx.RootCertDir, ctx.Addresses)
	if err != nil {
		return err
	}

	if len(certificatesData) > 0x20000 {
		errors.New("Too many certificates! Aborting")
	}

	// pad certificatesData to flash page
	for len(certificatesData)%int(payloadSize) != 0 {
		certificatesData = append(certificatesData, 0)
	}

	log.Println(string(certificatesData))

	return flashChunk(CertificatesOffset, certificatesData, false)
}

func flashFirmware(ctx *context.Context) error {
	FirmwareOffset := 0x0000

	log.Printf("Flashing firmware from '%v'", ctx.FirmwareFile)

	fwData, err := ioutil.ReadFile(ctx.FirmwareFile)
	if err != nil {
		return err
	}

	return flashChunk(FirmwareOffset, fwData, true)
}

func flashChunk(offset int, buffer []byte, doChecksum bool) error {
	chunkSize := int(payloadSize)
	bufferLength := len(buffer)

	if err := flasher.Erase(uint32(offset), uint32(bufferLength)); err != nil {
		return err
	}

	for i := 0; i < bufferLength; i += chunkSize {
		fmt.Printf("\rFlashing: " + strconv.Itoa((i*100)/bufferLength) + "%%")
		start := i
		end := i + chunkSize
		if end > bufferLength {
			end = bufferLength
		}
		if err := flasher.Write(uint32(offset+i), buffer[start:end]); err != nil {
			return err
		}
	}

	fmt.Println("")

	if doChecksum {
		return flasher.Md5sum(buffer)
	} else {
		return nil
	}
}
