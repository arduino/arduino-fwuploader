package avrdude

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/arduino/FirmwareUploader/utils/context"
	"github.com/arduino/arduino-cli/executils"
	"github.com/pkg/errors"
)

type Avrdude struct {
}

func (b *Avrdude) Flash(ctx *context.Context, filename string) error {
	log.Println("Flashing " + filename)
	err := invokeAvrdude([]string{ctx.ProgrammerPath, "-C" + filepath.Join(filepath.Dir(ctx.ProgrammerPath), "..", "etc/avrdude.conf"), "-v", "-patmega4809", "-cxplainedmini_updi", "-Pusb", "-b115200", "-e", "-D", "-Uflash:w:" + filename + ":i", "-Ufuse8:w:0x00:m"})

	time.Sleep(3 * time.Second)

	return err
}

func (b *Avrdude) DumpAndFlash(ctx *context.Context, filename string) (string, error) {
	dir, err := ioutil.TempDir("", "wifiFlasher_dump")
	if err != nil {
		return "", errors.WithMessage(err, "creating temp dir to store current sketch")
	}

	log.Println("Reading existing sketch from the baord, to restore it later")
	err = invokeAvrdude([]string{ctx.ProgrammerPath, "-C" + filepath.Join(filepath.Dir(ctx.ProgrammerPath), "..", "etc/avrdude.conf"), "-v", "-patmega4809", "-cxplainedmini_updi", "-Pusb", "-b115200", "-D", "-Uflash:r:" + filepath.Join(dir, "dump.bin") + ":i"})
	if err != nil {
		return "", err
	}
	log.Println("Original sketch saved at " + filepath.Join(dir, "dump.bin"))

	log.Println("Flashing " + filename)
	err = invokeAvrdude([]string{ctx.ProgrammerPath, "-C" + filepath.Join(filepath.Dir(ctx.ProgrammerPath), "..", "etc/avrdude.conf"), "-v", "-patmega4809", "-cxplainedmini_updi", "-Pusb", "-b115200", "-e", "-D", "-Uflash:w:" + filename + ":i", "-Ufuse8:w:0x00:m"})

	time.Sleep(3 * time.Second)

	return filepath.Join(dir, "dump.bin"), err
}

func invokeAvrdude(args []string) error {
	cmd, err := executils.NewProcess(args...)
	if err != nil {
		return err
	}
	cmd.RedirectStdoutTo(os.Stdout)
	cmd.RedirectStderrTo(os.Stderr)
	return cmd.Run()
}
