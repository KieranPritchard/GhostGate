package commands

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
)

// startTLSServer is a shared helper that creates a TLS listener from in-memory PEM bytes
// and serves the given handler. It returns any error from http.Server.Serve.
func startTLSServer(port string, handler http.Handler, certPem, keyPem []byte) error {
	tlsCert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return fmt.Errorf("failed to parse TLS key pair: %w", err)
	}

	tlsCfg := &tls.Config{Certificates: []tls.Certificate{tlsCert}}
	ln, err := tls.Listen("tcp", ":"+port, tlsCfg)
	if err != nil {
		return fmt.Errorf("failed to open TLS listener on :%s: %w", port, err)
	}

	server := &http.Server{Handler: handler}
	return server.Serve(ln)
}

// tlsListenerFromFiles creates a net.Listener backed by TLS using cert/key files from disk.
func tlsListenerFromFiles(port, certFile, keyFile string) (net.Listener, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}
	return tls.Listen("tcp", ":"+port, tlsCfg)
}
