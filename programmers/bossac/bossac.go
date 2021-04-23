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
