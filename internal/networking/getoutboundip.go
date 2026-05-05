package networking

import (
	"log"
	"net"
)

// Gets the outbound address I am after
func GetOutboundIP() net.IP {
	// We use Google's public DNS as a destination, but any IP works.
	// No connection is actually established.
	conn, err := net.Dial("udp", "8.8.8.8:80")
	
	// Catches the error
	if err != nil {
		// Logs the error
		log.Fatal(err)
	}
	// Closes the file when done
	defer conn.Close()

	// Stores the local address
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	// Returns local ip
	return localAddr.IP
}