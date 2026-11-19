package media

import (
	"image"
	"image/color"
	_ "image/gif"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"

	// 注册 webp 解码器：既让 EnsureThumbnail 能缩放 webp，也让使用
	// image.Decode 的 myui.DecodeBitmap 能显示 webp 原图与缩略图。
	"autocarsystem/internal/i18n"
	_ "golang.org/x/image/webp"
)

const thumbMaxSide = 320

// ThumbnailRel 返回缩略图相对路径。保留原相对路径的目录结构并统一用 .jpg，
// 避免不同子目录下的同名图片在 thumb/ 中互相覆盖，也避免 webp/gif 被编码成
// JPEG 却保留原扩展名导致解码失败。
func ThumbnailRel(relPath string) string {
	clean := filepath.FromSlash(filepath.Clean(relPath))
	dir := filepath.Dir(clean)
	if dir == "." || dir == string(filepath.Separator) {
		dir = ""
	}
	name := strings.TrimSuffix(filepath.Base(clean), filepath.Ext(clean)) + ".jpg"
	return filepath.ToSlash(filepath.Join("thumb", dir, name))
}

func EnsureThumbnail(finalDir string, relPath string) (string, error) {
	src := filepath.Join(finalDir, filepath.FromSlash(relPath))
	thumbRel := ThumbnailRel(relPath)
	target := filepath.Join(finalDir, filepath.FromSlash(thumbRel))

	if _, err := os.Stat(target); err == nil {
		return thumbRel, nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", i18n.Errorf("创建缩略图目录失败: %w", err)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return "", i18n.Errorf("打开原图失败: %w", err)
	}
	defer srcFile.Close()

	img, _, err := image.Decode(srcFile)
	if err != nil {
		// 编码成 JPEG 的占位缩略图，而不是把原文件字节复制成 .jpg ——
		// 否则解码器会把任意格式当作 JPEG，产生坏图。
		placeholder := image.NewRGBA(image.Rect(0, 0, 1, 1))
		if encErr := encodeJPEG(target, placeholder); encErr != nil {
			return "", i18n.Errorf("生成缩略图占位图失败: %w (解码原图: %v)", encErr, err)
		}
		return thumbRel, nil
	}

	resized := resizeImage(img, thumbMaxSide)

	// 缩略图统一编码为 JPEG（目标扩展名固定为 .jpg），与源格式无关。
	if err := encodeJPEG(target, resized); err != nil {
		return "", err
	}
	return thumbRel, nil
}

// encodeJPEG 把图片编码为 JPEG。先把带 alpha 的图像压到白底 RGB，避免
// PNG/WebP 的透明区域在 JPEG 中变成黑块。
func encodeJPEG(target string, img image.Image) error {
	out, err := os.Create(target)
	if err != nil {
		return i18n.Errorf("创建缩略图失败: %w", err)
	}
	defer out.Close()
	if err := jpeg.Encode(out, flattenAlpha(img), &jpeg.Options{Quality: 82}); err != nil {
		return i18n.Errorf("编码缩略图失败: %w", err)
	}
	return nil
}

func flattenAlpha(src image.Image) image.Image {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := src.At(x, y).RGBA()
			if a != 0xffff {
				// 与白底合成
				r = r*a/0xffff + 0xffff*(0xffff-a)/0xffff
				g = g*a/0xffff + 0xffff*(0xffff-a)/0xffff
				b = b*a/0xffff + 0xffff*(0xffff-a)/0xffff
				a = 0xffff
			}
			dst.SetRGBA(x, y, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)})
		}
	}
	return dst
}

func resizeImage(src image.Image, maxSide int) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 {
		return src
	}
	scale := float64(maxSide) / float64(max(w, h))
	if scale >= 1 {
		return src
	}
	newW := int(float64(w) * scale)
	newH := int(float64(h) * scale)
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	for y := 0; y < newH; y++ {
		sy := bounds.Min.Y + y*h/newH
		for x := 0; x < newW; x++ {
			sx := bounds.Min.X + x*w/newW
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
