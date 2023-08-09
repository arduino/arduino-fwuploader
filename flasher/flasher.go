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

package flasher

// FlashResult contains the result of the flashing procedure
type FlashResult struct {
	Programmer *ExecOutput `json:"programmer"`
	Flasher    *ExecOutput `json:"flasher,omitempty"`
	Version    string      `json:"version,omitempty"`
}

// ExecOutput contains the stdout and stderr output, they are used to store the output of the flashing and upload
type ExecOutput struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

func (r *FlashResult) Data() interface{} {
	return r
}

func (r *FlashResult) String() string {
	// The output is already printed via os.Stdout/os.Stdin
	return ""
}
