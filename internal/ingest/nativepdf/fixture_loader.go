package nativepdf

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

// FixtureLoader loads raw-artifact fixtures into the nativepdf interfaces.
type FixtureLoader struct{}

// NewFixtureLoader returns a fixture-backed native loader for tests and goldens.
func NewFixtureLoader() *FixtureLoader {
	return &FixtureLoader{}
}

// OpenPath reads a raw-artifact fixture from disk.
func (l *FixtureLoader) OpenPath(_ context.Context, path string, opts OpenOptions) (DocumentHandle, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read native fixture: %w", err)
	}
	return l.open(data, filepath.Base(path), opts)
}

// OpenReader reads a raw-artifact fixture from a stream.
func (l *FixtureLoader) OpenReader(_ context.Context, name string, r io.Reader, opts OpenOptions) (DocumentHandle, error) {
	if r == nil {
		return nil, fmt.Errorf("native fixture reader is nil")
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read native fixture: %w", err)
	}
	return l.open(data, name, opts)
}

func (l *FixtureLoader) open(data []byte, fallbackName string, opts OpenOptions) (DocumentHandle, error) {
	var fixture fixtureDocument
	if err := json.Unmarshal(data, &fixture); err != nil {
		return nil, fmt.Errorf("decode native fixture: %w", err)
	}

	handle, err := fixture.toHandle(fallbackName, opts)
	if err != nil {
		return nil, err
	}
	return handle, nil
}

type fixtureDocument struct {
	Metadata fixtureDocumentMetadata `json:"metadata"`
	Pages    []fixturePage           `json:"pages"`
}

func (f fixtureDocument) toHandle(fallbackName string, opts OpenOptions) (DocumentHandle, error) {
	metadata := f.Metadata.toModel()
	if metadata.FileName == "" {
		metadata.FileName = fallbackName
	}

	selected := make(map[int]struct{}, len(opts.Pages))
	for _, page := range opts.Pages {
		if page > 0 {
			selected[page] = struct{}{}
		}
	}

	pages := make([]*fixturePageHandle, 0, len(f.Pages))
	for idx, pageFixture := range f.Pages {
		page, err := pageFixture.toHandle(idx)
		if err != nil {
			return nil, err
		}
		if len(selected) > 0 {
			if _, ok := selected[int(page.metadata.Number)]; !ok {
				continue
			}
		}
		pages = append(pages, page)
	}

	if len(selected) > 0 {
		metadata.PageCount = len(pages)
	} else if metadata.PageCount == 0 {
		metadata.PageCount = len(pages)
	}
	if len(pages) != metadata.PageCount {
		return nil, fmt.Errorf("native fixture page_count=%d does not match pages=%d", metadata.PageCount, len(pages))
	}

	return &fixtureDocumentHandle{
		metadata: metadata,
		pages:    pages,
	}, nil
}

type fixtureDocumentMetadata struct {
	FileName         string   `json:"file_name,omitempty"`
	Author           *string  `json:"author,omitempty"`
	Title            *string  `json:"title,omitempty"`
	CreationDate     *string  `json:"created_at,omitempty"`
	ModificationDate *string  `json:"modified_at,omitempty"`
	Producer         *string  `json:"producer,omitempty"`
	Creator          *string  `json:"creator,omitempty"`
	Subject          *string  `json:"subject,omitempty"`
	Language         *string  `json:"language,omitempty"`
	Keywords         []string `json:"keywords,omitempty"`
	PageCount        int      `json:"page_count,omitempty"`
}

func (f fixtureDocumentMetadata) toModel() model.DocumentMetadata {
	return model.DocumentMetadata{
		FileName:         f.FileName,
		Author:           f.Author,
		Title:            f.Title,
		CreationDate:     f.CreationDate,
		ModificationDate: f.ModificationDate,
		Producer:         f.Producer,
		Creator:          f.Creator,
		Subject:          f.Subject,
		Language:         f.Language,
		Keywords:         append([]string(nil), f.Keywords...),
		PageCount:        f.PageCount,
	}
}

