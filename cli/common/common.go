/*
	Copyright 2021 Arduino SA

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

package common

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/serialutils"
	"github.com/arduino/arduino-fwuploader/cli/feedback"
	"github.com/arduino/arduino-fwuploader/cli/globals"
	"github.com/arduino/arduino-fwuploader/indexes"
	"github.com/arduino/arduino-fwuploader/indexes/download"
	"github.com/arduino/arduino-fwuploader/indexes/firmwareindex"
	programmer "github.com/arduino/arduino-fwuploader/programmers"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/sirupsen/logrus"
)

// AdditionalPackageIndexURLs is a list of additional package_index.json URLs that
// are loaded together with the main index.
var AdditionalPackageIndexURLs []string

// AdditionalFirmwareIndexURLs is a list of additional module_firmware_index.json URLs that
// are loaded together with the main index.
var AdditionalFirmwareIndexURLs []string

// InitIndexes downloads and parses the package_index.json and firmwares_index.json
func InitIndexes() (*packagemanager.PackageManager, *firmwareindex.Index) {
	// Load main package index and optional additional indexes
	pmbuilder := packagemanager.NewBuilder(nil, nil, nil, nil, "")
	if err := indexes.GetPackageIndex(pmbuilder, globals.PackageIndexGZURL); err != nil {
		feedback.Fatal(fmt.Sprintf("Can't load package index: %s", err), feedback.ErrGeneric)
	}
	for _, indexURL := range AdditionalPackageIndexURLs {
		if err := indexes.GetPackageIndex(pmbuilder, indexURL); err != nil {
			feedback.Fatal(fmt.Sprintf("Can't load firmware index: %s", err), feedback.ErrGeneric)
		}
	}

	// Load main firmware index and optional additional indexes
	firmwareIndex, err := indexes.GetFirmwareIndex(globals.ModuleFirmwareIndexGZURL, true)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("Can't load firmware index: %s", err), feedback.ErrGeneric)
	}
	if pluginIndex, err := indexes.GetFirmwareIndex(globals.PluginFirmwareIndexGZURL, true); err != nil {
		feedback.Fatal(fmt.Sprintf("Can't load (plugin) firmware index: %s", err), feedback.ErrGeneric)
	} else {
		firmwareIndex.MergeWith(pluginIndex)
	}
	for _, additionalURL := range AdditionalFirmwareIndexURLs {
		additionalIndex, err := indexes.GetFirmwareIndex(additionalURL, false)
		if err != nil {
			feedback.Fatal(fmt.Sprintf("Can't load firmware index: %s", err), feedback.ErrGeneric)
		}
		firmwareIndex.MergeWith(additionalIndex)
	}

	return pmbuilder.Build(), firmwareIndex
}

// CheckFlags runs a basic check, errors if the flags are not defined
func CheckFlags(fqbn, address string) {
	if fqbn == "" {
		feedback.Fatal("Error during firmware flashing: missing board fqbn", feedback.ErrBadArgument)
	}

	if address == "" {
		feedback.Fatal("Error during firmware flashing: missing board address", feedback.ErrBadArgument)
	}
	logrus.Debugf("fqbn: %s, address: %s", fqbn, address)
}

// GetBoard is an helper function useful to get the IndexBoard,
// the struct that contains all the infos to make all the operations possible
func GetBoard(firmwareIndex *firmwareindex.Index, fqbn string) *firmwareindex.IndexBoard {
	board := firmwareIndex.GetBoard(fqbn)
	if board == nil {
		feedback.Fatal(fmt.Sprintf("Can't find board with %s fqbn", fqbn), feedback.ErrBadArgument)
	}
	logrus.Debugf("got board: %s", board.Fqbn)
	return board
}

// DownloadRequiredToolsForBoard is an helper function that downloads the correct tool to flash a board,
// it returns the path of the downloaded tool
func DownloadRequiredToolsForBoard(pm *packagemanager.PackageManager, board *firmwareindex.IndexBoard) *paths.Path {
	if !board.IsPlugin() {
		// Just download the upload tool for integrated uploaders
		return downloadTool(pm, board.Uploader)
	}

	// Download the plugin
	toolDir := downloadTool(pm, board.UploaderPlugin)

	// Also download the other additional tools
	for _, tool := range board.AdditionalTools {
		_ = downloadTool(pm, tool)
	}

	return toolDir
}

func downloadTool(pm *packagemanager.PackageManager, tool string) *paths.Path {
	toolRelease := indexes.GetToolRelease(pm, tool)
	if toolRelease == nil {
		feedback.Fatal(fmt.Sprintf("Error getting upload tool %s", tool), feedback.ErrGeneric)
	}
	toolDir, err := download.DownloadTool(toolRelease)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("Error downloading tool %s: %s", tool, err), feedback.ErrGeneric)
	}
	logrus.Debugf("upload tool downloaded in %s", toolDir.String())
	return toolDir
}

// FlashSketch is the business logic that handles the flashing procedure,
// it returns using a buffer the stdout and the stderr of the programmer
func FlashSketch(board *firmwareindex.IndexBoard, sketch string, uploadToolDir *paths.Path, address string) (programmerOut, programmerErr *bytes.Buffer, err error) {
	bootloaderPort, err := GetNewAddress(board, address)
	if err != nil {
		return nil, nil, err
	}

	uploaderCommand := board.GetUploaderCommand()
	uploaderCommand = strings.ReplaceAll(uploaderCommand, "{tool_dir}", filepath.FromSlash(uploadToolDir.String()))
	uploaderCommand = strings.ReplaceAll(uploaderCommand, "{serial.port.file}", bootloaderPort)
	uploaderCommand = strings.ReplaceAll(uploaderCommand, "{loader.sketch}", sketch) // we leave that name here because it's only a template,

	logrus.Debugf("uploading with command: %s", uploaderCommand)
	commandLine, err := properties.SplitQuotedString(uploaderCommand, "\"", false)
	if err != nil {
		feedback.Fatal(fmt.Sprintf(`Error splitting command line "%s": %s`, uploaderCommand, err), feedback.ErrGeneric)
	}

	// Flash the actual sketch
	programmerOut = new(bytes.Buffer)
	programmerErr = new(bytes.Buffer)
	if feedback.GetFormat() == feedback.JSON {
		err = programmer.Flash(commandLine, programmerOut, programmerErr)
	} else {
		err = programmer.Flash(commandLine, os.Stdout, os.Stderr)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("error during sketch flashing: %s", err)
	}
	return programmerOut, programmerErr, err
}

// GetNewAddress is a function used to reset a board and put it in bootloader mode
// it could happen that the board is assigned to a different serial port, after the reset,
// this fuction handles also this possibility
func GetNewAddress(board *firmwareindex.IndexBoard, oldAddress string) (string, error) {
	// Check if board needs a 1200bps touch for upload
	bootloaderPort := oldAddress
	if board.UploadTouch {
		logrus.Info("Putting board into bootloader mode")
		newUploadPort, err := serialutils.Reset(oldAddress, board.UploadWait, nil, false)
		if err != nil {
			return "", fmt.Errorf("error during sketch flashing: missing board address. %s", err)
		}
		if newUploadPort != "" {
			logrus.Infof("Found port to upload: %s", newUploadPort)
			bootloaderPort = newUploadPort
		}
	}
	return bootloaderPort, nil
}
