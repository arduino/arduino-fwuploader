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
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

var DefaultIndexGZURL = []string{
	"https://downloads.arduino.cc/packages/package_index.json.gz",
	"http://downloads-dev.arduino.cc/arduino-fwuploader/boards/module_firmware_index.json.gz",
}

func TestDownloadIndex(t *testing.T) {
	for _, u := range DefaultIndexGZURL {
		t.Logf("testing with index: %s", u)
		err := DownloadIndex(u)
		require.NoError(t, err)
		indexFolder := paths.TempDir().Join("fwuploader")
		require.DirExists(t, indexFolder.String())
		URL, err := utils.URLParse(u)
		require.NoError(t, err)
		indexPath := indexFolder.Join(path.Base(strings.ReplaceAll(URL.Path, ".gz", "")))
		require.FileExists(t, indexPath.String())
		sigURL, err := url.Parse(URL.String())
		require.NoError(t, err)
		sigURL.Path = strings.ReplaceAll(sigURL.Path, "gz", "sig")
		signaturePath := indexFolder.Join(path.Base(sigURL.Path)).String()
		require.FileExists(t, signaturePath)
		defer os.RemoveAll(indexFolder.String()) // cleanup afer tests
	}
}
