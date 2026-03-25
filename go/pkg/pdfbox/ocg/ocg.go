// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
//
// This package provides functionality equivalent to Apache PDFBox 3.0.4
// (https://pdfbox.apache.org/), implemented using pdfcpu.

package ocg

import (
	"fmt"

	pdfcpu_model "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

type OCGEntry struct {
	Name string
	Ref  int
}

// SetupOCProperties initializes /OCProperties if not present.
func SetupOCProperties(ctx *pdfcpu_model.Context) error {
	if ctx == nil {
		return fmt.Errorf("ocg: nil pdf context")
	}

	rootDict, err := ctx.Catalog()
	if err != nil {
		return err
	}

	if _, ok := rootDict.Find("OCProperties"); ok {
		return nil
	}

	rootDict.Update("OCProperties", types.Dict(
		map[string]types.Object{
			"OCGs": types.Array{},
			"D": types.Dict(
				map[string]types.Object{
					"AS":       types.Array{},
					"ON":       types.Array{},
					"Order":    types.Array{},
					"RBGroups": types.Array{},
				},
			),
		},
	))

	return nil
}

// AddOCG adds an OCG to /Catalog /OCProperties.
func AddOCG(ctx *pdfcpu_model.Context, name string) (*OCGEntry, error) {
	if err := SetupOCProperties(ctx); err != nil {
		return nil, err
	}

	rootDict, err := ctx.Catalog()
	if err != nil {
		return nil, err
	}

	ocPropsObj, _ := rootDict.Find("OCProperties")
	ocPropsDict, err := ctx.DereferenceDict(ocPropsObj)
	if err != nil {
		return nil, err
	}

	ocgDict := types.Dict(
		map[string]types.Object{
			"Name": types.StringLiteral(name),
			"Type": types.Name("OCG"),
			"Usage": types.Dict(
				map[string]types.Object{
					"PageElement": types.Dict(map[string]types.Object{"Subtype": types.Name("FG")}),
					"View":        types.Dict(map[string]types.Object{"ViewState": types.Name("ON")}),
					"Print":       types.Dict(map[string]types.Object{"PrintState": types.Name("ON")}),
					"Export":      types.Dict(map[string]types.Object{"ExportState": types.Name("ON")}),
				},
			),
		},
	)

	indRef, err := ctx.IndRefForNewObject(ocgDict)
	if err != nil {
		return nil, err
	}

	ocgs, err := arrayEntry(ctx, ocPropsDict, "OCGs")
	if err != nil {
		return nil, err
	}
	ocPropsDict.Update("OCGs", append(ocgs, *indRef))

	dObj, found := ocPropsDict.Find("D")
	if !found {
		dObj = types.Dict{}
		ocPropsDict.Update("D", dObj)
	}

	dict, err := ctx.DereferenceDict(dObj)
	if err != nil {
		return nil, err
	}

	on, err := arrayEntry(ctx, dict, "ON")
	if err != nil {
		return nil, err
	}
	dict.Update("ON", append(on, *indRef))

	order, err := arrayEntry(ctx, dict, "Order")
	if err != nil {
		return nil, err
	}
	dict.Update("Order", append(order, *indRef))

	as, err := arrayEntry(ctx, dict, "AS")
	if err != nil {
		return nil, err
	}
	if len(as) == 0 {
		dict.Update("AS", types.Array{
			types.Dict(
				map[string]types.Object{
					"Category": types.NewNameArray("View"),
					"Event":    types.Name("View"),
					"OCGs":     types.Array{*indRef},
				},
			),
			types.Dict(
				map[string]types.Object{
					"Category": types.NewNameArray("Print"),
					"Event":    types.Name("Print"),
					"OCGs":     types.Array{*indRef},
				},
			),
			types.Dict(
				map[string]types.Object{
					"Category": types.NewNameArray("Export"),
					"Event":    types.Name("Export"),
					"OCGs":     types.Array{*indRef},
				},
			),
		})
	}

	return &OCGEntry{Name: name, Ref: indRef.ObjectNumber.Value()}, nil
}

func arrayEntry(ctx *pdfcpu_model.Context, dict types.Dict, key string) (types.Array, error) {
	obj, found := dict.Find(key)
	if !found {
		return types.Array{}, nil
	}

	arr, err := ctx.DereferenceArray(obj)
	if err != nil {
		return nil, err
	}

	return arr, nil
}
