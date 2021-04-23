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

package sara

import (
	"fmt"
	"github.com/arduino/FirmwareUploader/programmers/bossac"
	"github.com/arduino/FirmwareUploader/utils/context"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

var f *Flasher
var payloadSize uint16
var programmer context.Programmer

func Run(ctx *context.Context) {
	programmer := bossac.NewBossac(ctx)

	if ctx.FWUploaderBinary != "" {
		log.Println("Flashing firmware uploader sara")
		if err := programmer.Flash(ctx.FWUploaderBinary, nil); err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Connecting to programmer")
	if _f, err := OpenFlasher(ctx.PortName); err != nil {
		log.Fatal(err)
	} else {
		f = _f
	}

	time.Sleep(2 * time.Second)

	// Synchronize with programmer
	log.Println("Sync with programmer")
	if err := f.Hello(); err != nil {
		log.Fatal(err)
	}

	// Check maximum supported payload size
	log.Println("Reading actual firmware version")

	if fwVersion, err := f.GetFwVersion(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Initial firmware version: " + fwVersion)
	}

	payloadSize = 128

	if ctx.FirmwareFile != "" {
		if err := flashFirmware(ctx); err != nil {
			log.Fatal(err)
		}
	}

	if fwVersion, err := f.GetFwVersion(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("After applying update firmware version: " + fwVersion)
	}

	f.Close()

	if ctx.BinaryToRestore != "" {
		log.Println("Restoring previous sketch")

		if err := programmer.Flash(ctx.BinaryToRestore, nil); err != nil {
			log.Fatal(err)
		}
	}
}

func flashFirmware(ctx *context.Context) error {
	FirmwareOffset := 0x0000

	log.Printf("Flashing firmware from '%v'", ctx.FirmwareFile)

	fwData, err := ioutil.ReadFile(ctx.FirmwareFile)
	if err != nil {
		return err
	}

	_, err = f.Expect("AT+ULSTFILE", "+ULSTFILE:", 1000)
	if err != nil {
		return err
	}

	_, err = f.Expect("AT+UDWNFILE=\"UPDATE.BIN\","+strconv.Itoa(len(fwData))+",\"FOAT\"", ">", 20000)
	if err != nil {
		return err
	}

	err = flashChunk(FirmwareOffset, fwData)
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	_, err = f.Expect("", "OK", 1000)
	if err != nil {
		return err
	}

	_, err = f.Expect("AT+UFWINSTALL", "OK", 60000)
	if err != nil {
		return err
	}

	time.Sleep(10 * time.Second)

	// wait up to 20 minutes trying to ping the module. After 20 minutes signal the error
	start := time.Now()
	for time.Since(start) < time.Minute*20 {
		err = f.Hello()
		if err == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return err
}

func flashChunk(offset int, buffer []byte) error {
	chunkSize := int(payloadSize)
	bufferLength := len(buffer)

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
		//time.Sleep(1 * time.Millisecond)
	}

	return nil
}
