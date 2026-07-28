package imaging

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestThumbKey(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"uploads/7/ab12-photo.jpg", "uploads/7/thumbs/ab12-photo.jpg"},
		{"uploads/7/ab12-photo.png", "uploads/7/thumbs/ab12-photo.jpg"},
		{"photo.gif", "thumbs/photo.jpg"},
		{"uploads/7/noext", "uploads/7/thumbs/noext.jpg"},
	}
	for _, tt := range tests {
		if got := ThumbKey(tt.in); got != tt.want {
			t.Errorf("ThumbKey(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestThumbnailable(t *testing.T) {
	// Every image type the upload allowlist accepts, including WebP, which
	// decodes through x/image.
	for _, ct := range []string{"image/jpeg", "image/png", "image/gif", "image/webp", "IMAGE/PNG", " image/webp "} {
		if !Thumbnailable(ct) {
			t.Errorf("Thumbnailable(%q) = false, want true", ct)
		}
	}
	for _, ct := range []string{"application/pdf", "text/plain", "image/svg+xml", ""} {
		if Thumbnailable(ct) {
			t.Errorf("Thumbnailable(%q) = true, want false", ct)
		}
	}
}

func TestThumbnailBoundsLongestSide(t *testing.T) {
	tests := []struct {
		name               string
		srcW, srcH, maxDim int
		wantW, wantH       int
	}{
		{"landscape", 800, 400, 320, 320, 160},
		{"portrait", 400, 800, 320, 160, 320},
		{"square", 500, 500, 100, 100, 100},
		// Already inside the bound: re-encoded, never enlarged.
		{"smaller than bound", 64, 32, 320, 64, 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Thumbnail(bytes.NewReader(pngOf(t, tt.srcW, tt.srcH)), tt.maxDim)
			if err != nil {
				t.Fatalf("Thumbnail: %v", err)
			}

			cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("decoding the thumbnail: %v", err)
			}
			if format != "jpeg" {
				t.Errorf("format = %q, want jpeg", format)
			}
			if cfg.Width != tt.wantW || cfg.Height != tt.wantH {
				t.Errorf("size = %dx%d, want %dx%d", cfg.Width, cfg.Height, tt.wantW, tt.wantH)
			}
		})
	}
}

// A transparent source must flatten onto white rather than JPEG's default black.
func TestThumbnailFlattensTransparencyOntoWhite(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 40, 40))
	// Fully transparent everywhere.
	data, err := Thumbnail(bytes.NewReader(encodePNG(t, src)), 20)
	if err != nil {
		t.Fatalf("Thumbnail: %v", err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decoding the thumbnail: %v", err)
	}
	r, g, b, _ := img.At(10, 10).RGBA()
	if r>>8 < 250 || g>>8 < 250 || b>>8 < 250 {
		t.Errorf("transparent pixel became (%d,%d,%d), want near-white", r>>8, g>>8, b>>8)
	}
}

// Averaging must actually mix the source, not just sample one corner: a
// half-black half-white image reduced to 1px is grey.
func TestThumbnailAveragesRatherThanSamples(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := range 100 {
		for x := range 100 {
			c := color.RGBA{A: 0xff}
			if x >= 50 {
				c = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
			}
			src.Set(x, y, c)
		}
	}

	data, err := Thumbnail(bytes.NewReader(encodePNG(t, src)), 1)
	if err != nil {
		t.Fatalf("Thumbnail: %v", err)
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decoding the thumbnail: %v", err)
	}

	r, _, _, _ := img.At(0, 0).RGBA()
	if got := r >> 8; got < 100 || got > 155 {
		t.Errorf("1px reduction of half-black/half-white = %d, want mid-grey", got)
	}
}

func TestThumbnailRejectsNonImages(t *testing.T) {
	_, err := Thumbnail(strings.NewReader("this is not an image"), 320)
	if err == nil {
		t.Fatal("Thumbnail accepted a non-image")
	}
}

func pngOf(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 0x80, A: 0xff})
		}
	}
	return encodePNG(t, img)
}

func encodePNG(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encoding the fixture: %v", err)
	}
	return buf.Bytes()
}

// webpFixture is a 60x40 lossless WebP. WebP is the one accepted upload format
// with no standard-library decoder, so a real file is the only way to prove the
// x/image decoder is actually wired in — a synthesised header would pass while
// the import was missing.
const webpFixture = "UklGRjIAAABXRUJQVlA4TCYAAAAvO8AJALkyRPQ/dhHR/wCRtk3R/fsefwgkbfG2fwBjAtCVByj/HA=="

func TestThumbnailDecodesWebP(t *testing.T) {
	raw, err := base64.StdEncoding.DecodeString(webpFixture)
	if err != nil {
		t.Fatal(err)
	}

	out, err := Thumbnail(bytes.NewReader(raw), 20)
	if err != nil {
		t.Fatalf("Thumbnail(webp) failed: %v", err)
	}

	cfg, format, err := image.DecodeConfig(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("thumbnail did not decode: %v", err)
	}
	if format != "jpeg" {
		t.Errorf("format = %q, want jpeg", format)
	}
	if cfg.Width != 20 || cfg.Height != 13 {
		t.Errorf("thumbnail is %dx%d, want 20x13", cfg.Width, cfg.Height)
	}
}
