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

package indexes

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/arduino/arduino-cli/arduino/cores/packageindex"
	"github.com/arduino/arduino-cli/arduino/security"
	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/arduino/go-paths-helper"
	rice "github.com/cmaglie/go.rice"
	"go.bug.st/downloader/v2"
)

// DownloadIndex will download the index in the os temp directory
func DownloadIndex(indexURL string) error {
	fwUploaderPath := paths.TempDir().Join("fwuploader")

	URL, err := utils.URLParse(indexURL)
	if err != nil {
		return fmt.Errorf("unable to parse URL %s: %s", indexURL, err)
	}

	// Download index
	var tmpGZIndex *paths.Path
	if tmpGZFile, err := paths.MkTempFile(nil, ""); err != nil {
		return fmt.Errorf("creating temp file for compressed index download: %s", err)
	} else {
		tmpGZIndex = paths.New(tmpGZFile.Name())
		defer os.Remove(tmpGZFile.Name())
		defer tmpGZIndex.Remove()
	}
	d, err := downloader.Download(tmpGZIndex.String(), URL.String())
	if err != nil {
		return fmt.Errorf("downloading index %s: %s", indexURL, err)
	}
	indexPath := fwUploaderPath.Join(path.Base(strings.ReplaceAll(URL.Path, ".gz", "")))
	if err := Download(d); err != nil || d.Error() != nil {
		return fmt.Errorf("downloading index %s: %s %s", URL, d.Error(), err)
	}

	// Extract the real index
	var tmpIndex *paths.Path
	if tmpFile, err := paths.MkTempFile(nil, ""); err != nil {
		return fmt.Errorf("creating temp file for index extraction: %s", err)
	} else {
		tmpIndex = paths.New(tmpFile.Name())
		defer os.Remove(tmpFile.Name())
		defer tmpIndex.Remove()
	}
	if err := paths.GUnzip(tmpGZIndex, tmpIndex); err != nil {
		return fmt.Errorf("unzipping %s", URL)
	}

	// Download Signature
	sigURL, err := url.Parse(URL.String())
	if err != nil {
		return fmt.Errorf("unable to parse URL %s: %s", sigURL, err)
	}
	sigURL.Path = strings.ReplaceAll(sigURL.Path, "gz", "sig")
	var tmpSig *paths.Path
	if t, err := paths.MkTempFile(nil, ""); err != nil {
		return fmt.Errorf("creating temp file for index signature download: %s", err)
	} else {
		tmpSig = paths.New(t.Name())
		defer tmpSig.Remove()
		defer os.Remove(t.Name())
	}
	d, err = downloader.Download(tmpSig.String(), sigURL.String())
	if err != nil {
		return fmt.Errorf("downloading index signature %s: %s", sigURL, err)
	}
	indexSigPath := fwUploaderPath.Join(path.Base(sigURL.Path))
	if err := Download(d); err != nil || d.Error() != nil {
		return fmt.Errorf("downloading index signature %s: %s %s", URL, d.Error(), err)
	}
	if err := verifySignature(tmpIndex, tmpSig, URL, sigURL); err != nil {
		return fmt.Errorf("signature verification failed: %s", err)
	}
	if err := fwUploaderPath.MkdirAll(); err != nil { //does not overwrite if dir already present
		return fmt.Errorf("can't create data directory %s: %s", fwUploaderPath, err)
	}
	if err := tmpIndex.CopyTo(indexPath); err != nil { //does overwrite
		return fmt.Errorf("saving downloaded index %s: %s", URL, err)
	}
	if tmpSig != nil {
		if err := tmpSig.CopyTo(indexSigPath); err != nil { //does overwrite
			return fmt.Errorf("saving downloaded index signature: %s", err)
		}
	}
	return nil
}

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

func verifySignature(targetPath, signaturePath *paths.Path, URL, sigURL *url.URL) error {
	var valid bool
	var err error
	if path.Base(URL.Path) == "package_index.json.gz" {
		valid, _, err = security.VerifyArduinoDetachedSignature(targetPath, signaturePath)
		// the signature verification is already done above
		if _, err = packageindex.LoadIndexNoSign(targetPath); err != nil {
			return fmt.Errorf("invalid package index: %s", err)
		}
	} else if path.Base(URL.Path) == "module_firmware_index.json.gz" {
		keysBox, err := rice.FindBox("gpg_keys")
		if err != nil {
			return fmt.Errorf("could not find bundled signature keys")
		}
		key, err := keysBox.Open("module_firmware_index_public.gpg.key")
		if err != nil {
			return fmt.Errorf("could not find bundled signature keys")
		}
		valid, _, err = security.VerifySignature(targetPath, signaturePath, key)
		//TODO missing something like packageindex.LoadIndexNoSign(targetPath) for firmware_module_index.json
	} else {
		return fmt.Errorf("index %s not supported", URL.Path)
	}
	if err != nil {
		return fmt.Errorf("signature verification error: %s for index %s", err, URL)
	}
	if !valid {
		return fmt.Errorf("index \"%s\" has an invalid signature", sigURL)
	}
	return nil
}
