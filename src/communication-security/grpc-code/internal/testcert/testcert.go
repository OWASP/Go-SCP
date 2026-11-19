// Package testcert creates throwaway TLS certificates for the gRPC example
// tests.
package testcert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Serial numbers only need to be unique per issuing CA.
const (
	caSerial     = 1
	serverSerial = 2
)

// Files holds a CA and a server certificate for 127.0.0.1 signed by it.
type Files struct {
	CAFile   string          // CA certificate, PEM
	CertFile string          // server certificate, PEM
	KeyFile  string          // server private key, PEM
	Server   tls.Certificate // server certificate and key, loaded
}

// New creates a CA and a server certificate for 127.0.0.1 in a temporary
// directory that is removed when the test ends. Each call creates a new,
// unrelated CA.
func New(t *testing.T) Files {
	t.Helper()
	dir := t.TempDir()
	notAfter := time.Now().Add(time.Hour)

	caKey := newKey(t)
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(caSerial),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotAfter:              notAfter,
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	caDER := createCertificate(t, caTemplate, caTemplate, caKey, caKey)
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatalf("testcert.New: %s", err)
	}

	serverKey := newKey(t)
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(serverSerial),
		Subject:      pkix.Name{CommonName: "127.0.0.1"},
		NotAfter:     notAfter,
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	serverDER := createCertificate(t, serverTemplate, caCert, serverKey, caKey)

	keyDER, err := x509.MarshalECPrivateKey(serverKey)
	if err != nil {
		t.Fatalf("testcert.New: %s", err)
	}

	files := Files{
		CAFile:   filepath.Join(dir, "ca.pem"),
		CertFile: filepath.Join(dir, "server.pem"),
		KeyFile:  filepath.Join(dir, "server.key"),
		Server:   tls.Certificate{Certificate: [][]byte{serverDER}, PrivateKey: serverKey},
	}
	writePEM(t, files.CAFile, "CERTIFICATE", caDER)
	writePEM(t, files.CertFile, "CERTIFICATE", serverDER)
	writePEM(t, files.KeyFile, "EC PRIVATE KEY", keyDER)
	return files
}

func newKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("testcert.New: %s", err)
	}
	return key
}

func createCertificate(
	t *testing.T,
	template, parent *x509.Certificate,
	key, parentKey *ecdsa.PrivateKey,
) []byte {
	t.Helper()
	der, err := x509.CreateCertificate(rand.Reader, template, parent, &key.PublicKey, parentKey)
	if err != nil {
		t.Fatalf("testcert.New: %s", err)
	}
	return der
}

func writePEM(t *testing.T, path, blockType string, der []byte) {
	t.Helper()
	data := pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der})
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("testcert.New: %s", err)
	}
}
