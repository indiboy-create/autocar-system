package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"autocarsystem/internal/domain"
	"autocarsystem/internal/i18n"
	"autocarsystem/internal/media"
)

func (s *Service) CreateYuvision(ctx context.Context, brandID int64, firmID int64, seriesID int64, categoryID int64, name string, description string) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, fmt.Errorf("%s", i18n.T("版型名称不能为空"))
	}

	ok, err := s.exists(ctx, "SELECT 1 FROM series WHERE id = ? AND brand_id = ? AND firm_id = ?", seriesID, brandID, firmID)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("%s", i18n.T("车系不存在或归属不匹配"))
	}

	ok, err = s.exists(ctx, "SELECT 1 FROM category WHERE id = ? AND enabled = 1", categoryID)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("%s", i18n.T("分类不存在或已禁用"))
	}

	now := time.Now().Format(time.RFC3339)
	result, err := s.store.Execute(ctx, `
INSERT INTO youvision(
    brand_id, firm_id, series_id, category_id,
    name, cover_image, images, description, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, '', '[]', ?, ?, ?)
`, brandID, firmID, seriesID, categoryID, name, description, now, now)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", i18n.T("新增版型失败"), err)
	}
	return result.LastInsertID, nil
}

func (s *Service) ListYuvisions(ctx context.Context, limit int) ([]domain.Yuvision, error) {
	rows, err := s.store.Query(ctx, `
SELECT id, brand_id, firm_id, series_id, category_id,
       name, cover_image, images, description, created_at, updated_at
FROM youvision
ORDER BY updated_at DESC, id DESC
LIMIT ?
`, limit)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("读取版型失败"), err)
	}

	items := make([]domain.Yuvision, 0, len(rows.Values))
	for _, values := range rows.Values {
		if len(values) < 11 {
			continue
		}
		items = append(items, domain.Yuvision{
			ID:          valueInt64(values[0]),
			BrandID:     valueInt64(values[1]),
			FirmID:      valueInt64(values[2]),
			SeriesID:    valueInt64(values[3]),
			CategoryID:  valueInt64(values[4]),
			Name:        valueString(values[5]),
			CoverImage:  valueString(values[6]),
			Images:      valueString(values[7]),
			Description: valueString(values[8]),
			CreatedAt:   valueString(values[9]),
			UpdatedAt:   valueString(values[10]),
		})
	}
	return items, nil
}

func (s *Service) UpdateYuvisionText(ctx context.Context, id int64, name string, description string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%s", i18n.T("版型名称不能为空"))
	}
	_, err := s.store.Execute(ctx,
		"UPDATE youvision SET name = ?, description = ?, updated_at = ? WHERE id = ?",
		name,
		description,
		time.Now().Format(time.RFC3339),
		id,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("更新版型失败"), err)
	}
	return nil
}

func (s *Service) DeleteYuvision(ctx context.Context, id int64) error {
	ok, err := s.exists(ctx, "SELECT 1 FROM gallery WHERE youvision_id = ? LIMIT 1", id)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("%s", i18n.T("该版型下存在图集，禁止删除"))
	}
	if _, err := s.store.Execute(ctx, "DELETE FROM youvision WHERE id = ?", id); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("删除版型失败"), err)
	}
	return nil
}

func (s *Service) CreateGallery(ctx context.Context, yuvisionID int64, title string) (int64, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return 0, fmt.Errorf("%s", i18n.T("图集标题不能为空"))
	}

	yuvision, err := s.getYuvision(ctx, yuvisionID)
	if err != nil {
		return 0, err
	}

	now := time.Now().Format(time.RFC3339)
	result, err := s.store.Execute(ctx, `
INSERT INTO gallery(
    youvision_id, brand_id, firm_id, series_id, category_id,
    title, cover_image, images, image_count, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, '', '[]', 0, ?, ?)
`, yuvisionID,
		yuvision.BrandID,
		yuvision.FirmID,
		yuvision.SeriesID,
		yuvision.CategoryID,
		title,
		now,
		now,
	)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", i18n.T("新增图集失败"), err)
	}
	return result.LastInsertID, nil
}

