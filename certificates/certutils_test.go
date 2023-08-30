package certificates_test

import (
	"testing"

	"github.com/arduino/arduino-fwuploader/certificates"
	"github.com/stretchr/testify/require"
)

func TestScrapeRootCertificatesFromURL(t *testing.T) {
	cert, err := certificates.ScrapeRootCertificatesFromURL("www.arduino.cc:443")
	require.NoError(t, err)
	require.Equal(t, cert.Issuer, cert.Subject)
}
