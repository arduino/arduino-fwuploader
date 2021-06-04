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

package fwindex

import (
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

var DefaultIndexGZURL = []string{
	"http://downloads-dev.arduino.cc/arduino-fwuploader/boards/module_firmware_index.json.gz",
}

func TestIndexParsing(t *testing.T) {
	// semver.WarnInvalidVersionWhenParsingRelaxed = true
	list, err := paths.New("testdata").ReadDir()
	require.NoError(t, err)
	for _, indexFile := range list {
		if indexFile.Ext() != ".json" {
			continue
		}
		t.Logf("testing with index: %s", indexFile)
		Index, e := LoadIndexNoSign(indexFile)
		require.NoError(t, e)
		require.NotEmpty(t, Index)

		Index, e = LoadIndex(indexFile)
		require.NoError(t, e)
		require.NotEmpty(t, Index)
	}
}
