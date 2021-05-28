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

package globals

var DefaultIndexURL = []string{
	"https://downloads.arduino.cc/packages/package_index.json",
	// "http://downloads-dev.arduino.cc/arduino-fwuploader/arduino-fwuploader/boards/board_index.json", // the index currently do not have the signature
	// There is no sugnature, and the path is not correct see fwuploader/fwuploader. Also add downloads-dev
}
