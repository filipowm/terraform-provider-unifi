package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ensurePath checks if a path exists and is a directory, if not it creates the directory. Returns true if the directories were created.
func ensurePath(path string) (bool, error) {
	// Check if output directory exists, if not create and perform extraction
	targetInfo, err := os.Stat(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return false, err
		}
		if err = os.MkdirAll(path, 0o755); err != nil {
			return false, err
		}
		return true, nil
	}
	if !targetInfo.IsDir() {
		return false, fmt.Errorf("%s isn't a directory", path)
	}
	return false, nil
}

func findProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// Walk up the directory tree until we find a go.mod file
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}
		if wd == "/" {
			break
		}
		wd = filepath.Dir(wd)
	}
	return "", errors.New("unable to find project root")
}

func findCodegenDir() (string, error) {
	root, err := findProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "codegen"), nil
}
