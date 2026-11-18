package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"autocarsystem/internal/domain"
)

func (s *Service) ListBrands(ctx context.Context, limit int) ([]domain.Brand, error) {
	rows, err := s.store.Query(ctx, `
SELECT id, name, firstletter, logo_url, logo_local_path,
       state, havenewenergy, api_sort, collected_at, user_updated_at, media_updated_at
FROM brand
ORDER BY id
LIMIT ?
`, limit)
	if err != nil {
		return nil, fmt.Errorf("读取品牌失败: %w", err)
	}

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
	return items, nil
}

func (s *Service) UpdateBrand(ctx context.Context, id int64, in domain.BrandEdit) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return fmt.Errorf("品牌名称不能为空")
	}
	_, err := s.store.Execute(ctx, `
UPDATE brand
SET name = ?, firstletter = ?, logo_url = ?, logo_local_path = ?,
    state = ?, havenewenergy = ?, api_sort = ?, user_updated_at = ?
WHERE id = ?
`, in.Name, in.FirstLetter, in.LogoURL, in.LogoLocalPath,
		in.State, in.HaveNewEnergy, in.APISort, nowString(), id)
	if err != nil {
		return fmt.Errorf("更新品牌失败: %w", err)
	}
	return nil
}

func (s *Service) ListFirms(ctx context.Context, limit int) ([]domain.Firm, error) {
	rows, err := s.store.Query(ctx, `
SELECT id, brand_id, name, py, link, logo_url, logo_local_path,
       collected_at, user_updated_at, media_updated_at
FROM firm
ORDER BY brand_id, id
LIMIT ?
`, limit)
	if err != nil {
		return nil, fmt.Errorf("读取厂商失败: %w", err)
	}

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
	return items, nil
}

func (s *Service) UpdateFirm(ctx context.Context, id int64, in domain.FirmEdit) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return fmt.Errorf("厂商名称不能为空")
	}
	_, err := s.store.Execute(ctx, `
UPDATE firm
SET name = ?, py = ?, link = ?, logo_url = ?, logo_local_path = ?, user_updated_at = ?
WHERE id = ?
`, in.Name, in.PY, in.Link, in.LogoURL, in.LogoLocalPath, nowString(), id)
	if err != nil {
		return fmt.Errorf("更新厂商失败: %w", err)
	}
	return nil
}

func (s *Service) ListSeries(ctx context.Context, limit int) ([]domain.Series, error) {
	rows, err := s.store.Query(ctx, `
SELECT id, brand_id, firm_id, name, group_name, state, price_min, price_max,
       place, py, logo_url, logo_local_path, level_name, newenergy,
       is_foreign_car, is_top, fuel_types, cartypename, tag1,
       "rank", uv_rank, hot_rank, collected_at, user_updated_at, media_updated_at
FROM series
ORDER BY brand_id, firm_id, id
LIMIT ?
`, limit)
	if err != nil {
		return nil, fmt.Errorf("读取车系失败: %w", err)
	}

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
	return items, nil
}

func (s *Service) UpdateSeries(ctx context.Context, id int64, in domain.SeriesEdit) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return fmt.Errorf("车系名称不能为空")
	}
	_, err := s.store.Execute(ctx, `
UPDATE series
SET name = ?, group_name = ?, state = ?, price_min = ?, price_max = ?,
    place = ?, py = ?, logo_url = ?, logo_local_path = ?, level_name = ?,
    newenergy = ?, is_foreign_car = ?, is_top = ?, fuel_types = ?, cartypename = ?,
    tag1 = ?, "rank" = ?, uv_rank = ?, hot_rank = ?, user_updated_at = ?
WHERE id = ?
`, in.Name, in.GroupName, in.State, in.PriceMin, in.PriceMax,
		in.Place, in.PY, in.LogoURL, in.LogoLocalPath, in.LevelName,
		in.NewEnergy, in.IsForeignCar, in.IsTop, in.FuelTypes, in.CartypeName,
		in.Tag1, in.Rank, in.UVRank, in.HotRank, nowString(), id)
	if err != nil {
		return fmt.Errorf("更新车系失败: %w", err)
	}
	return nil
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
		return nil, fmt.Errorf("读取年款失败: %w", err)
	}

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
	return items, nil
}

func (s *Service) UpdateSpec(ctx context.Context, id int64, in domain.SpecEdit) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return fmt.Errorf("年款名称不能为空")
	}
	_, err := s.store.Execute(ctx, `
UPDATE spec
SET group_name = ?, name = ?, series_name = ?, model_year = ?, body_form = ?,
    price_min = ?, price_max = ?, news_price = ?, news_id = ?, dealer_id = ?,
    sale_type = ?, image_url = ?, image_local_path = ?, image_sha256 = ?,
    white_image_url = ?, white_image_local_path = ?, white_image_sha256 = ?,
    user_updated_at = ?
WHERE id = ?
`, in.GroupName, in.Name, in.SeriesName, in.ModelYear, in.BodyForm,
		in.PriceMin, in.PriceMax, in.NewsPrice, in.NewsID, in.DealerID,
		in.SaleType, in.ImageURL, in.ImageLocalPath, in.ImageSHA256,
		in.WhiteImageURL, in.WhiteImageLocalPath, in.WhiteImageSHA256,
		nowString(), id)
	if err != nil {
		return fmt.Errorf("更新年款失败: %w", err)
	}
	return nil
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