func (s *Service) ListGalleries(ctx context.Context, limit int) ([]domain.Gallery, error) {
	rows, err := s.store.Query(ctx, `
SELECT id, youvision_id, brand_id, firm_id, series_id, category_id,
       title, cover_image, images, image_count, created_at, updated_at
FROM gallery
ORDER BY updated_at DESC, id DESC
LIMIT ?
`, limit)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("读取图集失败"), err)
	}

	items := make([]domain.Gallery, 0, len(rows.Values))
	for _, values := range rows.Values {
		if len(values) < 12 {
			continue
		}
		items = append(items, domain.Gallery{
			ID:          valueInt64(values[0]),
			YouvisionID: valueInt64(values[1]),
			BrandID:     valueInt64(values[2]),
			FirmID:      valueInt64(values[3]),
			SeriesID:    valueInt64(values[4]),
			CategoryID:  valueInt64(values[5]),
			Title:       valueString(values[6]),
			CoverImage:  valueString(values[7]),
			Images:      valueString(values[8]),
			ImageCount:  int(valueInt64(values[9])),
			CreatedAt:   valueString(values[10]),
			UpdatedAt:   valueString(values[11]),
		})
	}
	return items, nil
}

func (s *Service) UpdateGalleryTitle(ctx context.Context, id int64, title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return fmt.Errorf("%s", i18n.T("图集标题不能为空"))
	}
	_, err := s.store.Execute(ctx,
		"UPDATE gallery SET title = ?, updated_at = ? WHERE id = ?",
		title,
		time.Now().Format(time.RFC3339),
		id,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("更新图集失败"), err)
	}
	return nil
}

func (s *Service) DeleteGallery(ctx context.Context, id int64) error {
	gallery, err := s.getGallery(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.store.Execute(ctx, "DELETE FROM gallery WHERE id = ?", id); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("删除图集失败"), err)
	}

	dir := filepath.Join(s.paths.Gallery, fmt.Sprint(gallery.SeriesID), fmt.Sprint(gallery.ID))
	if err := os.RemoveAll(dir); err != nil {
		_ = media.AppendError(s.paths.MediaError, "delete-gallery-dir", dir, err)
	}
	return nil
}

func (s *Service) AddYuvisionImage(ctx context.Context, id int64, srcPath string) (string, error) {
	item, err := s.getYuvision(ctx, id)
	if err != nil {
		return "", err
	}
	images, err := parseImages(item.Images)
	if err != nil {
		return "", err
	}

	dir := filepath.Join(s.paths.Yuvision, fmt.Sprint(item.SeriesID), fmt.Sprint(item.ID))
	saved, err := media.SaveUpload(s.paths.CacheTmp, dir, srcPath, "youvision")
	if err != nil {
		return "", err
	}
	thumbRel := media.ThumbnailRel(saved.RelPath)
	thumbAbs := filepath.Join(dir, filepath.FromSlash(thumbRel))
	if _, thumbErr := media.EnsureThumbnail(dir, saved.RelPath); thumbErr != nil {
		_ = media.AppendError(s.paths.MediaError, "make-yuvision-thumb", saved.AbsPath, thumbErr)
	}
	if containsString(images, saved.RelPath) {
		return saved.RelPath, nil
	}

	images = append(images, saved.RelPath)
	data, err := marshalImages(images)
	if err != nil {
		return "", err
	}
	cover := item.CoverImage
	if cover == "" {
		cover = saved.RelPath
	}

	_, err = s.store.Execute(ctx,
		"UPDATE youvision SET images = ?, cover_image = ?, updated_at = ? WHERE id = ?",
		data,
		cover,
		time.Now().Format(time.RFC3339),
		id,
	)
	if err != nil {
		if saved.Created {
			_ = media.RemoveFile(saved.AbsPath)
			_ = media.RemoveFile(thumbAbs)
			_ = media.AppendError(s.paths.MediaError, "rollback-yuvision-image", saved.AbsPath, err)
		}
		return "", fmt.Errorf("%s: %w", i18n.T("更新版型图片失败"), err)
	}
	return saved.RelPath, nil
}

