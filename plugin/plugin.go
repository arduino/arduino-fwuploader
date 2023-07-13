/*
	arduino-fwuploader
	Copyright (c) 2023 Arduino LLC.  All right reserved.

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

package plugin

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/go-paths-helper"
	semver "go.bug.st/relaxed-semver"
	"gopkg.in/yaml.v3"
)

// FwUploadr is an helper to run a fwuploader-plugin.
type FwUploader struct {
	pluginPath *paths.Path
	apiVersion int
}

// NewFWUploaderPlugin creates a new FWUploader, pluginDir must point to the plugin directory.
func NewFWUploaderPlugin(pluginDir *paths.Path) (*FwUploader, error) {
	files, err := pluginDir.ReadDirRecursiveFiltered(
		paths.FilterNames(),
		paths.AndFilter(
			paths.FilterOutDirectories(),
			paths.FilterOutPrefixes("LICENSE"),
		),
	)
	if err != nil {
		return nil, err
	}
	if len(files) != 1 {
		return nil, fmt.Errorf("invalid uploader-plugin in %s: multiple files in the root dir", pluginDir)
	}
	uploader := &FwUploader{
		pluginPath: files[0],
	}

	apiVersion, err := uploader.QueryAPIVersion()
	if err != nil {
		return nil, fmt.Errorf("error getting plugin version %s: %w", pluginDir, err)
	}
	uploader.apiVersion = apiVersion
	return uploader, nil
}

// QueryAPIVersion queries the plugin API version
func (uploader *FwUploader) QueryAPIVersion() (int, error) {
	proc, err := executils.NewProcessFromPath(nil, uploader.pluginPath, "version")
	if err != nil {
		return 0, err
	}
	stdout, _, err := proc.RunAndCaptureOutput(context.Background())
	if err != nil {
		return 0, err
	}

	var result struct {
		ApiVersion int `yaml:"plugin_api_version"`
	}
	if err := yaml.Unmarshal(stdout, &result); err != nil {
		return 0, err
	}

	return result.ApiVersion, nil
}

// GetFirmwareVersion runs the plugin to obtain the version of the installed firmware
func (uploader *FwUploader) GetFirmwareVersion(portAddress, fqbn, LogLevel string, verbose bool, stdout, stderr io.Writer) (*GetFirmwareVersionResult, error) {
	args := []string{"firmware", "get-version"}
	if portAddress != "" {
		args = append(args, "-p", portAddress)
	}
	if fqbn != "" {
		args = append(args, "-b", fqbn)
	}
	if verbose {
		args = append(args, "-v")
	}
	if LogLevel != "" {
		args = append(args, "--log-level", LogLevel)
	}
	execStdout, execStderr, execErr := uploader.exec(stdout, stderr, args...)

	res := &GetFirmwareVersionResult{
		Stdout: execStdout.Bytes(),
		Stderr: execStderr.Bytes(),
	}
	fwVersionPrefix := "FIRMWARE-VERSION: "
	fwErrorPrefix := "GET-VERSION-ERROR: "
	for _, line := range strings.Split(execStdout.String(), "\n") {
		if strings.HasPrefix(line, fwVersionPrefix) {
			version := strings.TrimPrefix(line, fwVersionPrefix)
			res.FirmwareVersion = semver.ParseRelaxed(version)
		}
		if strings.HasPrefix(line, fwErrorPrefix) {
			res.Error = strings.TrimPrefix(line, fwErrorPrefix)
		}
	}

	if res.Error != "" {
		if execErr != nil {
			execErr = fmt.Errorf("%s: %w", res.Error, execErr)
		} else {
			execErr = errors.New(res.Error)
		}
	}
	return res, execErr
}

// GetFirmwareVersionResult contains the result of GetFirmwareVersion command
type GetFirmwareVersionResult struct {
	FirmwareVersion *semver.RelaxedVersion
	Error           string
	Stdout          []byte
	Stderr          []byte
}

// FlashFirmware runs the plugin to flash the selected firmware
func (uploader *FwUploader) FlashFirmware(portAddress, fqbn, LogLevel string, verbose bool, firmwarePath *paths.Path, stdout, stderr io.Writer) (*FlashFirmwareResult, error) {
	args := []string{"firmware", "flash", firmwarePath.String()}
	if portAddress != "" {
		args = append(args, "-p", portAddress)
	}
	if fqbn != "" {
		args = append(args, "-b", fqbn)
	}
	if verbose {
		args = append(args, "-v")
	}
	if LogLevel != "" {
		args = append(args, "--log-level", LogLevel)
	}
	execStdout, execStderr, execErr := uploader.exec(stdout, stderr, args...)

	res := &FlashFirmwareResult{
		Stdout: execStdout.Bytes(),
		Stderr: execStderr.Bytes(),
	}
	fwErrorPrefix := "FLASH-FIRMWARE-ERROR: "
	for _, line := range strings.Split(execStdout.String(), "\n") {
		if strings.HasPrefix(line, fwErrorPrefix) {
			res.Error = strings.TrimPrefix(line, fwErrorPrefix)
		}
	}
	if res.Error != "" {
		if execErr != nil {
			execErr = fmt.Errorf("%s: %w", res.Error, execErr)
		} else {
			execErr = errors.New(res.Error)
		}
	}
	return res, execErr
}

// GetFirmwareVersionResult contains the result of GetFirmwareVersion command
type FlashFirmwareResult struct {
	Error  string
	Stdout []byte
	Stderr []byte
}

func (uploader *FwUploader) exec(stdout, stderr io.Writer, args ...string) (*bytes.Buffer, *bytes.Buffer, error) {
	stdoutBuffer := &bytes.Buffer{}
	stderrBuffer := &bytes.Buffer{}

	proc, err := executils.NewProcessFromPath(nil, uploader.pluginPath, args...)
	if err != nil {
		return stdoutBuffer, stderrBuffer, err
	}

	if stdout != nil {
		proc.RedirectStdoutTo(io.MultiWriter(stdoutBuffer, stdout))
	} else {
		proc.RedirectStdoutTo(stdoutBuffer)
	}

	if stderr != nil {
		proc.RedirectStderrTo(io.MultiWriter(stderrBuffer, stderr))
	} else {
		proc.RedirectStderrTo(stderr)
	}

	execErr := proc.RunWithinContext(context.Background())
	return stdoutBuffer, stderrBuffer, execErr
}

// FlashCertificates writes the given certificates bundle in PEM format.
func (uploader *FwUploader) FlashCertificates(portAddress, fqbn, LogLevel string, verbose bool, certsPath *paths.Path, stdout, stderr io.Writer) (*FlashCertificatesResult, error) {
	args := []string{"cert", "flash", certsPath.String()}
	if portAddress != "" {
		args = append(args, "-p", portAddress)
	}
	if fqbn != "" {
		args = append(args, "-b", fqbn)
	}
	if verbose {
		args = append(args, "-v")
	}
	if LogLevel != "" {
		args = append(args, "--log-level", LogLevel)
	}
	execStdout, execStderr, execErr := uploader.exec(stdout, stderr, args...)

	res := &FlashCertificatesResult{
		Stdout: execStdout.Bytes(),
		Stderr: execStderr.Bytes(),
	}
	fwErrorPrefix := "FLASH-CERTIFICATES-ERROR: "
	for _, line := range strings.Split(execStdout.String(), "\n") {
		if strings.HasPrefix(line, fwErrorPrefix) {
			res.Error = strings.TrimPrefix(line, fwErrorPrefix)
		}
	}
	if res.Error != "" {
		if execErr != nil {
			execErr = fmt.Errorf("%s: %w", res.Error, execErr)
		} else {
			execErr = errors.New(res.Error)
		}
	}
	return res, execErr
}

// FlashCertificatesResult contains the result of GetFirmwareVersion command
type FlashCertificatesResult struct {
	Error  string
	Stdout []byte
	Stderr []byte
}
