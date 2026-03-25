// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package generators

import (
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	htmlgen "github.com/opendataloader-project/opendataloader-pdf-go/internal/generators/html"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/generators/markdown"
	textgen "github.com/opendataloader-project/opendataloader-pdf-go/internal/generators/text"
)

type Generator interface {
	Generate(doc *entities.Document) (string, error)
}

func GetGenerator(format string, config *api.Config) Generator {
	switch format {
	case api.FormatText:
		return textgen.NewTextGenerator(config)
	case api.FormatHTML:
		return htmlgen.NewHtmlGenerator(config)
	case api.FormatMarkdown:
		return markdown.NewMarkdownGenerator(config, false, false)
	case api.FormatMarkdownWithHTML:
		return markdown.NewMarkdownHTMLGenerator(config)
	case api.FormatMarkdownWithImages:
		return markdown.NewMarkdownGenerator(config, false, true)
	case api.FormatJSON, api.FormatPDF:
		return nil
	default:
		return nil
	}
}
