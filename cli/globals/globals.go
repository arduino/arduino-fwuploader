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

package globals

import (
	"embed"

	"github.com/arduino/go-paths-helper"
)

var (
	PackageIndexGZURL        = "https://downloads.arduino.cc/packages/package_index.json.gz"
	PluginFirmwareIndexGZURL = "https://downloads.arduino.cc/arduino-fwuploader/boards/plugin_firmware_index.json.gz"
	FwUploaderPath           = paths.TempDir().Join("fwuploader")
	Verbose                  bool
	LogLevel                 string
)

//go:embed keys/*
var Keys embed.FS
