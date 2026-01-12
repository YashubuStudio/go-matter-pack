package store

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Store defines a simple interface for persisting data.
type Store interface {
	Load(ctx context.Context, v any) error
	Save(ctx context.Context, v any) error
}

// JSONFileStore persists data to a JSON file.
type JSONFileStore struct {
	path string
}

// NewJSONFileStore returns a JSONFileStore pointing at the given path.
func NewJSONFileStore(path string) *JSONFileStore {
	return &JSONFileStore{path: path}
}

// Load populates v with data from the JSON file, if it exists.
func (s *JSONFileStore) Load(_ context.Context, v any) error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, v)
}

// Save writes v as JSON to the file path.
func (s *JSONFileStore) Save(_ context.Context, v any) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}
