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
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/arduino/FirmwareUploader/cli/version"
	"github.com/arduino/FirmwareUploader/indexes"
	"github.com/arduino/FirmwareUploader/modules/nina"
	"github.com/arduino/FirmwareUploader/modules/sara"
	"github.com/arduino/FirmwareUploader/modules/winc"
	"github.com/arduino/FirmwareUploader/utils"
	"github.com/arduino/FirmwareUploader/utils/context"
	v "github.com/arduino/FirmwareUploader/version"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/go-paths-helper"
	"github.com/mattn/go-colorable"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	ctx          = &context.Context{}
	outputFormat string
	verbose      bool
	logFile      string
	logFormat    string
	logLevel     string
)

func NewCommand() *cobra.Command {
	// FirmwareUploader is the root command
	firmwareUploaderCli := &cobra.Command{
		Use:              "FirmwareUploader",
		Short:            "FirmwareUploader.",
		Long:             "FirmwareUploader (FirmwareUploader).",
		Example:          "  " + os.Args[0] + " <command> [flags...]",
		Args:             cobra.NoArgs,
		Run:              run,
		PersistentPreRun: preRun,
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

	firmwareUploaderCli.PersistentFlags().StringVar(&outputFormat, "format", "text", "The output format, can be {text|json}.")

	firmwareUploaderCli.PersistentFlags().StringVar(&logFile, "log-file", "", "Path to the file where logs will be written")
	firmwareUploaderCli.PersistentFlags().StringVar(&logFormat, "log-format", "", "The output format for the logs, can be {text|json}.")
	firmwareUploaderCli.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Messages with this level and above will be logged. Valid levels are: trace, debug, info, warn, error, fatal, panic")
	firmwareUploaderCli.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print the logs on the standard output.")

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

// Convert the string passed to the `--log-level` option to the corresponding
// logrus formal level.
func toLogLevel(s string) (t logrus.Level, found bool) {
	t, found = map[string]logrus.Level{
		"trace": logrus.TraceLevel,
		"debug": logrus.DebugLevel,
		"info":  logrus.InfoLevel,
		"warn":  logrus.WarnLevel,
		"error": logrus.ErrorLevel,
		"fatal": logrus.FatalLevel,
		"panic": logrus.PanicLevel,
	}[s]

	return
}

func parseFormatString(arg string) (feedback.OutputFormat, bool) {
	f, found := map[string]feedback.OutputFormat{
		"json": feedback.JSON,
		"text": feedback.Text,
	}[arg]

	return f, found
}

func preRun(cmd *cobra.Command, args []string) {
	// Prepare logging
	if verbose {
		// if we print on stdout, do it in full colors
		logrus.SetOutput(colorable.NewColorableStdout())
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors: true,
		})
	} else {
		logrus.SetOutput(ioutil.Discard)
	}

	// Normalize the format strings
	logFormat = strings.ToLower(logFormat)
	if logFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	logFile := ""
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("Unable to open file for logging: %s", logFile)
			os.Exit(errorcodes.ErrBadCall)
		}

		// Use a hook so we don't get color codes in the log file
		if outputFormat == "json" {
			logrus.AddHook(lfshook.NewHook(file, &logrus.JSONFormatter{}))
		} else {
			logrus.AddHook(lfshook.NewHook(file, &logrus.TextFormatter{}))
		}
	}

	// Configure logging filter
	if lvl, found := toLogLevel(logLevel); !found {
		feedback.Errorf("Invalid option for --log-level: %s", logLevel)
		os.Exit(errorcodes.ErrBadArgument)
	} else {
		logrus.SetLevel(lvl)
	}

	indexes.DownloadIndex()

	//
	// Prepare the Feedback system
	//

	// normalize the format strings
	outputFormat = strings.ToLower(outputFormat)
	// check the right output format was passed
	format, found := parseFormatString(outputFormat)
	if !found {
		feedback.Errorf("Invalid output format: %s", outputFormat)
		os.Exit(errorcodes.ErrBadCall)
	}

	// use the output format to configure the Feedback
	feedback.SetFormat(format)

	logrus.Info(v.VersionInfo)

	if outputFormat != "text" {
		cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
			logrus.Warn("Calling help on JSON format")
			feedback.Error("Invalid Call : should show Help, but it is available only in TEXT mode.")
			os.Exit(errorcodes.ErrBadCall)
		})
	}
}
