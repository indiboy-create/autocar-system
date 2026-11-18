package media

import (
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const thumbMaxSide = 320

func ThumbnailRel(relPath string) string {
	name := filepath.Base(filepath.FromSlash(relPath))
	return filepath.ToSlash(filepath.Join("thumb", name))
}

func EnsureThumbnail(finalDir string, relPath string) (string, error) {
	src := filepath.Join(finalDir, filepath.FromSlash(relPath))
	thumbRel := ThumbnailRel(relPath)
	target := filepath.Join(finalDir, filepath.FromSlash(thumbRel))

	if _, err := os.Stat(target); err == nil {
		return thumbRel, nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", fmt.Errorf("创建缩略图目录失败: %w", err)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return "", fmt.Errorf("打开原图失败: %w", err)
	}
	defer srcFile.Close()

	img, _, err := image.Decode(srcFile)
	if err != nil {
		if err := copyFile(src, target); err != nil {
			return "", err
		}
		return thumbRel, nil
	}

	resized := resizeImage(img, thumbMaxSide)
	ext := strings.ToLower(filepath.Ext(src))

	out, err := os.Create(target)
	if err != nil {
		return "", fmt.Errorf("创建缩略图失败: %w", err)
	}
	defer out.Close()

	if ext == ".png" {
		if err := png.Encode(out, resized); err != nil {
			return "", fmt.Errorf("编码 PNG 缩略图失败: %w", err)
		}
		return thumbRel, nil
	}
	if err := jpeg.Encode(out, resized, &jpeg.Options{Quality: 82}); err != nil {
		return "", fmt.Errorf("编码缩略图失败: %w", err)
	}
	return thumbRel, nil
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

func copyFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开复制源文件失败: %w", err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("创建复制目标文件失败: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}
	return nil
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
