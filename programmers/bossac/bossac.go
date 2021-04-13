package bossac

import (
	"log"
	"os"
	"time"

	"github.com/arduino/FirmwareUpdater/utils/context"
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

func (b *Bossac) invoke(args ...string) error {
	cmd, err := executils.NewProcessFromPath(b.bossacPath, args...)
	if err != nil {
		return err
	}
	cmd.RedirectStdoutTo(os.Stdout)
	cmd.RedirectStderrTo(os.Stderr)
	return cmd.Run()
}
