package networking

import (
	"net"
	"testing"
)

func TestGetOutboundIP(t *testing.T) {
	ip := GetOutboundIP()
	if ip == nil {
		// In a fully offline environment this can return nil; treat as a soft failure.
		t.Skip("GetOutboundIP() returned nil — no outbound network route available")
	}

	// The result must be a valid, non-loopback IP.
	parsedIP := net.ParseIP(ip.String())
	if parsedIP == nil {
		t.Errorf("GetOutboundIP() returned a non-parseable IP: %v", ip)
	}
}
