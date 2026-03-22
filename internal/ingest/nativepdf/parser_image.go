package nativepdf

import (
	"bytes"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func imageArtifactFromXObject(graph *objectGraph, ref pdfRef, name string, page model.PageMetadata, matrix matrix2D, sequence int) (*model.RawArtifact, bool) {
	if graph == nil {
		return nil, false
	}
	object := graph.objects[ref]
	if object == nil {
		return nil, false
	}
	dict, ok := object.Value.(pdfDict)
	if !ok || normalizedPDFName(dict["Subtype"]) != "Image" {
		return nil, false
	}

	width, ok := floatValue(dict["Width"])
	if !ok || width <= 0 {
		width = 1
	}
	height, ok := floatValue(dict["Height"])
	if !ok || height <= 0 {
		height = 1
	}

	stream, _ := parsePDFStreamObject(object.Raw)
	payload := stream.Payload
	filters := canonicalizeFilterNames(nameListValue(graph, dict["Filter"]))
	if len(filters) == 0 {
		filters = stream.Filters
	}
	format := inferImageFormat(filters, payload)
	bounds := matrix.transformBox(rectangleFromSize(width, height)).Normalize()
	bitsPerComponent, _ := intValueResolved(graph, dict["BitsPerComponent"])
	colorSpace := primaryNameValue(graph, dict["ColorSpace"])

	return &model.RawArtifact{
		Kind:             model.ArtifactKindImage,
		PageIndex:        page.Index,
		PageNumber:       page.Number,
		Sequence:         sequence,
		Bounds:           bounds,
		Format:           format,
		Data:             append([]byte(nil), payload...),
		ColorSpace:       colorSpace,
		BitsPerComponent: bitsPerComponent,
		Filters:          filters,
		Style: model.TextProperties{
			Content: name,
		},
	}, true
}

func inferImageFormat(filters []string, payload []byte) model.ImageFormat {
	if containsName(filters, "DCTDecode") || hasJPEGSignature(payload) {
		return model.ImageFormatJPEG
	}
	if hasPNGSignature(payload) {
		return model.ImageFormatPNG
	}
	return ""
}

func containsName(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func canonicalizeFilterNames(filters []string) []string {
	if len(filters) == 0 {
		return nil
	}
	out := make([]string, 0, len(filters))
	for _, filter := range filters {
		if canonical := canonicalFilterName(filter); canonical != "" {
			out = append(out, canonical)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func canonicalFilterName(name string) string {
	switch name {
	case "AHx":
		return "ASCIIHexDecode"
	case "A85":
		return "ASCII85Decode"
	case "LZW":
		return "LZWDecode"
	case "Fl":
		return "FlateDecode"
	case "RL":
		return "RunLengthDecode"
	case "CCF":
		return "CCITTFaxDecode"
	case "DCT":
		return "DCTDecode"
	case "JPX":
		return "JPXDecode"
	default:
		return name
	}
}

func nameListValue(g *objectGraph, value pdfValue) []string {
	switch typed := value.(type) {
	case pdfName, string:
		name := normalizedPDFName(typed)
		if name == "" {
			return nil
		}
		return []string{name}
	case pdfArray:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			out = append(out, nameListValue(g, item)...)
		}
		if len(out) == 0 {
			return nil
		}
		return out
	case pdfRef:
		if g == nil {
			return nil
		}
		object := g.objects[typed]
		if object == nil {
			return nil
		}
		return nameListValue(g, object.Value)
	default:
		return nil
	}
}

func primaryNameValue(g *objectGraph, value pdfValue) string {
	switch typed := value.(type) {
	case pdfName, string:
		return normalizedPDFName(typed)
	case pdfArray:
		for _, item := range typed {
			if name := primaryNameValue(g, item); name != "" {
				return name
			}
		}
		return ""
	case pdfRef:
		if g == nil {
			return ""
		}
		object := g.objects[typed]
		if object == nil {
			return ""
		}
		return primaryNameValue(g, object.Value)
	default:
		return ""
	}
}

func intValueResolved(g *objectGraph, value pdfValue) (int, bool) {
	switch typed := value.(type) {
	case pdfRef:
		if g == nil {
			return 0, false
		}
		object := g.objects[typed]
		if object == nil {
			return 0, false
		}
		return intValueResolved(g, object.Value)
	default:
		return intValue(value)
	}
}

func hasJPEGSignature(payload []byte) bool {
	return bytes.HasPrefix(payload, []byte{0xFF, 0xD8})
}

func hasPNGSignature(payload []byte) bool {
	return bytes.HasPrefix(payload, []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A})
}
