/*
	Copyright 2021 Arduino SA

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

package arguments

import (
	"github.com/spf13/cobra"
)

// Flags contains various common flags.
// This is useful so all flags used by commands that need
// this information are consistent with each other.
type Flags struct {
	Address string
	Fqbn    string
}

// AddToCommand adds the flags used to set address and fqbn to the specified Command
func (f *Flags) AddToCommand(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.Fqbn, "fqbn", "b", "", "Fully Qualified Board Name, e.g.: arduino:samd:mkr1000, arduino:mbed_nano:nanorp2040connect")
	cmd.Flags().StringVarP(&f.Address, "address", "a", "", "Upload port, e.g.: COM10, /dev/ttyACM0")
}
