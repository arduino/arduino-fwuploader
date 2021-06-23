/*
  arduino-fwuploader
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

package main

import (
	"log"
	"os"

	"github.com/arduino/arduino-fwuploader/cli"
	"github.com/spf13/cobra/doc"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide output folder")
	}

	cli := cli.NewCommand()
	cli.DisableAutoGenTag = true // Disable addition of auto-generated date stamp
	err := doc.GenMarkdownTree(cli, os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
}
