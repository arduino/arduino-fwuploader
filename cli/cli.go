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

package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/arduino/FirmwareUploader/cli/version"
	"github.com/arduino/FirmwareUploader/modules/nina"
	"github.com/arduino/FirmwareUploader/modules/sara"
	"github.com/arduino/FirmwareUploader/modules/winc"
	"github.com/arduino/FirmwareUploader/utils"
	"github.com/arduino/FirmwareUploader/utils/context"
	"github.com/arduino/go-paths-helper"
	"github.com/spf13/cobra"
)

var ctx = &context.Context{}

func NewCommand() *cobra.Command {
	// FirmwareUploader is the root command
	firmwareUploaderCli := &cobra.Command{
		Use:     "FirmwareUploader",
		Short:   "FirmwareUploader.",
		Long:    "FirmwareUploader (FirmwareUploader).",
		Example: "  " + os.Args[0] + " <command> [flags...]",
		Run:     run,
	}

	firmwareUploaderCli.AddCommand(version.NewCommand())

	firmwareUploaderCli.Flags().StringVar(&ctx.PortName, "port", "", "serial port to use for flashing")
	firmwareUploaderCli.Flags().StringVar(&ctx.RootCertDir, "certs", "", "root certificate directory")
	firmwareUploaderCli.Flags().StringSliceVar(&ctx.Addresses, "address", []string{}, "address (host:port) to fetch and flash root certificate for, multiple values allowed")
	firmwareUploaderCli.Flags().StringVar(&ctx.FirmwareFile, "firmware", "", "firmware file to flash")
	firmwareUploaderCli.Flags().BoolVar(&ctx.ReadAll, "read", false, "read all firmware and output to stdout")
	firmwareUploaderCli.Flags().StringVar(&ctx.FWUploaderBinary, "flasher", "", "firmware upload binary (precompiled for the right target)")
	firmwareUploaderCli.Flags().StringVar(&ctx.BinaryToRestore, "restore_binary", "", "binary to restore after the firmware upload (precompiled for the right target)")
	firmwareUploaderCli.Flags().StringVar(&ctx.ProgrammerPath, "programmer", "", "path of programmer in use (avrdude/bossac)")
	firmwareUploaderCli.Flags().StringVar(&ctx.Model, "model", "", "module model (winc, nina or sara)")
	firmwareUploaderCli.Flags().StringVar(&ctx.BoardName, "get_available_for", "", "Ask for available firmwares matching a given board")
	firmwareUploaderCli.Flags().IntVar(&ctx.Retries, "retries", 9, "Number of retries in case of upload failure")

	return firmwareUploaderCli
}

func run(cmd *cobra.Command, args []string) {
	if ctx.BoardName != "" {
		el, _ := json.Marshal(utils.GetCompatibleWith(ctx.BoardName, ""))
		fmt.Println(string(el))
		os.Exit(0)
	}

	if ctx.PortName == "" {
		log.Fatal("Please specify a serial port")
	}

	if ctx.BinaryToRestore != "" {
		// sanity check for BinaryToRestore
		f := paths.New(ctx.BinaryToRestore)
		info, err := f.Stat()
		if err != nil {
			log.Fatalf("Error opening restore_binary: %s", err)
		}
		if info.IsDir() {
			log.Fatalf("Error opening restore_binary: is a directory...")
		}
		if info.Size() == 0 {
			log.Println("WARNING: restore_binary is empty! Will not restore binary after upload.")
			ctx.BinaryToRestore = ""
		}
	}

	retry := 0
	for {
		var err error
		if ctx.Model == "nina" || strings.Contains(ctx.FirmwareFile, "NINA") || strings.Contains(ctx.FWUploaderBinary, "NINA") {
			err = nina.Run(ctx)
		} else if ctx.Model == "winc" || strings.Contains(ctx.FirmwareFile, "WINC") || strings.Contains(ctx.FWUploaderBinary, "WINC") {
			err = winc.Run(ctx)
		} else {
			err = sara.Run(ctx)
		}
		if err == nil {
			log.Println("Operation completed: success! :-)")
			break
		}
		log.Println("Error: " + err.Error())

		if retry >= ctx.Retries {
			log.Fatal("Operation failed. :-(")
		}

		retry++
		log.Println("Waiting 1 second before retrying...")
		time.Sleep(time.Second)
		log.Printf("Retrying upload (%d of %d)", retry, ctx.Retries)
	}
}
