package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverExpandsDirectoriesDeterministically(t *testing.T) {
	root := t.TempDir()
	explicit := filepath.Join(t.TempDir(), "explicit.PDF")

	mustWriteFile(t, filepath.Join(root, "zeta", "nested", "c.PDF"))
	mustWriteFile(t, filepath.Join(root, "alpha", "b.txt"))
	mustWriteFile(t, filepath.Join(root, "alpha", "a.JSON"))
	mustWriteFile(t, filepath.Join(root, "beta", "d.pdf"))
	mustWriteFile(t, filepath.Join(root, "beta", "e.md"))
	mustWriteFile(t, explicit)

	got, err := Discover([]string{
		explicit,
		root,
	})
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	want := []string{
		explicit,
		filepath.Join(root, "alpha", "a.JSON"),
		filepath.Join(root, "beta", "d.pdf"),
		filepath.Join(root, "zeta", "nested", "c.PDF"),
	}

	assertPaths(t, got, want)
}

func TestDiscoverSkipsUnsupportedFiles(t *testing.T) {
	root := t.TempDir()

	mustWriteFile(t, filepath.Join(root, "notes.txt"))
	mustWriteFile(t, filepath.Join(root, "report.docx"))
	mustWriteFile(t, filepath.Join(root, "fixture.json"))

	got, err := Discover([]string{
		filepath.Join(root, "notes.txt"),
		filepath.Join(root, "report.docx"),
		root,
	})
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	want := []string{
		filepath.Join(root, "fixture.json"),
	}

	assertPaths(t, got, want)
}

func TestDiscoverPreservesExplicitInputOrder(t *testing.T) {
	root := t.TempDir()

	first := filepath.Join(root, "first.pdf")
	second := filepath.Join(root, "second.json")
	mustWriteFile(t, first)
	mustWriteFile(t, second)
	mustWriteFile(t, filepath.Join(root, "folder", "third.pdf"))
	mustWriteFile(t, filepath.Join(root, "folder", "nested", "fourth.json"))

	got, err := Discover([]string{
		first,
		filepath.Join(root, "folder"),
		second,
	})
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	want := []string{
		first,
		filepath.Join(root, "folder", "nested", "fourth.json"),
		filepath.Join(root, "folder", "third.pdf"),
		second,
	}

	assertPaths(t, got, want)
}

func TestIsSupported(t *testing.T) {
	if !IsSupported("sample.PDF") {
		t.Fatalf("IsSupported(sample.PDF) = false, want true")
	}
	if !IsSupported("fixture.Json") {
		t.Fatalf("IsSupported(fixture.Json) = false, want true")
	}
	if IsSupported("notes.txt") {
		t.Fatalf("IsSupported(notes.txt) = true, want false")
	}
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func assertPaths(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d\n got: %v\nwant: %v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q\n got: %v\nwant: %v", i, got[i], want[i], got, want)
		}
	}
}
