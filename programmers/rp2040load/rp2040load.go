package rp2040load

import (
	"log"
	"os"
	"time"

	"github.com/arduino/FirmwareUploader/utils/context"
	"github.com/arduino/arduino-cli/arduino/serialutils"
	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
)

type RP2040Load struct {
	rp2040LoadPath *paths.Path
	portName       string
}

func NewRP2040Load(ctx *context.Context) *RP2040Load {
	return &RP2040Load{
		rp2040LoadPath: paths.New(ctx.ProgrammerPath),
		portName:       ctx.PortName,
	}
}

func (b *RP2040Load) Flash(filename string, cb *serialutils.ResetProgressCallbacks) error {
	log.Println("Entering board into bootloader mode")
	_, err := serialutils.Reset(b.portName, true, cb)
	if err != nil {
		return err
	}

	log.Println("Flashing " + filename)
	if err := b.invoke("-v", "-D", filename); err != nil {
		return errors.Errorf("Error flashing %s: %s", filename, err)
	}

	time.Sleep(5 * time.Second)

	return err
}

func (b *RP2040Load) invoke(args ...string) error {
	cmd, err := executils.NewProcessFromPath(b.rp2040LoadPath, args...)
	if err != nil {
		return err
	}
	cmd.RedirectStdoutTo(os.Stdout)
	cmd.RedirectStderrTo(os.Stderr)
	return cmd.Run()
}