type fixturePage struct {
	Metadata        fixturePageMetadata          `json:"metadata"`
	Artifacts       []fixtureArtifact            `json:"artifacts"`
	TableCandidates *fixtureTableCandidateSet    `json:"table_candidates,omitempty"`
	StructTree      *fixtureStructNode           `json:"struct_tree,omitempty"`
}

func (f fixturePage) toHandle(position int) (*fixturePageHandle, error) {
	metadata := f.Metadata.toModel()
	if metadata.Number <= 0 {
		metadata.Number = model.PageNumber(position + 1)
	}
	if position > 0 && metadata.Index == 0 {
		metadata.Index = model.PageIndex(position)
	}

	artifacts := make([]*model.RawArtifact, 0, len(f.Artifacts))
	seen := make(map[int]struct{}, len(f.Artifacts))
	for pos, artifactFixture := range f.Artifacts {
		artifact, err := artifactFixture.toModel(metadata)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[artifact.Sequence]; exists {
			return nil, fmt.Errorf("native fixture page %d has duplicate artifact sequence %d", metadata.Number, artifact.Sequence)
		}
		seen[artifact.Sequence] = struct{}{}
		if artifact.Sequence != pos {
			return nil, fmt.Errorf("native fixture page %d sequence %d is not contiguous at position %d", metadata.Number, artifact.Sequence, pos)
		}
		artifacts = append(artifacts, artifact)
	}

	tableCandidates, err := f.TableCandidates.toModel(metadata.Index)
	if err != nil {
		return nil, err
	}

	return &fixturePageHandle{
		metadata:        metadata,
		artifacts:       artifacts,
		tableCandidates: tableCandidates,
		structTree:      f.StructTree.toModel(),
	}, nil
}

type fixturePageMetadata struct {
	Number   model.PageNumber `json:"number,omitempty"`
	Index    model.PageIndex  `json:"index,omitempty"`
	Label    string           `json:"label,omitempty"`
	Width    float64          `json:"width,omitempty"`
	Height   float64          `json:"height,omitempty"`
	Rotation int              `json:"rotation,omitempty"`
	CropBox  model.Box        `json:"crop_box,omitempty"`
	MediaBox model.Box        `json:"media_box,omitempty"`
}

func (f fixturePageMetadata) toModel() model.PageMetadata {
	size := model.PageSize{
		Width:  f.Width,
		Height: f.Height,
	}
	bounds := f.MediaBox
	if bounds.IsZero() && (f.Width > 0 || f.Height > 0) {
		bounds = model.Box{Left: 0, Bottom: 0, Right: f.Width, Top: f.Height}
	}
	return model.PageMetadata{
		Number:   f.Number,
		Index:    f.Index,
		Label:    f.Label,
		Size:     size,
		Bounds:   bounds,
		Rotation: f.Rotation,
	}
}

type fixtureArtifact struct {
	Kind       model.ArtifactKind   `json:"kind"`
	PageIndex  *model.PageIndex     `json:"page_index,omitempty"`
	PageNumber *model.PageNumber    `json:"page_number,omitempty"`
	Sequence   int                  `json:"sequence"`
	Bounds     model.Box            `json:"bounds,omitempty"`
	Boxes      model.MultiBox       `json:"boxes,omitempty"`
	Text       string               `json:"text,omitempty"`
	Data       []byte               `json:"data,omitempty"`
	Format     model.ImageFormat    `json:"format,omitempty"`
	Style      fixtureTextStyle     `json:"style,omitempty"`
	Links      fixtureLinkField     `json:"links,omitempty"`
}

func (f fixtureArtifact) toModel(page model.PageMetadata) (*model.RawArtifact, error) {
	if f.Kind == "" {
		return nil, fmt.Errorf("native fixture page %d has artifact with empty kind", page.Number)
	}
	if f.Kind == model.ArtifactKindText && f.Text == "" {
		return nil, fmt.Errorf("native fixture page %d has text artifact with empty text", page.Number)
	}

	pageIndex := page.Index
	if f.PageIndex != nil {
		pageIndex = *f.PageIndex
	}
	pageNumber := page.Number
	if f.PageNumber != nil {
		pageNumber = *f.PageNumber
	}

	artifact := &model.RawArtifact{
		Kind:       f.Kind,
		PageIndex:  pageIndex,
		PageNumber: pageNumber,
		Sequence:   f.Sequence,
		Bounds:     f.Bounds,
		Boxes:      append(model.MultiBox(nil), f.Boxes...),
		Text:       f.Text,
		Data:       append([]byte(nil), f.Data...),
		Format:     f.Format,
		Style:      f.Style.toModel(),
		Links:      f.Links.toModel(),
	}
	return artifact, nil
}

