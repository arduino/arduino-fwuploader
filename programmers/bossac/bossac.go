package bossac

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/arduino/FirmwareUpdater/utils/context"
	"github.com/arduino/arduino-cli/arduino/serialutils"
	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
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

func (b *Bossac) Flash(filename string) error {
	log.Println("Entering board into bootloader mode")
	port, err := serialutils.Reset(b.portName, true)
	if err != nil {
		return err
	}

	log.Println("Flashing " + filename)
	err = b.invoke("-e", "-R", "-p", port, "-w", filename)

	b.portName, err = serialutils.WaitForNewSerialPortOrDefaultTo(port)
	log.Println("Board is back online " + b.portName)
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
	err = b.invoke("-u", "-r", "-p", port, filepath.Join(dir, "dump.bin"))

	log.Println("Original sketch saved at " + filepath.Join(dir, "dump.bin"))
	if err != nil {
		return "", err
	}

	log.Println("Flashing " + filename)
	err = b.invoke("-e", "-R", "-p", port, "-w", filename)

	ctx.PortName, err = serialutils.WaitForNewSerialPortOrDefaultTo(port)
	log.Println("Board is back online " + ctx.PortName)
	time.Sleep(1 * time.Second)

	return filepath.Join(dir, "dump.bin"), err
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
