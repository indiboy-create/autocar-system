package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"autocarsystem/internal/fileutil"
	"autocarsystem/internal/i18n"
)

type Prefs struct {
	Language           string `json:"language"`
	Appearance         string `json:"appearance"`
	ColorScheme        string `json:"color_scheme"`
	DefaultPage        string `json:"default_page"`
	AutoCheckUpdates   bool   `json:"auto_check_updates"`
	AutoInstallUpdates bool   `json:"auto_install_updates"`
	// AutoBackup 开启后按 BackupIntervalHours 周期性备份数据库；0 表示关闭。
	AutoBackup          bool `json:"auto_backup"`
	BackupIntervalHours int  `json:"backup_interval_hours"`
	// LastBackupAt 记录上次自动备份时间的 RFC3339，用于节流。
	LastBackupAt string `json:"last_backup_at"`
	// LastUpdateCheck 记录上次自动检查更新的时间（RFC3339），仅为兼容旧配置保留，
	// 运行时的检查时间已由 updater 插件的 updater.json 管理。
	LastUpdateCheck string `json:"last_update_check"`
	// UpdatePrefsMigrated 标记旧更新偏好是否已迁移到 updater 插件状态。
	UpdatePrefsMigrated bool `json:"update_prefs_migrated"`
	// SyncSpecs 控制同步时是否包含年款；默认关闭，需在设置->同步单独开启。
	SyncSpecs bool `json:"sync_specs"`
}

func DefaultPrefs() Prefs {
	return Prefs{
		Language:            "system",
		Appearance:          "system",
		ColorScheme:         "orange",
		DefaultPage:         "gallery",
		AutoCheckUpdates:    true,
		AutoInstallUpdates:  true,
		AutoBackup:          true,
		BackupIntervalHours: 72,
	}
}

func LoadPrefs(path string) (Prefs, error) {
	prefs := DefaultPrefs()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return prefs, nil
		}
		return prefs, i18n.Errorf("读取偏好设置失败: %w", err)
	}
	if err := json.Unmarshal(data, &prefs); err != nil {
		return prefs, i18n.Errorf("解析偏好设置失败: %w", err)
	}
	return prefs, nil
}

func SavePrefs(path string, prefs Prefs) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return i18n.Errorf("创建偏好设置目录失败: %w", err)
	}
	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return i18n.Errorf("序列化偏好设置失败: %w", err)
	}
	// 原子写：崩溃/断电时 prefs.json 只会被替换或保留旧内容，不会被截断。
	if err := fileutil.WriteFileAtomic(path, data, 0o644); err != nil {
		return i18n.Errorf("写入偏好设置失败: %w", err)
	}
	return nil
}
