package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile_Success(t *testing.T) {
	// Create a temporary directory to act as our workspace.
	tmpDir := t.TempDir()

	// Write known content to the source file.
	srcContent := []byte("GhostGate test payload content")
	srcPath := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(srcPath, srcContent, 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	dstPath := filepath.Join(tmpDir, "destination.txt")

	// Copy the file.
	if err := CopyFile(srcPath, dstPath); err != nil {
		t.Fatalf("CopyFile() returned unexpected error: %v", err)
	}

	// Verify destination exists.
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Error("destination file does not exist after CopyFile()")
	}

	// Verify content matches.
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}
	if string(dstContent) != string(srcContent) {
		t.Errorf("destination content = %q; want %q", dstContent, srcContent)
	}
}

func TestCopyFile_MissingSource(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "nonexistent.txt")
	dstPath := filepath.Join(tmpDir, "destination.txt")

	err := CopyFile(srcPath, dstPath)
	if err == nil {
		t.Error("CopyFile() with missing source expected an error, got nil")
	}
}

func TestCopyFile_InvalidDestinationDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid source file.
	srcPath := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(srcPath, []byte("data"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Destination in a non-existent directory.
	dstPath := filepath.Join(tmpDir, "nonexistent_dir", "destination.txt")

	err := CopyFile(srcPath, dstPath)
	if err == nil {
		t.Error("CopyFile() to invalid destination directory expected an error, got nil")
	}
}
