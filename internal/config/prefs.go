package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"autocarsystem/internal/fileutil"
)

type Prefs struct {
	Language           string `json:"language"`
	Appearance         string `json:"appearance"`
	DefaultPage        string `json:"default_page"`
	AutoCheckUpdates   bool   `json:"auto_check_updates"`
	AutoInstallUpdates bool   `json:"auto_install_updates"`
	// LastUpdateCheck 记录上次自动检查更新的时间（RFC3339），用于每日一次的节流。
	LastUpdateCheck string `json:"last_update_check"`
}

func DefaultPrefs() Prefs {
	return Prefs{
		Language:    "system",
		Appearance:  "system",
		DefaultPage: "gallery",
	}
}

func LoadPrefs(path string) (Prefs, error) {
	prefs := DefaultPrefs()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return prefs, nil
		}
		return prefs, fmt.Errorf("读取偏好设置失败: %w", err)
	}
	if err := json.Unmarshal(data, &prefs); err != nil {
		return prefs, fmt.Errorf("解析偏好设置失败: %w", err)
	}
	return prefs, nil
}

func SavePrefs(path string, prefs Prefs) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建偏好设置目录失败: %w", err)
	}
	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化偏好设置失败: %w", err)
	}
	// 原子写：崩溃/断电时 prefs.json 只会被替换或保留旧内容，不会被截断。
	if err := fileutil.WriteFileAtomic(path, data, 0o644); err != nil {
		return fmt.Errorf("写入偏好设置失败: %w", err)
	}
	return nil
}
