package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOConfig struct {
	Endpoint  string // host:port, no scheme
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	UseSSL    bool
	Public    bool   // set an anonymous read policy so URLs work in <img> tags
	PublicURL string // optional base URL override (CDN / custom domain)

	// PublicURLIsBucketRoot allows a PublicURL that does not end in the bucket
	// name, for CDNs whose origin is mapped to the bucket root.
	PublicURLIsBucketRoot bool
}

// MinIO is an S3-compatible object-storage backend (MinIO, AWS S3, etc.).
type MinIO struct {
	client  *minio.Client
	bucket  string
	baseURL string
}

func NewMinIO(ctx context.Context, cfg MinIOConfig) (*MinIO, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("storage: minio client: %w", err)
	}

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("storage: bucket check: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{Region: cfg.Region}); err != nil {
			return nil, fmt.Errorf("storage: make bucket %q: %w", cfg.Bucket, err)
		}
	}

	base, err := publicBase(cfg)
	if err != nil {
		return nil, err
	}

	m := &MinIO{client: client, bucket: cfg.Bucket, baseURL: base}

	if cfg.Public {
		if err := m.setPublicReadPolicy(ctx); err != nil {
			return nil, fmt.Errorf("storage: set public policy: %w", err)
		}
	}
	return m, nil
}

func (m *MinIO) Save(ctx context.Context, key string, r io.Reader, contentType string) (Object, error) {
	object := cleanKey(key)
	if object == "" {
		return Object{}, ErrInvalidKey
	}
	// Size -1 streams with a default part size, so callers need not know the length up front.
	info, err := m.client.PutObject(ctx, m.bucket, object, r, -1, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return Object{}, err
	}
	return Object{Key: object, URL: m.URL(object), Size: info.Size, ContentType: contentType}, nil
}

func (m *MinIO) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	object := cleanKey(key)
	if object == "" {
		return nil, ErrInvalidKey
	}
	return m.client.GetObject(ctx, m.bucket, object, minio.GetObjectOptions{})
}

func (m *MinIO) Delete(ctx context.Context, key string) error {
	object := cleanKey(key)
	if object == "" {
		return ErrInvalidKey
	}
	return m.client.RemoveObject(ctx, m.bucket, object, minio.RemoveObjectOptions{})
}

func (m *MinIO) URL(key string) string {
	return m.baseURL + "/" + cleanKey(key)
}

func (m *MinIO) Stat(ctx context.Context, key string) (Object, error) {
	object := cleanKey(key)
	if object == "" {
		return Object{}, ErrInvalidKey
	}
	info, err := m.client.StatObject(ctx, m.bucket, object, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).StatusCode == http.StatusNotFound {
			return Object{}, ErrNotFound
		}
		return Object{}, err
	}
	return Object{Key: object, URL: m.URL(object), Size: info.Size, ContentType: info.ContentType}, nil
}

func (m *MinIO) KeyFromURL(url string) (string, bool) {
	return keyFromURL(m.baseURL, url)
}

// PresignedGetURL returns a time-limited download URL, for private buckets where
// the object is not publicly readable.
func (m *MinIO) PresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	u, err := m.client.PresignedGetObject(ctx, m.bucket, cleanKey(key), expiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (m *MinIO) setPublicReadPolicy(ctx context.Context) error {
	policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, m.bucket)
	return m.client.SetBucketPolicy(ctx, m.bucket, policy)
}

// publicBase derives the base URL objects are advertised at. The endpoint
// fallback appends the bucket, so an override that omits it produces URLs that
// 404/403 on every read — a misconfiguration invisible on the write path. It is
// rejected here instead, at startup.
func publicBase(cfg MinIOConfig) (string, error) {
	if cfg.PublicURL == "" {
		scheme := "http"
		if cfg.UseSSL {
			scheme = "https"
		}
		return fmt.Sprintf("%s://%s/%s", scheme, cfg.Endpoint, cfg.Bucket), nil
	}

	base := strings.TrimRight(cfg.PublicURL, "/")
	if !cfg.PublicURLIsBucketRoot && path.Base(base) != cfg.Bucket {
		return "", fmt.Errorf(
			"storage: STORAGE_MINIO_PUBLIC_URL %q must end with the bucket name %q "+
				"(objects are stored under it); set STORAGE_MINIO_PUBLIC_URL_IS_BUCKET_ROOT=true "+
				"if a CDN already maps this URL to the bucket root",
			cfg.PublicURL, cfg.Bucket)
	}
	return base, nil
}
