package storage

import "testing"

// publicBase is the one place a misconfiguration is silent on the write path and
// only shows up as broken images in a client, so it is pinned here.
func TestPublicBase(t *testing.T) {
	tests := []struct {
		name    string
		cfg     MinIOConfig
		want    string
		wantErr bool
	}{
		{
			name: "endpoint fallback appends the bucket",
			cfg:  MinIOConfig{Endpoint: "localhost:9000", Bucket: "fleet"},
			want: "http://localhost:9000/fleet",
		},
		{
			name: "ssl fallback uses https",
			cfg:  MinIOConfig{Endpoint: "s3.example.com", Bucket: "fleet", UseSSL: true},
			want: "https://s3.example.com/fleet",
		},
		{
			name: "override ending in the bucket is accepted",
			cfg:  MinIOConfig{Endpoint: "minio:9000", Bucket: "fleet", PublicURL: "http://localhost:9010/fleet"},
			want: "http://localhost:9010/fleet",
		},
		{
			name: "trailing slash is trimmed",
			cfg:  MinIOConfig{Endpoint: "minio:9000", Bucket: "fleet", PublicURL: "http://localhost:9010/fleet/"},
			want: "http://localhost:9010/fleet",
		},
		{
			name:    "override missing the bucket is rejected",
			cfg:     MinIOConfig{Endpoint: "minio:9000", Bucket: "fleet", PublicURL: "http://localhost:9010"},
			wantErr: true,
		},
		{
			name:    "a different trailing segment is rejected",
			cfg:     MinIOConfig{Endpoint: "minio:9000", Bucket: "fleet", PublicURL: "https://cdn.example.com/media"},
			wantErr: true,
		},
		{
			name: "bucket-root opt-out allows a CDN mapping",
			cfg: MinIOConfig{
				Endpoint: "minio:9000", Bucket: "fleet",
				PublicURL: "https://cdn.example.com", PublicURLIsBucketRoot: true,
			},
			want: "https://cdn.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := publicBase(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("publicBase() = %q, want an error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("publicBase() error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("publicBase() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestKeyFromURL(t *testing.T) {
	const base = "http://localhost:9010/fleet"

	tests := []struct {
		url  string
		want string
		ok   bool
	}{
		{base + "/uploads/1/abc-photo.png", "uploads/1/abc-photo.png", true},
		{base + "/", "", false},
		{base, "", false},
		{"https://evil.example.com/uploads/1/abc.png", "", false},
		{"http://localhost:9010/other/uploads/1/abc.png", "", false},
		// Traversal must not escape the bucket namespace.
		{base + "/../../etc/passwd", "etc/passwd", true},
	}

	for _, tt := range tests {
		got, ok := keyFromURL(base, tt.url)
		if ok != tt.ok || got != tt.want {
			t.Errorf("keyFromURL(%q) = (%q, %v), want (%q, %v)", tt.url, got, ok, tt.want, tt.ok)
		}
	}
}
