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

package flash_firmware

import (
	"bytes"
	"os"

	"github.com/arduino/FirmwareUploader/flasher"
	programmer "github.com/arduino/FirmwareUploader/programmers"
	"github.com/arduino/arduino-cli/arduino/serialutils"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/spf13/cobra"
)

var (
	fqbn    string
	address string
	module  string
)

// NewCommand created a new `version` command
func NewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "flash-firmware",
		Short:   "Shows version number of FirmwareUploader.",
		Long:    "Shows the version number of FirmwareUploader which is installed on your system.",
		Example: "  " + os.Args[0] + " version",
		Args:    cobra.NoArgs,
		Run:     run,
	}

	command.Flags().StringVarP(&fqbn, "fqbn", "b", "", "Fully Qualified Board Name, e.g.: arduino:samd:mkr1000, arduino:mbed_nano:nanorp2040connect")
	command.Flags().StringVarP(&address, "address", "a", "", "Upload port, e.g.: COM10, /dev/ttyACM0")
	command.Flags().StringVarP(&module, "module", "m", "", "Firmware module ID, e.g.: WINC1500, NINA")
	return command
}

func run(cmd *cobra.Command, args []string) {
	if fqbn == "" {
		feedback.Errorf("Error during firmware flashing: missing board fqbn")
		os.Exit(errorcodes.ErrGeneric)
	}

	if address == "" {
		feedback.Errorf("Error during firmware flashing: missing board address")
		os.Exit(errorcodes.ErrGeneric)
	}

	if module == "" {
		// TODO: Get firmware ID for board if not provided
	}

	// TODO: Get firmware binary from given ID

	// TODO: Get uploader executable path

	// TODO: Get uploader command line
	commandLine := []string{""}

	// TODO: Build uploader command line using uploader path, eventual config path and Loader Sketch binary

	// TODO: Get 1200bps touch from upload properties
	use1200bpsTouch := false
	if use1200bpsTouch {
		feedback.Print("Putting board into bootloader mode")
		// TODO: Get waitForUploadPort from upload properties
		waitForUploadPort := false
		_, err := serialutils.Reset(address, waitForUploadPort, nil)
		if err != nil {
			// TODO
		}
	}

	// TODO: Flash loader Sketch
	flashOut := new(bytes.Buffer)
	flashErr := new(bytes.Buffer)
	// TODO: Maybe this can be done differently?
	var err error
	// TODO: OutputFormat is not stored globally, we must store it globally since we need it
	OutputFormat := "json"
	if OutputFormat == "json" {
		err = programmer.Flash(commandLine, flashOut, flashErr)
	} else {
		err = programmer.Flash(commandLine, os.Stdout, os.Stderr)
	}
	if err != nil {
		// TODO
	}

	// Get flasher depending on which module to use
	var f flasher.Flasher
	switch module {
	case "NINA":
		f, err = flasher.NewNinaFlasher(address)
	case "SARA":
		f, err = flasher.NewSaraFlasher(address)
	case "WINC":
		f, err = flasher.NewWincFlasher(address)
	}
	if err != nil {
		// TODO
	}
	defer f.Close()

	// TODO: Flash firmware
}
