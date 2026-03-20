package model

// NodeID identifies a semantic or layout node inside a document.
type NodeID int64

// ArtifactID identifies a raw artifact extracted from a page.
type ArtifactID int64

// PageIndex is a zero-based page index.
type PageIndex int

// PageNumber is a one-based page number.
type PageNumber int

// ElementType names a semantic or layout element.
type ElementType string

const (
	ElementTypeParagraph ElementType = "paragraph"
	ElementTypeHeading   ElementType = "heading"
	ElementTypeCaption   ElementType = "caption"
	ElementTypeTable     ElementType = "table"
	ElementTypeTextBlock ElementType = "text block"
	ElementTypeList      ElementType = "list"
	ElementTypeListItem  ElementType = "list item"
	ElementTypeImage     ElementType = "image"
	ElementTypeHeader    ElementType = "header"
	ElementTypeFooter    ElementType = "footer"
	ElementTypeTableRow  ElementType = "table row"
	ElementTypeTableCell ElementType = "table cell"
	ElementTypeUnknown   ElementType = ""
)

// Level captures a semantic or structural level label.
type Level string

// LinkField stores bidirectional and cross-node references used by processors.
type LinkField struct {
	ParentID        *NodeID
	PreviousID      *NodeID
	NextID          *NodeID
	LinkedContentID *NodeID
}

// IsZero reports whether no link fields are set.
func (l LinkField) IsZero() bool {
	return l.ParentID == nil && l.PreviousID == nil && l.NextID == nil && l.LinkedContentID == nil
}
