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
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

var DefaultIndexURL = []string{
	"https://downloads.arduino.cc/packages/package_index.json",
	// "http://downloads-dev.arduino.cc/arduino-fwuploader/arduino-fwuploader/boards/board_index.json", // the index currently do not have the signature
	// There is no sugnature, and the path is not correct see fwuploader/fwuploader. Also add downloads-dev
}

func TestDownloadIndex(t *testing.T) {
	for _, u := range DefaultIndexURL {
		t.Logf("testing with index: %s", u)
		err := DownloadIndex(u)
		require.NoError(t, err)
		indexPath := paths.TempDir().Join("fwuploader")
		require.DirExists(t, indexPath.String())
		URL, err := utils.URLParse(u)
		require.NoError(t, err)
		packageIndex := indexPath.Join(path.Base(URL.Path)).String()
		require.FileExists(t, packageIndex)
		sigURL, err := url.Parse(URL.String())
		require.NoError(t, err)
		sigURL.Path += ".sig"
		signature := indexPath.Join(path.Base(sigURL.Path)).String()
		require.FileExists(t, signature)
		os.RemoveAll(indexPath.String()) // cleanup afer tests
	}
}
