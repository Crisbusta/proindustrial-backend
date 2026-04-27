package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalProvider stores files on the local filesystem.
// Files are written to Dir and served via BaseURL (e.g. http://localhost:8080/uploads).
// Switching to S3/R2 later requires zero changes to callers — only swap the Provider.
type LocalProvider struct {
	Dir     string // absolute path to uploads directory
	BaseURL string // public base URL, e.g. "http://localhost:8080/uploads"
}

func NewLocalProvider(dir, baseURL string) (*LocalProvider, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("storage: cannot create uploads dir %q: %w", dir, err)
	}
	return &LocalProvider{Dir: dir, BaseURL: baseURL}, nil
}

func (p *LocalProvider) Upload(_ context.Context, key string, r io.Reader, _ int64, _ string) (string, error) {
	dst := filepath.Join(p.Dir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", fmt.Errorf("storage: mkdir %q: %w", filepath.Dir(dst), err)
	}

	f, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("storage: create file %q: %w", dst, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return "", fmt.Errorf("storage: write file %q: %w", dst, err)
	}
	return p.PublicURL(key), nil
}

func (p *LocalProvider) Delete(_ context.Context, key string) error {
	dst := filepath.Join(p.Dir, filepath.FromSlash(key))
	err := os.Remove(dst)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (p *LocalProvider) PublicURL(key string) string {
	return p.BaseURL + "/" + key
}
