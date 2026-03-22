package nativepdf

import "github.com/guswns531/opendataloader-pdf-go/internal/model"

const (
	maxStructTreeDepth = 32
	maxStructTreeNodes = 512
)

type structTreeParser struct {
	graph      *objectGraph
	pageByRef  map[pdfRef]model.PageIndex
	visited    map[pdfRef]struct{}
	nodeBudget int
}

func resolveStructTreesByPage(graph *objectGraph, plans []pagePlan) []*StructNode {
	if graph == nil || len(plans) == 0 {
		return nil
	}
	rootRef, ok := graph.catalogStructTreeRootRef()
	if !ok {
		return nil
	}
	rootDict, ok := resolveDictValue(graph, rootRef)
	if !ok {
		return nil
	}

	pageByRef := make(map[pdfRef]model.PageIndex, len(plans))
	for _, plan := range plans {
		pageByRef[plan.PageRef] = plan.Metadata.Index
	}

	parser := &structTreeParser{
		graph:      graph,
		pageByRef:  pageByRef,
		visited:    make(map[pdfRef]struct{}),
		nodeBudget: maxStructTreeNodes,
	}
	root := parser.parseStructElem(rootRef, rootDict, nil, 0)
	if root == nil {
		return nil
	}

	out := make([]*StructNode, len(plans))
	for i, plan := range plans {
		filtered := filterStructNodeForPage(root, plan.Metadata.Index)
		if filtered != nil {
			out[i] = filtered
		}
	}
	return out
}

func (g *objectGraph) catalogStructTreeRootRef() (pdfRef, bool) {
	if g == nil {
		return pdfRef{}, false
	}
	if g.hasRoot {
		if object := g.objects[g.rootRef]; object != nil {
			if dict, ok := object.Value.(pdfDict); ok {
				if ref, ok := dict["StructTreeRoot"].(pdfRef); ok {
					return ref, true
				}
			}
		}
	}
	for _, object := range g.objects {
		dict, ok := object.Value.(pdfDict)
		if !ok || nameValue(dict["Type"]) != "Catalog" {
			continue
		}
		if ref, ok := dict["StructTreeRoot"].(pdfRef); ok {
			return ref, true
		}
	}
	return pdfRef{}, false
}

func (p *structTreeParser) parseStructElem(ref pdfRef, dict pdfDict, inheritedPage *model.PageIndex, depth int) *StructNode {
	if p == nil || len(dict) == 0 || depth > maxStructTreeDepth || p.nodeBudget <= 0 {
		return nil
	}
	if ref != (pdfRef{}) {
		if _, seen := p.visited[ref]; seen {
			return nil
		}
		p.visited[ref] = struct{}{}
		defer delete(p.visited, ref)
	}

	pageIndex := p.pageIndexForDict(dict, inheritedPage)
	nodeType := normalizedPDFName(dict["S"])
	if nodeType == "" {
		nodeType = normalizedPDFName(dict["Type"])
	}
	if nodeType == "" {
		nodeType = "StructElem"
	}

	p.nodeBudget--
	node := &StructNode{
		Type:             nodeType,
		PageIndex:        pageIndex,
		Bounds:           structBoundsFromDict(dict),
		Kids:             make([]*StructNode, 0),
		ArtifactIDs:      nil,
		MarkedContentIDs: nil,
	}

	for _, child := range p.parseStructKids(dict["K"], pageIndex, depth+1) {
		if child == nil {
			continue
		}
		node.Kids = append(node.Kids, child)
	}

	if len(node.Kids) == 0 && pageIndex == nil && nodeType == "StructTreeRoot" {
		return nil
	}
	return node
}

