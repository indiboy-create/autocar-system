package service

import (
	"context"
	"fmt"
	"strings"

	"autocarsystem/internal/domain"
	"autocarsystem/internal/i18n"
)

func (s *Service) SearchGalleries(ctx context.Context, filter domain.GalleryFilter, limit int) ([]domain.Gallery, error) {
	where := []string{"1 = 1"}
	args := make([]any, 0, 8)

	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		pattern := "%" + keyword + "%"
		where = append(where, "(title LIKE ? OR CAST(id AS TEXT) LIKE ? OR CAST(youvision_id AS TEXT) LIKE ?)")
		args = append(args, pattern, pattern, pattern)
	}
	if filter.BrandID > 0 {
		where = append(where, "brand_id = ?")
		args = append(args, filter.BrandID)
	}
	if filter.FirmID > 0 {
		where = append(where, "firm_id = ?")
		args = append(args, filter.FirmID)
	}
	if filter.SeriesID > 0 {
		where = append(where, "series_id = ?")
		args = append(args, filter.SeriesID)
	}
	if filter.CategoryID > 0 {
		where = append(where, "category_id = ?")
		args = append(args, filter.CategoryID)
	}
	if filter.YouvisionID > 0 {
		where = append(where, "yuvision_id = ?")
		args = append(args, filter.YouvisionID)
	}

	order := "updated_at DESC, id DESC"
	switch filter.Sort {
	case "created_desc":
		order = "created_at DESC, id DESC"
	case "title_asc":
		order = "title COLLATE NOCASE ASC, id DESC"
	case "id_desc":
		order = "id DESC"
	case "image_count_desc":
		order = "image_count DESC, id DESC"
	}

	args = append(args, limitRows(limit))
	sql := `
SELECT id, youvision_id, brand_id, firm_id, series_id, category_id,
       title, cover_image, images, image_count, created_at, updated_at
FROM gallery
WHERE ` + strings.Join(where, " AND ") + `
ORDER BY ` + order + `
LIMIT ?`

	rows, err := s.store.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("搜索图集失败"), err)
	}

	items := make([]domain.Gallery, 0, len(rows.Values))
	for _, v := range rows.Values {
		if len(v) < 12 {
			continue
		}
		items = append(items, domain.Gallery{
			ID:          valueInt64(v[0]),
			YouvisionID: valueInt64(v[1]),
			BrandID:     valueInt64(v[2]),
			FirmID:      valueInt64(v[3]),
			SeriesID:    valueInt64(v[4]),
			CategoryID:  valueInt64(v[5]),
			Title:       valueString(v[6]),
			CoverImage:  valueString(v[7]),
			Images:      valueString(v[8]),
			ImageCount:  int(valueInt64(v[9])),
			CreatedAt:   valueString(v[10]),
			UpdatedAt:   valueString(v[11]),
		})
	}
	return items, nil
}
