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

// Creates a temporary self-signed tls certificate stored entirely in memory
func GenerateInMemoryCert() ([]byte, []byte, error) {
	// Generates a private rsa key
	private, err := rsa.GenerateKey(rand.Reader, 2048)

	// CHecks if an error occured with the key generation
	if err != nil {
		// Returns the nil errors
		return nil, nil, err
	}

	// Stores the timing for the certificate
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour) // Makes it valid for the year

	// Creates the serial number limit
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	// Generates the serial number
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)

	if err != nil {
		return nil, nil, err
	}

	// Stores the template for the certificate
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"GhostGate Framework"},
		},
		NotBefore: notBefore,
		NotAfter: notAfter,
		KeyUsage: x509.KeyUsageDataEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("0.0.0.0")},
	}

	// Creates the bytes to create the cert
	derBytes , err := x509.CreateCertificate(rand.Reader, &template, &template, &private.PublicKey, private)

	// Returns the errors to the main
	if err != nil {
		return nil, nil, err
	}

	// PEM encode the certificate
	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	// PEM encode the private key
	privBytes := x509.MarshalPKCS1PrivateKey(private)
	keyPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})

	// Returns the keys
	return certPem, keyPem, nil
}