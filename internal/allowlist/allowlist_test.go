package allowlist

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndContains(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "senders.txt")
	err := os.WriteFile(path, []byte(`
# comment

Alice@Example.com
bob@example.org
`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	list, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if list.Len() != 2 {
		t.Fatalf("got %d addresses, want 2", list.Len())
	}
	if !list.Contains("alice@example.com") {
		t.Fatal("expected Alice to match case-insensitively")
	}
	if list.Contains("nobody@example.com") {
		t.Fatal("unexpected sender match")
	}
}

func TestRejectDisplayName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "senders.txt")
	if err := os.WriteFile(path, []byte(`Alice <alice@example.com>`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected display name to be rejected")
	}
}
