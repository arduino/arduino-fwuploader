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

	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-fwuploader/cli/globals"
	"github.com/stretchr/testify/require"
)

func TestGetPackageIndex(t *testing.T) {
	pmb := packagemanager.NewBuilder(nil, nil, nil, nil, "")
	err := GetPackageIndex(pmb, globals.PackageIndexGZURL)
	require.NoError(t, err)
}

func TestGetFirmwareIndex(t *testing.T) {
	index, err := GetFirmwareIndex(globals.PluginFirmwareIndexGZURL, true)
	require.NoError(t, err)
	require.NotNil(t, index)
}
