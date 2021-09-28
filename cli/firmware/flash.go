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

package firmware

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
	"github.com/arduino/arduino-fwuploader/indexes/firmwareindex"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	commonFlags arguments.Flags // contains fqbn and address
	module      string
	retries     uint8
	fwFile      string
)

// NewFlashCommand creates a new `flash` command
func NewFlashCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "flash",
		Short: "Flashes firmwares to board.",
		Long:  "Flashes specified module firmware to board at specified address. Module name and version can be omitted to install latest version.",
		Example: "" +
			"  " + os.Args[0] + " firmware flash --fqbn arduino:samd:mkr1000 --address COM10 --module WINC1500@19.5.2\n" +
			"  " + os.Args[0] + " firmware flash -b arduino:samd:mkr1000 -a COM10 -m WINC15000\n" +
			"  " + os.Args[0] + " firmware flash -b arduino:samd:mkr1000 -a COM10\n" +
			"  " + os.Args[0] + " firmware flash -b arduino:samd:mkr1000 -a COM10 -i firmware.bin\n",
		Args: cobra.NoArgs,
		Run:  runFlash,
	}
	commonFlags.AddToCommand(command)
	command.Flags().StringVarP(&module, "module", "m", "", "Firmware module ID, e.g.: WINC1500, NINA")
	command.Flags().Uint8Var(&retries, "retries", 9, "Number of retries in case of upload failure (default 9)")
	command.Flags().StringVarP(&fwFile, "input-file", "i", "", "Path of the firmware to upload")
	return command
}

func runFlash(cmd *cobra.Command, args []string) {

	packageIndex, firmwareIndex := common.InitIndexes()
	common.CheckFlags(commonFlags.Fqbn, commonFlags.Address)
	board := common.GetBoard(firmwareIndex, commonFlags.Fqbn)
	uploadToolDir := common.GetUploadToolDir(packageIndex, board)

	// Get module name if not specified
	moduleName := ""
	moduleVersion := ""
	if module == "" {
		moduleName = board.Module
	} else {
		moduleSplit := strings.Split(module, "@")
		if len(moduleSplit) == 2 {
			moduleName = moduleSplit[0]
			moduleVersion = moduleSplit[1]
		} else {
			moduleName = module
		}
	}
	// Normalize module name
	moduleName = strings.ToUpper(moduleName)

	var firmwareFilePath *paths.Path
	var err error
	// If a local firmware file has been specified
	if fwFile != "" {
		firmwareFilePath = paths.New(fwFile)
		if !firmwareFilePath.Exist() {
			feedback.Errorf("firmware file not found in %s", firmwareFilePath)
			os.Exit(errorcodes.ErrGeneric)
		}
	} else {
		// Download the firmware
		var firmware *firmwareindex.IndexFirmware
		if moduleVersion == "" {
			firmware = board.LatestFirmware
		} else {
			firmware = board.GetFirmware(moduleVersion)
		}
		logrus.Debugf("module name: %s, firmware version: %s", firmware.Module, firmware.Version.String())
		if firmware == nil {
			feedback.Errorf("Error getting firmware for board: %s", commonFlags.Fqbn)
			os.Exit(errorcodes.ErrGeneric)
		}
		firmwareFilePath, err = download.DownloadFirmware(firmware)
		if err != nil {
			feedback.Errorf("Error downloading firmware from %s: %s", firmware.URL, err)
			os.Exit(errorcodes.ErrGeneric)
		}
		logrus.Debugf("firmware file downloaded in %s", firmwareFilePath.String())
	}

	loaderSketchPath, err := download.DownloadSketch(board.LoaderSketch)
	if err != nil {
		feedback.Errorf("Error downloading loader sketch from %s: %s", board.LoaderSketch.URL, err)
		os.Exit(errorcodes.ErrGeneric)
	}
	logrus.Debugf("loader sketch downloaded in %s", loaderSketchPath.String())

	loaderSketch := strings.ReplaceAll(loaderSketchPath.String(), loaderSketchPath.Ext(), "")

	for retry := 1; retry <= int(retries); retry++ {
		err = updateFirmware(board, loaderSketch, moduleName, uploadToolDir, firmwareFilePath)
		if err == nil {
			logrus.Info("Operation completed: success! :-)")
			break
		}
		feedback.Error(err)
		if retry == int(retries) {
			logrus.Fatal("Operation failed. :-(")
		}
		logrus.Info("Waiting 1 second before retrying...")
		time.Sleep(time.Second)
		logrus.Infof("Retrying upload (%d of %d)", retry, retries)
	}
}

func updateFirmware(board *firmwareindex.IndexBoard, loaderSketch, moduleName string, uploadToolDir, firmwareFile *paths.Path) error {
	programmerOut, programmerErr, err := common.FlashSketch(board, loaderSketch, uploadToolDir, commonFlags.Address)
	if err != nil {
		return err
	}
	// Wait a bit after flashing the loader sketch for the board to become
	// available again.
	logrus.Debug("sleeping for 3 sec")
	time.Sleep(3 * time.Second)

	// Get flasher depending on which module to use
	var f flasher.Flasher

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
		feedback.Errorf("Error during firmware flashing: %s", err)
		os.Exit(errorcodes.ErrGeneric)
	}
	if err != nil {
		feedback.Errorf("Error during firmware flashing: %s", err)
		return err
	}
	defer f.Close()

	// now flash the actual firmware
	flasherOut := new(bytes.Buffer)
	flasherErr := new(bytes.Buffer)
	if feedback.GetFormat() == feedback.JSON {
		err = f.FlashFirmware(firmwareFile, flasherOut)
	} else {
		f.SetProgressCallback(printProgress)
		err = f.FlashFirmware(firmwareFile, os.Stdout)
	}
	if err != nil {
		flasherErr.Write([]byte(fmt.Sprintf("Error during firmware flashing: %s", err)))
		return err
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
	return nil
}

// callback used to print the progress
func printProgress(progress int) {
	fmt.Printf("Flashing progress: %d%%\r", progress)
}
