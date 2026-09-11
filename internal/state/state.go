package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type State struct {
	mu        sync.Mutex
	Processed map[uint32]struct{} `json:"-"`
	path      string
}

type diskState struct {
	Processed []uint32 `json:"processed_uids"`
}

func Open(path string) (*State, error) {
	s := &State{Processed: make(map[uint32]struct{}), path: path}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read state: %w", err)
	}

	var d diskState
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("parse state: %w", err)
	}
	for _, uid := range d.Processed {
		s.Processed[uid] = struct{}{}
	}
	return s, nil
}

func (s *State) Has(uid uint32) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.Processed[uid]
	return ok
}

func (s *State) Add(uid uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Processed[uid] = struct{}{}
	return s.saveLocked()
}

func (s *State) saveLocked() error {
	uids := make([]uint32, 0, len(s.Processed))
	for uid := range s.Processed {
		uids = append(uids, uid)
	}

	data, err := json.MarshalIndent(diskState{Processed: uids}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".state-*.tmp")
	if err != nil {
		return fmt.Errorf("create state temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return fmt.Errorf("chmod state temp file: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write state temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync state temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close state temp file: %w", err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		return fmt.Errorf("replace state file: %w", err)
	}
	return nil
}
