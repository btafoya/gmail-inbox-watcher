package envfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDoesNotOverrideExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("TEST_GIW=fromfile\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_GIW", "existing")
	if err := Load(path); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("TEST_GIW"); got != "existing" {
		t.Fatalf("got %q", got)
	}
}
