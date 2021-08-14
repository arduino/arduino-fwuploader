/*
	arduino-fwuploader
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
	"testing"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packageindex"
	"github.com/arduino/arduino-cli/arduino/resources"
	"github.com/arduino/arduino-fwuploader/cli/globals"
	"github.com/arduino/arduino-fwuploader/indexes/firmwareindex"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
)

var defaultIndexGZURL = []string{
	"https://downloads.arduino.cc/packages/package_index.json.gz",
	"https://downloads.arduino.cc/arduino-fwuploader/boards/module_firmware_index.json.gz",
}

func TestDownloadIndex(t *testing.T) {
	defer os.RemoveAll(globals.FwUploaderPath.String()) // cleanup after tests
	for _, u := range defaultIndexGZURL {
		t.Logf("testing with index: %s", u)
		indexPath, err := DownloadIndex(u)
		require.NoError(t, err)
		require.DirExists(t, globals.FwUploaderPath.String())
		require.FileExists(t, indexPath.String())
		signaturePath := globals.FwUploaderPath.Join(indexPath.Base() + ".sig").String()
		require.FileExists(t, signaturePath)
	}
}

func TestDownloadTool(t *testing.T) {
	toolRelease := &cores.ToolRelease{
		Version: semver.ParseRelaxed("1.7.0-arduino3"),
		Tool: &cores.Tool{
			Name: "bossac",
		},
		Flavors: []*cores.Flavor{
			{
				OS: "i686-mingw32",
				Resource: &resources.DownloadResource{
					URL:             "http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-windows.tar.gz",
					ArchiveFileName: "bossac-1.7.0-arduino3-windows.tar.gz",
					Checksum:        "SHA-256:62745cc5a98c26949ec9041ef20420643c561ec43e99dae659debf44e6836526",
					Size:            3607421,
				},
			},
			{
				OS: "x86_64-apple-darwin",
				Resource: &resources.DownloadResource{
					URL:             "http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-osx.tar.gz",
					ArchiveFileName: "bossac-1.7.0-arduino3-osx.tar.gz",
					Checksum:        "SHA-256:adb3c14debd397d8135e9e970215c6972f0e592c7af7532fa15f9ce5e64b991f",
					Size:            75510,
				},
			},
			{
				OS: "x86_64-pc-linux-gnu",
				Resource: &resources.DownloadResource{
					URL:             "http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-linux64.tar.gz",
					ArchiveFileName: "bossac-1.7.0-arduino3-linux64.tar.gz",
					Checksum:        "SHA-256:1ae54999c1f97234a5c603eb99ad39313b11746a4ca517269a9285afa05f9100",
					Size:            207271,
				},
			},
			{
				OS: "i686-pc-linux-gnu",
				Resource: &resources.DownloadResource{
					URL:             "http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-linux32.tar.gz",
					ArchiveFileName: "bossac-1.7.0-arduino3-linux32.tar.gz",
					Checksum:        "SHA-256:4ac4354746d1a09258f49a43ef4d1baf030d81c022f8434774268b00f55d3ec3",
					Size:            193577,
				},
			},
			{
				OS: "arm-linux-gnueabihf",
				Resource: &resources.DownloadResource{
					URL:             "http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-linuxarm.tar.gz",
					ArchiveFileName: "bossac-1.7.0-arduino3-linuxarm.tar.gz",
					Checksum:        "SHA-256:626c6cc548046901143037b782bf019af1663bae0d78cf19181a876fb9abbb90",
					Size:            193941,
				},
			},
			{
				OS: "aarch64-linux-gnu",
				Resource: &resources.DownloadResource{
					URL:             "http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-linuxaarch64.tar.gz",
					ArchiveFileName: "bossac-1.7.0-arduino3-linuxaarch64.tar.gz",
					Checksum:        "SHA-256:a098b2cc23e29f0dc468416210d097c4a808752cd5da1a7b9b8b7b931a04180b",
					Size:            268365,
				},
			},
		},
	}
	defer os.RemoveAll(globals.FwUploaderPath.String())
	indexFile := paths.New("testdata/package_index.json")
	t.Logf("testing with index: %s", indexFile)
	index, e := packageindex.LoadIndexNoSign(indexFile)
	require.NoError(t, e)
	require.NotEmpty(t, index)
	toolDir, err := DownloadTool(toolRelease)
	require.NoError(t, err)
	require.NotEmpty(t, toolDir)
	require.DirExists(t, toolDir.String())
	toolDirContent, err := toolDir.ReadDir()
	require.NoError(t, err)
	require.True(t, len(toolDirContent) > 0)
}

func TestDownloadFirmware(t *testing.T) {
	defer os.RemoveAll(globals.FwUploaderPath.String())
	indexFile := paths.New("testdata/module_firmware_index.json")
	t.Logf("testing with index: %s", indexFile)
	index, e := firmwareindex.LoadIndexNoSign(indexFile)
	require.NoError(t, e)
	require.NotEmpty(t, index)
	firmwarePath, err := DownloadFirmware(index.Boards[0].Firmwares[0])
	require.NoError(t, err)
	require.NotEmpty(t, firmwarePath)
	require.FileExists(t, firmwarePath.String())
}

func TestDownloadLoaderSketch(t *testing.T) {
	defer os.RemoveAll(globals.FwUploaderPath.String())
	indexFile := paths.New("testdata/module_firmware_index.json")
	t.Logf("testing with index: %s", indexFile)
	index, e := firmwareindex.LoadIndexNoSign(indexFile)
	require.NoError(t, e)
	require.NotEmpty(t, index)
	loaderPath, err := DownloadLoaderSketch(index.Boards[0].LoaderSketch)
	require.NoError(t, err)
	require.NotEmpty(t, loaderPath)
	require.FileExists(t, loaderPath.String())
}
