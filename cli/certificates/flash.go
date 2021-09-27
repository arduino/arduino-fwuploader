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
	"strings"
	"time"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-fwuploader/cli/arguments"
	"github.com/arduino/arduino-fwuploader/cli/common"
	"github.com/arduino/arduino-fwuploader/flasher"
	"github.com/arduino/arduino-fwuploader/indexes/download"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	commonFlags      arguments.Flags
	certificateURLs  []string
	certificatePaths []string
)

// NewFlashCommand creates a new `flash` command
func NewFlashCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "flash",
		Short: "Flashes certificates to board.",
		Long:  "Flashes specified certificates to board at specified address.",
		Example: "" +
			"  " + os.Args[0] + " certificates flash --fqbn arduino:samd:mkr1000 --address COM10 --url arduino.cc:443 --file /home/me/Digicert.cer\n" +
			"  " + os.Args[0] + " certificates flash -b arduino:samd:mkr1000 -a COM10 -u arduino.cc:443 -u google.cc:443\n" +
			"  " + os.Args[0] + " certificates flash -b arduino:samd:mkr1000 -a COM10 -f /home/me/VeriSign.cer -f /home/me/Digicert.cer\n",
		Args: cobra.NoArgs,
		Run:  runFlash,
	}
	commonFlags.AddToCommand(command)
	command.Flags().StringSliceVarP(&certificateURLs, "url", "u", []string{}, "List of urls to download root certificates, e.g.: arduino.cc:443")
	command.Flags().StringSliceVarP(&certificatePaths, "file", "f", []string{}, "List of paths to certificate file, e.g.: /home/me/Digicert.cer")
	return command
}

func runFlash(cmd *cobra.Command, args []string) {

	packageIndex, firmwareIndex := common.InitIndexes()
	common.CheckFlags(commonFlags.Fqbn, commonFlags.Address)
	board := common.GetBoard(firmwareIndex, commonFlags.Fqbn)
	uploadToolDir := common.GetUploadToolDir(packageIndex, board)

	if len(certificateURLs) == 0 && len(certificatePaths) == 0 {
		feedback.Errorf("Error during certificates flashing: no certificates provided")
		os.Exit(errorcodes.ErrBadArgument)
	}

	loaderSketchPath, err := download.DownloadSketch(board.LoaderSketch)
	if err != nil {
		feedback.Errorf("Error downloading loader sketch from %s: %s", board.LoaderSketch.URL, err)
		os.Exit(errorcodes.ErrGeneric)
	}
	logrus.Debugf("loader sketch downloaded in %s", loaderSketchPath.String())

	loaderSketch := strings.ReplaceAll(loaderSketchPath.String(), loaderSketchPath.Ext(), "")

	programmerOut, programmerErr, err := common.FlashSketch(board, loaderSketch, uploadToolDir, commonFlags.Address)
	if err != nil {
		feedback.Error(err)
		os.Exit(errorcodes.ErrGeneric)
	}

	// Wait a bit after flashing the loader sketch for the board to become
	// available again.
	logrus.Debug("sleeping for 3 sec")
	time.Sleep(3 * time.Second)

	// Get flasher depending on which module to use
	var f flasher.Flasher
	moduleName := board.Module

	// This matches the baudrate used in the FirmwareUpdater.ino sketch
	// https://github.com/arduino-libraries/WiFiNINA/blob/master/examples/Tools/FirmwareUpdater/FirmwareUpdater.ino
	const baudRate = 1000000
	switch moduleName {
	case "NINA":
		// we use address and not bootloaderPort because the board should not be in bootloader mode
		f, err = flasher.NewNinaFlasher(commonFlags.Address, baudRate, 30)
	case "WINC1500":
		f, err = flasher.NewWincFlasher(commonFlags.Address, baudRate, 30)
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
