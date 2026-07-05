package networking

import (
	"net"
)

// GetOutboundIP returns the machine's preferred outbound IP address.
// It dials a UDP address to determine the local interface used for external traffic;
// no actual connection is established.
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}
