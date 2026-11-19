package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"autocarsystem/internal/domain"
	"autocarsystem/internal/i18n"

	"github.com/egoist/mygo/plugins/sqlite"
)

func (s *Service) ListBrands(ctx context.Context, limit int) ([]domain.Brand, error) {
	rows, err := s.store.Query(ctx, `
SELECT id, name, firstletter, logo_url, logo_local_path,
       state, havenewenergy, api_sort, collected_at, user_updated_at, media_updated_at
FROM brand
ORDER BY id
LIMIT ?
`, limitRows(limit))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("读取品牌失败"), err)
	}
	return scanBrandRows(rows), nil
}

func scanBrandRows(rows sqlite.Rows) []domain.Brand {
	items := make([]domain.Brand, 0, len(rows.Values))
	for _, v := range rows.Values {
		if len(v) < 11 {
			continue
		}
		items = append(items, domain.Brand{
			ID:             valueInt64(v[0]),
			Name:           valueString(v[1]),
			FirstLetter:    valueString(v[2]),
			LogoURL:        valueString(v[3]),
			LogoLocalPath:  valueString(v[4]),
			State:          valueInt64(v[5]),
			HaveNewEnergy:  valueInt64(v[6]),
			APISort:        valueInt64(v[7]),
			CollectedAt:    valueString(v[8]),
			UserUpdatedAt:  valueString(v[9]),
			MediaUpdatedAt: valueString(v[10]),
		})
	}
	return items
}

// ListBrandsPage 返回从 offset 开始的一页品牌。多取一条用于判断是否还有下一页，
// 避免额外的 COUNT 查询；more 为 true 表示后面还有数据。
func (s *Service) ListBrandsPage(ctx context.Context, offset, limit int) ([]domain.Brand, bool, error) {
	if limit <= 0 {
		limit = 60
	}
	rows, err := s.store.Query(ctx, `
SELECT id, name, firstletter, logo_url, logo_local_path,
       state, havenewenergy, api_sort, collected_at, user_updated_at, media_updated_at
FROM brand
ORDER BY id
LIMIT ? OFFSET ?
`, limit+1, offset)
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", i18n.T("读取品牌失败"), err)
	}

	items := scanBrandRows(rows)
	more := len(items) > limit
	if more {
		items = items[:limit]
	}
	return items, more, nil
}

// SearchBrandsPage 按关键词在数据库中搜索品牌，返回从 offset 开始的一页。
// 支持名称、首字母与精确 ID 匹配，多取一条判断是否有下一页。
func (s *Service) SearchBrandsPage(ctx context.Context, keyword string, offset, limit int) ([]domain.Brand, bool, error) {
	keyword = strings.TrimSpace(keyword)
	if limit <= 0 {
		limit = 60
	}
	const columns = `
SELECT id, name, firstletter, logo_url, logo_local_path,
       state, havenewenergy, api_sort, collected_at, user_updated_at, media_updated_at
FROM brand`
	const order = `
ORDER BY id
LIMIT ? OFFSET ?`

	var rows sqlite.Rows
	var err error
	if keyword == "" {
		rows, err = s.store.Query(ctx, columns+order, limit+1, offset)
	} else if id, parseErr := strconv.ParseInt(keyword, 10, 64); parseErr == nil {
		rows, err = s.store.Query(ctx, columns+`
WHERE id = ?`+order, id, limit+1, offset)
	} else {
		like := "%" + keyword + "%"
		rows, err = s.store.Query(ctx, columns+`
WHERE name LIKE ? OR firstletter LIKE ?`+order, like, like, limit+1, offset)
	}
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", i18n.T("搜索品牌失败"), err)
	}
	items := scanBrandRows(rows)
	more := len(items) > limit
	if more {
		items = items[:limit]
	}
	return items, more, nil
}

func (s *Service) ListFirms(ctx context.Context, limit int) ([]domain.Firm, error) {
	rows, err := s.store.Query(ctx, `
SELECT id, brand_id, name, py, link, logo_url, logo_local_path,
       collected_at, user_updated_at, media_updated_at
FROM firm
ORDER BY brand_id, id
LIMIT ?
`, limitRows(limit))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("读取厂商失败"), err)
	}

	return scanFirmRows(rows), nil
}

