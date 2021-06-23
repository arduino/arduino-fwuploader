/*
  arduino-fwuploader
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

package certificates

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/arduino/serialutils"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-fwuploader/flasher"
	"github.com/arduino/arduino-fwuploader/indexes"
	"github.com/arduino/arduino-fwuploader/indexes/download"
	programmer "github.com/arduino/arduino-fwuploader/programmers"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	fqbn             string
	address          string
	certificateURLs  []string
	certificatePaths []string
)

// NewCommand created a new `version` command
func NewFlashCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "flash",
		Short: "Flashes certificates to board.",
		Long:  "Flashes specified certificates to board at specified address.",
		Example: "" +
			"  " + os.Args[0] + " flash --fqbn arduino:samd:mkr1000 --address COM10 --url arduino.cc:443 --file /home/me/Digicert.cer\n" +
			"  " + os.Args[0] + " flash -b arduino:samd:mkr1000 -a COM10 -u arduino.cc:443 -u google.cc:443\n" +
			"  " + os.Args[0] + " flash -b arduino:samd:mkr1000 -a COM10 -f /home/me/VeriSign.cer -f /home/me/Digicert.cer\n",
		Args: cobra.NoArgs,
		Run:  run,
	}

	command.Flags().StringVarP(&fqbn, "fqbn", "b", "", "Fully Qualified Board Name, e.g.: arduino:samd:mkr1000, arduino:mbed_nano:nanorp2040connect")
	command.Flags().StringVarP(&address, "address", "a", "", "Upload port, e.g.: COM10, /dev/ttyACM0")
	command.Flags().StringSliceVarP(&certificateURLs, "url", "u", []string{}, "List of urls to download root certificates, e.g.: arduino.cc:443")
	command.Flags().StringSliceVarP(&certificatePaths, "file", "f", []string{}, "List of paths to certificate file, e.g.: /home/me/Digicert.cer")
	return command
}

func run(cmd *cobra.Command, args []string) {
	packageIndex, err := indexes.GetPackageIndex()
	if err != nil {
		feedback.Errorf("Can't load package index: %s", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	firmwareIndex, err := indexes.GetFirmwareIndex()
	if err != nil {
		feedback.Errorf("Can't load firmware index: %s", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	if fqbn == "" {
		feedback.Errorf("Error during certificates flashing: missing board fqbn")
		os.Exit(errorcodes.ErrBadArgument)
	}

	if address == "" {
		feedback.Errorf("Error during certificates flashing: missing board address")
		os.Exit(errorcodes.ErrBadArgument)
	}

	if len(certificateURLs) == 0 && len(certificatePaths) == 0 {
		feedback.Errorf("Error during certificates flashing: no certificates provided")
		os.Exit(errorcodes.ErrBadArgument)
	}

	board := firmwareIndex.GetBoard(fqbn)
	if board == nil {
		feedback.Errorf("Can't find board with %s fqbn", fqbn)
		os.Exit(errorcodes.ErrBadArgument)
	}

	toolRelease := indexes.GetToolRelease(packageIndex, board.Uploader)
	if toolRelease == nil {
		feedback.Errorf("Error getting upload tool %s for board %s", board.Uploader, board.Fqbn)
		os.Exit(errorcodes.ErrGeneric)
	}
	uploadToolDir, err := download.DownloadTool(toolRelease)
	if err != nil {
		feedback.Errorf("Error downloading tool %s: %s", board.Uploader, err)
		os.Exit(errorcodes.ErrGeneric)
	}

	loaderSketchPath, err := download.DownloadLoaderSketch(board.LoaderSketch)
	if err != nil {
		feedback.Errorf("Error downloading loader sketch from %s: %s", board.LoaderSketch.URL, err)
		os.Exit(errorcodes.ErrGeneric)
	}

	loaderSketch := strings.ReplaceAll(loaderSketchPath.String(), loaderSketchPath.Ext(), "")

	uploaderCommand := board.GetUploaderCommand()
	uploaderCommand = strings.ReplaceAll(uploaderCommand, "{tool_dir}", filepath.FromSlash(uploadToolDir.String()))
	uploaderCommand = strings.ReplaceAll(uploaderCommand, "{serial.port.file}", address)
	uploaderCommand = strings.ReplaceAll(uploaderCommand, "{loader.sketch}", loaderSketch)

	commandLine, err := properties.SplitQuotedString(uploaderCommand, "\"", false)
	if err != nil {
		feedback.Errorf(`Error splitting command line "%s": %s`, uploaderCommand, err)
		os.Exit(errorcodes.ErrGeneric)
	}

	// Check if board needs a 1200bps touch for upload
	uploadPort := address
	if board.UploadTouch {
		logrus.Info("Putting board into bootloader mode")
		newUploadPort, err := serialutils.Reset(address, board.UploadWait, nil)
		if err != nil {
			feedback.Errorf("Error during certificates flashing: missing board address")
			os.Exit(errorcodes.ErrGeneric)
		}
		if newUploadPort != "" {
			logrus.Infof("Found port to upload Loader: %s", newUploadPort)
			uploadPort = newUploadPort
		}
	}

	// Flash loader Sketch
	programmerOut := new(bytes.Buffer)
	programmerErr := new(bytes.Buffer)
	if feedback.GetFormat() == feedback.JSON {
		err = programmer.Flash(commandLine, programmerOut, programmerErr)
	} else {
		err = programmer.Flash(commandLine, os.Stdout, os.Stderr)
	}
	if err != nil {
		feedback.Errorf("Error during certificates flashing: %s", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	// Wait a bit after flashing the loader sketch for the board to become
	// available again.
	time.Sleep(2 * time.Second)

	// Get flasher depending on which module to use
	var f flasher.Flasher
	moduleName := board.Module
	switch moduleName {
	case "NINA":
		f, err = flasher.NewNinaFlasher(uploadPort)
	case "WINC1500":
		f, err = flasher.NewWincFlasher(uploadPort)
	default:
		err = fmt.Errorf("unknown module: %s", moduleName)
	}
	if err != nil {
		feedback.Errorf("Error during certificates flashing: %s", err)
		os.Exit(errorcodes.ErrGeneric)
	}
	defer f.Close()

	// now flash the certificate
	flasherOut := new(bytes.Buffer)
	flasherErr := new(bytes.Buffer)
	certFileList := paths.NewPathList(certificatePaths...)
	if feedback.GetFormat() == feedback.JSON {
		err = f.FlashCertificates(&certFileList, certificateURLs, flasherOut)
	} else {
		err = f.FlashCertificates(&certFileList, certificateURLs, os.Stdout)
	}
	if err != nil {
		feedback.Errorf("Error during certificates flashing: %s", err)
		flasherErr.Write([]byte(fmt.Sprintf("Error during certificates flashing: %s", err)))
	}

	// Print the results
	feedback.PrintResult(&flasher.FlashResult{
		Programmer: (&flasher.ExecOutput{
			Stdout: programmerOut.String(),
			Stderr: programmerErr.String(),
		}),
		Flasher: (&flasher.ExecOutput{
			Stdout: flasherOut.String(),
			Stderr: flasherErr.String(),
		}),
	})
	// Exit if something went wrong but after printing
	if err != nil {
		os.Exit(errorcodes.ErrGeneric)
	}
}
