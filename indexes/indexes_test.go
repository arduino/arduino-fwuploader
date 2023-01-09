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
	"testing"

	"github.com/arduino/arduino-cli/arduino/cores/packageindex"
	"github.com/arduino/arduino-fwuploader/cli/globals"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestGetToolRelease(t *testing.T) {
	indexFile := paths.New("testdata/package_index.json")
	index, err := packageindex.LoadIndexNoSign(indexFile)
	require.NoError(t, err)
	toolRelease := GetToolRelease(index, "arduino:bossac@1.7.0-arduino3")

	require.Equal(t, toolRelease.Version.String(), "1.7.0-arduino3")
	require.Equal(t, toolRelease.Tool.Name, "bossac")
	require.NotEmpty(t, toolRelease.Flavors)
}

func TestGetPackageIndex(t *testing.T) {
	index, err := GetPackageIndex()
	require.NoError(t, err)
	require.NotNil(t, index)
}

func TestGetFirmwareIndex(t *testing.T) {
	index, err := GetFirmwareIndex()
	require.NoError(t, err)
	require.NotNil(t, index)
	require.NoDirExists(t, globals.FwUploaderPath.String())
}
