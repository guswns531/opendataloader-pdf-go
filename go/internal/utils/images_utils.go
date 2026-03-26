// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func NormalizeImageFormat(format string) string {
	switch strings.ToLower(strings.TrimPrefix(format, ".")) {
	case "jpg":
		return "jpeg"
	case "jpeg", "png":
		return strings.ToLower(strings.TrimPrefix(format, "."))
	default:
		return "png"
	}
}

func SaveImageExternal(data []byte, outputDir, baseName string, pageNumber, imageIndex int, format string) (string, error) {
	if len(data) == 0 {
		return "", nil
	}
	if outputDir == "" {
		return "", fmt.Errorf("outputDir is required")
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", err
	}

	format = NormalizeImageFormat(format)
	if baseName == "" {
		baseName = "image"
	}
	fileName := fmt.Sprintf("%s_page_%03d_image_%03d.%s", baseName, pageNumber, imageIndex, format)
	fullPath := filepath.Join(outputDir, fileName)
	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
		return "", err
	}
	return fullPath, nil
}

func IsImageFileExists(fileName string) bool {
	if strings.TrimSpace(fileName) == "" {
		return false
	}
	info, err := os.Stat(fileName)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
