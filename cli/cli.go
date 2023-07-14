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

package cli

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/arduino/arduino-fwuploader/cli/certificates"
	"github.com/arduino/arduino-fwuploader/cli/common"
	"github.com/arduino/arduino-fwuploader/cli/feedback"
	"github.com/arduino/arduino-fwuploader/cli/firmware"
	"github.com/arduino/arduino-fwuploader/cli/globals"
	"github.com/arduino/arduino-fwuploader/cli/version"
	v "github.com/arduino/arduino-fwuploader/version"
	"github.com/mattn/go-colorable"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	outputFormat           string
	logFile                string
	logFormat              string
	additionalFirmwareURLs []string
	additionalPackageURLs  []string
)

func NewCommand() *cobra.Command {
	// arduino-fwuploader is the root command
	rootCmd := &cobra.Command{
		Use:              "arduino-fwuploader",
		Short:            "arduino-fwuploader.",
		Long:             "Arduino Firmware Uploader (arduino-fwuploader).",
		Example:          "  " + os.Args[0] + " <command> [flags...]",
		Args:             cobra.NoArgs,
		PersistentPreRun: preRun,
	}

	rootCmd.AddCommand(version.NewCommand())
	rootCmd.AddCommand(firmware.NewCommand())
	rootCmd.AddCommand(certificates.NewCommand())

	rootCmd.PersistentFlags().StringVar(&outputFormat, "format", "text", "The output format, can be {text|json}.")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", "", "Path to the file where logs will be written")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "", "The output format for the logs, can be {text|json}.")
	rootCmd.PersistentFlags().StringVar(&globals.LogLevel, "log-level", "info", "Messages with this level and above will be logged. Valid levels are: trace, debug, info, warn, error, fatal, panic")
	rootCmd.PersistentFlags().BoolVarP(&globals.Verbose, "verbose", "v", false, "Print the logs on the standard output.")
	rootCmd.PersistentFlags().StringArrayVarP(&additionalFirmwareURLs, "additional-fw-index", "F", nil, "Additional firmwares index URLs (useful for testing purposes)")
	rootCmd.PersistentFlags().StringArrayVarP(&additionalPackageURLs, "additional-packages-index", "P", nil, "Additional packages index URLs (useful for testing purposes)")
	return rootCmd
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

func preRun(cmd *cobra.Command, args []string) {
	// Prepare the Feedback system
	// check the right output format was passed
	format, found := feedback.ParseOutputFormat(outputFormat)
	if !found {
		feedback.Fatal(fmt.Sprintf("Invalid output format: %s", outputFormat), feedback.ErrBadArgument)
	}
	feedback.SetFormat(format)

	// Prepare logging
	if globals.Verbose {
		// if we print on stdout, do it in full colors
		logrus.SetOutput(colorable.NewColorableStdout())
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors: true,
		})
	} else {
		logrus.SetOutput(ioutil.Discard)
	}

	// Normalize the format strings
	if logFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	logFile := ""
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			feedback.Fatal(fmt.Sprintf("Unable to open file for logging: %s", logFile), feedback.ErrBadArgument)
		}

		// Use a hook so we don't get color codes in the log file
		if outputFormat == "json" {
			logrus.AddHook(lfshook.NewHook(file, &logrus.JSONFormatter{}))
		} else {
			logrus.AddHook(lfshook.NewHook(file, &logrus.TextFormatter{}))
		}
	}

	// Configure logging filter
	if lvl, found := toLogLevel(globals.LogLevel); !found {
		feedback.Fatal(fmt.Sprintf("Invalid option for --log-level: %s", globals.LogLevel), feedback.ErrBadArgument)
	} else {
		logrus.SetLevel(lvl)
	}

	logrus.Info(v.VersionInfo)

	// Setup additional indexes
	common.AdditionalPackageIndexURLs = append(common.AdditionalPackageIndexURLs, additionalPackageURLs...)
	common.AdditionalFirmwareIndexURLs = append(common.AdditionalFirmwareIndexURLs, additionalFirmwareURLs...)
}
