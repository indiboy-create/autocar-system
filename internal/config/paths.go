package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// executable 返回当前主程序路径；测试可替换以验证目录规则。
var executable = os.Executable

type Paths struct {
	Root       string
	Data       string
	Media      string
	DBPath     string
	PrefsPath  string
	BackupDir  string
	CacheTmp   string
	SyncDir    string
	RawDir     string
	LogsDir    string
	UpdateDir  string
	Autohome   string
	Yuvision   string
	Gallery    string
	StatePath  string
	Conflicts  string
	MediaError string
	AppLog     string
	SyncLog    string
}

// Load 以主程序所在目录作为根目录，data 与 media 均位于主程序同级。
func Load() (Paths, error) {
	exe, err := executable()
	if err != nil {
		return Paths{}, fmt.Errorf("获取主程序路径失败: %w", err)
	}

	root, err := filepath.Abs(filepath.Dir(exe))
	if err != nil {
		return Paths{}, fmt.Errorf("解析主程序目录失败: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}

	if err := ensureWritableRoot(root); err != nil {
		return Paths{}, err
	}

	return newPaths(root), nil
}

func ensureWritableRoot(root string) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("创建主程序目录失败: %w", err)
	}
	f, err := os.CreateTemp(root, ".writecheck-*")
	if err != nil {
		return fmt.Errorf("主程序目录不可写: %w", err)
	}
	name := f.Name()
	if err := f.Close(); err != nil {
		_ = os.Remove(name)
		return fmt.Errorf("写入检查失败: %w", err)
	}
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("清理写入检查文件失败: %w", err)
	}
	return nil
}

func newPaths(root string) Paths {
	data := filepath.Join(root, "data")
	media := filepath.Join(root, "media")
	syncDir := filepath.Join(data, "sync")
	logsDir := filepath.Join(data, "logs")

	return Paths{
		Root:       root,
		Data:       data,
		Media:      media,
		DBPath:     filepath.Join(data, "car.db"),
		PrefsPath:  filepath.Join(data, "prefs.json"),
		BackupDir:  filepath.Join(data, "backup"),
		CacheTmp:   filepath.Join(data, "cache", "tmp"),
		SyncDir:    syncDir,
		RawDir:     filepath.Join(syncDir, "raw"),
		LogsDir:    logsDir,
		UpdateDir:  filepath.Join(data, "update"),
		Autohome:   filepath.Join(media, "autohome"),
		Yuvision:   filepath.Join(media, "youvision"),
		Gallery:    filepath.Join(media, "gallery"),
		StatePath:  filepath.Join(syncDir, "state.json"),
		Conflicts:  filepath.Join(syncDir, "conflicts.jsonl"),
		MediaError: filepath.Join(logsDir, "media-errors.jsonl"),
		AppLog:     filepath.Join(logsDir, "app.log"),
		SyncLog:    filepath.Join(logsDir, "sync.log"),
	}
}

func (p Paths) EnsureDirs() error {
	dirs := []string{
		p.Data,
		p.BackupDir,
		p.CacheTmp,
		p.SyncDir,
		p.RawDir,
		p.LogsDir,
		p.UpdateDir,
		p.Media,
		p.Autohome,
		p.Yuvision,
		p.Gallery,
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("创建目录失败 %s: %w", dir, err)
		}
	}
	return nil
}
