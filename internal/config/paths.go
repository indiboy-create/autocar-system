package config

import (
	"autocarsystem/internal/i18n"
	"os"
	"path/filepath"
	"strings"
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
	Autocar    string
	Yuvision   string
	Gallery    string
	StatePath  string
	Conflicts  string
	MediaError string
	AppLog     string
	SyncLog    string
}

// stableRoot 保护实时预览下的数据：mygo dev 的可执行文件位于 .mygo/dev 下，
// 每次热重建都会清空该目录。若从项目根目录启动，就把 data/media 放到项目根，
// 避免自动重建删除或覆盖已有数据。发布版路径不受影响。
func stableRoot(root string) string {
	parts := strings.FieldsFunc(filepath.ToSlash(root), func(r rune) bool { return r == '/' })
	devBuild := false
	for i := 0; i+1 < len(parts); i++ {
		if parts[i] == ".mygo" && parts[i+1] == "dev" {
			devBuild = true
			break
		}
	}
	if !devBuild {
		return root
	}
	wd, err := os.Getwd()
	if err != nil {
		return root
	}
	for _, marker := range []string{"mygo.json", "go.mod"} {
		if _, err := os.Stat(filepath.Join(wd, marker)); err == nil {
			return wd
		}
	}
	return root
}

// Load 以主程序所在目录作为根目录，data 与 media 均位于主程序同级。
func Load() (Paths, error) {
	exe, err := executable()
	if err != nil {
		return Paths{}, i18n.Errorf("获取主程序路径失败: %w", err)
	}

	root, err := filepath.Abs(filepath.Dir(exe))
	if err != nil {
		return Paths{}, i18n.Errorf("解析主程序目录失败: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	root = stableRoot(root)

	if err := ensureWritableRoot(root); err != nil {
		return Paths{}, err
	}

	return newPaths(root), nil
}

func ensureWritableRoot(root string) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return i18n.Errorf("创建主程序目录失败: %w", err)
	}
	f, err := os.CreateTemp(root, ".writecheck-*")
	if err != nil {
		return i18n.Errorf("主程序目录不可写: %w", err)
	}
	name := f.Name()
	if err := f.Close(); err != nil {
		_ = os.Remove(name)
		return i18n.Errorf("写入检查失败: %w", err)
	}
	if err := os.Remove(name); err != nil {
		return i18n.Errorf("清理写入检查文件失败: %w", err)
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
		Autocar:    filepath.Join(media, "autocar"),
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
		p.Autocar,
		p.Yuvision,
		p.Gallery,
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return i18n.Errorf("创建目录失败 %s: %w", dir, err)
		}
	}
	return nil
}
