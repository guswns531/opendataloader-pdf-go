package nativepdf

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
)

func TestUnavailableLoaderReturnsBackendUnavailable(t *testing.T) {
	loader := NewUnavailableLoader()

	_, err := loader.OpenPath(context.Background(), "sample.pdf", OpenOptions{})
	if !errors.Is(err, ErrBackendUnavailable) {
		t.Fatalf("OpenPath() error = %v, want ErrBackendUnavailable", err)
	}
}

func TestUnavailableLoaderWorksThroughNativeIngestor(t *testing.T) {
	ingestor := NewIngestor(NewUnavailableLoader())

	_, err := ingestor.Ingest(nil, core.Source{
		Name:   "sample.pdf",
		Reader: strings.NewReader("%PDF"),
	})
	if !errors.Is(err, ErrBackendUnavailable) {
		t.Fatalf("Ingest() error = %v, want ErrBackendUnavailable", err)
	}
}