type fixtureTextStyle struct {
	Font       string  `json:"font,omitempty"`
	FontSize   float64 `json:"font_size,omitempty"`
	TextColor  string  `json:"text_color,omitempty"`
	Content    string  `json:"content,omitempty"`
	HiddenText bool    `json:"hidden_text,omitempty"`
	Bold       bool    `json:"bold,omitempty"`
	Italic     bool    `json:"italic,omitempty"`
	Underline  bool    `json:"underline,omitempty"`
}

func (f fixtureTextStyle) toModel() model.TextProperties {
	return model.TextProperties{
		Font:       f.Font,
		FontSize:   f.FontSize,
		TextColor:  f.TextColor,
		Content:    f.Content,
		HiddenText: f.HiddenText,
		Bold:       f.Bold,
		Italic:     f.Italic,
		Underline:  f.Underline,
	}
}

type fixtureLinkField struct {
	ParentID        *model.NodeID `json:"parent_id,omitempty"`
	PreviousID      *model.NodeID `json:"previous_id,omitempty"`
	NextID          *model.NodeID `json:"next_id,omitempty"`
	LinkedContentID *model.NodeID `json:"linked_content_id,omitempty"`
}

func (f fixtureLinkField) toModel() model.LinkField {
	return model.LinkField{
		ParentID:        f.ParentID,
		PreviousID:      f.PreviousID,
		NextID:          f.NextID,
		LinkedContentID: f.LinkedContentID,
	}
}

type fixtureTableCandidateSet struct {
	HorizontalLines []fixtureLineSegment `json:"horizontal_lines,omitempty"`
	VerticalLines   []fixtureLineSegment `json:"vertical_lines,omitempty"`
	Rectangles      []model.Box          `json:"rectangles,omitempty"`
}

func (f *fixtureTableCandidateSet) toModel(pageIndex model.PageIndex) (*TableCandidateSet, error) {
	if f == nil {
		return nil, nil
	}
	set := &TableCandidateSet{
		HorizontalLines: make([]LineSegment, 0, len(f.HorizontalLines)),
		VerticalLines:   make([]LineSegment, 0, len(f.VerticalLines)),
		Rectangles:      append([]model.Box(nil), f.Rectangles...),
	}
	for _, line := range f.HorizontalLines {
		set.HorizontalLines = append(set.HorizontalLines, line.toModel(pageIndex))
	}
	for _, line := range f.VerticalLines {
		set.VerticalLines = append(set.VerticalLines, line.toModel(pageIndex))
	}
	return set, nil
}

type fixtureLineSegment struct {
	Start     model.Point `json:"start"`
	End       model.Point `json:"end"`
	Width     float64     `json:"width,omitempty"`
	StrokeRGB string      `json:"stroke_rgb,omitempty"`
}

func (f fixtureLineSegment) toModel(pageIndex model.PageIndex) LineSegment {
	return LineSegment{
		Start:     f.Start,
		End:       f.End,
		Width:     f.Width,
		StrokeRGB: f.StrokeRGB,
		PageIndex: pageIndex,
	}
}

type fixtureStructNode struct {
	Type        string              `json:"type"`
	PageIndex   *model.PageIndex    `json:"page_index,omitempty"`
	Bounds      model.Box           `json:"bounds,omitempty"`
	Kids        []fixtureStructNode `json:"kids,omitempty"`
	ArtifactIDs []model.ArtifactID  `json:"artifact_ids,omitempty"`
}

