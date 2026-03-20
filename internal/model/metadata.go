package model

// DocumentMetadata stores document-level metadata and provenance.
type DocumentMetadata struct {
	FileName         string
	Author           *string
	Title            *string
	CreationDate     *string
	ModificationDate *string
	Producer         *string
	Creator          *string
	Subject          *string
	Language         *string
	Keywords         []string
	PageCount        int
}

// PageMetadata stores page-level metadata and geometry.
type PageMetadata struct {
	Index    PageIndex
	Number   PageNumber
	Label    string
	Size     PageSize
	Bounds   Box
	Rotation int
}

// TextProperties stores common text styling and content fields.
type TextProperties struct {
	Font       string
	FontSize   float64
	TextColor  string
	Content    string
	HiddenText bool
	Bold       bool
	Italic     bool
	Underline  bool
}

// HasContent reports whether the text property set contains text.
func (t TextProperties) HasContent() bool {
	return t.Content != ""
}
