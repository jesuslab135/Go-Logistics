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

	// PrivateBucket holds objects that must never be anonymously readable. It is
	// a second bucket rather than per-object ACLs because the public/private
	// split is decided by an upload's purpose and never changes afterwards: one
	// bucket policy is one thing to get right, where N object ACLs are N.
	PrivateBucket string

	// SignedURLTTL bounds a presigned read URL's lifetime.
	SignedURLTTL time.Duration

	// SigningEndpoint is the host signed URLs are addressed to, when that differs
	// from Endpoint. Presigning signs the host, so a URL signed for the address
	// the API dials — "minio:9000" on a container network — is unusable from a
	// browser, and the host cannot be swapped afterwards without breaking the
	// signature. Signing is an offline computation, so a client pointed at the
	// browser-facing address signs correctly without ever connecting to it.
	SigningEndpoint string
	SigningUseSSL   bool
}

// MinIO is an S3-compatible object-storage backend (MinIO, AWS S3, etc.).
type MinIO struct {
	client *minio.Client
	// signer presigns read URLs. It is the same client unless the browser reaches
	// object storage at a different address than the API does.
	signer        *minio.Client
	bucket        string
	privateBucket string
	baseURL       string
	signedTTL     time.Duration
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

	privateBucket := cfg.PrivateBucket
	if privateBucket == "" {
		privateBucket = cfg.Bucket + "-private"
	}
	if privateBucket == cfg.Bucket {
		return nil, fmt.Errorf(
			"storage: STORAGE_MINIO_PRIVATE_BUCKET %q must differ from STORAGE_MINIO_BUCKET — "+
				"the public bucket carries an anonymous-read policy, so sharing it would publish every private object",
			privateBucket)
	}

	for _, bucket := range []string{cfg.Bucket, privateBucket} {
		if err := ensureBucket(ctx, client, bucket, cfg.Region); err != nil {
			return nil, err
		}
	}

	base, err := publicBase(cfg)
	if err != nil {
		return nil, err
	}

	ttl := cfg.SignedURLTTL
	if ttl <= 0 {
		ttl = defaultSignedURLTTL
	}

	signer := client
	if cfg.SigningEndpoint != "" && cfg.SigningEndpoint != cfg.Endpoint {
		signer, err = minio.New(cfg.SigningEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
			Secure: cfg.SigningUseSSL,
			Region: cfg.Region,
		})
		if err != nil {
			return nil, fmt.Errorf("storage: minio signing client: %w", err)
		}
	}

	m := &MinIO{
		client:        client,
		signer:        signer,
		bucket:        cfg.Bucket,
		privateBucket: privateBucket,
		baseURL:       base,
		signedTTL:     ttl,
	}

	if cfg.Public {
		if err := m.setPublicReadPolicy(ctx); err != nil {
			return nil, fmt.Errorf("storage: set public policy: %w", err)
		}
	}
	return m, nil
}

func ensureBucket(ctx context.Context, client *minio.Client, bucket, region string) error {
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("storage: bucket check %q: %w", bucket, err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: region}); err != nil {
			return fmt.Errorf("storage: make bucket %q: %w", bucket, err)
		}
	}
	return nil
}

// bucketFor routes a key to the bucket its visibility requires. Every operation
// goes through this, so a private key can never be written to, read from or
// deleted in the public bucket by an operation that forgot to ask.
func (m *MinIO) bucketFor(key string) string {
	if VisibilityOf(key) == VisibilityPrivate {
		return m.privateBucket
	}
	return m.bucket
}

func (m *MinIO) Save(ctx context.Context, key string, r io.Reader, contentType string) (Object, error) {
	object := cleanKey(key)
	if object == "" {
		return Object{}, ErrInvalidKey
	}
	// Size -1 streams with a default part size, so callers need not know the length up front.
	info, err := m.client.PutObject(ctx, m.bucketFor(object), object, r, -1, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return Object{}, err
	}
	// A private object has no durable URL, only a signed one resolved per read.
	url := ""
	if VisibilityOf(object) == VisibilityPublic {
		url = m.URL(object)
	}
	return Object{Key: object, URL: url, Size: info.Size, ContentType: contentType}, nil
}

func (m *MinIO) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	object := cleanKey(key)
	if object == "" {
		return nil, ErrInvalidKey
	}
	return m.client.GetObject(ctx, m.bucketFor(object), object, minio.GetObjectOptions{})
}

func (m *MinIO) Delete(ctx context.Context, key string) error {
	object := cleanKey(key)
	if object == "" {
		return ErrInvalidKey
	}
	return m.client.RemoveObject(ctx, m.bucketFor(object), object, minio.RemoveObjectOptions{})
}

func (m *MinIO) URL(key string) string {
	return m.baseURL + "/" + cleanKey(key)
}

func (m *MinIO) Stat(ctx context.Context, key string) (Object, error) {
	object := cleanKey(key)
	if object == "" {
		return Object{}, ErrInvalidKey
	}
	info, err := m.client.StatObject(ctx, m.bucketFor(object), object, minio.StatObjectOptions{})
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

// ReadURL returns the URL a client should read key at. A public object keeps its
// permanent URL; a private one is presigned per call and expires, so the caller
// is handed the deadline rather than left to discover it as a 403.
//
// Presigning signs the host, so the signature is only valid against the address
// it was generated for. That address is STORAGE_MINIO_SIGNING_ENDPOINT when set,
// otherwise STORAGE_MINIO_ENDPOINT, and it must be the one the browser can
// reach. A CDN base (STORAGE_MINIO_PUBLIC_URL) cannot be substituted afterwards
// without invalidating the signature.
func (m *MinIO) ReadURL(ctx context.Context, key string) (string, time.Time, error) {
	object := cleanKey(key)
	if object == "" {
		return "", time.Time{}, ErrInvalidKey
	}
	if VisibilityOf(object) == VisibilityPublic {
		return m.URL(object), time.Time{}, nil
	}

	u, err := m.signer.PresignedGetObject(ctx, m.privateBucket, object, m.signedTTL, nil)
	if err != nil {
		return "", time.Time{}, err
	}
	return u.String(), time.Now().Add(m.signedTTL), nil
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
