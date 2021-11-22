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
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestIndexParsing(t *testing.T) {
	indexFile := paths.New("testdata/module_firmware_index.json")
	t.Logf("testing with index: %s", indexFile)
	index, e := LoadIndexNoSign(indexFile)
	require.NoError(t, e)
	require.NotEmpty(t, index)

	index, e = LoadIndex(indexFile)
	require.NoError(t, e)
	require.NotEmpty(t, index)
}

func TestGetBoard(t *testing.T) {
	indexFile := paths.New("testdata/module_firmware_index.json")
	t.Logf("testing with index: %s", indexFile)
	index, e := LoadIndexNoSign(indexFile)
	require.NoError(t, e)
	require.NotEmpty(t, index)

	board := index.GetBoard("arduino:samd:mkr1000")
	require.NotNil(t, board)
	require.Equal(t, board.Fqbn, "arduino:samd:mkr1000")

	board = index.GetBoard("arduino:samd:nano_33_iot")
	require.NotNil(t, board)
	require.Equal(t, board.Fqbn, "arduino:samd:nano_33_iot")

	board = index.GetBoard("arduino:avr:nessuno")
	require.Nil(t, board)
}

func TestGetLatestFirmware(t *testing.T) {
	indexFile := paths.New("testdata/module_firmware_index.json")
	t.Logf("testing with index: %s", indexFile)
	index, e := LoadIndexNoSign(indexFile)
	require.NoError(t, e)
	require.NotEmpty(t, index)

	firmware := index.GetBoard("arduino:samd:mkr1000").LatestFirmware
	require.Equal(t, firmware.Version.String(), "19.6.1")
}

func TestGetFirmware(t *testing.T) {
	indexFile := paths.New("testdata/module_firmware_index.json")
	t.Logf("testing with index: %s", indexFile)
	index, e := LoadIndexNoSign(indexFile)
	require.NoError(t, e)
	require.NotEmpty(t, index)

	firmware := index.GetBoard("arduino:samd:mkr1000").GetFirmware("19.6.1")
	require.Equal(t, firmware.Version.String(), "19.6.1")

	firmware = index.GetBoard("arduino:samd:mkr1000").GetFirmware("19.5.2")
	require.Equal(t, firmware.Version.String(), "19.5.2")

	firmware = index.GetBoard("arduino:samd:mkr1000").GetFirmware("0.0.0")
	require.Nil(t, firmware)
}
