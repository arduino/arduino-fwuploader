package avrdude

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/arduino/FirmwareUpdater/utils/context"
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

func (b *Avrdude) Flash(filename string) error {
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

	time.Sleep(3 * time.Second)

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
