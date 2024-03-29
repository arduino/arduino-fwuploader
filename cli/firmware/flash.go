/*
	arduino-fwuploader
	Copyright (c) 2021 Arduino LLC.  All right reserved.

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published
	by the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package firmware

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/arduino/arduino-fwuploader/cli/arguments"
	"github.com/arduino/arduino-fwuploader/cli/common"
	"github.com/arduino/arduino-fwuploader/cli/feedback"
	"github.com/arduino/arduino-fwuploader/cli/globals"
	"github.com/arduino/arduino-fwuploader/flasher"
	"github.com/arduino/arduino-fwuploader/indexes/download"
	"github.com/arduino/arduino-fwuploader/indexes/firmwareindex"
	"github.com/arduino/arduino-fwuploader/plugin"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	commonFlags arguments.Flags // contains fqbn and address
	module      string
	retries     int
	fwFile      string
)

// NewFlashCommand creates a new `flash` command
func NewFlashCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "flash",
		Short: "Flashes firmwares to board.",
		Long:  "Flashes specified module firmware to board at specified address. Module name and version can be omitted to install latest version.",
		Example: "" +
			"  " + os.Args[0] + " firmware flash --fqbn arduino:samd:mkrwifi1010 --address COM10 --module NINA@1.4.8\n" +
			"  " + os.Args[0] + " firmware flash -b arduino:renesas_uno:unor4wifi -a COM10 -m ESP32-S3\n" +
			"  " + os.Args[0] + " firmware flash -b arduino:renesas_uno:unor4wifi -a COM10\n" +
			"  " + os.Args[0] + " firmware flash -b arduino:samd:mkrwifi1010 -a COM10 -i firmware.bin\n",
		Args: cobra.NoArgs,
		Run:  runFlash,
	}
	commonFlags.AddToCommand(command)
	command.Flags().StringVarP(&module, "module", "m", "", "Firmware module ID, e.g.: ESP32-S3, NINA")
	command.Flags().IntVar(&retries, "retries", 9, "Number of retries in case of upload failure (default 9)")
	command.Flags().StringVarP(&fwFile, "input-file", "i", "", "Path of the firmware to upload")
	return command
}

func runFlash(cmd *cobra.Command, args []string) {
	// at the end cleanup the fwuploader temp dir
	defer globals.FwUploaderPath.RemoveAll()

	if retries < 1 {
		feedback.Fatal("Number of retries should be at least 1", feedback.ErrBadArgument)
	}

	common.CheckFlags(commonFlags.Fqbn, commonFlags.Address)
	packageIndex, firmwareIndex := common.InitIndexes()
	board := common.GetBoard(firmwareIndex, commonFlags.Fqbn)
	uploadToolDir := common.DownloadRequiredToolsForBoard(packageIndex, board)

	// Get module name if not specified
	moduleName := board.Module
	moduleVersion := ""
	if module != "" {
		split := strings.SplitN(module, "@", 2)
		moduleName = split[0]
		if len(split) == 2 {
			moduleVersion = split[1]
		}
	}
	// Normalize module name
	moduleName = strings.ToUpper(moduleName)

	var firmwareFilePath *paths.Path
	// If a local firmware file has been specified
	if fwFile != "" {
		firmwareFilePath = paths.New(fwFile)
		if !firmwareFilePath.Exist() {
			feedback.Fatal(fmt.Sprintf("firmware file not found in %s", firmwareFilePath), feedback.ErrGeneric)
		}
	} else {
		// Download the firmware
		var firmware *firmwareindex.IndexFirmware
		if moduleVersion == "" {
			firmware = board.LatestFirmware()
		} else {
			firmware = board.GetFirmware(moduleVersion)
		}
		if firmware == nil {
			feedback.Fatal(fmt.Sprintf("Error getting firmware for board: %s", commonFlags.Fqbn), feedback.ErrGeneric)
		}
		logrus.Debugf("module name: %s, firmware version: %s", firmware.Module, firmware.Version.String())
		if fwPath, err := download.DownloadFirmware(firmware); err != nil {
			feedback.Fatal(fmt.Sprintf("Error downloading firmware from %s: %s", firmware.URL, err), feedback.ErrGeneric)
		} else {
			firmwareFilePath = fwPath
		}
		logrus.Debugf("firmware file downloaded in %s", firmwareFilePath.String())
	}

	uploader, err := plugin.NewFWUploaderPlugin(uploadToolDir)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("Could not open uploader plugin: %s", err), feedback.ErrGeneric)
	}

	retry := 0
	for {
		retry++
		logrus.Infof("Uploading firmware (try %d of %d)", retry, retries)

		res, err := updateFirmware(uploader, firmwareFilePath)
		if err == nil {
			feedback.PrintResult(res)
			logrus.Info("Operation completed: success! :-)")
			break
		}
		logrus.Error(err)

		if retry == retries {
			logrus.Fatal("Operation failed. :-(")
		}

		logrus.Info("Waiting 1 second before retrying...")
		time.Sleep(time.Second)
	}
}

func updateFirmware(uploader *plugin.FwUploader, fwPath *paths.Path) (*flasher.FlashResult, error) {
	var stdout, stderr io.Writer
	if feedback.GetFormat() == feedback.Text {
		stdout = os.Stdout
		stderr = os.Stderr
	}
	res, err := uploader.FlashFirmware(commonFlags.Address, commonFlags.Fqbn, globals.LogLevel, globals.Verbose, fwPath, stdout, stderr)
	if err != nil {
		return nil, fmt.Errorf("couldn't update firmware: %s", err)
	}
	return &flasher.FlashResult{
		Programmer: &flasher.ExecOutput{
			Stdout: string(res.Stdout),
			Stderr: string(res.Stderr),
		},
	}, nil
}

// callback used to print the progress
func printProgress(progress int) {
	fmt.Printf("Flashing progress: %d%%\r", progress)
}