func (s *Service) AddGalleryImage(ctx context.Context, id int64, srcPath string) (string, error) {
	item, err := s.getGallery(ctx, id)
	if err != nil {
		return "", err
	}
	images, err := parseImages(item.Images)
	if err != nil {
		return "", err
	}

	dir := filepath.Join(s.paths.Gallery, fmt.Sprint(item.SeriesID), fmt.Sprint(item.ID))
	saved, err := media.SaveUpload(s.paths.CacheTmp, dir, srcPath, "gallery")
	if err != nil {
		return "", err
	}
	thumbRel := media.ThumbnailRel(saved.RelPath)
	thumbAbs := filepath.Join(dir, filepath.FromSlash(thumbRel))
	if _, thumbErr := media.EnsureThumbnail(dir, saved.RelPath); thumbErr != nil {
		_ = media.AppendError(s.paths.MediaError, "make-gallery-thumb", saved.AbsPath, thumbErr)
	}
	if containsString(images, saved.RelPath) {
		return saved.RelPath, nil
	}

	images = append(images, saved.RelPath)
	data, err := marshalImages(images)
	if err != nil {
		return "", err
	}
	cover := item.CoverImage
	if cover == "" {
		cover = saved.RelPath
	}

	_, err = s.store.Execute(ctx,
		"UPDATE gallery SET images = ?, cover_image = ?, image_count = ?, updated_at = ? WHERE id = ?",
		data,
		cover,
		len(images),
		time.Now().Format(time.RFC3339),
		id,
	)
	if err != nil {
		if saved.Created {
			_ = media.RemoveFile(saved.AbsPath)
			_ = media.RemoveFile(thumbAbs)
			_ = media.AppendError(s.paths.MediaError, "rollback-gallery-image", saved.AbsPath, err)
		}
		return "", fmt.Errorf("%s: %w", i18n.T("更新图集图片失败"), err)
	}
	return saved.RelPath, nil
}

func (s *Service) RemoveYuvisionImage(ctx context.Context, id int64, relPath string) error {
	item, err := s.getYuvision(ctx, id)
	if err != nil {
		return err
	}
	images, err := parseImages(item.Images)
	if err != nil {
		return err
	}
	if !containsString(images, relPath) {
		return fmt.Errorf("%s", i18n.T("图片不属于该版型"))
	}

	remaining := make([]string, 0, len(images))
	for _, image := range images {
		if image != relPath {
			remaining = append(remaining, image)
		}
	}
	data, err := marshalImages(remaining)
	if err != nil {
		return err
	}
	cover := item.CoverImage
	if cover == relPath {
		cover = ""
		if len(remaining) > 0 {
			cover = remaining[0]
		}
	}

	if _, err := s.store.Execute(ctx,
		"UPDATE youvision SET images = ?, cover_image = ?, updated_at = ? WHERE id = ?",
		data,
		cover,
		time.Now().Format(time.RFC3339),
		id,
	); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("移除版型图片失败"), err)
	}

	base := filepath.Join(s.paths.Yuvision, fmt.Sprint(item.SeriesID), fmt.Sprint(item.ID))
	target := filepath.Join(base, filepath.FromSlash(relPath))
	if err := media.RemoveFile(target); err != nil {
		_ = media.AppendError(s.paths.MediaError, "remove-yuvision-image", target, err)
	}
	thumb := filepath.Join(base, filepath.FromSlash(media.ThumbnailRel(relPath)))
	if err := media.RemoveFile(thumb); err != nil {
		_ = media.AppendError(s.paths.MediaError, "remove-yuvision-thumb", thumb, err)
	}
	return nil
}

