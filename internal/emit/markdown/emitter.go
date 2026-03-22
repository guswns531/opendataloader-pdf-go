package markdown

import (
	"fmt"
	"io"
	"strings"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

// Emitter renders a document graph as deterministic Markdown.
type Emitter struct{}

// New returns a Markdown emitter.
func New() *Emitter {
	return &Emitter{}
}

// Name returns the emitter name.
func (e *Emitter) Name() string {
	return "markdown"
}

// Format returns the emitter output format.
func (e *Emitter) Format() core.OutputFormat {
	return core.OutputFormatMarkdown
}

// Emit writes a deterministic Markdown view of the document graph.
func (e *Emitter) Emit(_ *core.ProcessingContext, document *core.Document, w io.Writer) error {
	if document == nil {
		return fmt.Errorf("markdown emitter: nil document")
	}
	if w == nil {
		return fmt.Errorf("markdown emitter: nil writer")
	}

	var b strings.Builder
	writeLine(&b, 0, "# Document")
	writeDocumentMetadata(&b, document.Metadata)

	if len(document.Kids) > 0 {
		writeLine(&b, 0, "## Document Content")
		for _, kid := range document.Kids {
			writeContent(&b, kid, 0)
		}
	}

	if len(document.Pages) > 0 {
		writeLine(&b, 0, "## Pages")
		for _, page := range document.Pages {
			writePage(&b, page)
		}
	}

	_, err := io.WriteString(w, b.String())
	return err
}

func writeDocumentMetadata(b *strings.Builder, metadata model.DocumentMetadata) {
	writeField(b, 0, "File", metadata.FileName)
	writeIntLine(b, 0, "Page count", metadata.PageCount)
	writeStringField(b, 0, "Author", metadata.Author)
	writeStringField(b, 0, "Title", metadata.Title)
	writeStringField(b, 0, "Creation date", metadata.CreationDate)
	writeStringField(b, 0, "Modification date", metadata.ModificationDate)
	writeStringField(b, 0, "Producer", metadata.Producer)
	writeStringField(b, 0, "Creator", metadata.Creator)
	writeStringField(b, 0, "Subject", metadata.Subject)
	writeStringField(b, 0, "Language", metadata.Language)
	if len(metadata.Keywords) > 0 {
		writeLine(b, 0, "- Keywords: %s", strings.Join(metadata.Keywords, ", "))
	}
}

func writePage(b *strings.Builder, page *model.Page) {
	if page == nil {
		return
	}
	writeLine(b, 0, "### Page %d", page.Metadata.Number)
	writeIntLine(b, 1, "Index", int(page.Metadata.Index))
	writeIntLine(b, 1, "Number", int(page.Metadata.Number))
	writeField(b, 1, "Label", page.Metadata.Label)
	if page.Metadata.Size.Width != 0 || page.Metadata.Size.Height != 0 {
		writeLine(b, 1, "- Size: %s", formatSize(page.Metadata.Size))
	}
	if !page.Metadata.Bounds.IsZero() {
		writeLine(b, 1, "- Bounds: %s", formatBox(page.Metadata.Bounds))
	}
	writeIntLine(b, 1, "Rotation", page.Metadata.Rotation)

	if len(page.Artifacts) > 0 {
		writeLine(b, 1, "#### Artifacts")
		for _, artifact := range page.Artifacts {
			writeArtifact(b, artifact)
		}
	}

	if len(page.Kids) > 0 {
		writeLine(b, 1, "#### Content")
		for _, kid := range page.Kids {
			writeContent(b, kid, 1)
		}
	}
}

func writeArtifact(b *strings.Builder, artifact *model.RawArtifact) {
	if artifact == nil {
		return
	}
	title := "Artifact"
	if artifact.ID != 0 {
		title = fmt.Sprintf("Artifact %d", artifact.ID)
	}
	if artifact.Kind != "" {
		title += fmt.Sprintf(" (%s)", artifact.Kind)
	}
	writeLine(b, 1, "- %s", title)
	writeIntLine(b, 2, "ID", int(artifact.ID))
	writeIntLine(b, 2, "Page index", int(artifact.PageIndex))
	writeIntLine(b, 2, "Page number", int(artifact.PageNumber))
	if !artifact.Bounds.IsZero() {
		writeLine(b, 2, "- Bounds: %s", formatBox(artifact.Bounds))
	}
	if len(artifact.Boxes) > 0 {
		writeLine(b, 2, "- Boxes: %s", formatMultiBox(artifact.Boxes))
	}
	writeField(b, 2, "Text", artifact.Text)
	writeField(b, 2, "Format", string(artifact.Format))
	if len(artifact.Data) > 0 {
		writeLine(b, 2, "- Data bytes: %d", len(artifact.Data))
	}
	writeField(b, 2, "ColorSpace", artifact.ColorSpace)
	if artifact.BitsPerComponent > 0 {
		writeLine(b, 2, "- BitsPerComponent: %d", artifact.BitsPerComponent)
	}
	if len(artifact.Filters) > 0 {
		writeField(b, 2, "Filters", strings.Join(artifact.Filters, ","))
	}
	writeTextProperties(b, 2, artifact.Style)
	writeLinkField(b, 2, artifact.Links)
	writeIntLine(b, 2, "Sequence", artifact.Sequence)
}

func writeContent(b *strings.Builder, element model.ContentElement, depth int) {
	if element == nil {
		return
	}

	base := element.NodeBase()
	label := contentLabel(element)
	if base != nil && base.ID != 0 {
		label = fmt.Sprintf("%s #%d", label, base.ID)
	}
	writeLine(b, depth, "- %s", label)

	if base == nil {
		return
	}
	writeIntLine(b, depth+1, "ID", int(base.ID))
	writeField(b, depth+1, "Type", string(base.Type))
	if base.Level != "" {
		writeLine(b, depth+1, "- Level: %s", base.Level)
	}
	writeIntLine(b, depth+1, "Index", base.Index)
	writeIntLine(b, depth+1, "Page index", int(base.PageIndex))
	writeIntLine(b, depth+1, "Page number", int(base.PageNumber))
	if !base.Bounds.IsZero() {
		writeLine(b, depth+1, "- Bounds: %s", formatBox(base.Bounds))
	}
	if len(base.Boxes) > 0 {
		writeLine(b, depth+1, "- Boxes: %s", formatMultiBox(base.Boxes))
	}
	writeLinkField(b, depth+1, base.Links)

	switch node := element.(type) {
	case *model.Paragraph:
		writeTextProperties(b, depth+1, node.TextProperties)
	case *model.Heading:
		writeTextProperties(b, depth+1, node.TextProperties)
		writeIntLine(b, depth+1, "Heading level", node.HeadingLevel)
	case *model.Caption:
		writeTextProperties(b, depth+1, node.TextProperties)
	case *model.TextBlock:
		for _, kid := range node.Kids {
			writeContent(b, kid, depth+1)
		}
	case *model.List:
		writeField(b, depth+1, "Numbering style", node.NumberingStyle)
		writeIntLine(b, depth+1, "Number of items", node.NumberOfItems)
		for i, item := range node.ListItems {
			writeListItem(b, item, depth+1, i+1)
		}
	case *model.ListItem:
		writeTextProperties(b, depth+1, node.TextProperties)
		for _, kid := range node.Kids {
			writeContent(b, kid, depth+1)
		}
	case *model.Image:
		writeField(b, depth+1, "Source", node.Source)
		writeField(b, depth+1, "Format", string(node.Format))
		if len(node.Data) > 0 {
			writeLine(b, depth+1, "- Data bytes: %d", len(node.Data))
		}
	case *model.HeaderFooter:
		for _, kid := range node.Kids {
			writeContent(b, kid, depth+1)
		}
	case *model.Table:
		writeIntLine(b, depth+1, "Rows", node.NumberOfRows)
		writeIntLine(b, depth+1, "Columns", node.NumberOfColumns)
		if node.PreviousTableID != nil {
			writeLine(b, depth+1, "- Previous table ID: %d", *node.PreviousTableID)
		}
		if node.NextTableID != nil {
			writeLine(b, depth+1, "- Next table ID: %d", *node.NextTableID)
		}
		for _, row := range node.Rows {
			writeTableRow(b, row, depth+1)
		}
	case *model.TableCell:
		writeIntLine(b, depth+1, "Row number", node.RowNumber)
		writeIntLine(b, depth+1, "Column number", node.ColumnNumber)
		writeIntLine(b, depth+1, "Row span", node.RowSpan)
		writeIntLine(b, depth+1, "Column span", node.ColumnSpan)
		for _, kid := range node.Kids {
			writeContent(b, kid, depth+1)
		}
	}
}

func writeListItem(b *strings.Builder, item *model.ListItem, depth, index int) {
	if item == nil {
		return
	}
	label := fmt.Sprintf("List item %d", index)
	if item.ID != 0 {
		label = fmt.Sprintf("%s #%d", label, item.ID)
	}
	writeLine(b, depth, "- %s", label)
	writeTextProperties(b, depth+1, item.TextProperties)
	for _, kid := range item.Kids {
		writeContent(b, kid, depth+1)
	}
}

func writeTableRow(b *strings.Builder, row model.TableRow, depth int) {
	writeLine(b, depth, "- Row %d", row.RowNumber)
	writeField(b, depth+1, "Type", string(row.Type))
	for _, cell := range row.Cells {
		writeContent(b, cell, depth+1)
	}
}

func writeTextProperties(b *strings.Builder, depth int, props model.TextProperties) {
	writeField(b, depth, "Content", props.Content)
	writeField(b, depth, "Font", props.Font)
	if props.FontSize != 0 {
		writeLine(b, depth, "- Font size: %s", trimFloat(props.FontSize))
	}
	writeField(b, depth, "Text color", props.TextColor)
	if props.Bold {
		writeLine(b, depth, "- Bold: true")
	}
	if props.Italic {
		writeLine(b, depth, "- Italic: true")
	}
	if props.Underline {
		writeLine(b, depth, "- Underline: true")
	}
	if props.HiddenText {
		writeLine(b, depth, "- Hidden text: true")
	}
}

func writeLinkField(b *strings.Builder, depth int, links model.LinkField) {
	if links.IsZero() {
		return
	}
	writeLine(b, depth, "- Links")
	writeNodeLinkField(b, depth+1, "Parent", links.ParentID)
	writeNodeLinkField(b, depth+1, "Previous", links.PreviousID)
	writeNodeLinkField(b, depth+1, "Next", links.NextID)
	writeNodeLinkField(b, depth+1, "Linked content", links.LinkedContentID)
}

func writeNodeLinkField(b *strings.Builder, depth int, name string, id *model.NodeID) {
	if id == nil {
		return
	}
	writeLine(b, depth, "- %s: %d", name, *id)
}

func writeField(b *strings.Builder, depth int, name, value string) {
	if value == "" {
		return
	}
	writeLine(b, depth, "- %s: %s", name, value)
}

func writeStringField(b *strings.Builder, depth int, name string, value *string) {
	if value == nil || *value == "" {
		return
	}
	writeLine(b, depth, "- %s: %s", name, *value)
}

func writeIntLine(b *strings.Builder, depth int, name string, value int) {
	writeLine(b, depth, "- %s: %d", name, value)
}

func contentLabel(element model.ContentElement) string {
	switch element.(type) {
	case *model.Paragraph:
		return "Paragraph"
	case *model.Heading:
		return "Heading"
	case *model.Caption:
		return "Caption"
	case *model.TextBlock:
		return "Text block"
	case *model.List:
		return "List"
	case *model.ListItem:
		return "List item"
	case *model.Image:
		return "Image"
	case *model.HeaderFooter:
		return "Header/footer"
	case *model.Table:
		return "Table"
	case *model.TableCell:
		return "Table cell"
	default:
		if element != nil && element.ContentType() != "" {
			return string(element.ContentType())
		}
		return "Content"
	}
}

func writeLine(b *strings.Builder, depth int, format string, args ...any) {
	for i := 0; i < depth; i++ {
		b.WriteString("  ")
	}
	fmt.Fprintf(b, format+"\n", args...)
}

func formatBox(box model.Box) string {
	return fmt.Sprintf("%.2f, %.2f, %.2f, %.2f", box.Left, box.Bottom, box.Right, box.Top)
}

func formatSize(size model.PageSize) string {
	return fmt.Sprintf("%.2f x %.2f", size.Width, size.Height)
}

func formatMultiBox(boxes model.MultiBox) string {
	parts := make([]string, 0, len(boxes))
	for _, box := range boxes {
		parts = append(parts, formatBox(box))
	}
	return strings.Join(parts, "; ")
}

func trimFloat(value float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
}
