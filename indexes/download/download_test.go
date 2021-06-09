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

package download

import (
	"os"
	"strings"
	"testing"

	"github.com/arduino/FirmwareUploader/cli/globals"
	"github.com/arduino/FirmwareUploader/indexes/firmwareindex"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packageindex"
	"github.com/arduino/arduino-cli/arduino/resources"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
)

// TODO not working because of include loop
// func TestDownloadTool(t *testing.T) {
// 	defer os.RemoveAll(globals.FwUploaderPath.String()) // cleanup after tests
// 	t.Logf("testing with index: %s", packageIndexGZURL)
// 	err := DownloadIndex(packageIndexGZURL)
// 	require.NoError(t, err)
// 	require.DirExists(t, globals.FwUploaderPath.String())
// 	URL, err := utils.URLParse(packageIndexGZURL)
// 	require.NoError(t, err)
// 	indexPath := globals.FwUploaderPath.Join(path.Base(strings.ReplaceAll(URL.Path, ".gz", "")))
// 	require.FileExists(t, indexPath.String())
// 	sigURL, err := url.Parse(URL.String())
// 	require.NoError(t, err)
// 	sigURL.Path = strings.ReplaceAll(sigURL.Path, "gz", "sig")
// 	signaturePath := globals.FwUploaderPath.Join(path.Base(sigURL.Path)).String()
// 	require.FileExists(t, signaturePath)
// }

func TestDownloadTool(t *testing.T) {
	defer os.RemoveAll(globals.FwUploaderPath.String())
	// semver.WarnInvalidVersionWhenParsingRelaxed = true
	list, err := paths.New("testdata").ReadDir()
	require.NoError(t, err)
	list.FilterSuffix("package_index.json")
	for _, indexFile := range list {
		t.Logf("testing with index: %s", indexFile)
		index, e := packageindex.LoadIndexNoSign(indexFile)
		require.NoError(t, e)
		require.NotEmpty(t, index)
		tool := GetToolRelease(index, "arduino:bossac@1.7.0-arduino3")
		toolPath, err := DownloadTool(tool)
		// TODO verify that the tool is installed
		require.NoError(t, err)
		require.NotEmpty(t, toolPath)
		require.FileExists(t, toolPath.String())
	}
}

func TestDownloadFirmware(t *testing.T) {
	defer os.RemoveAll(globals.FwUploaderPath.String())
	list, err := paths.New("../testdata").ReadDir() // TODO fix this
	require.NoError(t, err)
	list.FilterSuffix("module_firmware_index.json")
	for _, indexFile := range list {
		t.Logf("testing with index: %s", indexFile)
		index, e := firmwareindex.LoadIndexNoSign(indexFile)
		require.NoError(t, e)
		require.NotEmpty(t, index)
		firmwarePath, err := DownloadFirmware(index.Boards[0].Firmwares[0])
		require.NoError(t, err)
		require.NotEmpty(t, firmwarePath)
		require.FileExists(t, firmwarePath.String())
	}
}

func TestDownloadLoaderSketch(t *testing.T) {
	defer os.RemoveAll(globals.FwUploaderPath.String())
	list, err := paths.New("../testdata").ReadDir() // TODO fix this
	require.NoError(t, err)
	list.FilterSuffix("module_firmware_index.json")
	for _, indexFile := range list {
		t.Logf("testing with index: %s", indexFile)
		index, e := firmwareindex.LoadIndexNoSign(indexFile)
		require.NoError(t, e)
		require.NotEmpty(t, index)
		loaderPath, err := DownloadLoaderSketch(index.Boards[0].LoaderSketch)
		require.NoError(t, err)
		require.NotEmpty(t, loaderPath)
		require.FileExists(t, loaderPath.String())
	}
}

func GetToolRelease(index *packageindex.Index, toolID string) *cores.ToolRelease { // TODO put this logic in index.go in cli
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
