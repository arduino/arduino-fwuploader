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
	"log"
	"os"
	"strings"
	"time"

	"github.com/arduino/arduino-fwuploader/cli/common"
	"github.com/arduino/arduino-fwuploader/cli/feedback"
	"github.com/arduino/arduino-fwuploader/cli/globals"
	"github.com/arduino/arduino-fwuploader/flasher"
	"github.com/arduino/arduino-fwuploader/indexes/download"
	"github.com/arduino/arduino-fwuploader/indexes/firmwareindex"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	semver "go.bug.st/relaxed-semver"
)

// NewGetVersionCommand creates a new `get-version` command
func NewGetVersionCommand() *cobra.Command {

	command := &cobra.Command{
		Use:   "get-version",
		Short: "Gets the version of the firmware the board is using.",
		Long:  "Flashes a sketch to a board to obtain the firmware version used by the board",
		Example: "" +
			"  " + os.Args[0] + " firmware get-version --fqbn arduino:samd:mkr1000 --address COM10\n" +
			"  " + os.Args[0] + " firmware get-version -b arduino:samd:mkr1000 -a COM10\n",
		Args: cobra.NoArgs,
		Run:  runGetVersion,
	}
	commonFlags.AddToCommand(command)
	return command
}

func runGetVersion(cmd *cobra.Command, args []string) {
	// at the end cleanup the fwuploader temp dir
	defer globals.FwUploaderPath.RemoveAll()

	packageIndex, firmwareIndex := common.InitIndexes()
	common.CheckFlags(commonFlags.Fqbn, commonFlags.Address)
	board := common.GetBoard(firmwareIndex, commonFlags.Fqbn)
	uploadToolDir := common.GetUploadToolDir(packageIndex, board)

	versionSketchPath, err := download.DownloadSketch(board.VersionSketch)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("Error downloading loader sketch from %s: %s", board.LoaderSketch.URL, err), feedback.ErrGeneric)
	}
	logrus.Debugf("version sketch downloaded in %s", versionSketchPath.String())

	versionSketch := strings.ReplaceAll(versionSketchPath.String(), versionSketchPath.Ext(), "")

	programmerOut, programmerErr, err := common.FlashSketch(board, versionSketch, uploadToolDir, commonFlags.Address)
	if err != nil {
		feedback.FatalError(err, feedback.ErrGeneric)
	}

	// Wait a bit after flashing the sketch for the board to become available again.
	logrus.Debug("sleeping for 3 sec")
	time.Sleep(3 * time.Second)

	currentVersion, err := getVersion(board)
	if err != nil {
		feedback.FatalError(err, feedback.ErrGeneric)
	}
	if feedback.GetFormat() == feedback.Text {
		fmt.Printf("Firmware version installed: %s", currentVersion)
	} else {
		// Print the results
		feedback.PrintResult(&flasher.FlashResult{
			Programmer: (&flasher.ExecOutput{
				Stdout: programmerOut.String(),
				Stderr: programmerErr.String(),
			}),
			Version: currentVersion,
		})
	}
}

func getVersion(board *firmwareindex.IndexBoard) (fwVersion string, err error) {

	// 9600 is the baudrate used in the CheckVersion sketch
	port, err := flasher.OpenSerial(commonFlags.Address, 9600, 2)
	if err != nil {
		feedback.FatalError(err, feedback.ErrGeneric)
	}

	buff := make([]byte, 200)
	serialResult := make([]byte, 0)
	for {
		n, err := port.Read(buff)
		if err != nil {
			log.Fatal(err)
			break
		}
		serialResult = append(serialResult, buff[:n]...)
		if n == 0 { // exit when done reading from serial
			break
		}
		logrus.Info(string(buff[:n]))
	}
	lines := strings.Split(string(serialResult), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Firmware version installed: ") {
			version := strings.TrimSpace(strings.Replace(line, "Firmware version installed: ", "", 1))
			semver := semver.ParseRelaxed(version)
			return semver.String(), nil
		}
		if strings.HasPrefix(line, "Communication with WiFi module failed!") {
			return "", fmt.Errorf("communication with WiFi module failed")
		}
	}
	return "", fmt.Errorf("could not find the version string to parse")
}
