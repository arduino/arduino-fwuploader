/*
	arduino-fwuploader
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

package flasher

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/binary"
	"time"

	"github.com/sirupsen/logrus"
)

func calculateNameSha1(cert *x509.Certificate) (b []byte, err error) {
	nameSha1 := sha1.New()

	var subjectDistinguishedNameSequence pkix.RDNSequence

	if _, err = asn1.Unmarshal(cert.RawSubject, &subjectDistinguishedNameSequence); err != nil {
		logrus.Error(err)
		return nil, err
	}

	for _, dn := range subjectDistinguishedNameSequence {
		nameSha1.Write([]byte(dn[0].Value.(string)))
	}

	b = nameSha1.Sum(nil)

	return
}

func convertTime(time time.Time) ([]byte, error) {
	asn1Bytes, err := asn1.Marshal(time)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	b := bytes.Repeat([]byte{0x00}, 20) // value must be zero bytes
	copy(b, asn1Bytes[2:])              // copy but drop the first two bytes

	return b, err
}

func modulusN(publicKey rsa.PublicKey) []byte {
	return publicKey.N.Bytes()
}

func publicExponent(publicKey rsa.PublicKey) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(publicKey.E))
	// strip leading zeros
	for b[0] == 0 {
		b = b[1:]
	}
	return b
}

func uint16ToBytes(i int) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(i))
	return b
}
