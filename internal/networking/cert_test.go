package networking

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net"
	"testing"
	"time"
)

func TestGenerateInMemoryCert(t *testing.T) {
	certPEM, keyPEM, err := GenerateInMemoryCert()
	if err != nil {
		t.Fatalf("GenerateInMemoryCert() returned unexpected error: %v", err)
	}

	// Both PEM blobs must be non-empty.
	if len(certPEM) == 0 {
		t.Error("certPEM is empty")
	}
	if len(keyPEM) == 0 {
		t.Error("keyPEM is empty")
	}

	// The certificate PEM must decode cleanly.
	certBlock, rest := pem.Decode(certPEM)
	if certBlock == nil {
		t.Fatal("failed to decode certPEM as PEM block")
	}
	if certBlock.Type != "CERTIFICATE" {
		t.Errorf("certPEM block type = %q; want %q", certBlock.Type, "CERTIFICATE")
	}
	_ = rest

	// The key PEM must decode cleanly.
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		t.Fatal("failed to decode keyPEM as PEM block")
	}
	if keyBlock.Type != "RSA PRIVATE KEY" {
		t.Errorf("keyPEM block type = %q; want %q", keyBlock.Type, "RSA PRIVATE KEY")
	}

	// The pair must be parseable by tls.X509KeyPair.
	_, err = tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Errorf("tls.X509KeyPair() failed: %v", err)
	}

	// Parse the raw certificate and inspect its properties.
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		t.Fatalf("x509.ParseCertificate() failed: %v", err)
	}

	// Must be self-signed (issuer == subject).
	if cert.Issuer.String() != cert.Subject.String() {
		t.Errorf("cert is not self-signed: issuer=%q subject=%q", cert.Issuer, cert.Subject)
	}

	// Organisation must match.
	if len(cert.Subject.Organization) == 0 || cert.Subject.Organization[0] != "GhostGate Framework" {
		t.Errorf("unexpected organisation: %v", cert.Subject.Organization)
	}

	// Validity window must be approximately 1 year.
	duration := cert.NotAfter.Sub(cert.NotBefore)
	expectedDuration := 365 * 24 * time.Hour
	// Allow ±1 minute of tolerance.
	if duration < expectedDuration-time.Minute || duration > expectedDuration+time.Minute {
		t.Errorf("cert validity window = %v; want ~%v", duration, expectedDuration)
	}

	// IP SANs must include 127.0.0.1.
	loopback := net.ParseIP("127.0.0.1")
	found := false
	for _, ip := range cert.IPAddresses {
		if ip.Equal(loopback) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("cert IP SANs %v do not include 127.0.0.1", cert.IPAddresses)
	}
}
