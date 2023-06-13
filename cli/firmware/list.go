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

package firmware

import (
	"os"

	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/arduino-fwuploader/cli/feedback"
	"github.com/arduino/arduino-fwuploader/indexes"
	"github.com/spf13/cobra"
	semver "go.bug.st/relaxed-semver"
)

func newListCommand() *cobra.Command {
	var fqbn *string

	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "List available firmwares",
		Long:    "Displays the availale firmwares, is it possible to filter results for a specific board.",
		Example: "  " + os.Args[0] + " firmware list -b arduino:samd:mkr1000",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			list(*fqbn)
		},
	}
	fqbn = listCmd.Flags().StringP("fqbn", "b", "", "Filter result for the specified board FQBN")
	return listCmd
}

type FirmwareResult struct {
	BoardName       string                 `json:"board_name"`
	BoardFQBN       string                 `json:"board_fqbn"`
	Module          string                 `json:"module"`
	FirmwareVersion *semver.RelaxedVersion `json:"firmware_version"`
	Latest          bool
}

type FirmwareListResult []*FirmwareResult

func list(fqbn string) {
	firmwareIndex, err := indexes.GetFirmwareIndex()
	if err != nil {
		feedback.FatalError(err, feedback.ErrGeneric)
	}

	res := FirmwareListResult{}
	for _, board := range firmwareIndex.Boards {
		if fqbn == "" || board.Fqbn == fqbn {
			for _, firmware := range board.Firmwares {
				res = append(res, &FirmwareResult{
					BoardName:       board.Name,
					BoardFQBN:       board.Fqbn,
					Module:          board.Module,
					FirmwareVersion: firmware.Version,
					Latest:          board.LatestFirmware == firmware,
				})
			}
		}
	}

	feedback.PrintResult(res)
}

func (f FirmwareListResult) String() string {
	if len(f) == 0 {
		return "No firmwares available."
	}
	t := table.New()
	t.SetHeader("Board", "FQBN", "Module", "", "Version")
	for _, fw := range f {
		latest := ""
		if fw.Latest {
			latest = "✔"
		}
		t.AddRow(fw.BoardName, fw.BoardFQBN, fw.Module, latest, fw.FirmwareVersion)
	}
	return t.Render()
}

func (f FirmwareListResult) Data() interface{} {
	return f
}
