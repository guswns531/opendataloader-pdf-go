package textclean

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

// Cleaner rewrites invalid or unprintable text using a configured replacement string.
type Cleaner struct {
	Replacement string
}

// New returns a new Cleaner.
func New(replacement string) Cleaner {
	return Cleaner{Replacement: replacement}
}

// String cleans a single string.
func (c Cleaner) String(value string) string {
	if value == "" {
		return value
	}

	original := value
	var b strings.Builder
	b.Grow(len(value))

	changed := false
	for len(value) > 0 {
		r, size := utf8.DecodeRuneInString(value)
		value = value[size:]

		if needsReplacement(r) {
			b.WriteString(c.Replacement)
			changed = true
			continue
		}

		b.WriteRune(r)
	}

	if !changed {
		return original
	}
	return b.String()
}

// TextProperties cleans the content string on a text property set.
func (c Cleaner) TextProperties(props *model.TextProperties) bool {
	if props == nil {
		return false
	}

	cleaned := c.String(props.Content)
	if cleaned == props.Content {
		return false
	}

	props.Content = cleaned
	return true
}

// RawArtifact cleans the text-bearing fields on a raw artifact.
func (c Cleaner) RawArtifact(artifact *model.RawArtifact) bool {
	if artifact == nil {
		return false
	}

	changed := false
	if cleaned := c.String(artifact.Text); cleaned != artifact.Text {
		artifact.Text = cleaned
		changed = true
	}
	if c.TextProperties(&artifact.Style) {
		changed = true
	}
	return changed
}

// Page cleans the text-bearing fields on a page.
func (c Cleaner) Page(page *model.Page) bool {
	if page == nil {
		return false
	}

	changed := false
	for _, artifact := range page.Artifacts {
		if c.RawArtifact(artifact) {
			changed = true
		}
	}
	if c.ContentElements(page.Kids) {
		changed = true
	}
	return changed
}

// Document cleans all text-bearing fields reachable from a document.
func (c Cleaner) Document(document *model.Document) bool {
	if document == nil {
		return false
	}

	changed := false
	for _, page := range document.Pages {
		if c.Page(page) {
			changed = true
		}
	}
	for _, artifact := range document.Artifacts {
		if c.RawArtifact(artifact) {
			changed = true
		}
	}
	if c.ContentElements(document.Kids) {
		changed = true
	}
	return changed
}

// ContentElements cleans a slice of content elements in place.
func (c Cleaner) ContentElements(elements []model.ContentElement) bool {
	changed := false
	for _, element := range elements {
		if c.ContentElement(element) {
			changed = true
		}
	}
	return changed
}

// ContentElement cleans a single content element and its descendants.
func (c Cleaner) ContentElement(element model.ContentElement) bool {
	switch node := element.(type) {
	case *model.Paragraph:
		if node == nil {
			return false
		}
		return c.TextNode(&node.TextNode)
	case *model.Heading:
		if node == nil {
			return false
		}
		return c.TextNode(&node.TextNode)
	case *model.Caption:
		if node == nil {
			return false
		}
		return c.TextNode(&node.TextNode)
	case *model.ListItem:
		if node == nil {
			return false
		}
		changed := c.TextNode(&node.TextNode)
		if c.ContentElements(node.Kids) {
			changed = true
		}
		return changed
	case *model.TextBlock:
		if node == nil {
			return false
		}
		return c.ContentElements(node.Kids)
	case *model.List:
		if node == nil {
			return false
		}
		changed := false
		for _, item := range node.ListItems {
			if c.ContentElement(item) {
				changed = true
			}
		}
		return changed
	case *model.HeaderFooter:
		if node == nil {
			return false
		}
		return c.ContentElements(node.Kids)
	case *model.Table:
		if node == nil {
			return false
		}
		changed := false
		for i := range node.Rows {
			if c.TableRow(&node.Rows[i]) {
				changed = true
			}
		}
		return changed
	case *model.TableCell:
		if node == nil {
			return false
		}
		return c.ContentElements(node.Kids)
	default:
		return false
	}
}

// TextNode cleans the shared text fields on a text-bearing node.
func (c Cleaner) TextNode(node *model.TextNode) bool {
	if node == nil {
		return false
	}
	return c.TextProperties(&node.TextProperties)
}

// TableRow cleans the cells reachable from a table row.
func (c Cleaner) TableRow(row *model.TableRow) bool {
	if row == nil {
		return false
	}

	changed := false
	for _, cell := range row.Cells {
		if c.ContentElement(cell) {
			changed = true
		}
	}
	return changed
}

func needsReplacement(r rune) bool {
	if r == '\n' || r == '\r' || r == '\t' {
		return false
	}
	return r == utf8.RuneError || !unicode.IsPrint(r)
}
