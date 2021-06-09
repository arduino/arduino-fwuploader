/*
  FirmwareUploader
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

package indexes

import (
	"strings"

	"github.com/arduino/FirmwareUploader/cli/globals"
	"github.com/arduino/FirmwareUploader/indexes/download"
	"github.com/arduino/FirmwareUploader/indexes/firmwareindex"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packageindex"
	"github.com/arduino/arduino-cli/arduino/resources"
	"github.com/arduino/go-paths-helper"
	semver "go.bug.st/relaxed-semver"
)

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

func downloadIndexes() (*paths.Path, error) {

}

func GetPackageIndex() (*packageindex.Index, error) {
	if err := download.DownloadIndex(globals.PackageIndexGZURL)

}

func GetFirmwareIndex() (*firmwareindex.Index, error) {

}

// download indexes in /tmp/fwuloader/package_index.json etc..
// for _, u := range globals.DefaultIndexGZURL {
// 	indexes.DownloadIndex(u)
// }

// list, err := globals.FwUploaderPath.ReadDir()
// if err != nil {
// 	feedback.Errorf("Can't read fwuploader directory: %s", err)
// }
// for _, indexFile := range list {
// 	if indexFile.Ext() != ".json" {
// 		continue
// 	}
// 	if indexFile.String() == "package_index.json" {
// 		PackageIndex, e := packageindex.LoadIndexNoSign(indexFile) // TODO fare funzione che ti ritorna le strutture dati, e fa tutto quello che ci sta dietro.
// 	} else if indexFile.String() == "module_firmware_index.json" {
// 		ModuleFWIndex, e := firmwareindex.LoadIndexNoSign(indexFile)
// 	} else {
// 		feedback.Errorf("Unknown index: %s", indexFile.String())
// 	}
// }

// //TODO ⬇️ study in the CLI how the indexes are passed to other modules
