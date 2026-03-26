// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// Ported from veraPDF (https://github.com/veraPDF/veraPDF-library)
// Original copyright: veraPDF Consortium
// Original license: Mozilla Public License 2.0

package pdfbox

import (
	"fmt"
	"strings"

	pdfloader "github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/loader"
	pdfmodel "github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/model"
	pdfcpuTypes "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

type AnnotationFeature struct {
	Type    string
	SubType string
	BBox    [4]float64
}

type ImageFeature struct {
	Width            int
	Height           int
	ColorSpace       string
	BitsPerComponent int
}

type PDFFeatures struct {
	HasStructureTree bool
	HasMetadata      bool
	HasDocumentTitle bool
	HasLanguage      bool
	FontsEmbedded    bool
	HasOutputIntent  bool
	ColorSpaces      []string
	Annotations      []AnnotationFeature
	Images           []ImageFeature
}

type FeatureExtractor struct{}

func (e *FeatureExtractor) Extract(pdfPath string) (*PDFFeatures, error) {
	if pdfPath == "" {
		return nil, fmt.Errorf("pdf path is empty")
	}
	doc, err := pdfloader.Open(pdfPath, "")
	if err != nil {
		return nil, fmt.Errorf("open pdf: %w", err)
	}
	defer doc.Close()

	features := &PDFFeatures{
		FontsEmbedded: true,
		ColorSpaces:   []string{},
		Annotations:   []AnnotationFeature{},
		Images:        []ImageFeature{},
	}

	if err := extractDocumentFeatures(doc, features); err != nil {
		return nil, err
	}
	if err := extractPageFeatures(doc, features); err != nil {
		return nil, err
	}

	features.ColorSpaces = uniqueStrings(features.ColorSpaces)
	return features, nil
}

func extractDocumentFeatures(doc *pdfmodel.PDDocument, features *PDFFeatures) error {
	rootDict, err := doc.Context.Catalog()
	if err != nil {
		return fmt.Errorf("load catalog: %w", err)
	}

	features.HasMetadata = hasNonNilEntry(rootDict, "Metadata")
	features.HasStructureTree = hasNonNilEntry(rootDict, "StructTreeRoot")
	if outputIntents, err := doc.Context.DereferenceArray(objectForKey(rootDict, "OutputIntents")); err == nil {
		features.HasOutputIntent = len(outputIntents) > 0
	}

	if lang, err := dereferenceTextEntry(doc, rootDict, "Lang"); err == nil {
		features.HasLanguage = strings.TrimSpace(lang) != ""
	}

	if title := strings.TrimSpace(doc.Context.Title); title != "" {
		features.HasDocumentTitle = true
		return nil
	}

	if doc.Context.Info != nil {
		infoDict, err := doc.Context.DereferenceDict(*doc.Context.Info)
		if err == nil && infoDict != nil {
			if title, err := dereferenceTextEntry(doc, infoDict, "Title"); err == nil {
				features.HasDocumentTitle = strings.TrimSpace(title) != ""
			}
		}
	}

	return nil
}

func extractPageFeatures(doc *pdfmodel.PDDocument, features *PDFFeatures) error {
	for pageNr := 1; pageNr <= doc.PageCount(); pageNr++ {
		pageDict, _, attrs, err := doc.Context.PageDict(pageNr, false)
		if err != nil {
			return fmt.Errorf("load page %d: %w", pageNr, err)
		}

		if annots, err := extractAnnotations(doc, pageDict); err == nil {
			features.Annotations = append(features.Annotations, annots...)
		}

		resourceDict := attrs.Resources
		if resourceDict == nil {
			resourceDict, _ = doc.Context.DereferenceDict(objectForKey(pageDict, "Resources"))
		}
		if resourceDict == nil {
			continue
		}

		if err := extractFonts(doc, resourceDict, features); err != nil {
			return fmt.Errorf("extract fonts on page %d: %w", pageNr, err)
		}
		if err := extractColorSpaces(doc, resourceDict, features); err != nil {
			return fmt.Errorf("extract colorspaces on page %d: %w", pageNr, err)
		}
		if err := extractXObjectFeatures(doc, resourceDict, features); err != nil {
			return fmt.Errorf("extract xobjects on page %d: %w", pageNr, err)
		}
	}

	return nil
}

func extractAnnotations(doc *pdfmodel.PDDocument, pageDict pdfcpuTypes.Dict) ([]AnnotationFeature, error) {
	arr, err := doc.Context.DereferenceArray(objectForKey(pageDict, "Annots"))
	if err != nil || arr == nil {
		return nil, nil
	}

	out := make([]AnnotationFeature, 0, len(arr))
	for _, obj := range arr {
		dict, err := doc.Context.DereferenceDict(obj)
		if err != nil || dict == nil {
			continue
		}

		ann := AnnotationFeature{
			Type:    "Annot",
			SubType: nameEntry(dict, "Subtype"),
		}
		if rect, err := doc.Context.DereferenceArray(objectForKey(dict, "Rect")); err == nil && len(rect) >= 4 {
			ann.BBox = [4]float64{
				numberAt(doc, rect, 0),
				numberAt(doc, rect, 1),
				numberAt(doc, rect, 2),
				numberAt(doc, rect, 3),
			}
		}
		out = append(out, ann)
	}

	return out, nil
}

func extractFonts(doc *pdfmodel.PDDocument, resourceDict pdfcpuTypes.Dict, features *PDFFeatures) error {
	fontObj, _ := resourceDict.Find("Font")
	fonts, err := doc.Context.DereferenceDict(fontObj)
	if err != nil || fonts == nil {
		return nil
	}

	for _, obj := range fonts {
		fontDict, err := doc.Context.DereferenceDict(obj)
		if err != nil || fontDict == nil {
			continue
		}
		if !fontIsEmbedded(doc, fontDict) {
			features.FontsEmbedded = false
		}
	}
	return nil
}

func extractColorSpaces(doc *pdfmodel.PDDocument, resourceDict pdfcpuTypes.Dict, features *PDFFeatures) error {
	csObj, _ := resourceDict.Find("ColorSpace")
	if csDict, err := doc.Context.DereferenceDict(csObj); err == nil && csDict != nil {
		for _, value := range csDict {
			if name := colorSpaceName(doc, value); name != "" {
				features.ColorSpaces = append(features.ColorSpaces, name)
			}
		}
	}
	return nil
}

func extractXObjectFeatures(doc *pdfmodel.PDDocument, resourceDict pdfcpuTypes.Dict, features *PDFFeatures) error {
	xObj, _ := resourceDict.Find("XObject")
	xObjects, err := doc.Context.DereferenceDict(xObj)
	if err != nil || xObjects == nil {
		return nil
	}

	for _, obj := range xObjects {
		dict, err := doc.Context.DereferenceDict(obj)
		if err != nil || dict == nil {
			continue
		}

		subtype := nameEntry(dict, "Subtype")
		if subtype == "Image" {
			image := ImageFeature{
				Width:            intEntry(dict, "Width"),
				Height:           intEntry(dict, "Height"),
				ColorSpace:       colorSpaceName(doc, objectForKey(dict, "ColorSpace")),
				BitsPerComponent: intEntry(dict, "BitsPerComponent"),
			}
			if image.ColorSpace != "" {
				features.ColorSpaces = append(features.ColorSpaces, image.ColorSpace)
			}
			features.Images = append(features.Images, image)
			continue
		}

		if subtype == "Form" {
			if resources, err := doc.Context.DereferenceDict(objectForKey(dict, "Resources")); err == nil && resources != nil {
				if err := extractFonts(doc, resources, features); err != nil {
					return err
				}
				if err := extractColorSpaces(doc, resources, features); err != nil {
					return err
				}
				if err := extractXObjectFeatures(doc, resources, features); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func fontIsEmbedded(doc *pdfmodel.PDDocument, fontDict pdfcpuTypes.Dict) bool {
	if fontDescriptorHasEmbeddedFont(doc, fontDict.DictEntry("FontDescriptor")) {
		return true
	}

	descendants, err := doc.Context.DereferenceArray(objectForKey(fontDict, "DescendantFonts"))
	if err != nil {
		return false
	}
	for _, descendant := range descendants {
		descendantDict, err := doc.Context.DereferenceDict(descendant)
		if err != nil || descendantDict == nil {
			continue
		}
		if fontDescriptorHasEmbeddedFont(doc, descendantDict.DictEntry("FontDescriptor")) {
			return true
		}
	}

	return false
}

func fontDescriptorHasEmbeddedFont(doc *pdfmodel.PDDocument, descriptor pdfcpuTypes.Dict) bool {
	if descriptor == nil {
		return false
	}
	for _, key := range []string{"FontFile", "FontFile2", "FontFile3"} {
		if hasNonNilEntry(descriptor, key) {
			return true
		}
		if obj, found := descriptor.Find(key); found {
			if deref, err := doc.Context.Dereference(obj); err == nil && deref != nil {
				return true
			}
		}
	}
	return false
}

func colorSpaceName(doc *pdfmodel.PDDocument, obj pdfcpuTypes.Object) string {
	if obj == nil {
		return ""
	}

	deref, err := doc.Context.Dereference(obj)
	if err != nil || deref == nil {
		return ""
	}

	switch v := deref.(type) {
	case pdfcpuTypes.Name:
		return strings.TrimPrefix(v.Value(), "/")
	case pdfcpuTypes.Array:
		if len(v) == 0 {
			return ""
		}
		if name, ok := v[0].(pdfcpuTypes.Name); ok {
			return strings.TrimPrefix(name.Value(), "/")
		}
	}

	return ""
}

func dereferenceTextEntry(doc *pdfmodel.PDDocument, dict pdfcpuTypes.Dict, key string) (string, error) {
	obj, found := dict.Find(key)
	if !found || obj == nil {
		return "", nil
	}
	return doc.Context.DereferenceText(obj)
}

func hasNonNilEntry(dict pdfcpuTypes.Dict, key string) bool {
	obj, found := dict.Find(key)
	return found && obj != nil
}

func objectForKey(dict pdfcpuTypes.Dict, key string) pdfcpuTypes.Object {
	obj, _ := dict.Find(key)
	return obj
}

func nameEntry(dict pdfcpuTypes.Dict, key string) string {
	if value := dict.NameEntry(key); value != nil {
		return strings.TrimPrefix(*value, "/")
	}
	return ""
}

func intEntry(dict pdfcpuTypes.Dict, key string) int {
	if value := dict.IntEntry(key); value != nil {
		return *value
	}
	return 0
}

func numberAt(doc *pdfmodel.PDDocument, arr pdfcpuTypes.Array, idx int) float64 {
	if idx < 0 || idx >= len(arr) {
		return 0
	}
	value, err := doc.Context.DereferenceNumber(arr[idx])
	if err != nil {
		return 0
	}
	return value
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return values
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