func (p *structTreeParser) parseStructKids(value pdfValue, inheritedPage *model.PageIndex, depth int) []*StructNode {
	if depth > maxStructTreeDepth || p.nodeBudget <= 0 {
		return nil
	}
	switch typed := value.(type) {
	case pdfArray:
		out := make([]*StructNode, 0, len(typed))
		for _, item := range typed {
			out = append(out, p.parseStructKids(item, inheritedPage, depth)...)
		}
		return out
	case pdfRef:
		object := p.graph.objects[typed]
		if object == nil {
			return nil
		}
		if dict, ok := object.Value.(pdfDict); ok {
			switch normalizedPDFName(dict["Type"]) {
			case "MCR", "OBJR":
				if node := p.parseMarkedContentNode(dict, inheritedPage); node != nil {
					return []*StructNode{node}
				}
				return nil
			}
			if normalizedPDFName(dict["S"]) != "" || normalizedPDFName(dict["Type"]) == "StructElem" || normalizedPDFName(dict["Type"]) == "StructTreeRoot" {
				if node := p.parseStructElem(typed, dict, inheritedPage, depth); node != nil {
					return []*StructNode{node}
				}
				return nil
			}
		}
		return p.parseStructKids(object.Value, inheritedPage, depth)
	case pdfDict:
		switch normalizedPDFName(typed["Type"]) {
		case "MCR", "OBJR":
			if node := p.parseMarkedContentNode(typed, inheritedPage); node != nil {
				return []*StructNode{node}
			}
			return nil
		}
		if normalizedPDFName(typed["S"]) != "" || normalizedPDFName(typed["Type"]) == "StructElem" {
			if node := p.parseStructElem(pdfRef{}, typed, inheritedPage, depth); node != nil {
				return []*StructNode{node}
			}
		}
		return nil
	case int:
		if node := markedContentIDNode(typed, inheritedPage); node != nil {
			return []*StructNode{node}
		}
		return nil
	case float64:
		if node := markedContentIDNode(int(typed), inheritedPage); node != nil {
			return []*StructNode{node}
		}
		return nil
	default:
		return nil
	}
}

func (p *structTreeParser) parseMarkedContentNode(dict pdfDict, inheritedPage *model.PageIndex) *StructNode {
	pageIndex := p.pageIndexForDict(dict, inheritedPage)
	node := &StructNode{
		Type:             normalizedPDFName(dict["Type"]),
		PageIndex:        pageIndex,
		Bounds:           structBoundsFromDict(dict),
		MarkedContentIDs: nil,
	}
	if node.Type == "" {
		node.Type = "MCR"
	}
	if mcid, ok := intValue(dict["MCID"]); ok {
		node.MarkedContentIDs = []int{mcid}
	}
	if len(node.MarkedContentIDs) == 0 && node.PageIndex == nil {
		return nil
	}
	p.nodeBudget--
	return node
}

func markedContentIDNode(mcid int, pageIndex *model.PageIndex) *StructNode {
	if mcid < 0 {
		return nil
	}
	return &StructNode{
		Type:             "MCR",
		PageIndex:        pageIndex,
		MarkedContentIDs: []int{mcid},
	}
}

func (p *structTreeParser) pageIndexForDict(dict pdfDict, inheritedPage *model.PageIndex) *model.PageIndex {
	if ref, ok := dict["Pg"].(pdfRef); ok {
		if index, ok := p.pageByRef[ref]; ok {
			indexCopy := index
			return &indexCopy
		}
	}
	if inheritedPage == nil {
		return nil
	}
	indexCopy := *inheritedPage
	return &indexCopy
}

func structBoundsFromDict(dict pdfDict) model.Box {
	bounds, ok := arrayToBox(dict["BBox"])
	if !ok {
		return model.Box{}
	}
	return bounds.Normalize()
}

func filterStructNodeForPage(node *StructNode, pageIndex model.PageIndex) *StructNode {
	if node == nil {
		return nil
	}
	cloned := &StructNode{
		Type:             node.Type,
		PageIndex:        nil,
		Bounds:           node.Bounds,
		Kids:             make([]*StructNode, 0, len(node.Kids)),
		ArtifactIDs:      append([]model.ArtifactID(nil), node.ArtifactIDs...),
		MarkedContentIDs: append([]int(nil), node.MarkedContentIDs...),
	}
	if node.PageIndex != nil && *node.PageIndex == pageIndex {
		indexCopy := *node.PageIndex
		cloned.PageIndex = &indexCopy
	}
	for _, kid := range node.Kids {
		filtered := filterStructNodeForPage(kid, pageIndex)
		if filtered != nil {
			cloned.Kids = append(cloned.Kids, filtered)
		}
	}
	if cloned.PageIndex == nil && len(cloned.Kids) == 0 && len(cloned.MarkedContentIDs) == 0 && cloned.Type != "StructTreeRoot" {
		return nil
	}
	return cloned
}
