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

package certificates

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"io"
	"os"

	"github.com/arduino/arduino-fwuploader/certificates"
	"github.com/arduino/arduino-fwuploader/cli/arguments"
	"github.com/arduino/arduino-fwuploader/cli/common"
	"github.com/arduino/arduino-fwuploader/cli/feedback"
	"github.com/arduino/arduino-fwuploader/cli/globals"
	"github.com/arduino/arduino-fwuploader/flasher"
	"github.com/arduino/arduino-fwuploader/plugin"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	commonFlags arguments.Flags
)

// NewFlashCommand creates a new `flash` command
func NewFlashCommand() *cobra.Command {
	var certificateURLs []string
	var certificatePaths []string
	command := &cobra.Command{
		Use:   "flash",
		Short: "Flashes certificates to board.",
		Long:  "Flashes specified certificates to board at specified address.",
		Example: "" +
			"  " + os.Args[0] + " certificates flash --fqbn arduino:samd:mkrwifi1010 --address COM10 --url arduino.cc:443 --file /home/me/Digicert.cer\n" +
			"  " + os.Args[0] + " certificates flash -b arduino:renesas_uno:unor4wifi -a COM10 -u arduino.cc:443 -u google.com:443\n" +
			"  " + os.Args[0] + " certificates flash -b arduino:samd:mkrwifi1010 -a COM10 -f /home/me/VeriSign.cer -f /home/me/Digicert.cer\n",
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runFlash(certificateURLs, certificatePaths)
		},
	}
	commonFlags.AddToCommand(command)
	command.Flags().StringSliceVarP(&certificateURLs, "url", "u", []string{}, "List of urls to download root certificates, e.g.: arduino.cc:443")
	command.Flags().StringSliceVarP(&certificatePaths, "file", "f", []string{}, "List of paths to certificate file, e.g.: /home/me/Digicert.cer")
	return command
}

func runFlash(certificateURLs, certificatePaths []string) {
	// at the end cleanup the fwuploader temp dir
	defer globals.FwUploaderPath.RemoveAll()

	common.CheckFlags(commonFlags.Fqbn, commonFlags.Address)
	if len(certificateURLs) == 0 && len(certificatePaths) == 0 {
		feedback.Fatal("Error during certificates flashing: no certificates provided", feedback.ErrBadArgument)
	}

	packageIndex, firmwareIndex := common.InitIndexes()
	board := common.GetBoard(firmwareIndex, commonFlags.Fqbn)
	uploadToolDir := common.DownloadRequiredToolsForBoard(packageIndex, board)

	uploader, err := plugin.NewFWUploaderPlugin(uploadToolDir)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("Could not open uploader plugin: %s", err), feedback.ErrGeneric)
	}

	res, flashErr := flashCertificates(uploader, certificateURLs, certificatePaths)
	feedback.PrintResult(res)
	if flashErr != nil {
		feedback.Fatal(fmt.Sprintf("Error during certificates flashing: %s", flashErr), feedback.ErrGeneric)
	}
}

func flashCertificates(uploader *plugin.FwUploader, certificateURLs, certificatePaths []string) (*flasher.FlashResult, error) {
	tmp, err := paths.MkTempDir("", "")
	if err != nil {
		return nil, err
	}
	defer tmp.RemoveAll()
	certsBundle := tmp.Join("certs.pem")

	stdoutBuffer := &bytes.Buffer{}
	stderrBuffer := &bytes.Buffer{}
	var stdout io.Writer = stdoutBuffer
	var stderr io.Writer = stdoutBuffer
	if feedback.GetFormat() == feedback.Text {
		stdout = io.MultiWriter(os.Stdout, stdoutBuffer)
		stderr = io.MultiWriter(os.Stderr, stderrBuffer)
	}

	var allCerts []*x509.Certificate
	for _, certPath := range certificatePaths {
		logrus.Infof("Converting and flashing certificate %s", certPath)
		stdout.Write([]byte(fmt.Sprintf("Converting and flashing certificate %s\n", certPath)))

		certs, err := certificates.LoadCertificatesFromFile(paths.New(certPath))
		if err != nil {
			return nil, err
		}
		allCerts = append(allCerts, certs...)
	}

	for _, URL := range certificateURLs {
		logrus.Infof("Converting and flashing certificate from %s", URL)
		stdout.Write([]byte(fmt.Sprintf("Converting and flashing certificate from %s\n", URL)))
		rootCert, err := certificates.ScrapeRootCertificatesFromURL(URL)
		if err != nil {
			return nil, err
		}
		allCerts = append(allCerts, rootCert)
	}

	f, err := certsBundle.Create()
	if err != nil {
		return nil, err
	}
	defer f.Close() // Defer close if an error occurs while writing file
	for _, cert := range allCerts {
		_, err := f.Write(certificates.EncodeCertificateAsPEM(cert))
		if err != nil {
			return nil, err
		}
	}
	if err := f.Close(); err != nil {
		return nil, err
	}

	_, err = uploader.FlashCertificates(commonFlags.Address, commonFlags.Fqbn, globals.LogLevel, globals.Verbose, certsBundle, stdout, stderr)
	return &flasher.FlashResult{
		Flasher: &flasher.ExecOutput{
			Stdout: stdoutBuffer.String(),
			Stderr: stderrBuffer.String(),
		},
	}, err
}
