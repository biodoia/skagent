package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DocLoader loads documentation from the filesystem
type DocLoader struct {
	docsPath string
}

// NewDocLoader creates a new documentation loader
func NewDocLoader(docsPath string) *DocLoader {
	return &DocLoader{
		docsPath: docsPath,
	}
}

// LoadSpecKitDocs loads all markdown files from the docs directory
func (d *DocLoader) LoadSpecKitDocs() (string, error) {
	// Use os.ReadDir instead of deprecated ioutil.ReadDir
	entries, err := os.ReadDir(d.docsPath)
	if err != nil {
		return "", fmt.Errorf("failed to read docs directory: %w", err)
	}

	var content strings.Builder
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".md" {
			data, err := os.ReadFile(filepath.Join(d.docsPath, entry.Name()))
			if err != nil {
				return "", fmt.Errorf("failed to read file %s: %w", entry.Name(), err)
			}
			content.Write(data)
			content.WriteString("\n\n")
		}
	}

	return content.String(), nil
}

// LoadFile loads a single documentation file
func (d *DocLoader) LoadFile(filename string) (string, error) {
	data, err := os.ReadFile(filepath.Join(d.docsPath, filename))
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	return string(data), nil
}

// ListDocs returns a list of available documentation files
func (d *DocLoader) ListDocs() ([]string, error) {
	entries, err := os.ReadDir(d.docsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read docs directory: %w", err)
	}

	var docs []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
			docs = append(docs, entry.Name())
		}
	}
	return docs, nil
}
