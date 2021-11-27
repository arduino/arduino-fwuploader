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

import (
	"os"

	"github.com/arduino/arduino-cli/cli/feedback"
	v "github.com/arduino/arduino-fwuploader/version"
	"github.com/spf13/cobra"
)

// NewCommand created a new `version` command
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Shows version number of arduino-fwuploader.",
		Long:    "Shows the version number of arduino-fwuploader which is installed on your system.",
		Example: "  " + os.Args[0] + " version",
		Args:    cobra.NoArgs,
		Run:     run,
	}
}

func run(cmd *cobra.Command, args []string) {
	feedback.Print(v.VersionInfo)
}
