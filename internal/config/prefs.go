package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Prefs struct {
	Language           string `json:"language"`
	Appearance         string `json:"appearance"`
	DefaultPage        string `json:"default_page"`
	AutoCheckUpdates   bool   `json:"auto_check_updates"`
	AutoInstallUpdates bool   `json:"auto_install_updates"`
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
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("写入偏好设置失败: %w", err)
	}
	return nil
}
