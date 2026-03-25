//go:build tools

package tools

import (
	_ "github.com/goccy/go-json"
	_ "github.com/pdfcpu/pdfcpu/pkg/api"
	_ "github.com/spf13/cobra"
	_ "github.com/spf13/pflag"
	_ "github.com/stretchr/testify/assert"
)
