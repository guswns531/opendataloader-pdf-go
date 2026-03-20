package model

// Document is the top-level in-memory representation for a converted PDF.
type Document struct {
	Metadata       DocumentMetadata
	Pages          []*Page
	Kids           []ContentElement
	Artifacts      []*RawArtifact
	nextNodeID     NodeID
	nextArtifactID ArtifactID
}

// Page groups raw artifacts and processed content for a single page.
type Page struct {
	Metadata  PageMetadata
	Artifacts []*RawArtifact
	Kids      []ContentElement
}

// NewDocument creates a document skeleton with stable ID counters.
func NewDocument(metadata DocumentMetadata) *Document {
	return &Document{
		Metadata:       metadata,
		Pages:          make([]*Page, 0),
		Kids:           make([]ContentElement, 0),
		Artifacts:      make([]*RawArtifact, 0),
		nextNodeID:     1,
		nextArtifactID: 1,
	}
}

// NewNodeID allocates a new document-local node identifier.
func (d *Document) NewNodeID() NodeID {
	if d == nil {
		return 0
	}
	id := d.nextNodeID
	d.nextNodeID++
	return id
}

// NewArtifactID allocates a new document-local artifact identifier.
func (d *Document) NewArtifactID() ArtifactID {
	if d == nil {
		return 0
	}
	id := d.nextArtifactID
	d.nextArtifactID++
	return id
}

// AddPage appends a page to the document.
func (d *Document) AddPage(page *Page) *Page {
	if d == nil || page == nil {
		return page
	}
	if page.Metadata.Index < 0 {
		page.Metadata.Index = PageIndex(len(d.Pages))
	}
	if page.Metadata.Number <= 0 {
		page.Metadata.Number = PageNumber(len(d.Pages) + 1)
	}
	d.Pages = append(d.Pages, page)
	return page
}

// PageByNumber looks up a page by its one-based page number.
func (d *Document) PageByNumber(number PageNumber) (*Page, bool) {
	if d == nil {
		return nil, false
	}
	for _, page := range d.Pages {
		if page != nil && page.Metadata.Number == number {
			return page, true
		}
	}
	return nil, false
}

// PageByIndex looks up a page by its zero-based page index.
func (d *Document) PageByIndex(index PageIndex) (*Page, bool) {
	if d == nil {
		return nil, false
	}
	for _, page := range d.Pages {
		if page != nil && page.Metadata.Index == index {
			return page, true
		}
	}
	return nil, false
}

// EnsurePage returns an existing page or creates a new placeholder page.
func (d *Document) EnsurePage(number PageNumber) *Page {
	if d == nil {
		return nil
	}
	if page, ok := d.PageByNumber(number); ok {
		return page
	}
	page := &Page{
		Metadata: PageMetadata{
			Index:  PageIndex(len(d.Pages)),
			Number: number,
		},
		Artifacts: make([]*RawArtifact, 0),
		Kids:      make([]ContentElement, 0),
	}
	d.Pages = append(d.Pages, page)
	if int(d.Metadata.PageCount) < len(d.Pages) {
		d.Metadata.PageCount = len(d.Pages)
	}
	return page
}
