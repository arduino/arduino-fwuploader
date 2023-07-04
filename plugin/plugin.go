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
	"context"
	"fmt"

	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/go-paths-helper"
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
		paths.FilterOutDirectories(),
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
