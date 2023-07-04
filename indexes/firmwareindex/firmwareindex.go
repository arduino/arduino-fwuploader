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

package firmwareindex

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/arduino/arduino-cli/arduino/security"
	"github.com/arduino/arduino-fwuploader/cli/globals"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	semver "go.bug.st/relaxed-semver"
)

// Index represents Boards struct as seen from module_firmware_index.json file.
type Index struct {
	Boards    []*IndexBoard
	IsTrusted bool
}

// IndexBoard represents a single entry from module_firmware_index.json file.
type IndexBoard struct {
	Fqbn      string           `json:"fqbn"`
	Firmwares []*IndexFirmware `json:"firmware"`
	Module    string           `json:"module"`
	Name      string           `json:"name"`

	// Fields required for integrated uploaders (deprecated)
	LoaderSketch    *IndexSketch          `json:"loader_sketch"`
	VersionSketch   *IndexSketch          `json:"version_sketch"`
	Uploader        string                `json:"uploader"`
	UploadTouch     bool                  `json:"upload.use_1200bps_touch"`
	UploadWait      bool                  `json:"upload.wait_for_upload_port"`
	UploaderCommand *IndexUploaderCommand `json:"uploader.command"`

	// Fields required for plugin uploaders
	UploaderPlugin  string   `json:"uploader_plugin"`
	AdditionalTools []string `json:"additional_tools"`
}

// IndexUploaderCommand represents the command-line to use for different OS
type IndexUploaderCommand struct {
	Linux   string `json:"linux"`
	Windows string `json:"windows"`
	Macosx  string `json:"macosx"`
}

// IndexFirmware represents a single Firmware version from module_firmware_index.json file.
type IndexFirmware struct {
	Version  *semver.RelaxedVersion `json:"version"`
	URL      string                 `json:"url"`
	Checksum string                 `json:"checksum"`
	Size     json.Number            `json:"size"`
	Module   string                 `json:"module"`
}

// IndexSketch represents a sketch used to manage firmware on a board.
type IndexSketch struct {
	URL      string      `json:"url"`
	Checksum string      `json:"checksum"`
	Size     json.Number `json:"size"`
}

// LoadIndex reads a module_firmware_index.json from a file and returns the corresponding Index structure.
func LoadIndex(jsonIndexFile *paths.Path) (*Index, error) {
	index, err := LoadIndexNoSign(jsonIndexFile)
	if err != nil {
		return nil, err
	}

	jsonSignatureFile := jsonIndexFile.Parent().Join(jsonIndexFile.Base() + ".sig")
	arduinoKeyringFile, err := globals.Keys.Open("keys/module_firmware_index_public.gpg.key")
	if err != nil {
		return nil, fmt.Errorf("could not find bundled signature keys: %s", err)

	}
	defer arduinoKeyringFile.Close()
	trusted, _, err := security.VerifySignature(jsonIndexFile, jsonSignatureFile, arduinoKeyringFile)
	if err != nil {
		logrus.
			WithField("index", jsonIndexFile).
			WithField("signatureFile", jsonSignatureFile).
			WithError(err).Infof("Checking signature")
		return nil, err
	}
	logrus.
		WithField("index", jsonIndexFile).
		WithField("signatureFile", jsonSignatureFile).
		WithField("trusted", trusted).Infof("Checking signature")
	index.IsTrusted = trusted
	return index, nil
}

// LoadIndexNoSign reads a module_firmware_index.json from a file and returns the corresponding Index structure.
func LoadIndexNoSign(jsonIndexFile *paths.Path) (*Index, error) {
	buff, err := jsonIndexFile.ReadFile()
	if err != nil {
		return nil, err
	}
	var index Index
	err = json.Unmarshal(buff, &index.Boards)
	if err != nil {
		return nil, err
	}

	index.IsTrusted = true
	return &index, nil
}

// GetBoard returns the IndexBoard for the given FQBN
func (i *Index) GetBoard(fqbn string) *IndexBoard {
	for _, b := range i.Boards {
		if b.Fqbn == fqbn {
			return b
		}
	}
	return nil
}

// GetFirmware returns the specified IndexFirmware version for this board.
// Returns nil if version is not found.
func (b *IndexBoard) GetFirmware(version string) *IndexFirmware {
	v := semver.ParseRelaxed(version)
	for _, firmware := range b.Firmwares {
		if firmware.Version.Equal(v) {
			return firmware
		}
	}
	return nil
}

// GetUploaderCommand returns the command to use for the upload
func (b *IndexBoard) GetUploaderCommand() string {
	if runtime.GOOS == "windows" && b.UploaderCommand.Windows != "" {
		return b.UploaderCommand.Linux
	} else if runtime.GOOS == "darwin" && b.UploaderCommand.Macosx != "" {
		return b.UploaderCommand.Macosx
	}
	// The linux uploader command is considere to be the generic one
	return b.UploaderCommand.Linux
}

// LatestFirmware returns the latest firmware version for the IndexBoard
func (b *IndexBoard) LatestFirmware() *IndexFirmware {
	var latest *IndexFirmware
	for _, firmware := range b.Firmwares {
		if latest == nil || firmware.Version.GreaterThan(latest.Version) {
			latest = firmware
		}
	}
	return latest
}

// IsPlugin returns true if the IndexBoard uses the plugin system
func (b *IndexBoard) IsPlugin() bool {
	return b.UploaderPlugin != ""
}
