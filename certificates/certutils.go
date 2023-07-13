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
	"fmt"

	"github.com/sirupsen/logrus"
)

// ScrapeRootCertificatesFromURL downloads from a webserver the root certificate
// required to connect to that server from the TLS handshake response.
func ScrapeRootCertificatesFromURL(URL string) (*x509.Certificate, error) {
	conn, err := tls.Dial("tcp", URL, &tls.Config{
		InsecureSkipVerify: true,
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
