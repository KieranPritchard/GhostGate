package util

import (
	"bufio"
	"embed"
	"io/fs"
	"math/rand"
	"time"
)

//go:embed resources/*
var embeddedUserAgentFiles embed.FS

// Type as fs.FS interface so tests can assign fstest.MapFS to it
var userAgentFiles fs.FS = embeddedUserAgentFiles

func loadUserAgents() ([]string, error)  {
	// Functions to load user agents file
	
	// Stores the agents
	agents := make([]string, 0)

	// Opens the directory and closes when done
	file, err := userAgentFiles.Open("resources/user_agents.txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Creates a scanner for the file
	scanner := bufio.NewScanner(file)

	// Scans in each of the files
	for scanner.Scan() {
		line := scanner.Text()
		// Process the line
		agents = append(agents, line)
	}

	// Checks for scanning error
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return agents, nil
}

func GetRandomHeader() (string, error) {
	// Function to get random header

	// Loads in the user agents 
	userAgents, err := loadUserAgents()
	if err != nil {
		return "", err
	} 
	
	// Gets a random header
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(userAgents))

	// Returns the index
	return userAgents[index], nil
}