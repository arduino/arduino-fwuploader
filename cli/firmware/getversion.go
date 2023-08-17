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

	"github.com/arduino/arduino-fwuploader/cli/common"
	"github.com/arduino/arduino-fwuploader/cli/feedback"
	"github.com/arduino/arduino-fwuploader/cli/globals"
	"github.com/arduino/arduino-fwuploader/flasher"
	"github.com/arduino/arduino-fwuploader/plugin"
	"github.com/spf13/cobra"
)

// NewGetVersionCommand creates a new `get-version` command
func NewGetVersionCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "get-version",
		Short: "Gets the version of the firmware the board is using.",
		Long:  "Flashes a sketch to a board to obtain the firmware version used by the board",
		Example: "" +
			"  " + os.Args[0] + " firmware get-version --fqbn arduino:samd:mkrwifi1010 --address COM10\n" +
			"  " + os.Args[0] + " firmware get-version -b arduino:renesas_uno:unor4wifi -a COM10\n",
		Args: cobra.NoArgs,
		Run:  runGetVersion,
	}
	commonFlags.AddToCommand(command)
	return command
}

func runGetVersion(cmd *cobra.Command, args []string) {
	// at the end cleanup the fwuploader temp dir
	defer globals.FwUploaderPath.RemoveAll()

	common.CheckFlags(commonFlags.Fqbn, commonFlags.Address)
	packageIndex, firmwareIndex := common.InitIndexes()
	board := common.GetBoard(firmwareIndex, commonFlags.Fqbn)
	uploadToolDir := common.DownloadRequiredToolsForBoard(packageIndex, board)

	uploader, err := plugin.NewFWUploaderPlugin(uploadToolDir)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("Could not open uploader plugin: %s", err), feedback.ErrGeneric)
	}

	result := getVersion(uploader)
	if feedback.GetFormat() == feedback.Text {
		fmt.Printf("Firmware version installed: %s", result.Version)
	} else {
		feedback.PrintResult(result)
	}
}

func getVersion(uploader *plugin.FwUploader) *flasher.FlashResult {
	var stdout, stderr io.Writer
	if feedback.GetFormat() == feedback.Text {
		stdout = os.Stdout
		stderr = os.Stderr
	}
	res, err := uploader.GetFirmwareVersion(commonFlags.Address, commonFlags.Fqbn, globals.LogLevel, globals.Verbose, stdout, stderr)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("Couldn't get firmware version: %s", err), feedback.ErrGeneric)
	}

	return &flasher.FlashResult{
		Programmer: (&flasher.ExecOutput{
			Stdout: string(res.Stdout),
			Stderr: string(res.Stderr),
		}),
		Version: res.FirmwareVersion.String(),
	}
}
