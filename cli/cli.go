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

package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/arduino/arduino-fwuploader/cli/certificates"
	"github.com/arduino/arduino-fwuploader/cli/firmware"
	"github.com/arduino/arduino-fwuploader/cli/version"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	v "github.com/arduino/arduino-fwuploader/version"
	"github.com/mattn/go-colorable"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	verbose      bool
	logFile      string
	logFormat    string
	logLevel     string
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
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Messages with this level and above will be logged. Valid levels are: trace, debug, info, warn, error, fatal, panic")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print the logs on the standard output.")

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

	// Prepare the Feedback system

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
