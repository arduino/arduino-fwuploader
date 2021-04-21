package bossac

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/arduino/FirmwareUploader/utils/context"
	"github.com/arduino/arduino-cli/arduino/serialutils"
	"github.com/arduino/arduino-cli/executils"
	"github.com/pkg/errors"
)

type Bossac struct {
}

func (b *Bossac) Flash(ctx *context.Context, filename string) error {
	log.Println("Entering board into bootloader mode")
	port, err := serialutils.Reset(ctx.PortName, true)
	if err != nil {
		return err
	}

	log.Println("Flashing " + filename)
	err = invokeBossac([]string{ctx.ProgrammerPath, "-e", "-R", "-p", port, "-w", filename})

	ctx.PortName, err = serialutils.WaitForNewSerialPortOrDefaultTo(port)
	log.Println("Board is back online " + ctx.PortName)
	time.Sleep(1 * time.Second)

	return err
}

func (b *Bossac) DumpAndFlash(ctx *context.Context, filename string) (string, error) {
	dir, err := ioutil.TempDir("", "wifiFlasher_dump")
	if err != nil {
		return "", errors.WithMessage(err, "creating temp dir to store current sketch")
	}

	log.Println("Entering board into bootloader mode")
	port, err := serialutils.Reset(ctx.PortName, true)
	if err != nil {
		return "", err
	}

	log.Println("Reading existing sketch from the baord, to restore it later")
	err = invokeBossac([]string{ctx.ProgrammerPath, "-u", "-r", "-p", port, filepath.Join(dir, "dump.bin")})

	log.Println("Original sketch saved at " + filepath.Join(dir, "dump.bin"))
	if err != nil {
		return "", err
	}

	log.Println("Flashing " + filename)
	err = invokeBossac([]string{ctx.ProgrammerPath, "-e", "-R", "-p", port, "-w", filename})

	ctx.PortName, err = serialutils.WaitForNewSerialPortOrDefaultTo(port)
	log.Println("Board is back online " + ctx.PortName)
	time.Sleep(1 * time.Second)

	return filepath.Join(dir, "dump.bin"), err
}

func invokeBossac(args []string) error {
	cmd, err := executils.NewProcess(args...)
	if err != nil {
		return err
	}
	cmd.RedirectStdoutTo(os.Stdout)
	cmd.RedirectStderrTo(os.Stderr)
	return cmd.Run()
}
