package utils

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func GetProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	projectRootMarker := "rfid-backend"

	// Check if currentDir contains the project root marker
	if strings.Contains(currentDir, projectRootMarker) {
		// Find the index of project root marker in the path
		index := strings.Index(currentDir, projectRootMarker)
		// Add the length of projectRootMarker to the index to include it in the final path
		projectRoot := currentDir[:index+len(projectRootMarker)]
		return filepath.Clean(projectRoot), nil
	}
	log.Printf("config filepath:%s", filepath.Clean(currentDir))
	// If the current directory does not contain the project root marker,
	// it is assumed to be the project root itself
	return filepath.Clean(currentDir), nil
}
