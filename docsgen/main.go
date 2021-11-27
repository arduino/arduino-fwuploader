/*
	Source: https://github.com/arduino/tooling-project-assets/blob/main/workflow-templates/assets/cobra/docsgen/main.go

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

// Package main generates Markdown documentation for the project's CLI.
package main

import (
	"os"

	"github.com/arduino/arduino-fwuploader/cli"
	"github.com/spf13/cobra/doc"
)

func main() {
	if len(os.Args) < 2 {
		print("error: Please provide the output folder argument")
		os.Exit(1)
	}

	os.MkdirAll(os.Args[1], 0755) // Create the output folder if it doesn't already exist

	cli := cli.NewCommand()
	cli.DisableAutoGenTag = true // Disable addition of auto-generated date stamp
	err := doc.GenMarkdownTree(cli, os.Args[1])
	if err != nil {
		panic(err)
	}
}
