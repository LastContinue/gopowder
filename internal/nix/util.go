package nix

import (
	"path/filepath"
	"strings"
)

func expandTilde(path string, homeFinder func() (string, error)) (string, error) {

	home, err := homeFinder()
	if err != nil {
		return "", err
	}

	if path == "~" {
		return home, nil
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:]), nil
	}

	// This should catch everything else including
	// paths with no '~' and blank/empty paths
	return path, nil
}

func green(s string) string {
	// ANSI escape codes: green text
	return "\x1b[32m" + s + "\x1b[0m"
}
