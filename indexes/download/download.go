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

package download

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"path"
	"strings"

	"github.com/arduino/FirmwareUploader/cli/globals"
	"github.com/arduino/FirmwareUploader/indexes/firmwareindex"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"go.bug.st/downloader/v2"
)

func DownloadTool(toolRelease *cores.ToolRelease) (*paths.Path, error) {
	resource := toolRelease.GetCompatibleFlavour()
	installDir := globals.FwUploaderPath.Join(
		"tools",
		toolRelease.Tool.Name,
		toolRelease.Version.String())
	installDir.Parent().MkdirAll()
	if err := resource.Install(paths.TempDir(), paths.TempDir(), installDir); err != nil {
		logrus.Error(err)
		return nil, err
	}
	return installDir, nil
}

func DownloadFirmware(firmware *firmwareindex.IndexFirmware) (*paths.Path, error) {
	firmwarePath := globals.FwUploaderPath.Join(
		"firmwares",
		firmware.Module,
		firmware.Version,
		path.Base(firmware.URL))
	firmwarePath.Parent().MkdirAll()
	if err := firmwarePath.WriteFile(nil); err != nil {
		logrus.Error(err)
		return nil, err
	}
	d, err := downloader.Download(firmwarePath.String(), firmware.URL)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	if err := Download(d); err != nil {
		logrus.Error(err)
		return nil, err
	}
	if err := VerifyFileChecksum(firmware.Checksum, firmwarePath); err != nil {
		logrus.Error(err)
		return nil, err
	}
	size, _ := firmware.Size.Int64()
	if err := VerifyFileSize(size, firmwarePath); err != nil {
		logrus.Error(err)
		return nil, err
	}
	return firmwarePath, nil
}

func DownloadLoaderSketch(loader *firmwareindex.IndexLoaderSketch) (*paths.Path, error) {
	loaderPath := globals.FwUploaderPath.Join(
		"loader",
		path.Base(loader.URL))
	loaderPath.Parent().MkdirAll()
	if err := loaderPath.WriteFile(nil); err != nil {
		logrus.Error(err)
		return nil, err
	}
	d, err := downloader.Download(loaderPath.String(), loader.URL)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	if err := Download(d); err != nil {
		logrus.Error(err)
		return nil, err
	}
	if err := VerifyFileChecksum(loader.Checksum, loaderPath); err != nil {
		logrus.Error(err)
		return nil, err
	}
	size, _ := loader.Size.Int64()
	if err := VerifyFileSize(size, loaderPath); err != nil {
		logrus.Error(err)
		return nil, err
	}
	return loaderPath, nil
}

// Download will take a downloader.Downloader as parameter. It will Download the file specified in the downloader
func Download(d *downloader.Downloader) error {
	if d == nil {
		// This signal means that the file is already downloaded
		return nil
	}
	if err := d.Run(); err != nil {
		return fmt.Errorf("failed to download file from %s : %s", d.URL, err)
	}
	// The URL is not reachable for some reason
	if d.Resp.StatusCode >= 400 && d.Resp.StatusCode <= 599 {
		return fmt.Errorf(d.Resp.Status)
	}
	return nil
}

// taken and adapted from https://github.com/arduino/arduino-cli/blob/59b6277a4d6731a1c1579d43aef6df2a46a771d5/arduino/resources/checksums.go
func VerifyFileChecksum(checksum string, filePath *paths.Path) error {
	if checksum == "" {
		return fmt.Errorf("missing checksum for: %s", filePath)
	}
	split := strings.SplitN(checksum, ":", 2)
	if len(split) != 2 {
		return fmt.Errorf("invalid checksum format: %s", checksum)
	}
	digest, err := hex.DecodeString(split[1])
	if err != nil {
		return fmt.Errorf("invalid hash '%s': %s", split[1], err)
	}

	// names based on: https://docs.oracle.com/javase/8/docs/technotes/guides/security/StandardNames.html#MessageDigest
	var algo hash.Hash
	switch split[0] {
	case "SHA-256":
		algo = crypto.SHA256.New()
	case "SHA-1":
		algo = crypto.SHA1.New()
	case "MD5":
		algo = crypto.MD5.New()
	default:
		return fmt.Errorf("unsupported hash algorithm: %s", split[0])
	}

	file, err := filePath.Open()
	if err != nil {
		return fmt.Errorf("opening file: %s", err)
	}
	defer file.Close()
	if _, err := io.Copy(algo, file); err != nil {
		return fmt.Errorf("computing hash: %s", err)
	}
	if bytes.Compare(algo.Sum(nil), digest) != 0 {
		return fmt.Errorf("archive hash differs from hash in index")
	}

	return nil
}

// taken and adapted from https://github.com/arduino/arduino-cli/blob/59b6277a4d6731a1c1579d43aef6df2a46a771d5/arduino/resources/checksums.go
func VerifyFileSize(size int64, filePath *paths.Path) error {
	info, err := filePath.Stat()
	if err != nil {
		return fmt.Errorf("getting archive info: %s", err)
	}
	if info.Size() != size {
		return fmt.Errorf("fetched archive size differs from size specified in index")
	}

	return nil
}
