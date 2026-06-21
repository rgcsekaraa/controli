package controli

import (
	"path/filepath"
	"testing"
)

func TestResolveDownloadPathStaysInsideRoot(t *testing.T) {
	root := t.TempDir()
	got, err := resolveDownloadPath(root, filepath.Join("reports", "out.txt"))
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "reports", "out.txt")
	if got != want {
		t.Fatalf("resolveDownloadPath() = %q, want %q", got, want)
	}
}

func TestResolveDownloadPathRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{"..", "../secret.txt", filepath.Join("reports", "..", "..", "secret.txt")} {
		if got, err := resolveDownloadPath(root, path); err == nil {
			t.Fatalf("resolveDownloadPath(%q) = %q, want error", path, got)
		}
	}
}

func TestResolveDownloadPathRejectsAbsolutePath(t *testing.T) {
	root := t.TempDir()
	absolute := filepath.Join(root, "file.txt")
	if got, err := resolveDownloadPath(root, absolute); err == nil {
		t.Fatalf("resolveDownloadPath(%q) = %q, want error", absolute, got)
	}
}
