package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"autocarsystem/internal/domain"
	"autocarsystem/internal/i18n"
)

// BackupDatabase 立即生成一份数据库备份，命名与迁移备份一致，便于统一列表管理。
func (s *Service) BackupDatabase(ctx context.Context) (domain.BackupInfo, error) {
	if s.paths.BackupDir == "" {
		return domain.BackupInfo{}, errors.New(i18n.T("备份目录未配置"))
	}
	if err := os.MkdirAll(s.paths.BackupDir, 0o755); err != nil {
		return domain.BackupInfo{}, i18n.Errorf("创建备份目录失败: %w", err)
	}

	name := fmt.Sprintf("car.db.bak-%s", time.Now().Format("20060102-150405"))
	backupPath := filepath.Join(s.paths.BackupDir, name)
	for i := 2; fileExists(backupPath); i++ {
		backupPath = filepath.Join(s.paths.BackupDir, fmt.Sprintf("%s-%d", name, i))
	}

	if err := s.store.Backup(ctx, backupPath); err != nil {
		return domain.BackupInfo{}, err
	}
	info, err := os.Stat(backupPath)
	if err != nil {
		return domain.BackupInfo{}, i18n.Errorf("读取备份文件失败: %w", err)
	}
	return domain.BackupInfo{
		Name:    filepath.Base(backupPath),
		Path:    backupPath,
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (s *Service) ListBackups() ([]domain.BackupInfo, error) {
	if s.paths.BackupDir == "" {
		return []domain.BackupInfo{}, nil
	}
	entries, err := os.ReadDir(s.paths.BackupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []domain.BackupInfo{}, nil
		}
		return nil, fmt.Errorf("%s: %w", i18n.T("读取备份目录失败"), err)
	}

	items := make([]domain.BackupInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "car.db.bak-") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		items = append(items, domain.BackupInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(s.paths.BackupDir, entry.Name()),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ModTime.After(items[j].ModTime)
	})
	return items, nil
}
