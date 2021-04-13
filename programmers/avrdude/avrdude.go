package avrdude

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/arduino/FirmwareUpdater/utils/context"
	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
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

func (b *Avrdude) DumpAndFlash(filename string) (string, error) {
	dir, err := ioutil.TempDir("", "wifiFlasher_dump")
	if err != nil {
		return "", errors.WithMessage(err, "creating temp dir to store current sketch")
	}

	log.Println("Reading existing sketch from the baord, to restore it later")
	err = b.invoke(
		fmt.Sprintf("-C%s", b.configPath),
		"-v",
		"-patmega4809",
		"-cxplainedmini_updi",
		"-Pusb",
		"-b115200",
		"-D",
		"-Uflash:r:"+filepath.Join(dir, "dump.bin")+":i")
	if err != nil {
		return "", err
	}
	log.Println("Original sketch saved at " + filepath.Join(dir, "dump.bin"))

	log.Println("Flashing " + filename)
	err = b.invoke(
		fmt.Sprintf("-C%s", b.configPath),
		"-v",
		"-patmega4809",
		"-cxplainedmini_updi",
		"-Pusb",
		"-b115200",
		"-e",
		"-D",
		"-Uflash:w:"+filename+":i",
		"-Ufuse8:w:0x00:m")

	time.Sleep(3 * time.Second)

	return filepath.Join(dir, "dump.bin"), err
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
