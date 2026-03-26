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

package ocg

import (
	"fmt"

	pdfcpu_model "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

type OptionalContentGroup struct {
	Name    string
	Visible bool
	Ref     int
}

type OCGEntry = OptionalContentGroup

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
					"OFF":      types.Array{},
					"ListMode": types.Name("VisiblePages"),
					"Name":     types.StringLiteral("Default"),
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
func AddOCG(ctx *pdfcpu_model.Context, name string) (*OptionalContentGroup, error) {
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
	dict.Update("AS", appendOrCreateUsageApplications(as, *indRef))

	return &OptionalContentGroup{Name: name, Visible: true, Ref: indRef.ObjectNumber.Value()}, nil
}

func EnableOCG(ctx *pdfcpu_model.Context, group *OptionalContentGroup, enabled bool) error {
	if ctx == nil {
		return fmt.Errorf("ocg: nil pdf context")
	}
	if group == nil {
		return fmt.Errorf("ocg: nil optional content group")
	}

	rootDict, err := ctx.Catalog()
	if err != nil {
		return err
	}

	ocPropsObj, found := rootDict.Find("OCProperties")
	if !found {
		return fmt.Errorf("ocg: OCProperties missing")
	}
	ocPropsDict, err := ctx.DereferenceDict(ocPropsObj)
	if err != nil {
		return err
	}

	dObj, found := ocPropsDict.Find("D")
	if !found {
		return fmt.Errorf("ocg: default OCG config missing")
	}
	dict, err := ctx.DereferenceDict(dObj)
	if err != nil {
		return err
	}

	on, err := arrayEntry(ctx, dict, "ON")
	if err != nil {
		return err
	}
	off, err := arrayEntry(ctx, dict, "OFF")
	if err != nil {
		return err
	}

	targetRef := *types.NewIndirectRef(group.Ref, 0)
	filtered := make(types.Array, 0, len(on))
	for _, obj := range on {
		indRef, ok := obj.(types.IndirectRef)
		if ok && indRef.ObjectNumber.Value() == group.Ref {
			continue
		}
		filtered = append(filtered, obj)
	}
	if enabled {
		filtered = append(filtered, targetRef)
	}
	dict.Update("ON", filtered)

	filteredOff := make(types.Array, 0, len(off))
	for _, obj := range off {
		indRef, ok := obj.(types.IndirectRef)
		if ok && indRef.ObjectNumber.Value() == group.Ref {
			continue
		}
		filteredOff = append(filteredOff, obj)
	}
	if !enabled {
		filteredOff = append(filteredOff, targetRef)
	}
	dict.Update("OFF", filteredOff)
	group.Visible = enabled
	return nil
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

func appendOrCreateUsageApplications(existing types.Array, indRef types.IndirectRef) types.Array {
	if len(existing) == 0 {
		return types.Array{
			usageApplication("View", indRef),
			usageApplication("Print", indRef),
			usageApplication("Export", indRef),
		}
	}

	out := make(types.Array, 0, len(existing))
	for _, obj := range existing {
		d, ok := obj.(types.Dict)
		if !ok {
			out = append(out, obj)
			continue
		}

		ocgs, ok := d["OCGs"].(types.Array)
		if !ok {
			d["OCGs"] = types.Array{indRef}
		} else if !containsIndirectRef(ocgs, indRef.ObjectNumber.Value()) {
			d["OCGs"] = append(ocgs, indRef)
		}
		out = append(out, d)
	}

	return out
}

func usageApplication(event string, indRef types.IndirectRef) types.Dict {
	return types.Dict(
		map[string]types.Object{
			"Category": types.NewNameArray(event),
			"Event":    types.Name(event),
			"OCGs":     types.Array{indRef},
		},
	)
}

func containsIndirectRef(arr types.Array, objNr int) bool {
	for _, obj := range arr {
		indRef, ok := obj.(types.IndirectRef)
		if ok && indRef.ObjectNumber.Value() == objNr {
			return true
		}
	}

	return false
}
