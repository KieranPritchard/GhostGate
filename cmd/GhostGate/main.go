package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

// Gets the outbound address I am after
func getOutboundIP() net.IP {
	// We use Google's public DNS as a destination, but any IP works.
	// No connection is actually established.
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

// Function for staging the payload directory
func stagePayloadDirectory(port string, stagingDir string)  {
	// Creates the staging diredctory if it doesnt exist
	if _, err := os.Stat(stagingDir); os.IsNotExist(err) {
		// Stores the errors from creating the directory
		err := os.MkdirAll(stagingDir, 0755)

		// Checks if there is a error
		if err != nil {
			log.Fatal("Error creating directory:", err)
		}

		// Add a possible section for adding files from a folder here
	}

	// Creates the file server handler
	fileServer := http.FileServer(http.Dir(stagingDir))

	// Outputs information
	fmt.Printf("[*] Go Payload Staging Server running on port %s\n", port)
	fmt.Printf("[*] Serving files from: %s\n", stagingDir)
	fmt.Printf("[*] Target download example: curl http://%s:%s/%s/file\n", getOutboundIP(), port, stagingDir)

	// Starts the server
	err := http.ListenAndServe(":"+port, fileServer)

	// Checks if there is an error
	if err != nil {
		log.Fatal("Server failed:", err)
	}
}

func main(){
	/*
		Flags for each function which the listner can carry out
	*/

	// Stores the flag set for setting up a payload staging directory
	stageDirectoryOption := flag.NewFlagSet("stageDir", flag.ExitOnError)

	// Stores the flags for this flagset
	stageDirectoryPort := stageDirectoryOption.String("p", "8080", "Specifies the port number to host the server")
	stageDirectoryDir := stageDirectoryOption.String("f", "payloads", "Specifies the file path of the directory")

	// Checks if the user has provided a subcommand
	if len(os.Args) < 2 {
		// Outputs an invaild command
		fmt.Println("Unexpected input")
		// Exits the program
		os.Exit(1)
	}

	// Switch to select the command to be used
	switch os.Args[1] {
	case "stageDir":
		// Parse the flags starting from the 3rd argument (index 2)
		stageDirectoryOption.Parse(os.Args[2:])

		// Passes the flags into the function
		stagePayloadDirectory(*stageDirectoryPort, *stageDirectoryDir)
	}
}