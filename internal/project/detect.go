package project

import (
	"os"
	"path/filepath"
	"strings"
)

func DetectProjectType(dir string) string {
	if fileContains(filepath.Join(dir, "composer.json"), `"drupal/`) {
		return "drupal"
	}
	if fileExists(filepath.Join(dir, "package.json")) {
		return "node"
	}
	if fileExists(filepath.Join(dir, "go.mod")) {
		return "go"
	}
	if fileExists(filepath.Join(dir, "Cargo.toml")) {
		return "rust"
	}
	if fileExists(filepath.Join(dir, "pyproject.toml")) {
		return "python"
	}
	if fileExists(filepath.Join(dir, "requirements.txt")) {
		return "python"
	}
	return "generic"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func fileContains(path string, substr string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), substr)
}
