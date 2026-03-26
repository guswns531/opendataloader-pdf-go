// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
//
// This package provides functionality equivalent to Apache PDFBox 3.0.4
// (https://pdfbox.apache.org/), implemented using pdfcpu.

package model

import (
	"fmt"
	"os"
	"sync"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	pdfmodel "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

var disableConfigDir sync.Once

// PDDocument represents a loaded PDF document.
type PDDocument struct {
	Path     string
	Password string
	Ctx      *pdfmodel.Context
	Context  *pdfmodel.Context

	file     *os.File
	pageDims []float64Dim
}

type float64Dim struct {
	Width  float64
	Height float64
}

type PDPage struct {
	Number int
	Width  float64
	Height float64
	Rotate int
}

func Open(path, password string) (*PDDocument, error) {
	disableConfigDir.Do(api.DisableConfigDir)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	conf := pdfmodel.NewDefaultConfiguration()
	if password != "" {
		conf.UserPW = password
		conf.OwnerPW = password
	}

	ctx, err := api.ReadValidateAndOptimize(f, conf)
	if err != nil {
		_ = f.Close()
		return nil, err
	}

	dims, err := ctx.PageDims()
	if err != nil {
		_ = f.Close()
		return nil, err
	}

	pageDims := make([]float64Dim, 0, len(dims))
	for _, dim := range dims {
		pageDims = append(pageDims, float64Dim{
			Width:  dim.Width,
			Height: dim.Height,
		})
	}

	return &PDDocument{
		Path:     path,
		Password: password,
		Ctx:      ctx,
		Context:  ctx,
		file:     f,
		pageDims: pageDims,
	}, nil
}

func (d *PDDocument) PageCount() int {
	if d == nil || d.Context == nil {
		return 0
	}
	return d.Context.PageCount
}

func (d *PDDocument) GetPage(pageIdx int) (*PDPage, error) {
	if d == nil || d.Context == nil {
		return nil, fmt.Errorf("pdf document is not open")
	}
	if pageIdx < 0 || pageIdx >= d.PageCount() {
		return nil, fmt.Errorf("page index out of range: %d", pageIdx)
	}

	_, _, attrs, err := d.Context.PageDict(pageIdx+1, false)
	if err != nil {
		return nil, err
	}

	dim := d.pageDims[pageIdx]
	return &PDPage{
		Number: pageIdx,
		Width:  dim.Width,
		Height: dim.Height,
		Rotate: attrs.Rotate,
	}, nil
}

func (d *PDDocument) Close() error {
	if d == nil || d.file == nil {
		return nil
	}
	err := d.file.Close()
	d.file = nil
	d.Ctx = nil
	d.Context = nil
	d.pageDims = nil
	return err
}
