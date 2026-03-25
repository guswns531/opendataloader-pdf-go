// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
//
// This package provides functionality equivalent to Apache PDFBox 3.0.4
// (https://pdfbox.apache.org/), implemented using pdfcpu.

package annotation

import (
	"fmt"

	pdfcpu "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/color"
	pdfcpu_model "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

type SquareAnnotation struct {
	X, Y     float64
	Width    float64
	Height   float64
	Color    [3]float64
	Opacity  float64
	Contents string
	OCGRef   int
}

// AddSquareAnnotation adds a colored square annotation to a PDF page.
func AddSquareAnnotation(ctx *pdfcpu_model.Context, pageIdx int, ann *SquareAnnotation) error {
	if ctx == nil {
		return fmt.Errorf("annotation: nil pdf context")
	}
	if ann == nil {
		return fmt.Errorf("annotation: nil square annotation")
	}
	if pageIdx < 0 || pageIdx >= ctx.PageCount {
		return fmt.Errorf("annotation: page index out of range: %d", pageIdx)
	}

	pageNr := pageIdx + 1
	pageIndRef, err := ctx.PageDictIndRef(pageNr)
	if err != nil {
		return err
	}

	pageDict, err := ctx.DereferenceDict(*pageIndRef)
	if err != nil {
		return err
	}

	opacity := ann.Opacity
	if opacity == 0 {
		opacity = 0.4
	}

	rect := types.RectForWidthAndHeight(ann.X, ann.Y, ann.Width, ann.Height)
	strokeColor := &color.SimpleColor{
		R: float32(ann.Color[0]),
		G: float32(ann.Color[1]),
		B: float32(ann.Color[2]),
	}

	renderer := pdfcpu_model.NewSquareAnnotation(
		*rect,
		ann.Contents,
		"",
		"",
		pdfcpu_model.AnnPrint,
		strokeColor,
		"",
		nil,
		&opacity,
		"",
		"",
		strokeColor,
		0, 0, 0, 0,
		1,
		pdfcpu_model.BSSolid,
		false,
		0,
	)

	_, dict, err := pdfcpu.AddAnnotation(ctx, pageIndRef, pageDict, pageNr, renderer, false)
	if err != nil {
		return err
	}

	if ann.OCGRef > 0 {
		dict.Update("OC", *types.NewIndirectRef(ann.OCGRef, 0))
	}

	return nil
}
