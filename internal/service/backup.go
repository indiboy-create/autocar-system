package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"autocarsystem/internal/domain"
	"autocarsystem/internal/i18n"
)

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
