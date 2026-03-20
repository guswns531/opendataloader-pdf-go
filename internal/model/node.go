package model

// ContentElement is implemented by all semantic and layout content nodes.
type ContentElement interface {
	ContentType() ElementType
	NodeBase() *BaseNode
}

// BaseNode carries fields shared by semantic and layout elements.
type BaseNode struct {
	ID         NodeID
	Type       ElementType
	Level      Level
	Index      int
	PageIndex  PageIndex
	PageNumber PageNumber
	Bounds     Box
	Boxes      MultiBox
	Links      LinkField
}

// ContentType returns the node type.
func (n *BaseNode) ContentType() ElementType {
	if n == nil {
		return ElementTypeUnknown
	}
	return n.Type
}

// NodeBase returns the receiver for interface access.
func (n *BaseNode) NodeBase() *BaseNode {
	return n
}

// TextNode is the shared base for text-bearing content elements.
type TextNode struct {
	BaseNode
	TextProperties
}

// Paragraph represents paragraph content.
type Paragraph struct {
	TextNode
}

// Heading represents a heading with an explicit heading level.
type Heading struct {
	TextNode
	HeadingLevel int
}

// Caption represents a caption linked to another content element.
type Caption struct {
	TextNode
}

// TextBlock groups multiple content elements into a text block.
type TextBlock struct {
	BaseNode
	Kids []ContentElement
}

// List represents a list container.
type List struct {
	BaseNode
	NumberingStyle string
	NumberOfItems  int
	ListItems      []*ListItem
}

// ListItem represents a single list item.
type ListItem struct {
	TextNode
	Kids []ContentElement
}

// Image represents an extracted image node.
type Image struct {
	BaseNode
	Source string
	Data   string
	Format ImageFormat
}

// HeaderFooter represents a header or footer node.
type HeaderFooter struct {
	BaseNode
	Kids []ContentElement
}

// Table represents a table container.
type Table struct {
	BaseNode
	NumberOfRows    int
	NumberOfColumns int
	PreviousTableID *NodeID
	NextTableID     *NodeID
	Rows            []TableRow
}

// TableRow represents a table row.
type TableRow struct {
	Type      ElementType
	RowNumber int
	Cells     []*TableCell
}

// TableCell represents a table cell and can contain nested content.
type TableCell struct {
	BaseNode
	RowNumber    int
	ColumnNumber int
	RowSpan      int
	ColumnSpan   int
	Kids         []ContentElement
}
