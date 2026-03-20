package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var supportedExtensions = map[string]struct{}{
	".pdf":  {},
	".json": {},
}

// Discover expands the provided input paths into supported PDF and JSON files.
//
// File inputs are preserved in the order they were provided. Directory inputs
// are traversed recursively in a deterministic order based on lexical
// directory entry sorting.
func Discover(inputs []string) ([]string, error) {
	discovered := make([]string, 0, len(inputs))
	for _, input := range inputs {
		paths, err := discoverPath(input)
		if err != nil {
			return nil, err
		}
		discovered = append(discovered, paths...)
	}
	return discovered, nil
}

// IsSupported reports whether path looks like a supported discovery input.
func IsSupported(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := supportedExtensions[ext]
	return ok
}

func discoverPath(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat %q: %w", path, err)
	}

	if info.IsDir() {
		return discoverDir(path)
	}

	if !IsSupported(path) {
		return nil, nil
	}

	return []string{path}, nil
}

func discoverDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %q: %w", dir, err)
	}

	discovered := make([]string, 0, len(entries))
	for _, entry := range entries {
		childPath := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			children, err := discoverDir(childPath)
			if err != nil {
				return nil, err
			}
			discovered = append(discovered, children...)
			continue
		}

		if IsSupported(childPath) {
			discovered = append(discovered, childPath)
		}
	}

	return discovered, nil
}
