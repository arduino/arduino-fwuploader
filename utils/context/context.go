package context

import (
	"fmt"

	"github.com/arduino/arduino-cli/arduino/serialutils"
)

type addressFlags []string

func (af *addressFlags) String() string {
	return fmt.Sprint(*af)
}

func (af *addressFlags) Set(value string) error {
	*af = append(*af, value)
	return nil
}

type Context struct {
	PortName         string
	RootCertDir      string
	Addresses        addressFlags
	FirmwareFile     string
	FWUploaderBinary string
	ReadAll          bool
	BinaryToRestore  string
	ProgrammerPath   string
	Model            string
	Compatible       string
	Retries          int
}

type Programmer interface {
	Flash(filename string, cb *serialutils.ResetProgressCallbacks) error
}
