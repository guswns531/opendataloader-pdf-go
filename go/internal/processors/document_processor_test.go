package processors

import (
	"path/filepath"
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
)

func TestParsePageRangePreservesOrderAndDuplicates(t *testing.T) {
	got, explicit, err := parsePageRange("3,1,5-7,6", 10)
	if err != nil {
		t.Fatalf("parsePageRange returned error: %v", err)
	}
	if !explicit {
		t.Fatal("expected explicit selection")
	}
	want := []int{3, 1, 5, 6, 7, 6}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}

func TestParsePageRangeFiltersOutOfBoundsPages(t *testing.T) {
	got, explicit, err := parsePageRange("1,11", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !explicit {
		t.Fatal("expected explicit selection")
	}
	want := []int{1}
	if len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestParsePageRangeFiltersOutOfBoundsRanges(t *testing.T) {
	got, explicit, err := parsePageRange("2-12", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !explicit {
		t.Fatal("expected explicit selection")
	}
	want := []int{2, 3, 4, 5, 6, 7, 8, 9, 10}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}

func TestResolveImageDirUsesConfiguredDirectory(t *testing.T) {
	cfg := api.DefaultConfig()
	cfg.ImageDir = "/tmp/images"
	cfg.ImageOutput = api.ImageOutputExternal

	got := resolveImageDir("/tmp/sample.pdf", cfg)
	if got != "/tmp/images" {
		t.Fatalf("expected configured image dir, got %q", got)
	}
}

func TestResolveImageDirDefaultsToOutputFolder(t *testing.T) {
	cfg := api.DefaultConfig()
	cfg.OutputDir = "/tmp/out"
	cfg.ImageOutput = api.ImageOutputExternal

	got := resolveImageDir("/docs/sample.pdf", cfg)
	want := filepath.Join("/tmp/out", "sample_images")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