func scanFirmRows(rows sqlite.Rows) []domain.Firm {
	items := make([]domain.Firm, 0, len(rows.Values))
	for _, v := range rows.Values {
		if len(v) < 10 {
			continue
		}
		items = append(items, domain.Firm{
			ID:             valueInt64(v[0]),
			BrandID:        valueInt64(v[1]),
			Name:           valueString(v[2]),
			PY:             valueString(v[3]),
			Link:           valueString(v[4]),
			LogoURL:        valueString(v[5]),
			LogoLocalPath:  valueString(v[6]),
			CollectedAt:    valueString(v[7]),
			UserUpdatedAt:  valueString(v[8]),
			MediaUpdatedAt: valueString(v[9]),
		})
	}
	return items
}

// ListFirmsPage 返回从 offset 开始的一页厂商，多取一条判断是否有下一页。
func (s *Service) ListFirmsPage(ctx context.Context, offset, limit int) ([]domain.Firm, bool, error) {
	if limit <= 0 {
		limit = 60
	}
	rows, err := s.store.Query(ctx, `
SELECT id, brand_id, name, py, link, logo_url, logo_local_path,
       collected_at, user_updated_at, media_updated_at
FROM firm
ORDER BY brand_id, id
LIMIT ? OFFSET ?
`, limit+1, offset)
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", i18n.T("读取厂商失败"), err)
	}

	items := scanFirmRows(rows)
	more := len(items) > limit
	if more {
		items = items[:limit]
	}
	return items, more, nil
}

// SearchFirmsPage 按关键词在数据库中搜索厂商，支持名称、拼音与精确 ID 匹配。
func (s *Service) SearchFirmsPage(ctx context.Context, keyword string, offset, limit int) ([]domain.Firm, bool, error) {
	keyword = strings.TrimSpace(keyword)
	if limit <= 0 {
		limit = 60
	}
	const columns = `
SELECT id, brand_id, name, py, link, logo_url, logo_local_path,
       collected_at, user_updated_at, media_updated_at
FROM firm`
	const order = `
ORDER BY brand_id, id
LIMIT ? OFFSET ?`
	var rows sqlite.Rows
	var err error
	if keyword == "" {
		rows, err = s.store.Query(ctx, columns+order, limit+1, offset)
	} else if id, parseErr := strconv.ParseInt(keyword, 10, 64); parseErr == nil {
		rows, err = s.store.Query(ctx, columns+`
WHERE id = ?`+order, id, limit+1, offset)
	} else {
		like := "%" + keyword + "%"
		rows, err = s.store.Query(ctx, columns+`
WHERE name LIKE ? OR py LIKE ?`+order, like, like, limit+1, offset)
	}
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", i18n.T("搜索厂商失败"), err)
	}
	items := scanFirmRows(rows)
	more := len(items) > limit
	if more {
		items = items[:limit]
	}
	return items, more, nil
}

