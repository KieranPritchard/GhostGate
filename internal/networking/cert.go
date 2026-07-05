package networking

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"
)

// GenerateInMemoryCert creates a temporary self-signed TLS certificate and private key,
// stored entirely in memory as PEM-encoded bytes.
func GenerateInMemoryCert() ([]byte, []byte, error) {
	// Generate a 2048-bit RSA private key
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Certificate validity window
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	// Generate a random serial number
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, err
	}

	// Collect IPs to add as Subject Alternative Names.
	// Always include loopback; also add the machine's real outbound LAN IP
	// so that clients connecting via the local network can complete the TLS handshake.
	ipSANs := []net.IP{net.ParseIP("127.0.0.1")}
	if lanIP := GetOutboundIP(); len(lanIP) > 0 {
		ipSANs = append(ipSANs, lanIP)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"GhostGate Framework"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
		// KeyUsageKeyEncipherment is correct for RSA TLS servers.
		// KeyUsageDataEncipherment is for payload encryption, not TLS handshakes.
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           ipSANs,
	}

	// Sign the certificate with its own private key (self-signed)
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &private.PublicKey, private)
	if err != nil {
		return nil, nil, err
	}

	// PEM-encode the certificate
	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	// PEM-encode the private key
	privBytes := x509.MarshalPKCS1PrivateKey(private)
	keyPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})

	return certPem, keyPem, nil
}
