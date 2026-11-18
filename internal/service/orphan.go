package service

import (
	"autocarsystem/internal/i18n"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"autocarsystem/internal/media"
)

func (s *Service) ScanOrphanGalleryDirs(ctx context.Context) ([]string, error) {
	rows, err := s.store.Query(ctx, "SELECT series_id, id FROM gallery")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("读取图集目录引用失败"), err)
	}
	known := make(map[string]bool, len(rows.Values))
	for _, v := range rows.Values {
		if len(v) < 2 {
			continue
		}
		known[fmt.Sprintf("%d/%d", valueInt64(v[0]), valueInt64(v[1]))] = true
	}

	root := s.paths.Gallery
	if root == "" {
		return []string{}, nil
	}
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("%s: %w", i18n.T("访问图集媒体目录失败"), err)
	}

	seriesEntries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("扫描图集媒体目录失败"), err)
	}

	orphans := make([]string, 0)
	for _, seriesEntry := range seriesEntries {
		if !seriesEntry.IsDir() {
			continue
		}
		seriesDir := filepath.Join(root, seriesEntry.Name())
		galleryEntries, err := os.ReadDir(seriesDir)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", i18n.T("扫描车系图集目录失败"), err)
		}
		for _, galleryEntry := range galleryEntries {
			if !galleryEntry.IsDir() {
				continue
			}
			rel := filepath.ToSlash(filepath.Join(seriesEntry.Name(), galleryEntry.Name()))
			if !known[rel] {
				orphans = append(orphans, rel)
			}
		}
	}

	sort.Strings(orphans)
	return orphans, nil
}

func (s *Service) CleanupOrphanGalleryDirs(ctx context.Context) (int, error) {
	orphans, err := s.ScanOrphanGalleryDirs(ctx)
	if err != nil {
		return 0, err
	}

	removed := 0
	var firstErr error
	for _, rel := range orphans {
		target := filepath.Join(s.paths.Gallery, filepath.FromSlash(rel))
		if err := os.RemoveAll(target); err != nil {
			_ = media.AppendError(s.paths.MediaError, "cleanup-orphan-gallery-dir", target, err)
			if firstErr == nil {
				firstErr = fmt.Errorf("%s %s: %w", i18n.T("清理孤儿目录失败"), rel, err)
			}
			continue
		}
		removed++
	}
	return removed, firstErr
}
