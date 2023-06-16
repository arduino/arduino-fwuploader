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

package version

import "fmt"

var (
	defaultVersionString = "0.0.0-git"
	versionString        = ""
	commit               = ""
	date                 = ""
	// VersionInfo contains info regarding the version
	VersionInfo *info
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

// Data implements feedback.Result interface
func (i *info) Data() interface{} {
	return i
}

func init() {
	if versionString == "" {
		versionString = defaultVersionString
	}
	VersionInfo = newInfo("arduino-fwuploader")
}
