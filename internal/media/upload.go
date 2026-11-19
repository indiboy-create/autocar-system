package media

import (
	"autocarsystem/internal/i18n"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxUploadSize = 50 * 1024 * 1024

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

type SavedUpload struct {
	RelPath string
	AbsPath string
	Created bool
}

func SaveUpload(cacheDir string, finalDir string, srcPath string, kind string) (SavedUpload, error) {
	ext := strings.ToLower(filepath.Ext(srcPath))
	if !allowedExtensions[ext] {
		return SavedUpload{}, i18n.Errorf("不支持的图片格式: %s", ext)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return SavedUpload{}, i18n.Errorf("打开图片失败: %w", err)
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return SavedUpload{}, i18n.Errorf("读取图片信息失败: %w", err)
	}
	if info.Size() == 0 {
		return SavedUpload{}, i18n.Errorf("图片文件为空")
	}
	if info.Size() > maxUploadSize {
		return SavedUpload{}, i18n.Errorf("图片超过大小限制: %d bytes", maxUploadSize)
	}

	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return SavedUpload{}, i18n.Errorf("创建临时目录失败: %w", err)
	}
	tmp, err := os.CreateTemp(cacheDir, "upload-*"+ext)
	if err != nil {
		return SavedUpload{}, i18n.Errorf("创建临时图片失败: %w", err)
	}
	tmpPath := tmp.Name()
	removeTmp := func() { _ = os.Remove(tmpPath) }

	if _, err := io.Copy(tmp, src); err != nil {
		_ = tmp.Close()
		removeTmp()
		return SavedUpload{}, i18n.Errorf("复制临时图片失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		removeTmp()
		return SavedUpload{}, i18n.Errorf("关闭临时图片失败: %w", err)
	}

	hash, err := hashFile(tmpPath)
	if err != nil {
		removeTmp()
		return SavedUpload{}, err
	}

	name := hash + ext
	if kind == "gallery" {
		name = time.Now().Format("20060102150405") + "-" + name
	}
	subDir := "detail"
	if kind == "gallery" {
		subDir = "focus"
	}

	rel := filepath.ToSlash(filepath.Join(subDir, name))
	targetDir := filepath.Join(finalDir, subDir)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		removeTmp()
		return SavedUpload{}, i18n.Errorf("创建图片目录失败: %w", err)
	}
	target := filepath.Join(targetDir, name)

	if _, err := os.Stat(target); err == nil {
		removeTmp()
		return SavedUpload{RelPath: rel, AbsPath: target, Created: false}, nil
	}

	if err := os.Rename(tmpPath, target); err != nil {
		removeTmp()
		return SavedUpload{}, i18n.Errorf("移动图片失败: %w", err)
	}

	return SavedUpload{RelPath: rel, AbsPath: target, Created: true}, nil
}

func RemoveFile(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", i18n.Errorf("打开临时图片计算哈希失败: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", i18n.Errorf("计算图片哈希失败: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// EnsureSeeded 把随程序打包的 resources/media 整棵目录释放到媒体根目录。
// 只补建缺失的目录和文件，已有文件一律保留，不删除、不覆盖。
func EnsureSeeded(mediaRoot string, seed fs.FS) error {
	const prefix = "resources/media"
	return fs.WalkDir(seed, prefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return i18n.Errorf("读取预置媒体资源失败: %w", err)
		}
		rel, err := filepath.Rel(prefix, path)
		if err != nil {
			return i18n.Errorf("解析预置媒体路径失败: %w", err)
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(mediaRoot, filepath.FromSlash(rel))
		if d.IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return i18n.Errorf("创建预置媒体目录失败: %w", err)
			}
			return nil
		}
		if _, err := os.Stat(target); err == nil {
			return nil
		} else if !os.IsNotExist(err) {
			return i18n.Errorf("检查预置媒体文件失败: %w", err)
		}
		data, err := fs.ReadFile(seed, path)
		if err != nil {
			return i18n.Errorf("读取预置媒体文件失败: %w", err)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return i18n.Errorf("创建预置媒体目录失败: %w", err)
		}
		tmp, err := os.CreateTemp(filepath.Dir(target), ".media-seed-*")
		if err != nil {
			return i18n.Errorf("创建临时媒体文件失败: %w", err)
		}
		tmpName := tmp.Name()
		defer os.Remove(tmpName)
		if _, err := tmp.Write(data); err != nil {
			_ = tmp.Close()
			return i18n.Errorf("写入预置媒体文件失败: %w", err)
		}
		if err := tmp.Close(); err != nil {
			return i18n.Errorf("关闭预置媒体文件失败: %w", err)
		}
		if err := os.Rename(tmpName, target); err != nil {
			return i18n.Errorf("安装预置媒体文件失败: %w", err)
		}
		return nil
	})
}