func (s *Service) ListSeries(ctx context.Context, limit int) ([]domain.Series, error) {
	rows, err := s.store.Query(ctx, seriesQuerySQL+`
ORDER BY brand_id, firm_id, id
LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("读取车系失败"), err)
	}
	return scanSeriesRows(rows), nil
}

// SearchSeries 在数据库中按关键词搜索车系，支持名称、拼音、车组名称和 ID 匹配。
// limit 控制返回数量上限。
func (s *Service) SearchSeries(ctx context.Context, keyword string, limit int) ([]domain.Series, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return s.ListSeries(ctx, limit)
	}

	like := "%" + keyword + "%"

	// 判断是否纯数字，数字时按精确 ID 匹配
	var idMatch int64
	if n, err := fmt.Sscanf(keyword, "%d", &idMatch); err == nil && n == 1 && fmt.Sprintf("%d", idMatch) == keyword {
		rows, err := s.store.Query(ctx, seriesQuerySQL+`
WHERE id = ?
ORDER BY brand_id, firm_id, id
LIMIT ?`, idMatch, limit)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", i18n.T("搜索车系失败"), err)
		}
		return scanSeriesRows(rows), nil
	}

	rows, err := s.store.Query(ctx, seriesQuerySQL+`
WHERE name LIKE ? OR py LIKE ? OR group_name LIKE ?
ORDER BY brand_id, firm_id, id
LIMIT ?`, like, like, like, limit)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("搜索车系失败"), err)
	}
	return scanSeriesRows(rows), nil
}

// SearchSeriesPage 按关键词在数据库中搜索车系，返回从 offset 开始的一页。
// 支持名称、拼音、车组名称与精确 ID 匹配，多取一条判断是否有下一页。
func (s *Service) SearchSeriesPage(ctx context.Context, keyword string, offset, limit int) ([]domain.Series, bool, error) {
	keyword = strings.TrimSpace(keyword)
	if limit <= 0 {
		limit = 60
	}
	const order = `
ORDER BY brand_id, firm_id, id
LIMIT ? OFFSET ?`
	var rows sqlite.Rows
	var err error
	if keyword == "" {
		rows, err = s.store.Query(ctx, seriesQuerySQL+order, limit+1, offset)
	} else if id, parseErr := strconv.ParseInt(keyword, 10, 64); parseErr == nil {
		rows, err = s.store.Query(ctx, seriesQuerySQL+`
WHERE id = ?`+order, id, limit+1, offset)
	} else {
		like := "%" + keyword + "%"
		rows, err = s.store.Query(ctx, seriesQuerySQL+`
WHERE name LIKE ? OR py LIKE ? OR group_name LIKE ?`+order, like, like, like, limit+1, offset)
	}
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", i18n.T("搜索车系失败"), err)
	}
	items := scanSeriesRows(rows)
	more := len(items) > limit
	if more {
		items = items[:limit]
	}
	return items, more, nil
}

const seriesQuerySQL = `
SELECT id, brand_id, firm_id, name, group_name, state, price_min, price_max,
       place, py, logo_url, logo_local_path, level_name, newenergy,
       is_foreign_car, is_top, fuel_types, cartypename, tag1,
       "rank", uv_rank, hot_rank, collected_at, user_updated_at, media_updated_at
FROM series`

func scanSeriesRows(rows sqlite.Rows) []domain.Series {
	items := make([]domain.Series, 0, len(rows.Values))
	for _, v := range rows.Values {
		if len(v) < 25 {
			continue
		}
		items = append(items, domain.Series{
			ID:             valueInt64(v[0]),
			BrandID:        valueInt64(v[1]),
			FirmID:         valueInt64(v[2]),
			Name:           valueString(v[3]),
			GroupName:      valueString(v[4]),
			State:          valueInt64(v[5]),
			PriceMin:       valueFloat64(v[6]),
			PriceMax:       valueFloat64(v[7]),
			Place:          valueString(v[8]),
			PY:             valueString(v[9]),
			LogoURL:        valueString(v[10]),
			LogoLocalPath:  valueString(v[11]),
			LevelName:      valueString(v[12]),
			NewEnergy:      valueInt64(v[13]),
			IsForeignCar:   valueInt64(v[14]),
			IsTop:          valueInt64(v[15]),
			FuelTypes:      valueString(v[16]),
			CartypeName:    valueString(v[17]),
			Tag1:           valueString(v[18]),
			Rank:           valueInt64(v[19]),
			UVRank:         valueInt64(v[20]),
			HotRank:        valueInt64(v[21]),
			CollectedAt:    valueString(v[22]),
			UserUpdatedAt:  valueString(v[23]),
			MediaUpdatedAt: valueString(v[24]),
		})
	}
	return items
}

// SeriesFilter 定义筛选查询参数。
type SeriesFilter struct {
	BrandID  int64
	FirmID   int64
	SeriesID int64 // 车型（车系）ID，0 表示不限
	Level    string
	Energy   string // "", "新能源", "燃油"
	Keyword  string // 匹配车系名称 / 拼音 / 车组，纯数字时按 ID 精确匹配
	Limit    int
}

// FilterSeries 按品牌、厂商、级别、能源类型在数据库中筛选车系。
func (s *Service) FilterSeries(ctx context.Context, f SeriesFilter) ([]domain.Series, error) {
	var (
		conditions []string
		args       []any
	)

	if f.BrandID != 0 {
		conditions = append(conditions, "brand_id = ?")
		args = append(args, f.BrandID)
	}
	if f.FirmID != 0 {
		conditions = append(conditions, "firm_id = ?")
		args = append(args, f.FirmID)
	}
	if f.SeriesID != 0 {
		conditions = append(conditions, "id = ?")
		args = append(args, f.SeriesID)
	}
	if level := strings.TrimSpace(f.Level); level != "" {
		conditions = append(conditions, "LOWER(level_name) LIKE ?")
		args = append(args, "%"+strings.ToLower(level)+"%")
	}
	if kw := strings.TrimSpace(f.Keyword); kw != "" {
		// 纯数字按 ID 精确匹配，与 SearchSeries 的关键词行为一致。
		if id, err := strconv.ParseInt(kw, 10, 64); err == nil {
			conditions = append(conditions, "id = ?")
			args = append(args, id)
		} else {
			like := "%" + strings.ToLower(kw) + "%"
			conditions = append(conditions, "(LOWER(name) LIKE ? OR LOWER(py) LIKE ? OR LOWER(group_name) LIKE ?)")
			args = append(args, like, like, like)
		}
	}
	switch f.Energy {
	case i18n.Data("新能源"):
		conditions = append(conditions, "newenergy != 0")
	case i18n.Data("燃油"):
		conditions = append(conditions, "newenergy = 0")
	}

	sql := seriesQuerySQL
	if len(conditions) > 0 {
		sql += " WHERE " + strings.Join(conditions, " AND ")
	}
	sql += "\nORDER BY brand_id, firm_id, id\nLIMIT ?"
	args = append(args, f.Limit)

	rows, err := s.store.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("筛选车系失败"), err)
	}
	return scanSeriesRows(rows), nil
}

// ListSeriesPage 返回从 offset 开始的一页车系，多取一条判断是否有下一页。
func (s *Service) ListSeriesPage(ctx context.Context, offset, limit int) ([]domain.Series, bool, error) {
	if limit <= 0 {
		limit = 60
	}
	rows, err := s.store.Query(ctx, seriesQuerySQL+`
ORDER BY brand_id, firm_id, id
LIMIT ? OFFSET ?`, limit+1, offset)
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", i18n.T("读取车系失败"), err)
	}
	items := scanSeriesRows(rows)
	more := len(items) > limit
	if more {
		items = items[:limit]
	}
	return items, more, nil
}

func (s *Service) ListSpecs(ctx context.Context, limit int) ([]domain.Spec, error) {
	rows, err := s.store.Query(ctx, `
SELECT id, brand_id, firm_id, series_id, group_name, name, series_name,
       model_year, body_form, price_min, price_max, news_price, news_id, dealer_id,
       sale_type, image_url, image_local_path, image_sha256,
       white_image_url, white_image_local_path, white_image_sha256,
       collected_at, user_updated_at, media_updated_at
FROM spec
ORDER BY series_id, id
LIMIT ?
`, limit)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T("读取年款失败"), err)
	}

	return scanSpecRows(rows), nil
}

func scanSpecRows(rows sqlite.Rows) []domain.Spec {
	items := make([]domain.Spec, 0, len(rows.Values))
	for _, v := range rows.Values {
		if len(v) < 24 {
			continue
		}
		items = append(items, domain.Spec{
			ID:                  valueInt64(v[0]),
			BrandID:             valueInt64(v[1]),
			FirmID:              valueInt64(v[2]),
			SeriesID:            valueInt64(v[3]),
			GroupName:           valueString(v[4]),
			Name:                valueString(v[5]),
			SeriesName:          valueString(v[6]),
			ModelYear:           valueString(v[7]),
			BodyForm:            valueString(v[8]),
			PriceMin:            valueFloat64(v[9]),
			PriceMax:            valueFloat64(v[10]),
			NewsPrice:           valueFloat64(v[11]),
			NewsID:              valueInt64(v[12]),
			DealerID:            valueInt64(v[13]),
			SaleType:            valueString(v[14]),
			ImageURL:            valueString(v[15]),
			ImageLocalPath:      valueString(v[16]),
			ImageSHA256:         valueString(v[17]),
			WhiteImageURL:       valueString(v[18]),
			WhiteImageLocalPath: valueString(v[19]),
			WhiteImageSHA256:    valueString(v[20]),
			CollectedAt:         valueString(v[21]),
			UserUpdatedAt:       valueString(v[22]),
			MediaUpdatedAt:      valueString(v[23]),
		})
	}
	return items
}

// ListSpecsPage 返回从 offset 开始的一页年款，多取一条判断是否有下一页。
func (s *Service) ListSpecsPage(ctx context.Context, offset, limit int) ([]domain.Spec, bool, error) {
	if limit <= 0 {
		limit = 60
	}
	rows, err := s.store.Query(ctx, `SELECT id, brand_id, firm_id, series_id, group_name, name, series_name,
       model_year, body_form, price_min, price_max, news_price, news_id, dealer_id,
       sale_type, image_url, image_local_path, image_sha256,
       white_image_url, white_image_local_path, white_image_sha256,
       collected_at, user_updated_at, media_updated_at
FROM spec
ORDER BY brand_id, firm_id, series_id, id
LIMIT ? OFFSET ?`, limit+1, offset)
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", i18n.T("读取年款失败"), err)
	}
	items := scanSpecRows(rows)
	more := len(items) > limit
	if more {
		items = items[:limit]
	}
	return items, more, nil
}

// SearchSpecsPage 按关键词在数据库中搜索年款，支持名称、车系名、车组名、
// 年款与精确 ID 匹配，返回从 offset 开始的一页，多取一条判断是否有下一页。
func (s *Service) SearchSpecsPage(ctx context.Context, keyword string, offset, limit int) ([]domain.Spec, bool, error) {
	keyword = strings.TrimSpace(keyword)
	if limit <= 0 {
		limit = 60
	}
	const columns = `
SELECT id, brand_id, firm_id, series_id, group_name, name, series_name,
       model_year, body_form, price_min, price_max, news_price, news_id, dealer_id,
       sale_type, image_url, image_local_path, image_sha256,
       white_image_url, white_image_local_path, white_image_sha256,
       collected_at, user_updated_at, media_updated_at
FROM spec`
	const order = `
ORDER BY brand_id, firm_id, series_id, id
LIMIT ? OFFSET ?`
	var rows sqlite.Rows
	var err error
	if keyword == "" {
		rows, err = s.store.Query(ctx, columns+order, limit+1, offset)
	} else if id, parseErr := strconv.ParseInt(keyword, 10, 64); parseErr == nil {
		rows, err = s.store.Query(ctx, columns+`
WHERE id = ?`+order, id, limit+1, offset)
	} else {
		like := "%" + keyword + "%"
		rows, err = s.store.Query(ctx, columns+`
WHERE name LIKE ? OR series_name LIKE ? OR group_name LIKE ? OR model_year LIKE ?`+order,
			like, like, like, like, limit+1, offset)
	}
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", i18n.T("搜索年款失败"), err)
	}
	items := scanSpecRows(rows)
	more := len(items) > limit
	if more {
		items = items[:limit]
	}
	return items, more, nil
}

// limitRows 把非正数 limit 视为“不限条数”：SQLite 的负 LIMIT 返回全部行。
// 调用方用 0 表示不截断，避免各列表各写一个魔法上限。
func limitRows(n int) int {
	if n <= 0 {
		return -1
	}
	return n
}

func valueFloat64(v any) float64 {
	switch value := v.(type) {
	case float64:
		return value
	case int64:
		return float64(value)
	case bool:
		if value {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func nowString() string {
	return time.Now().Format(time.RFC3339)
}
