package state

import (
	"path/filepath"
	"testing"
)

func TestOpenAddReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")

	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Has(42) {
		t.Fatal("UID should not exist initially")
	}
	if err := s.Add(42); err != nil {
		t.Fatal(err)
	}

	s2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if !s2.Has(42) {
		t.Fatal("UID was not persisted")
	}
}
