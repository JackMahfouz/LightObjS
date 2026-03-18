package distributed

import (
	"LightObjS/internal/domain"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// DiskStore implements domain.Storage using a flat file hierarchy on disk.
type DiskStore struct {
	Root string
}

// NewDiskStore creates a new DiskStore at the given root directory.
func NewDiskStore(root string) *DiskStore {
	return &DiskStore{Root: root}
}

// getPath returns the full path for the data file and metadata file based on the key.
func (s *DiskStore) getPath(key string) (string, string) {
	hash := sha256.Sum256([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	// Use first 4 characters for directory structure to distribute files
	dir := filepath.Join(s.Root, hashStr[0:2], hashStr[2:4])
	base := filepath.Join(dir, hashStr)

	return base + ".data", base + ".json"
}

func (s *DiskStore) Upload(ctx context.Context, obj *domain.Object, data []byte) error {
	dataPath, metaPath := s.getPath(obj.Key)

	if err := os.MkdirAll(filepath.Dir(dataPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write data
	if err := os.WriteFile(dataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	// Write metadata
	metaBytes, err := json.Marshal(obj)
	if err != nil {
		// Clean up data file
		os.Remove(dataPath)
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metaPath, metaBytes, 0644); err != nil {
		os.Remove(dataPath)
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

func (s *DiskStore) Download(ctx context.Context, key string) (*domain.Object, []byte, error) {
	dataPath, metaPath := s.getPath(key)

	// Read metadata
	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("object not found: %s", key)
		}
		return nil, nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var obj domain.Object
	if err := json.Unmarshal(metaBytes, &obj); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	// Read data
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read data: %w", err)
	}

	return &obj, data, nil
}

func (s *DiskStore) Delete(ctx context.Context, key string) error {
	dataPath, metaPath := s.getPath(key)

	// We attempt to delete both, even if one fails
	errData := os.Remove(dataPath)
	errMeta := os.Remove(metaPath)

	if errData != nil && !os.IsNotExist(errData) {
		return fmt.Errorf("failed to delete data: %w", errData)
	}
	if errMeta != nil && !os.IsNotExist(errMeta) {
		return fmt.Errorf("failed to delete metadata: %w", errMeta)
	}

	return nil
}
