package nativepdf

import "strings"

func resolveFontNames(g *objectGraph, resources pdfDict) map[string]string {
	if g == nil || len(resources) == 0 {
		return nil
	}

	fontsValue, ok := resources["Font"]
	if !ok {
		return nil
	}
	fontsDict, ok := resolveDictValue(g, fontsValue)
	if !ok || len(fontsDict) == 0 {
		return nil
	}

	out := make(map[string]string, len(fontsDict))
	for alias, value := range fontsDict {
		fontDict, ok := resolveDictValue(g, value)
		if !ok {
			continue
		}
		name := fontNameFromDict(fontDict)
		if name == "" {
			continue
		}
		out[alias] = name
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func resolveImageXObjectRefs(g *objectGraph, resources pdfDict) map[string]pdfRef {
	if g == nil || len(resources) == 0 {
		return nil
	}

	xobjectsValue, ok := resources["XObject"]
	if !ok {
		return nil
	}
	xobjectsDict, ok := resolveDictValue(g, xobjectsValue)
	if !ok || len(xobjectsDict) == 0 {
		return nil
	}

	out := make(map[string]pdfRef, len(xobjectsDict))
	for name, value := range xobjectsDict {
		ref, ok := value.(pdfRef)
		if !ok || !isImageXObjectRef(g, ref) {
			continue
		}
		out[name] = ref
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func resolveDictValue(g *objectGraph, value pdfValue) (pdfDict, bool) {
	switch typed := value.(type) {
	case pdfDict:
		return typed, true
	case pdfRef:
		if g == nil {
			return nil, false
		}
		object := g.objects[typed]
		if object == nil {
			return nil, false
		}
		dict, ok := object.Value.(pdfDict)
		return dict, ok
	default:
		return nil, false
	}
}

func fontNameFromDict(dict pdfDict) string {
	if len(dict) == 0 {
		return ""
	}
	if name := normalizedPDFName(dict["BaseFont"]); name != "" {
		return name
	}
	if name := normalizedPDFName(dict["Name"]); name != "" {
		return name
	}
	if subtype := normalizedPDFName(dict["Subtype"]); subtype != "" {
		return subtype
	}
	return ""
}

func isImageXObjectRef(g *objectGraph, ref pdfRef) bool {
	if g == nil {
		return false
	}
	object := g.objects[ref]
	if object == nil {
		return false
	}
	dict, ok := object.Value.(pdfDict)
	if !ok {
		return false
	}
	return normalizedPDFName(dict["Subtype"]) == "Image"
}

func normalizedPDFName(value pdfValue) string {
	switch typed := value.(type) {
	case pdfName:
		return strings.TrimPrefix(string(typed), "/")
	case string:
		return strings.TrimPrefix(typed, "/")
	default:
		return ""
	}
}
