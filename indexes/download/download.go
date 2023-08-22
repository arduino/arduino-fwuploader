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

package download

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"math/rand"
	"net/url"
	"path"
	"strings"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packageindex"
	"github.com/arduino/arduino-cli/arduino/security"
	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/arduino/arduino-fwuploader/cli/globals"
	"github.com/arduino/arduino-fwuploader/indexes/firmwareindex"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"go.bug.st/downloader/v2"
)

// DownloadTool downloads and returns the path on the local filesystem of a tool
func DownloadTool(toolRelease *cores.ToolRelease) (*paths.Path, error) {
	resource := toolRelease.GetCompatibleFlavour()
	if resource == nil {
		err := fmt.Errorf("tool %s not available for this OS", toolRelease.String())
		logrus.Error(err)
		return nil, err
	}
	installDir := globals.FwUploaderPath.Join(
		"tools",
		toolRelease.Tool.Name,
		toolRelease.Version.String())
	if err := installDir.MkdirAll(); err != nil {
		logrus.Error(err)
		return nil, err
	}
	downloadsDir := globals.FwUploaderPath.Join("downloads")
	archivePath, err := resource.ArchivePath(downloadsDir)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	defer downloadsDir.RemoveAll()
	d, err := downloader.Download(archivePath.String(), resource.URL, downloader.NoResume)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	if err := Download(d); err != nil {
		logrus.Error(err)
		return nil, err
	}
	if err := resource.Install(downloadsDir, paths.TempDir(), installDir); err != nil {
		logrus.Error(err)
		return nil, err
	}
	return installDir, nil
}

