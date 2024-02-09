package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func GetProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	projectRootMarker := "dingus"

	if strings.Contains(currentDir, projectRootMarker) {
		index := strings.Index(currentDir, projectRootMarker)

		projectRoot := currentDir[:index+len(projectRootMarker)]

		return filepath.Clean(projectRoot), nil
	}

	return filepath.Clean(currentDir), nil
}
