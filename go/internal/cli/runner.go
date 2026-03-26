/*
 * Copyright 2025-2026 Hancom Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
	_ "github.com/opendataloader-project/opendataloader-pdf-go/internal/processors"
)

type UsageError struct {
	err error
}

func (e *UsageError) Error() string {
	return e.err.Error()
}

func (e *UsageError) Unwrap() error {
	return e.err
}

func Run(opts *CLIOptions, args []string) error {
	if opts == nil {
		opts = &CLIOptions{}
	}

	config, err := opts.BuildConfig(args)
	if err != nil {
		return &UsageError{err: err}
	}

	var errs []error
	for _, path := range args {
		if err := processPath(path, config); err != nil {
			errs = append(errs, err)
		}
	}
	api.Shutdown()

	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("one or more files failed: %w", errors.Join(errs...))
}

func processPath(path string, config *api.Config) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		return err
	}

	if info.IsDir() {
		var errs []error
		walkErr := filepath.WalkDir(path, func(current string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				errs = append(errs, walkErr)
				return nil
			}
			if d.IsDir() || !isPDFPath(current) {
				return nil
			}
			if err := api.ProcessFile(current, config); err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", current, err))
			}
			return nil
		})
		if walkErr != nil {
			errs = append(errs, walkErr)
		}
		if len(errs) == 0 {
			return nil
		}
		return errors.Join(errs...)
	}

	if !isPDFPath(path) {
		return fmt.Errorf("not a PDF file: %s", path)
	}
	return api.ProcessFile(path, config)
}

func isPDFPath(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".pdf")
}
