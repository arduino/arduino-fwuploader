package common

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-cli/arduino/cores/packageindex"
	"github.com/arduino/arduino-cli/arduino/serialutils"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-fwuploader/indexes"
	"github.com/arduino/arduino-fwuploader/indexes/download"
	"github.com/arduino/arduino-fwuploader/indexes/firmwareindex"
	programmer "github.com/arduino/arduino-fwuploader/programmers"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/sirupsen/logrus"
)

// InitIndexes does exactly what the name implies
func InitIndexes() (*packageindex.Index, *firmwareindex.Index) {
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
	return packageIndex, firmwareIndex
}

// CheckFlags runs a basic check, errors if the flags are not defined
func CheckFlags(fqbn, address string) {
	if fqbn == "" {
		feedback.Errorf("Error during firmware flashing: missing board fqbn")
		os.Exit(errorcodes.ErrBadArgument)
	}

	if address == "" {
		feedback.Errorf("Error during firmware flashing: missing board address")
		os.Exit(errorcodes.ErrBadArgument)
	}
	logrus.Debugf("fqbn: %s, address: %s", fqbn, address)
}

// GetBoard is an helper function useful to get the IndexBoard,
// the struct that contains all the infos to make all the operations possible
func GetBoard(firmwareIndex *firmwareindex.Index, fqbn string) *firmwareindex.IndexBoard {
	board := firmwareIndex.GetBoard(fqbn)
	if board == nil {
		feedback.Errorf("Can't find board with %s fqbn", fqbn)
		os.Exit(errorcodes.ErrBadArgument)
	}
	logrus.Debugf("got board: %s", board.Fqbn)
	return board
}

// GetUploadToolDir is an helper function that downloads the correct tool to flash a board,
// it returns the path of the downloaded tool
func GetUploadToolDir(packageIndex *packageindex.Index, board *firmwareindex.IndexBoard) *paths.Path {
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
	logrus.Debugf("upload tool downloaded in %s", uploadToolDir.String())
	return uploadToolDir
}

// flashSketch is the business logic that handles the flashing procedure,
// it returns using a buffer the out and the err of the programmer
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
		feedback.Errorf(`Error splitting command line "%s": %s`, uploaderCommand, err)
		os.Exit(errorcodes.ErrGeneric)
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

// getNewAddress is a function used to reset a board and put it in bootloader mode
// it could happen that the board is assigned to a different serial port, after the reset,
// this fuction handles also this possibility
func GetNewAddress(board *firmwareindex.IndexBoard, oldAddress string) (string, error) {
	// Check if board needs a 1200bps touch for upload
	bootloaderPort := oldAddress
	if board.UploadTouch {
		logrus.Info("Putting board into bootloader mode")
		newUploadPort, err := serialutils.Reset(oldAddress, board.UploadWait, nil)
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
