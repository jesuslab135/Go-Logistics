package handler

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"fleet/internal/http/dto"
	"fleet/internal/platform/storage"
)

func pngBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

// postUpload builds a multipart request with an optional purpose field.
func postUpload(t *testing.T, content []byte, filename, purpose string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if purpose != "" {
		if err := w.WriteField("purpose", purpose); err != nil {
			t.Fatalf("write purpose: %v", err)
		}
	}
	w.Close()

	r := gin.New()
	r.POST("/uploads", NewUploadHandler(storage.NewLocal(t.TempDir(), "/media"), nil).Upload)

	req := httptest.NewRequest(http.MethodPost, "/uploads", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestUploadPurposePolicies(t *testing.T) {
	pdf := []byte("%PDF-1.4\n% test document\n")
	img := pngBytes(t)

	tests := []struct {
		name     string
		content  []byte
		filename string
		purpose  string
		want     int
	}{
		{"image with no purpose uses the generic allowlist", img, "photo.png", "", http.StatusCreated},
		{"pdf with no purpose uses the generic allowlist", pdf, "doc.pdf", "", http.StatusCreated},
		{"image as a photo", img, "photo.png", "photo", http.StatusCreated},
		{"pdf as a document", pdf, "doc.pdf", "document", http.StatusCreated},
		{"image rejected as a document", img, "photo.png", "document", http.StatusUnsupportedMediaType},
		{"pdf rejected as a photo", pdf, "doc.pdf", "photo", http.StatusUnsupportedMediaType},
		{"unknown purpose is rejected", img, "photo.png", "bogus", http.StatusBadRequest},
		{"image as an asset photo", img, "truck.png", "asset_photo", http.StatusCreated},
		{"image as a company logo", img, "logo.png", "company_logo", http.StatusCreated},
		{"pdf as a receipt", pdf, "receipt.pdf", "receipt", http.StatusCreated},
		{"image as a signature", img, "sig.png", "signature", http.StatusCreated},
		{"image as inspection evidence", img, "damage.png", "inspection_photo", http.StatusCreated},
		{"pdf rejected as a company logo", pdf, "logo.pdf", "company_logo", http.StatusUnsupportedMediaType},
		{"pdf rejected as a signature", pdf, "sig.pdf", "signature", http.StatusUnsupportedMediaType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := postUpload(t, tt.content, tt.filename, tt.purpose)
			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d (body %s)", rec.Code, tt.want, rec.Body)
			}
		})
	}
}

// The purpose decides whether an object is world-readable, and the key is what
// carries that decision into every later read, write and delete. Getting this
// mapping wrong publishes receipts, so each purpose's visibility is pinned.
func TestUploadPurposeDecidesVisibility(t *testing.T) {
	tests := []struct {
		purpose string
		want    string
	}{
		{"asset_photo", "public"},
		{"company_logo", "public"},
		{"receipt", "private"},
		{"signature", "private"},
		{"inspection_photo", "private"},
		{"media", "private"},
		{"generic", "private"},
		// An unclassified legacy purpose must not publish: a caller that did not
		// say what the file is for has not established that publishing is safe.
		{"photo", "private"},
		{"document", "private"},
		{"", "private"},
	}

	for _, tt := range tests {
		name := tt.purpose
		if name == "" {
			name = "unspecified"
		}
		t.Run(name, func(t *testing.T) {
			content, filename := pngBytes(t), "file.png"
			if tt.purpose == "document" {
				content, filename = []byte("%PDF-1.4\n"), "file.pdf"
			}

			rec := postUpload(t, content, filename, tt.purpose)
			if rec.Code != http.StatusCreated {
				t.Fatalf("status = %d, want 201 (body %s)", rec.Code, rec.Body)
			}

			var body dto.UploadResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if body.Visibility != tt.want {
				t.Errorf("visibility = %q, want %q", body.Visibility, tt.want)
			}
			// The key must agree with the reported visibility, since the key is
			// what every later operation routes on.
			wantPrefix := "uploads/" + tt.want + "/"
			if !strings.HasPrefix(body.Key, wantPrefix) {
				t.Errorf("key = %q, want it under %q", body.Key, wantPrefix)
			}
		})
	}
}

func TestUploadRejectsUnknownPurposeWithStableCode(t *testing.T) {
	rec := postUpload(t, pngBytes(t), "photo.png", "bogus")

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error.Code != "invalid_upload_purpose" {
		t.Errorf("code = %q, want invalid_upload_purpose", body.Error.Code)
	}
}

// The declared Content-Type is not trusted: a PDF renamed .png is still a PDF.
func TestUploadIgnoresTheClaimedExtension(t *testing.T) {
	rec := postUpload(t, []byte("%PDF-1.4\n"), "actually-a-pdf.png", "photo")

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("status = %d, want 415 (body %s)", rec.Code, rec.Body)
	}
}
