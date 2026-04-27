package storage

import (
	"fmt"
	"path/filepath"
)

// Config holds the storage configuration loaded from environment variables.
type Config struct {
	Driver  string // "local" | "s3"
	BaseURL string // public base URL (used by local driver)
	Dir     string // upload directory (used by local driver)

	// S3 / R2 — only needed when Driver == "s3"
	S3Bucket    string
	S3Region    string
	S3Endpoint  string // empty for AWS; Cloudflare R2 endpoint otherwise
	S3AccessKey string
	S3SecretKey string
	S3PublicBase string // CDN or bucket public URL
}

// New returns the configured Provider.
// Adding a new backend = add a case here + implement storage.Provider.
func New(cfg Config) (Provider, error) {
	switch cfg.Driver {
	case "local", "":
		dir := cfg.Dir
		if dir == "" {
			dir = filepath.Join(".", "uploads")
		}
		return NewLocalProvider(dir, cfg.BaseURL)

	case "s3":
		return nil, fmt.Errorf(
			"storage: s3 driver is not yet active — " +
				"uncomment S3Provider in internal/storage/s3.go and add the aws-sdk-go-v2 dependency",
		)

	default:
		return nil, fmt.Errorf("storage: unknown driver %q (supported: local, s3)", cfg.Driver)
	}
}
