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

package avrdude

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/arduino/FirmwareUploader/utils/context"
	"github.com/arduino/arduino-cli/arduino/serialutils"
	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/go-paths-helper"
)

type Avrdude struct {
	avrdudePath *paths.Path
	configPath  *paths.Path
}

func NewAvrdude(ctx *context.Context) *Avrdude {
	avrdudePath := paths.New(ctx.ProgrammerPath)
	return &Avrdude{
		avrdudePath: avrdudePath,
		configPath:  avrdudePath.Parent().Join("..", "etc", "avrdude.conf"),
	}
}

func (b *Avrdude) Flash(filename string, cb *serialutils.ResetProgressCallbacks) error {
	log.Println("Flashing " + filename)
	err := b.invoke(
		fmt.Sprintf("-C%s", b.configPath),
		"-v",
		"-patmega4809",
		"-cxplainedmini_updi",
		"-Pusb",
		"-b115200",
		"-e",
		"-D",
		fmt.Sprintf("-Uflash:w:%s:i", filename),
		"-Ufuse8:w:0x00:m")

	time.Sleep(5 * time.Second)

	return err
}

func (b *Avrdude) invoke(args ...string) error {
	cmd, err := executils.NewProcessFromPath(b.avrdudePath, args...)
	if err != nil {
		return err
	}
	cmd.RedirectStdoutTo(os.Stdout)
	cmd.RedirectStderrTo(os.Stderr)
	return cmd.Run()
}
