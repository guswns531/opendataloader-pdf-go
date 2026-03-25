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

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/cli"
	"github.com/spf13/cobra"
)

type parseError struct {
	err error
}

func (e *parseError) Error() string {
	return e.err.Error()
}

func (e *parseError) Unwrap() error {
	return e.err
}

func main() {
	opts := &cli.CLIOptions{}
	rootCmd := &cobra.Command{
		Use:           "opendataloader-pdf [flags] <file.pdf> [...]",
		Short:         "PDF parser for AI-ready data extraction",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.ExportOptions {
				return cli.ExportOptionsJSON()
			}
			if len(args) == 0 {
				return cmd.Help()
			}
			return cli.Run(opts, args)
		},
	}
	rootCmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return &parseError{err: err}
	})
	cli.AddFlags(rootCmd, opts)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		var parseErr *parseError
		if errors.As(err, &parseErr) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
