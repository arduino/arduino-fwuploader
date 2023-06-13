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

package feedback

import (
	"encoding/json"
	"fmt"
	"os"
)

// ExitCode to be used for Fatal.
type ExitCode int

const (
	// Success (0 is the no-error return code in Unix)
	Success ExitCode = iota

	// ErrGeneric Generic error (1 is the reserved "catchall" code in Unix)
	ErrGeneric

	_ // (2 Is reserved in Unix)

	// ErrNoConfigFile is returned when the config file is not found (3)
	ErrNoConfigFile

	_ // (4 was ErrBadCall and has been removed)

	// ErrNetwork is returned when a network error occurs (5)
	ErrNetwork

	// ErrCoreConfig represents an error in the cli core config, for example some basic
	// files shipped with the installation are missing, or cannot create or get basic
	// directories vital for the CLI to work. (6)
	ErrCoreConfig

	// ErrBadArgument is returned when the arguments are not valid (7)
	ErrBadArgument
)

// OutputFormat is an output format
type OutputFormat int

const (
	// Text is the plain text format, suitable for interactive terminals
	Text OutputFormat = iota
	// JSON format
	JSON
)

var formats map[string]OutputFormat = map[string]OutputFormat{
	"json": JSON,
	"text": Text,
}

func (f OutputFormat) String() string {
	for res, format := range formats {
		if format == f {
			return res
		}
	}
	panic("unknown output format")
}

// ParseOutputFormat parses a string and returns the corresponding OutputFormat.
// The boolean returned is true if the string was a valid OutputFormat.
func ParseOutputFormat(in string) (OutputFormat, bool) {
	format, found := formats[in]
	return format, found
}

var (
	format         OutputFormat = Text
	formatSelected bool         = false
)

// Result is anything more complex than a sentence that needs to be printed
// for the user.
type Result interface {
	fmt.Stringer
	Data() interface{}
}

// ErrorResult is a result embedding also an error. In case of textual output
// the error will be printed on stderr.
type ErrorResult interface {
	Result
	ErrorString() string
}

// SetFormat can be used to change the output format at runtime
func SetFormat(f OutputFormat) {
	if formatSelected {
		panic("output format already selected")
	}
	format = f
	formatSelected = true
}

// GetFormat returns the output format currently set
func GetFormat() OutputFormat {
	return format
}

// FatalError outputs the error and exits with status exitCode.
func FatalError(err error, exitCode ExitCode) {
	Fatal(err.Error(), exitCode)
}

// FatalResult outputs the result and exits with status exitCode.
func FatalResult(res ErrorResult, exitCode ExitCode) {
	PrintResult(res)
	os.Exit(int(exitCode))
}

// Fatal outputs the errorMsg and exits with status exitCode.
func Fatal(errorMsg string, exitCode ExitCode) {
	if format == Text {
		fmt.Fprintln(os.Stderr, errorMsg)
		os.Exit(int(exitCode))
	}

	type FatalError struct {
		Error string `json:"error"`
	}
	res := &FatalError{
		Error: errorMsg,
	}
	var d []byte
	switch format {
	case JSON:
		d, _ = json.MarshalIndent(res, "", "  ")
	default:
		panic("unknown output format")
	}
	fmt.Fprintln(os.Stdout, string(d))
	os.Exit(int(exitCode))
}

// PrintResult is a convenient wrapper to provide feedback for complex data,
// where the contents can't be just serialized to JSON but requires more
// structure.
func PrintResult(res Result) {
	var data string
	var dataErr string
	switch format {
	case JSON:
		d, err := json.MarshalIndent(res.Data(), "", "  ")
		if err != nil {
			Fatal(fmt.Sprintf("Error during JSON encoding of the output: %v", err), ErrGeneric)
		}
		data = string(d)
	case Text:
		data = res.String()
		if resErr, ok := res.(ErrorResult); ok {
			dataErr = resErr.ErrorString()
		}
	default:
		panic("unknown output format")
	}
	if data != "" {
		fmt.Fprintln(os.Stdout, data)
	}
	if dataErr != "" {
		fmt.Fprintln(os.Stderr, dataErr)
	}
}
