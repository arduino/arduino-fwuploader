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

package version

import "fmt"

var (
	defaultVersionString = "0.0.0-git"
	versionString        = ""
	commit               = ""
	date                 = ""
	VersionInfo          *info
)

type info struct {
	Application   string `json:"Application"`
	VersionString string `json:"VersionString"`
	Commit        string `json:"Commit"`
	Date          string `json:"Date"`
}

func newInfo(application string) *info {
	return &info{
		Application:   application,
		VersionString: versionString,
		Commit:        commit,
		Date:          date,
	}
}

func (i *info) String() string {
	return fmt.Sprintf("%s Version: %s Commit: %s Date: %s", i.Application, i.VersionString, i.Commit, i.Date)
}

func init() {
	if versionString == "" {
		versionString = defaultVersionString
	}
	VersionInfo = newInfo("FirmwareUploader")
}