// DownloadFirmware downloads and returns the path on the local filesystem of a firmware
func DownloadFirmware(firmware *firmwareindex.IndexFirmware) (*paths.Path, error) {
	firmwarePath := globals.FwUploaderPath.Join(
		"firmwares",
		firmware.Module,
		firmware.Version.String(),
		path.Base(firmware.URL))
	if err := firmwarePath.Parent().MkdirAll(); err != nil {
		logrus.Error(err)
		return nil, err
	}
	d, err := downloader.Download(firmwarePath.String(), firmware.URL, downloader.NoResume)
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

// DownloadSketch downloads and returns the path on the local filesystem of a sketch
func DownloadSketch(loader *firmwareindex.IndexSketch) (*paths.Path, error) {
	sketchPath := globals.FwUploaderPath.Join(
		"sketch",
		path.Base(loader.URL))
	if err := sketchPath.Parent().MkdirAll(); err != nil {
		logrus.Error(err)
		return nil, err
	}
	d, err := downloader.Download(sketchPath.String(), loader.URL, downloader.NoResume)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	if err := Download(d); err != nil {
		logrus.Error(err)
		return nil, err
	}
	if err := VerifyFileChecksum(loader.Checksum, sketchPath); err != nil {
		logrus.Error(err)
		return nil, err
	}
	size, _ := loader.Size.Int64()
	if err := VerifyFileSize(size, sketchPath); err != nil {
		logrus.Error(err)
		return nil, err
	}
	return sketchPath, nil
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

// VerifyChecksum is taken and adapted from https://github.com/arduino/arduino-cli/blob/59b6277a4d6731a1c1579d43aef6df2a46a771d5/arduino/resources/checksums.go
func VerifyChecksum(checksum string, file io.Reader) error {
	if checksum == "" {
		return errors.New("missing checksum")
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

	if _, err := io.Copy(algo, file); err != nil {
		return fmt.Errorf("computing hash: %s", err)
	}
	if !bytes.Equal(algo.Sum(nil), digest) {
		return fmt.Errorf("archive hash differs from hash in index")
	}

	return nil
}

// VerifyFileChecksum checks if the passed checksum matches the passed file checksum
func VerifyFileChecksum(checksum string, filePath *paths.Path) error {
	file, err := filePath.Open()
	if err != nil {
		return fmt.Errorf("opening file: %s", err)
	}
	defer file.Close()
	if err = VerifyChecksum(checksum, file); err != nil {
		return fmt.Errorf("verifying checksum of file %s: %w", filePath, err)
	}
	return nil
}

// VerifyFileSize is taken and adapted from https://github.com/arduino/arduino-cli/blob/59b6277a4d6731a1c1579d43aef6df2a46a771d5/arduino/resources/checksums.go
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

// DownloadIndex will download the index in the os temp directory
func DownloadIndex(indexURL string) (*paths.Path, error) {
	indexArchiveURL, err := utils.URLParse(indexURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse URL %s: %s", indexURL, err)
	}
	// Create a directory to store temporary files before verifying them and copying to
	// their final directory
	tempDir, err := paths.MkTempDir("", fmt.Sprintf("%d", rand.Int()))
	if err != nil {
		return nil, err
	}
	defer tempDir.RemoveAll()

	// Download index
	tmpGZIndex := tempDir.Join("index.gz")
	d, err := downloader.Download(tmpGZIndex.String(), indexArchiveURL.String(), downloader.NoResume)
	if err != nil {
		return nil, fmt.Errorf("downloading index %s: %s", indexURL, err)
	}
	if err := Download(d); err != nil || d.Error() != nil {
		return nil, fmt.Errorf("downloading index %s: %s %s", indexArchiveURL, d.Error(), err)
	}

	// Extract the index to temporary file
	tmpIndex := tempDir.Join("index.json")
	if err := paths.GUnzip(tmpGZIndex, tmpIndex); err != nil {
		return nil, fmt.Errorf("unzipping %s", indexArchiveURL)
	}

	// Download signature file
	signatureURL, err := url.Parse(strings.ReplaceAll(indexURL, ".gz", ".sig"))
	if err != nil {
		return nil, fmt.Errorf("unable to parse URL %s: %s", signatureURL, err)
	}
	tmpSignature := tempDir.Join("index.json.sig")

	d, err = downloader.Download(tmpSignature.String(), signatureURL.String(), downloader.NoResume)
	if err != nil {
		return nil, fmt.Errorf("downloading index signature %s: %s", signatureURL, err)
	}
	if err := Download(d); err != nil || d.Error() != nil {
		return nil, fmt.Errorf("downloading index signature %s: %s %s", indexArchiveURL, d.Error(), err)
	}
	if err := verifyIndex(tmpIndex, indexArchiveURL); err != nil {
		return nil, fmt.Errorf("signature verification failed: %s", err)
	}
	if err := globals.FwUploaderPath.MkdirAll(); err != nil { //does not overwrite if dir already present
		return nil, fmt.Errorf("can't create data directory %s: %s", globals.FwUploaderPath, err)
	}
	indexSigPath := globals.FwUploaderPath.Join(path.Base(signatureURL.Path))
	indexPath := globals.FwUploaderPath.Join(path.Base(strings.ReplaceAll(indexArchiveURL.Path, ".gz", "")))
	if err := tmpIndex.CopyTo(indexPath); err != nil { //does overwrite
		return nil, fmt.Errorf("saving downloaded index %s: %s", indexArchiveURL, err)
	}
	if err := tmpSignature.CopyTo(indexSigPath); err != nil { //does overwrite
		return nil, fmt.Errorf("saving downloaded index signature: %s", err)
	}
	return indexPath, nil
}

// verifyIndex verifies if the signature is correct and the index is parsable.
func verifyIndex(indexPath *paths.Path, URL *url.URL) error {
	var valid bool
	var err error
	index := path.Base(URL.Path)
	signaturePath := paths.New(fmt.Sprintf("%s.sig", indexPath))
	if index == "package_index.json.gz" {
		valid, err = verifyPackageIndex(indexPath, signaturePath)
	} else if index == "plugin_firmware_index.json.gz" {
		valid, err = verifyPluginFirmwareIndex(indexPath, signaturePath)
	} else {
		return fmt.Errorf("index %s not supported", URL.Path)
	}

	if err != nil {
		return fmt.Errorf("signature verification error: %s for index %s", err, URL)
	}
	if !valid {
		return fmt.Errorf(`index "%s" has an invalid signature`, URL)
	}
	return nil
}

// verifyPackageIndex verify if the signature is valid for the provided package index
func verifyPackageIndex(indexPath, signaturePath *paths.Path) (bool, error) {
	valid, _, err := security.VerifyArduinoDetachedSignature(indexPath, signaturePath)
	if err != nil {
		return valid, err
	}
	// the signature verification is already done above
	if _, err := packageindex.LoadIndex(indexPath); err != nil {
		return valid, fmt.Errorf("invalid package index: %s", err)
	}
	return valid, nil
}

// verifyPluginFirmwareIndex verify if the signature is valid for the provided plugin firmware index
func verifyPluginFirmwareIndex(indexPath, signaturePath *paths.Path) (bool, error) {
	arduinoKeyringFile, err := globals.Keys.Open("keys/plugin_firmware_index_public.gpg.key")
	if err != nil {
		return false, fmt.Errorf("could not find bundled signature keys: %s", err)
	}
	defer arduinoKeyringFile.Close()
	valid, _, err := security.VerifySignature(indexPath, signaturePath, arduinoKeyringFile)
	if err != nil {
		return valid, nil
	}
	// the signature verification is already done above
	_, err = firmwareindex.LoadIndexNoSign(indexPath)
	if err != nil {
		return valid, nil
	}

	return valid, nil
}
