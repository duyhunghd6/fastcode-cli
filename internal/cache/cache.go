package cache

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"

	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

func init() {
	gob.Register([]types.ImportInfo{})
	gob.Register([]types.FunctionInfo{})
	gob.Register([]types.ClassInfo{})
	gob.Register(map[string]any{})
}

// IndexCache handles persisting and loading index data to/from disk.
type IndexCache struct {
	CacheDir string
}

// NewIndexCache creates a new cache manager.
func NewIndexCache(cacheDir string) *IndexCache {
	return &IndexCache{CacheDir: cacheDir}
}

// CachedIndex represents the serializable index data.
type CachedIndex struct {
	RepoName string
	Elements []types.CodeElement
	Vectors  map[string][]float32 // elementID → embedding
}

// Save writes the index data to disk.
func (c *IndexCache) Save(repoName string, data *CachedIndex) error {
	if err := os.MkdirAll(c.CacheDir, 0755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	path := c.cachePath(repoName)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create cache file: %w", err)
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("encode cache: %w", err)
	}

	return nil
}

// Load reads index data from disk.
func (c *IndexCache) Load(repoName string) (*CachedIndex, error) {
	path := c.cachePath(repoName)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open cache file: %w", err)
	}
	defer f.Close()

	var data CachedIndex
	dec := gob.NewDecoder(f)
	if err := dec.Decode(&data); err != nil {
		return nil, fmt.Errorf("decode cache: %w", err)
	}

	return &data, nil
}

// Exists returns true if a cache file exists for the repo.
func (c *IndexCache) Exists(repoName string) bool {
	_, err := os.Stat(c.cachePath(repoName))
	return err == nil
}

// Delete removes the cache file for a repo.
func (c *IndexCache) Delete(repoName string) error {
	return os.Remove(c.cachePath(repoName))
}

func (c *IndexCache) cachePath(repoName string) string {
	return filepath.Join(c.CacheDir, repoName+".gob")
}
