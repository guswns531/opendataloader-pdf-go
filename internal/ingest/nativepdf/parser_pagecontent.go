package nativepdf

import (
	"strings"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func shellArtifactsFromPagePlans(data []byte, plans []pagePlan) [][]*model.RawArtifact {
	if len(plans) == 0 {
		return nil
	}

	graph, _ := scanIndirectObjects(data)
	glyphMap := extractToUnicodeMap(data)
	out := make([][]*model.RawArtifact, len(plans))
	for i, plan := range plans {
		streams := decodedStreamsForPagePlan(data, graph, plan)
		if len(streams) == 0 {
			continue
		}
		if positioned := extractPositionedTextArtifactsFromStreams(streams, plan.Metadata, glyphMap, plan.FontNames, plan.FontMetrics); len(positioned) > 0 {
			out[i] = positioned
			continue
		}
		if texts := extractTextShellStringsFromStreams(streams, glyphMap); len(texts) > 0 {
			out[i] = buildTextArtifactsFromStrings(plan.Metadata, texts)
		}
	}
	if hasArtifacts(out) {
		return out
	}
	return nil
}

func decodedStreamsForPagePlan(data []byte, graph *objectGraph, plan pagePlan) [][]byte {
	if len(plan.ContentRefs) == 0 {
		return nil
	}
	if graph != nil {
		if streams := graph.decodedStreamsForRefs(plan.ContentRefs); len(streams) > 0 {
			return streams
		}
	}
	return decodedStreamsForRefsByScan(data, plan.ContentRefs)
}

func decodedStreamsForRefsByScan(data []byte, refs []pdfRef) [][]byte {
	if len(data) == 0 || len(refs) == 0 {
		return nil
	}
	out := make([][]byte, 0, len(refs))
	for _, ref := range refs {
		if decoded, ok := decodedStreamForRefByScan(data, ref); ok && len(decoded) > 0 {
			out = append(out, decoded)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func decodedStreamForRefByScan(data []byte, want pdfRef) ([]byte, bool) {
	offset := 0
	for {
		start, ref, ok := findNextObjectHeader(data, offset)
		if !ok {
			return nil, false
		}
		endRel := strings.Index(string(data[start:]), "endobj")
		if endRel < 0 {
			return nil, false
		}
		end := start + endRel
		if ref == want {
			raw := data[start:end]
			streams := extractDecodedStreams(raw)
			if len(streams) == 0 {
				return nil, false
			}
			return streams[0], true
		}
		offset = end + len("endobj")
	}
}

func extractPositionedTextArtifactsFromStreams(streams [][]byte, page model.PageMetadata, glyphMap map[string]string, fontAliases map[string]string, fontMetrics map[string]fontMetrics) []*model.RawArtifact {
	fragments := extractTextFragmentsFromStreamsWithFonts(streams, glyphMap, fontAliases, fontMetrics)
	if len(fragments) == 0 {
		return nil
	}

	pageBounds, pageBoundsOK := pageBoxFromMetadata(page)
	artifacts := make([]*model.RawArtifact, 0, len(fragments))
	for i, fragment := range fragments {
		text := strings.TrimSpace(fragment.text)
		if text == "" {
			continue
		}
		bounds := fragment.bounds
		if bounds.IsZero() || (pageBoundsOK && !bounds.Intersects(pageBounds)) {
			bounds = shellBoundsForText(page, i, text)
		}
		artifacts = append(artifacts, &model.RawArtifact{
			ID:              model.ArtifactID(i + 1),
			Kind:            model.ArtifactKindText,
			PageIndex:       page.Index,
			PageNumber:      page.Number,
			MarkedContentID: cloneIntPointer(fragment.markedContentID),
			Sequence:        i,
			Bounds:          bounds,
			Text:            text,
			Style: model.TextProperties{
				Font:     fragment.fontName,
				FontSize: fragment.fontSize,
				Content:  text,
			},
		})
	}
	if len(artifacts) == 0 {
		return nil
	}
	return artifacts
}

func extractTextShellStringsFromStreams(streams [][]byte, glyphMap map[string]string) []string {
	if len(streams) == 0 {
		return nil
	}
	out := make([]string, 0, len(streams))
	for _, stream := range streams {
		out = append(out, extractContentStreamStrings(stream, glyphMap)...)
	}
	return dedupeStrings(out)
}

func buildTextArtifactsFromStrings(page model.PageMetadata, texts []string) []*model.RawArtifact {
	if len(texts) == 0 {
		return nil
	}
	artifacts := make([]*model.RawArtifact, 0, len(texts))
	for i, text := range texts {
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		artifacts = append(artifacts, &model.RawArtifact{
			ID:         model.ArtifactID(i + 1),
			Kind:       model.ArtifactKindText,
			PageIndex:  page.Index,
			PageNumber: page.Number,
			Sequence:   i,
			Bounds:     shellBoundsForText(page, i, text),
			Text:       text,
			Style: model.TextProperties{
				Font:     "skeleton",
				FontSize: 12,
				Content:  text,
			},
		})
	}
	return artifacts
}
