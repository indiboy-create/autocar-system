package media

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 一个 64x40 的纯色 webp（由 PIL 生成），用于验证 webp 解码注册。
const testWebPBase64 = "UklGRkwAAABXRUJQVlA4IEAAAACQAwCdASpAACgAPm02mEkkIyKhIqgAgA2JZwB2AAAt4QkfWPDAAP7wG9//oUbKJin//9pZ/9Sz/6ln+KgAAAAA"

func TestThumbnailRelPreservesDirAndUsesJPG(t *testing.T) {
	cases := []struct{ in, want string }{
		{"focus/a.webp", "thumb/focus/a.jpg"},
		{"detail/a.png", "thumb/detail/a.jpg"},
		{"a.jpg", "thumb/a.jpg"},
		{"focus/sub/b.jpeg", "thumb/focus/sub/b.jpg"},
	}
	for _, tc := range cases {
		if got := ThumbnailRel(tc.in); got != tc.want {
			t.Errorf("ThumbnailRel(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestThumbnailRelNoCollisionAcrossDirs(t *testing.T) {
	a := ThumbnailRel("focus/photo.webp")
	b := ThumbnailRel("detail/photo.webp")
	if a == b {
		t.Fatalf("thumbnails collide: %q", a)
	}
}

func TestEnsureThumbnailWebP(t *testing.T) {
	dir := t.TempDir()
	srcRel := "focus/sample.webp"
	if err := os.MkdirAll(filepath.Join(dir, "focus"), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := base64.StdEncoding.DecodeString(testWebPBase64)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(srcRel)), data, 0o644); err != nil {
		t.Fatal(err)
	}

	thumbRel, err := EnsureThumbnail(dir, srcRel)
	if err != nil {
		t.Fatalf("EnsureThumbnail(webp): %v", err)
	}
	if !strings.HasSuffix(thumbRel, ".jpg") {
		t.Fatalf("thumb rel = %q, want .jpg", thumbRel)
	}
	f, err := os.Open(filepath.Join(dir, filepath.FromSlash(thumbRel)))
	if err != nil {
		t.Fatalf("open thumb: %v", err)
	}
	defer f.Close()
	img, format, err := image.Decode(f)
	if err != nil {
		t.Fatalf("decode thumb: %v (format=%s)", err, format)
	}
	if format != "jpeg" {
		t.Fatalf("thumb format = %s, want jpeg", format)
	}
	b := img.Bounds()
	if b.Dx() != 64 || b.Dy() != 40 {
		t.Fatalf("thumb size = %dx%d, want 64x40", b.Dx(), b.Dy())
	}
}

func TestEnsureThumbnailPNGAlphaFlattened(t *testing.T) {
	// image.Decode of the produced JPEG must succeed even for transparent PNG input.
	dir := t.TempDir()
	srcRel := "detail/alpha.png"
	if err := os.MkdirAll(filepath.Join(dir, "detail"), 0o755); err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 32, 24))
	if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(srcRel)), mustPNG(t, img), 0o644); err != nil {
		t.Fatal(err)
	}
	thumbRel, err := EnsureThumbnail(dir, srcRel)
	if err != nil {
		t.Fatalf("EnsureThumbnail(png alpha): %v", err)
	}
	f, err := os.Open(filepath.Join(dir, filepath.FromSlash(thumbRel)))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, _, err := image.Decode(f); err != nil {
		t.Fatalf("decode thumb: %v", err)
	}
}

func mustPNG(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
