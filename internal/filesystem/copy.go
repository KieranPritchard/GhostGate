package filesystem

import (
    "io"
    "os"
)

// Function to copy file
func CopyFile(src, dst string) error {
	// Stores the source file
	sourceFile, err := os.Open(src)

	// Catches the errors
	if err != nil {
		// Returns the error
		return err
	}
	// Closes the file when done
	defer sourceFile.Close()

	// Creates the destination file
	destinationFile, err := os.Create(dst)

	// Catches the errors
	if err != nil {
		// Returns the error
		return err
	}
	// Closes the file when done
	defer destinationFile.Close()

	// Copies the file to the destination
	_, err = io.Copy(destinationFile, sourceFile)
    // Catches the error
	if err != nil {
		// Returns the error
        return err
    }

	// Returns the nil
    return nil
}