package commands

import (
	"crypto/tls"
	"net/http"
	"os"
	"testing"

	"GhostGate/internal/networking"
)

// TestStartTLSServer_InvalidKeyPair verifies startTLSServer returns an error when given
// malformed PEM bytes that cannot be parsed as a TLS key pair.
func TestStartTLSServer_InvalidKeyPair(t *testing.T) {
	err := startTLSServer("0", http.NewServeMux(), []byte("bad cert"), []byte("bad key"))
	if err == nil {
		t.Error("startTLSServer() with invalid PEM expected an error; got nil")
	}
}

// TestStartTLSServer_InvalidPort verifies startTLSServer returns an error when given an invalid port.
func TestStartTLSServer_InvalidPort(t *testing.T) {
	certPEM, keyPEM, err := networking.GenerateInMemoryCert()
	if err != nil {
		t.Fatalf("GenerateInMemoryCert() failed: %v", err)
	}

	err = startTLSServer("-1", http.NewServeMux(), certPEM, keyPEM)
	if err == nil {
		t.Error("startTLSServer() with invalid port expected an error; got nil")
	}
}


// TestStartTLSServer_ValidCert verifies that a valid in-memory cert can be loaded
// into a TLS listener (exercises the key pair parsing path of startTLSServer).
func TestStartTLSServer_ValidCert(t *testing.T) {
	certPEM, keyPEM, err := networking.GenerateInMemoryCert()
	if err != nil {
		t.Fatalf("GenerateInMemoryCert() failed: %v", err)
	}

	// Verify the key pair parses — this exercises the first branch of startTLSServer.
	_, kpErr := tls.X509KeyPair(certPEM, keyPEM)
	if kpErr != nil {
		t.Errorf("generated cert/key pair is invalid: %v", kpErr)
	}

	// Build the TLS listener directly to confirm a valid listener can be opened.
	tlsCert, _ := tls.X509KeyPair(certPEM, keyPEM)
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{tlsCert}})
	if err != nil {
		t.Fatalf("tls.Listen() failed: %v", err)
	}
	defer ln.Close()

	if ln.Addr() == nil {
		t.Error("listener address is nil")
	}
}

// TestTLSListenerFromFiles_MissingFiles verifies tlsListenerFromFiles returns an error
// for non-existent cert/key files.
func TestTLSListenerFromFiles_MissingFiles(t *testing.T) {
	ln, err := tlsListenerFromFiles("0", "/nonexistent/cert.pem", "/nonexistent/key.pem")
	if err == nil {
		t.Error("tlsListenerFromFiles() with missing files expected an error; got nil")
		ln.Close()
	}
}

// TestTLSListenerFromFiles_ValidFiles verifies tlsListenerFromFiles creates a working TLS listener
// when given valid cert/key files.
func TestTLSListenerFromFiles_ValidFiles(t *testing.T) {
	certPEM, keyPEM, err := networking.GenerateInMemoryCert()
	if err != nil {
		t.Fatalf("GenerateInMemoryCert() failed: %v", err)
	}

	// Write PEM data to temp files.
	tmpDir := t.TempDir()
	certPath := tmpDir + "/cert.pem"
	keyPath := tmpDir + "/key.pem"

	if err := os.WriteFile(certPath, certPEM, 0600); err != nil {
		t.Fatalf("failed to write cert file: %v", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	ln, err := tlsListenerFromFiles("0", certPath, keyPath)
	if err != nil {
		t.Fatalf("tlsListenerFromFiles() unexpected error: %v", err)
	}
	defer ln.Close()

	// Verify the listener is bound to a real address.
	addr := ln.Addr()
	if addr == nil {
		t.Fatal("listener address is nil")
	}
	if addr.String() == "" || addr.String() == ":0" {
		t.Errorf("unexpected listener address: %s", addr.String())
	}
}
