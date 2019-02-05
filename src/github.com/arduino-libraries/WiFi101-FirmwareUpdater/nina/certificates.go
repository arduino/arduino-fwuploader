/*
  certificates.go - A firmware uploader for the WiFi101 module.
  Copyright (c) 2015 Arduino LLC.  All right reserved.

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
package nina

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"encoding/pem"
)

type CertEntry []byte

func ConvertCertificates(directory string, addresses []string) ([]byte, error) {
	var entryBytes []byte

	if directory != "" {
		cerFiles, err := filepath.Glob(path.Join(directory, "*.cer"))
		if err != nil {
			return nil, err
		}

		for _, cerFile := range cerFiles {
			cerEntry, err := EntryForFile(cerFile)

			if err != nil {
				log.Printf("Converting '%v' failed, skipping: %v\n", cerFile, err)
			} else {
				entryBytes = append(entryBytes, cerEntry...)
			}
		}
	}

	for _, address := range addresses {
		cerEntry, err := EntryForAddress(address)

		if err != nil {
			log.Printf("Converting address '%v' failed, skipping: %v\n", address, err)
		} else {
			entryBytes = append(entryBytes, cerEntry...)
		}
	}

	return entryBytes, nil
}

func EntryForFile(file string) (b CertEntry, err error) {
	cerData, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	certs, err := x509.ParseCertificates(cerData)
	if err != nil {
		return nil, err
	}

	if len(certs) < 1 {
		return nil, errors.New("No certificates in file")
	}

	cert := certs[0]

	return entryForCert(cert)
}

func EntryForAddress(address string) (b CertEntry, err error) {
	config := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", address, config)
	if err != nil {
		return nil, err
	}

	if err := conn.Handshake(); err != nil {
		return nil, err
	}

	peerCerts := conn.ConnectionState().PeerCertificates

	if len(peerCerts) == 0 {
		return nil, errors.New("No peer certificates")
	}

	rootCert := peerCerts[len(peerCerts)-1]

	conn.Close()

	return entryForCert(rootCert)
}

/* Write Root Certificate to flash. Must convert certificates to PEM and append them
-	SHA1_DIGEST_SIZE	--> NameSHA1 of the Root certificate.
-	uint16				--> N_SIZE (Byte count for the RSA modulus N).
-	uint16				--> E_SIZE (Byte count for the RSA public exponent E).
-	START_DATE			--> Start date of the root certificate(20 bytes).
-	EXPIRATION_DATE		--> Expiration date of the root certificate(20 bytes).
-	N_SIZE				--> RSA Modulus N.
-	E_SIZE				--> RSA Public exponent.
*/

func entryForCert(cert *x509.Certificate) (b CertEntry, err error) {
	return certToPEM(cert), nil
}

// CertToPEM is a utility function returns a PEM encoded x509 Certificate
func certToPEM(cert *x509.Certificate) []byte {
	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})

	return pemCert
}
