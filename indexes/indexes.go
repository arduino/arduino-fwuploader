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
	"strings"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packageindex"
	"github.com/arduino/arduino-cli/arduino/resources"
	"github.com/arduino/arduino-fwuploader/cli/globals"
	"github.com/arduino/arduino-fwuploader/indexes/download"
	"github.com/arduino/arduino-fwuploader/indexes/firmwareindex"
	"github.com/sirupsen/logrus"
	semver "go.bug.st/relaxed-semver"
)

// GetToolRelease returns a ToolRelease by searching the toolID in the index.
// Returns nil if no matching tool release is found
// Assumes toolID is formatted correctly as <packager>:<tool_name>@<version>
func GetToolRelease(index *packageindex.Index, toolID string) *cores.ToolRelease {
	split := strings.Split(toolID, ":")
	packageName := split[0]
	split = strings.Split(split[1], "@")
	toolName := split[0]
	version := semver.ParseRelaxed(split[1])
	for _, pack := range index.Packages {
		if pack.Name != packageName {
			continue
		}
		for _, tool := range pack.Tools {
			if tool.Name == toolName && tool.Version.Equal(version) {
				flavors := []*cores.Flavor{}
				for _, system := range tool.Systems {
					size, _ := system.Size.Int64()
					flavors = append(flavors, &cores.Flavor{
						OS: system.OS,
						Resource: &resources.DownloadResource{
							URL:             system.URL,
							ArchiveFileName: system.ArchiveFileName,
							Checksum:        system.Checksum,
							Size:            size,
						},
					})
				}
				return &cores.ToolRelease{
					Version: version,
					Flavors: flavors,
					Tool: &cores.Tool{
						Name: toolName,
					},
				}
			}
		}
	}
	return nil
}

// GetPackageIndex downloads and loads the Arduino package_index.json
func GetPackageIndex() (*packageindex.Index, error) {
	indexPath, err := download.DownloadIndex(globals.PackageIndexGZURL)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	in, err := packageindex.LoadIndex(indexPath)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return in, err
}

// GetFirmwareIndex downloads and loads the arduino-fwuploader module_firmware_index.json
func GetFirmwareIndex() (*firmwareindex.Index, error) {
	defer globals.FwUploaderPath.RemoveAll()
	indexPath, err := download.DownloadIndex(globals.ModuleFirmwareIndexGZURL)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	in, err := firmwareindex.LoadIndex(indexPath)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return in, err
}
