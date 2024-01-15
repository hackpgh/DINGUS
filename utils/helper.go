package utils

import (
	"os"
	"path/filepath"
)

func GetProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Join(currentDir, ".."), nil
}