func (s *Service) RemoveGalleryImage(ctx context.Context, id int64, relPath string) error {
	item, err := s.getGallery(ctx, id)
	if err != nil {
		return err
	}
	images, err := parseImages(item.Images)
	if err != nil {
		return err
	}
	if !containsString(images, relPath) {
		return fmt.Errorf("%s", i18n.T("图片不属于该图集"))
	}

	remaining := make([]string, 0, len(images))
	for _, image := range images {
		if image != relPath {
			remaining = append(remaining, image)
		}
	}
	data, err := marshalImages(remaining)
	if err != nil {
		return err
	}
	cover := item.CoverImage
	if cover == relPath {
		cover = ""
		if len(remaining) > 0 {
			cover = remaining[0]
		}
	}

	if _, err := s.store.Execute(ctx,
		"UPDATE gallery SET images = ?, cover_image = ?, image_count = ?, updated_at = ? WHERE id = ?",
		data,
		cover,
		len(remaining),
		time.Now().Format(time.RFC3339),
		id,
	); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("移除图集图片失败"), err)
	}

	base := filepath.Join(s.paths.Gallery, fmt.Sprint(item.SeriesID), fmt.Sprint(item.ID))
	target := filepath.Join(base, filepath.FromSlash(relPath))
	if err := media.RemoveFile(target); err != nil {
		_ = media.AppendError(s.paths.MediaError, "remove-gallery-image", target, err)
	}
	thumb := filepath.Join(base, filepath.FromSlash(media.ThumbnailRel(relPath)))
	if err := media.RemoveFile(thumb); err != nil {
		_ = media.AppendError(s.paths.MediaError, "remove-gallery-thumb", thumb, err)
	}
	return nil
}

func (s *Service) getYuvision(ctx context.Context, id int64) (domain.Yuvision, error) {
	rows, err := s.store.Query(ctx, `
SELECT id, brand_id, firm_id, series_id, category_id,
       name, cover_image, images, description, created_at, updated_at
FROM youvision WHERE id = ?
`, id)
	if err != nil {
		return domain.Yuvision{}, fmt.Errorf("%s: %w", i18n.T("读取版型失败"), err)
	}
	if len(rows.Values) == 0 || len(rows.Values[0]) < 11 {
		return domain.Yuvision{}, fmt.Errorf("%s", i18n.T("版型不存在"))
	}
	values := rows.Values[0]
	return domain.Yuvision{
		ID:          valueInt64(values[0]),
		BrandID:     valueInt64(values[1]),
		FirmID:      valueInt64(values[2]),
		SeriesID:    valueInt64(values[3]),
		CategoryID:  valueInt64(values[4]),
		Name:        valueString(values[5]),
		CoverImage:  valueString(values[6]),
		Images:      valueString(values[7]),
		Description: valueString(values[8]),
		CreatedAt:   valueString(values[9]),
		UpdatedAt:   valueString(values[10]),
	}, nil
}

func (s *Service) getGallery(ctx context.Context, id int64) (domain.Gallery, error) {
	rows, err := s.store.Query(ctx, `
SELECT id, youvision_id, brand_id, firm_id, series_id, category_id,
       title, cover_image, images, image_count, created_at, updated_at
FROM gallery WHERE id = ?
`, id)
	if err != nil {
		return domain.Gallery{}, fmt.Errorf("%s: %w", i18n.T("读取图集失败"), err)
	}
	if len(rows.Values) == 0 || len(rows.Values[0]) < 12 {
		return domain.Gallery{}, fmt.Errorf("%s", i18n.T("图集不存在"))
	}
	values := rows.Values[0]
	return domain.Gallery{
		ID:          valueInt64(values[0]),
		YouvisionID: valueInt64(values[1]),
		BrandID:     valueInt64(values[2]),
		FirmID:      valueInt64(values[3]),
		SeriesID:    valueInt64(values[4]),
		CategoryID:  valueInt64(values[5]),
		Title:       valueString(values[6]),
		CoverImage:  valueString(values[7]),
		Images:      valueString(values[8]),
		ImageCount:  int(valueInt64(values[9])),
		CreatedAt:   valueString(values[10]),
		UpdatedAt:   valueString(values[11]),
	}, nil
}

func (s *Service) exists(ctx context.Context, sql string, args ...any) (bool, error) {
	rows, err := s.store.Query(ctx, sql, args...)
	if err != nil {
		return false, err
	}
	return len(rows.Values) > 0, nil
}

func parseImages(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return []string{}, nil
	}
	var images []string
	if err := json.Unmarshal([]byte(value), &images); err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("解析图片列表失败"), err)
	}
	return images, nil
}

func marshalImages(images []string) (string, error) {
	data, err := json.Marshal(images)
	if err != nil {
		return "", fmt.Errorf("%s: %w", i18n.T("序列化图片列表失败"), err)
	}
	return string(data), nil
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
