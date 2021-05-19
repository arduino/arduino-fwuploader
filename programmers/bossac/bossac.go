/*
  FirmwareUploader
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

package bossac

import (
	"log"
	"os"
	"time"

	"github.com/arduino/FirmwareUploader/utils/context"
	"github.com/arduino/arduino-cli/arduino/serialutils"
	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/go-paths-helper"
)

type Bossac struct {
	bossacPath *paths.Path
	portName   string
}

func NewBossac(ctx *context.Context) *Bossac {
	return &Bossac{
		bossacPath: paths.New(ctx.ProgrammerPath),
		portName:   ctx.PortName,
	}
}

func (b *Bossac) Flash(filename string, cb *serialutils.ResetProgressCallbacks) error {
	log.Println("Entering board into bootloader mode")
	port, err := serialutils.Reset(b.portName, true, cb)
	if err != nil {
		return err
	}

	log.Println("Flashing " + filename)
	if port == "" {
		port = b.portName
	}
	err = b.invoke("-e", "-R", "-p", port, "-w", filename)

	time.Sleep(5 * time.Second)

	return err
}

func (b *Bossac) invoke(args ...string) error {
	cmd, err := executils.NewProcessFromPath(b.bossacPath, args...)
	if err != nil {
		return err
	}
	cmd.RedirectStdoutTo(os.Stdout)
	cmd.RedirectStderrTo(os.Stderr)
	return cmd.Run()
}
