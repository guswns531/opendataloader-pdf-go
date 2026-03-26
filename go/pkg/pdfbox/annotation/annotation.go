// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	LayerRef int
}

// AddSquareAnnotation adds a colored square annotation to a PDF page.
func AddSquareAnnotation(ctx *pdfcpu_model.Context, pageIdx int, ann *SquareAnnotation) error {
	if ctx == nil {
		return fmt.Errorf("annotation: nil pdf context")
	}
	if ann == nil {
		return fmt.Errorf("annotation: nil square annotation")
	}
	if pageIdx < 0 {
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

	layerRef := ann.OCGRef
	if layerRef == 0 {
		layerRef = ann.LayerRef
	}
	if layerRef > 0 {
		dict.Update("OC", types.Dict(
			map[string]types.Object{
				"Type": types.Name("OCMD"),
				"OCGs": types.Array{*types.NewIndirectRef(layerRef, 0)},
				"P":    types.Name("AllOn"),
				"VE":   types.Array{},
			},
		))
	}

	return nil
}
