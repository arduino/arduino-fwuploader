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

package indexes

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-fwuploader/cli/feedback"
	"github.com/arduino/arduino-fwuploader/indexes/download"
	"github.com/arduino/arduino-fwuploader/indexes/firmwareindex"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	semver "go.bug.st/relaxed-semver"
)

// GetToolRelease returns a ToolRelease by searching the toolID in the index.
// Returns nil if no matching tool release is found
// Assumes toolID is formatted correctly as <packager>:<tool_name>@<version>
func GetToolRelease(pm *packagemanager.PackageManager, toolID string) *cores.ToolRelease {
	split := strings.SplitN(toolID, ":", 2)
	packageName := split[0]
	split = strings.SplitN(split[1], "@", 2)
	toolName := split[0]
	version := semver.ParseRelaxed(split[1])

	pme, release := pm.NewExplorer()
	defer release()
	dep := &cores.ToolDependency{
		ToolName:     toolName,
		ToolVersion:  version,
		ToolPackager: packageName,
	}
	logrus.WithField("dep", dep).Debug("Tool dependency to download")
	toolRelease := pme.FindToolDependency(dep)
	if toolRelease == nil {
		feedback.Fatal(fmt.Sprintf("Can't find tool %s in index", dep), feedback.ErrGeneric)
	}
	logrus.WithField("tool", toolRelease.String()).Debug("Tool release to download")
	return toolRelease
}

// GetPackageIndex downloads and loads the Arduino package_index.json
func GetPackageIndex(pmbuilder *packagemanager.Builder, indexURL string) error {
	indexPath := paths.New(indexURL)
	if u, err := url.Parse(indexURL); err == nil && u.Scheme != "" {
		downloadedPath, err := download.DownloadIndex(indexURL)
		if err != nil {
			logrus.Error(err)
			return err
		}
		indexPath = downloadedPath
	}
	_, err := pmbuilder.LoadPackageIndexFromFile(indexPath)
	if err != nil {
		logrus.Error(err)
	}
	return err
}

// GetFirmwareIndex downloads and loads the arduino-fwuploader module_firmware_index.json
func GetFirmwareIndex(indexURL string, verifySignature bool) (*firmwareindex.Index, error) {
	indexPath := paths.New(indexURL)
	if u, err := url.Parse(indexURL); err == nil && u.Scheme != "" {
		downloadedPath, err := download.DownloadIndex(indexURL)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}
		indexPath = downloadedPath
	}

	var in *firmwareindex.Index
	var err error
	if verifySignature {
		in, err = firmwareindex.LoadIndex(indexPath)
	} else {
		in, err = firmwareindex.LoadIndexNoSign(indexPath)
	}
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return in, err
}
