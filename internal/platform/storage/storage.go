package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var ErrInvalidKey = errors.New("storage: invalid key")

type Object struct {
	Key         string `json:"key"`
	URL         string `json:"url"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

// Storage abstracts blob persistence so handlers depend on the interface, not on
// local disk vs. object storage. Keys are forward-slash relative paths
// (e.g. "media/2026/07/photo.jpg").
type Storage interface {
	Save(ctx context.Context, key string, r io.Reader, contentType string) (Object, error)
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	URL(key string) string
}

// FromEnv builds the storage backend selected by STORAGE_BACKEND (local|minio),
// defaulting to local disk. Both satisfy Storage, so callers are unaffected.
func FromEnv() (Storage, error) {
	switch strings.ToLower(env("STORAGE_BACKEND", "local")) {
	case "minio", "s3":
		return NewMinIO(context.Background(), MinIOConfig{
			Endpoint:  env("STORAGE_MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: os.Getenv("STORAGE_MINIO_ACCESS_KEY"),
			SecretKey: os.Getenv("STORAGE_MINIO_SECRET_KEY"),
			Bucket:    env("STORAGE_MINIO_BUCKET", "fleet"),
			Region:    env("STORAGE_MINIO_REGION", "us-east-1"),
			UseSSL:    env("STORAGE_MINIO_USE_SSL", "false") == "true",
			Public:    env("STORAGE_MINIO_PUBLIC", "false") == "true",
			PublicURL: os.Getenv("STORAGE_MINIO_PUBLIC_URL"),
		})
	default:
		return NewLocal(env("STORAGE_LOCAL_ROOT", "./uploads"), env("STORAGE_BASE_URL", "/media")), nil
	}
}

type Local struct {
	root    string
	baseURL string
}

func NewLocal(root, baseURL string) *Local {
	return &Local{root: root, baseURL: strings.TrimRight(baseURL, "/")}
}

func (l *Local) Save(_ context.Context, key string, r io.Reader, contentType string) (Object, error) {
	full, err := l.resolve(key)
	if err != nil {
		return Object{}, err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return Object{}, err
	}

	f, err := os.Create(full)
	if err != nil {
		return Object{}, err
	}
	defer f.Close()

	n, err := io.Copy(f, r)
	if err != nil {
		return Object{}, err
	}

	return Object{Key: cleanKey(key), URL: l.URL(key), Size: n, ContentType: contentType}, nil
}

func (l *Local) Open(_ context.Context, key string) (io.ReadCloser, error) {
	full, err := l.resolve(key)
	if err != nil {
		return nil, err
	}
	return os.Open(full)
}

func (l *Local) Delete(_ context.Context, key string) error {
	full, err := l.resolve(key)
	if err != nil {
		return err
	}
	if err := os.Remove(full); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (l *Local) URL(key string) string {
	return l.baseURL + "/" + cleanKey(key)
}

// resolve maps a key to an absolute path, refusing any key that would escape the
// storage root (path traversal).
func (l *Local) resolve(key string) (string, error) {
	clean := cleanKey(key)
	if clean == "" {
		return "", ErrInvalidKey
	}
	full := filepath.Join(l.root, filepath.FromSlash(clean))
	rootAbs, err := filepath.Abs(l.root)
	if err != nil {
		return "", err
	}
	fullAbs, err := filepath.Abs(full)
	if err != nil {
		return "", err
	}
	if fullAbs != rootAbs && !strings.HasPrefix(fullAbs, rootAbs+string(os.PathSeparator)) {
		return "", ErrInvalidKey
	}
	return full, nil
}

func cleanKey(key string) string {
	clean := path.Clean("/" + strings.ReplaceAll(key, "\\", "/"))
	return strings.TrimPrefix(clean, "/")
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
