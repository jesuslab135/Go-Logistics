// Package imaging derives thumbnails from uploaded images.
//
// The media gallery renders a grid, and serving full-size originals into it
// costs the client megabytes per screen. Thumbnails are produced once, at upload
// time, and stored beside the original under a derived key.
//
// Downscaling is an area average (each destination pixel is the mean of the
// source box it covers), which is the right filter for large reduction ratios
// and needs no third-party resampler. The only dependency outside the standard
// library is golang.org/x/image/webp, a decoder for the one accepted upload
// format the standard library cannot read.
package imaging

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"path"
	"strings"

	// Decoders for every format the upload allowlist accepts. WebP is not in the
	// standard library, so it brings x/image with it; the alternative was a
	// gallery where WebP items alone fall back to full-size originals.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

const (
	// DefaultMaxDim bounds the longest side of a thumbnail.
	DefaultMaxDim = 320

	// jpegQuality trades a little fidelity for a much smaller grid payload.
	jpegQuality = 82

	// maxPixels caps the decoded dimensions. A small, highly compressible file
	// can declare enormous dimensions ("decompression bomb"); decoding one
	// would allocate 4 bytes per pixel before any of our limits applied.
	maxPixels = 50_000_000

	// ThumbContentType is what Thumbnail always produces. Encoding everything
	// as JPEG keeps one key convention and one content type for the gallery.
	ThumbContentType = "image/jpeg"
)

// ErrUnsupported is returned when the source is not a format we can decode.
var ErrUnsupported = errors.New("imaging: unsupported image format")

// thumbnailable are the sniffed content types Thumbnail can handle. Output is
// always JPEG, so a format only needs a decoder to appear here.
var thumbnailable = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// Thumbnailable reports whether a thumbnail can be derived from contentType.
func Thumbnailable(contentType string) bool {
	return thumbnailable[strings.ToLower(strings.TrimSpace(contentType))]
}

// ThumbKey is the deterministic key of the thumbnail for an original object.
// Thumbnails live in a sibling "thumbs/" folder and always carry a .jpg suffix,
// so the pairing survives a re-upload and needs nothing recorded elsewhere.
func ThumbKey(key string) string {
	dir, file := path.Split(key)
	base := strings.TrimSuffix(file, path.Ext(file))
	if base == "" {
		base = "thumb"
	}
	return dir + "thumbs/" + base + ".jpg"
}

// Thumbnail decodes r and returns a JPEG whose longest side is at most maxDim.
// Images already within the bound are re-encoded rather than enlarged.
func Thumbnail(r io.Reader, maxDim int) ([]byte, error) {
	if maxDim < 1 {
		maxDim = DefaultMaxDim
	}

	// Buffer once: the config probe and the decode both need to read from the
	// start, and the caller's reader is not required to be seekable.
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnsupported, err)
	}
	if cfg.Width < 1 || cfg.Height < 1 {
		return nil, ErrUnsupported
	}
	if cfg.Width*cfg.Height > maxPixels {
		return nil, fmt.Errorf("imaging: image is %dx%d, above the %d pixel limit", cfg.Width, cfg.Height, maxPixels)
	}

	src, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnsupported, err)
	}

	dstW, dstH := fit(src.Bounds().Dx(), src.Bounds().Dy(), maxDim)
	thumb := downscale(flatten(src), dstW, dstH)

	var out bytes.Buffer
	if err := jpeg.Encode(&out, thumb, &jpeg.Options{Quality: jpegQuality}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// fit scales (w, h) down so the longest side is maxDim, preserving aspect ratio.
func fit(w, h, maxDim int) (int, int) {
	if w <= maxDim && h <= maxDim {
		return w, h
	}
	if w >= h {
		return maxDim, max(1, h*maxDim/w)
	}
	return max(1, w*maxDim/h), maxDim
}

// flatten composites src onto white in an RGBA buffer. It gives the downscaler
// direct pixel access, and it resolves transparency before JPEG encoding turns
// it into black.
func flatten(src image.Image) *image.RGBA {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Over)
	return dst
}

// downscale averages each destination pixel over the source box it covers.
func downscale(src *image.RGBA, dstW, dstH int) *image.RGBA {
	srcW, srcH := src.Bounds().Dx(), src.Bounds().Dy()
	if dstW == srcW && dstH == srcH {
		return src
	}

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	for dy := range dstH {
		y0 := dy * srcH / dstH
		y1 := max((dy+1)*srcH/dstH, y0+1)

		for dx := range dstW {
			x0 := dx * srcW / dstW
			x1 := max((dx+1)*srcW/dstW, x0+1)

			var r, g, b uint32
			var n uint32
			for y := y0; y < y1; y++ {
				for x := x0; x < x1; x++ {
					i := src.PixOffset(x, y)
					r += uint32(src.Pix[i])
					g += uint32(src.Pix[i+1])
					b += uint32(src.Pix[i+2])
					n++
				}
			}

			i := dst.PixOffset(dx, dy)
			dst.Pix[i] = uint8(r / n)
			dst.Pix[i+1] = uint8(g / n)
			dst.Pix[i+2] = uint8(b / n)
			// flatten already removed transparency.
			dst.Pix[i+3] = 0xff
		}
	}
	return dst
}
