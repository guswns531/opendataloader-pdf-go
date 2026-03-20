package model

// ArtifactKind classifies a raw page artifact.
type ArtifactKind string

const (
	ArtifactKindUnknown ArtifactKind = ""
	ArtifactKindText    ArtifactKind = "text"
	ArtifactKindImage   ArtifactKind = "image"
	ArtifactKindLine    ArtifactKind = "line"
	ArtifactKindPath    ArtifactKind = "path"
	ArtifactKindShape   ArtifactKind = "shape"
)

// ImageFormat identifies an image encoding.
type ImageFormat string

const (
	ImageFormatPNG  ImageFormat = "png"
	ImageFormatJPEG ImageFormat = "jpeg"
)

// RawArtifact represents a low-level extracted page artifact.
type RawArtifact struct {
	ID         ArtifactID
	Kind       ArtifactKind
	PageIndex  PageIndex
	PageNumber PageNumber
	Bounds     Box
	Boxes      MultiBox
	Text       string
	Format     ImageFormat
	Data       []byte
	Style      TextProperties
	Sequence   int
	Links      LinkField
}