func (f *fixtureStructNode) toModel() *StructNode {
	if f == nil {
		return nil
	}
	node := &StructNode{
		Type:        f.Type,
		PageIndex:   f.PageIndex,
		Bounds:      f.Bounds,
		Kids:        make([]*StructNode, 0, len(f.Kids)),
		ArtifactIDs: append([]model.ArtifactID(nil), f.ArtifactIDs...),
	}
	for _, kid := range f.Kids {
		node.Kids = append(node.Kids, kid.toModel())
	}
	return node
}

type fixtureDocumentHandle struct {
	metadata model.DocumentMetadata
	pages    []*fixturePageHandle
}

func (h *fixtureDocumentHandle) Metadata() model.DocumentMetadata {
	return h.metadata
}

func (h *fixtureDocumentHandle) PageCount() int {
	return len(h.pages)
}

func (h *fixtureDocumentHandle) Page(pageIndex int) (PageHandle, error) {
	if pageIndex < 0 || pageIndex >= len(h.pages) {
		return nil, fmt.Errorf("native fixture page %d out of range", pageIndex)
	}
	return h.pages[pageIndex], nil
}

func (h *fixtureDocumentHandle) Close() error {
	return nil
}

type fixturePageHandle struct {
	metadata        model.PageMetadata
	artifacts       []*model.RawArtifact
	tableCandidates *TableCandidateSet
	structTree      *StructNode
}

func (h *fixturePageHandle) Metadata() model.PageMetadata {
	return h.metadata
}

func (h *fixturePageHandle) Artifacts(_ context.Context, opts ArtifactOptions) ([]*model.RawArtifact, error) {
	if len(h.artifacts) == 0 {
		return nil, nil
	}
	if includeAllArtifacts(opts) {
		return cloneArtifacts(h.artifacts), nil
	}

	filtered := make([]*model.RawArtifact, 0, len(h.artifacts))
	for _, artifact := range h.artifacts {
		if artifact == nil {
			continue
		}
		if includeArtifactKind(artifact.Kind, opts) {
			filtered = append(filtered, cloneArtifact(artifact))
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		return filtered[i].Sequence < filtered[j].Sequence
	})
	return filtered, nil
}

func (h *fixturePageHandle) TableCandidates(_ context.Context) (*TableCandidateSet, error) {
	if h.tableCandidates == nil {
		return nil, nil
	}
	copied := *h.tableCandidates
	copied.HorizontalLines = append([]LineSegment(nil), h.tableCandidates.HorizontalLines...)
	copied.VerticalLines = append([]LineSegment(nil), h.tableCandidates.VerticalLines...)
	copied.Rectangles = append([]model.Box(nil), h.tableCandidates.Rectangles...)
	return &copied, nil
}

func (h *fixturePageHandle) StructTree(_ context.Context) (*StructNode, error) {
	return cloneStructNode(h.structTree), nil
}

func includeAllArtifacts(opts ArtifactOptions) bool {
	return !opts.IncludeText && !opts.IncludeImage && !opts.IncludeLine && !opts.IncludePath
}

func includeArtifactKind(kind model.ArtifactKind, opts ArtifactOptions) bool {
	switch kind {
	case model.ArtifactKindText:
		return opts.IncludeText
	case model.ArtifactKindImage:
		return opts.IncludeImage
	case model.ArtifactKindLine:
		return opts.IncludeLine
	case model.ArtifactKindPath, model.ArtifactKindShape:
		return opts.IncludePath
	default:
		return false
	}
}

func cloneArtifacts(artifacts []*model.RawArtifact) []*model.RawArtifact {
	out := make([]*model.RawArtifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		if artifact == nil {
			continue
		}
		out = append(out, cloneArtifact(artifact))
	}
	return out
}

func cloneStructNode(node *StructNode) *StructNode {
	if node == nil {
		return nil
	}
	cloned := &StructNode{
		Type:        node.Type,
		PageIndex:   node.PageIndex,
		Bounds:      node.Bounds,
		Kids:        make([]*StructNode, 0, len(node.Kids)),
		ArtifactIDs: append([]model.ArtifactID(nil), node.ArtifactIDs...),
	}
	for _, kid := range node.Kids {
		cloned.Kids = append(cloned.Kids, cloneStructNode(kid))
	}
	return cloned
}
