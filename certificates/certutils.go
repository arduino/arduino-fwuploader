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

package certificates

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

// ScrapeRootCertificatesFromURL downloads from a webserver the root certificate
// required to connect to that server from the TLS handshake response.
func ScrapeRootCertificatesFromURL(URL string) (*x509.Certificate, error) {
	conn, err := tls.Dial("tcp", URL, &tls.Config{
		InsecureSkipVerify: false,
	})
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	defer conn.Close()

	if err := conn.Handshake(); err != nil {
		logrus.Error(err)
		return nil, err
	}

	peerCertificates := conn.ConnectionState().PeerCertificates
	if len(peerCertificates) == 0 {
		err = fmt.Errorf("no peer certificates found at %s", URL)
		logrus.Error(err)
		return nil, err
	}

	rootCertificate := peerCertificates[len(peerCertificates)-1]
	return rootCertificate, nil
}

// LoadCertificatesFromFile read certificates from the given file. PEM and CER formats
// are supported.
func LoadCertificatesFromFile(certificateFile *paths.Path) ([]*x509.Certificate, error) {
	data, err := certificateFile.ReadFile()
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	var res []*x509.Certificate
	switch certificateFile.Ext() {
	case ".cer":
		cert, err := x509.ParseCertificate(data)
		if err != nil {
			logrus.Error(err)
		}
		res = append(res, cert)
		return res, err

	case ".pem":
		for {
			block, rest := pem.Decode(data)
			data = rest
			if block == nil && len(rest) > 0 {
				return nil, fmt.Errorf("invalid .pem data")
			}
			if block == nil {
				return res, nil
			}
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate: %w", err)
			}
			res = append(res, cert)
			if len(rest) == 0 {
				return res, nil
			}
		}
	default:
		return nil, fmt.Errorf("cert format %s not supported, please use .pem or .cer", certificateFile.Ext())
	}
}

// EncodeCertificateAsPEM returns the PEM encoding of the given certificate
func EncodeCertificateAsPEM(cert *x509.Certificate) []byte {
	pemBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}
	return pem.EncodeToMemory(pemBlock)
}
