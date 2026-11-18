package service

import (
	"context"
	"errors"
	"fmt"
)

var ErrIdentityConflict = errors.New("identity conflict")

type BrandInput struct {
	ID            int64
	Name          string
	FirstLetter   string
	LogoURL       string
	State         int64
	HaveNewEnergy int64
}

type FirmInput struct {
	ID    int64
	Brand int64
	Name  string
	Link  string
}

type SeriesInput struct {
	ID           int64
	BrandID      int64
	FirmID       int64
	GroupID      int64
	GroupName    string
	Name         string
	State        int64
	PriceMin     float64
	PriceMax     float64
	Place        string
	PY           string
	LogoURL      string
	LevelID      int64
	LevelName    string
	NewEnergy    int64
	IsForeignCar int64
	Rank         int64
	UVRank       int64
	HotRank      int64
	FuelTypes    string
	CartypeName  string
	Tag1         string
}

type SpecInput struct {
	ID            int64
	BrandID       int64
	FirmID        int64
	SeriesID      int64
	GroupName     string
	Name          string
	SeriesName    string
	ModelYear     string
	BodyForm      string
	PriceMin      float64
	PriceMax      float64
	NewsPrice     float64
	NewsID        int64
	DealerID      int64
	SaleType      string
	ImageURL      string
	WhiteImageURL string
}

func (s *Service) InsertBrand(ctx context.Context, in BrandInput) (bool, error) {
	if in.ID <= 0 || in.Name == "" {
		return false, fmt.Errorf("品牌 ID 或名称无效")
	}
	if rows, err := s.store.Query(ctx, "SELECT name FROM brand WHERE id = ?", in.ID); err != nil {
		return false, err
	} else if len(rows.Values) > 0 && valueString(rows.Values[0][0]) != in.Name {
		return false, fmt.Errorf("%w: brand id=%d", ErrIdentityConflict, in.ID)
	}

	result, err := s.store.Execute(ctx, `
INSERT OR IGNORE INTO brand(
    id, name, firstletter, logo_url, state, havenewenergy, api_sort, collected_at
) VALUES (?, ?, ?, ?, ?, ?, 0, ?)
`, in.ID, in.Name, in.FirstLetter, in.LogoURL, in.State, in.HaveNewEnergy, nowString())
	if err != nil {
		return false, fmt.Errorf("插入品牌失败: %w", err)
	}
	return result.Changes > 0, nil
}

func (s *Service) InsertFirm(ctx context.Context, in FirmInput) (bool, error) {
	if in.ID <= 0 || in.Brand <= 0 || in.Name == "" {
		return false, fmt.Errorf("厂商 ID、品牌 ID 或名称无效")
	}
	if rows, err := s.store.Query(ctx, "SELECT brand_id FROM firm WHERE id = ?", in.ID); err != nil {
		return false, err
	} else if len(rows.Values) > 0 && valueInt64(rows.Values[0][0]) != in.Brand {
		return false, fmt.Errorf("%w: firm id=%d", ErrIdentityConflict, in.ID)
	}
	if ok, err := s.exists(ctx, "SELECT 1 FROM brand WHERE id = ?", in.Brand); err != nil {
		return false, err
	} else if !ok {
		return false, fmt.Errorf("品牌不存在: %d", in.Brand)
	}

	result, err := s.store.Execute(ctx, `
INSERT OR IGNORE INTO firm(
    id, brand_id, name, link, collected_at
) VALUES (?, ?, ?, ?, ?)
`, in.ID, in.Brand, in.Name, in.Link, nowString())
	if err != nil {
		return false, fmt.Errorf("插入厂商失败: %w", err)
	}
	return result.Changes > 0, nil
}

func (s *Service) InsertSeries(ctx context.Context, in SeriesInput) (bool, error) {
	if in.ID <= 0 || in.BrandID <= 0 || in.FirmID <= 0 || in.Name == "" {
		return false, fmt.Errorf("车系 ID、归属或名称无效")
	}
	if rows, err := s.store.Query(ctx, "SELECT brand_id, firm_id FROM series WHERE id = ?", in.ID); err != nil {
		return false, err
	} else if len(rows.Values) > 0 && len(rows.Values[0]) >= 2 {
		if valueInt64(rows.Values[0][0]) != in.BrandID || valueInt64(rows.Values[0][1]) != in.FirmID {
			return false, fmt.Errorf("%w: series id=%d", ErrIdentityConflict, in.ID)
		}
	}
	if ok, err := s.exists(ctx, "SELECT 1 FROM brand WHERE id = ?", in.BrandID); err != nil {
		return false, err
	} else if !ok {
		return false, fmt.Errorf("品牌不存在: %d", in.BrandID)
	}
	if ok, err := s.exists(ctx, "SELECT 1 FROM firm WHERE id = ?", in.FirmID); err != nil {
		return false, err
	} else if !ok {
		return false, fmt.Errorf("厂商不存在: %d", in.FirmID)
	}

	result, err := s.store.Execute(ctx, `
INSERT OR IGNORE INTO series(
    id, brand_id, firm_id, group_id, group_name, name, state,
    price_min, price_max, place, py, logo_url, level_id, level_name,
    newenergy, is_foreign_car, "rank", uv_rank, hot_rank,
    fuel_types, cartypename, tag1, collected_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, in.ID, in.BrandID, in.FirmID, in.GroupID, in.GroupName, in.Name, in.State,
		in.PriceMin, in.PriceMax, in.Place, in.PY, in.LogoURL, in.LevelID, in.LevelName,
		in.NewEnergy, in.IsForeignCar, in.Rank, in.UVRank, in.HotRank,
		in.FuelTypes, in.CartypeName, in.Tag1, nowString())
	if err != nil {
		return false, fmt.Errorf("插入车系失败: %w", err)
	}
	return result.Changes > 0, nil
}

func (s *Service) InsertSpec(ctx context.Context, in SpecInput) (bool, error) {
	if in.ID <= 0 || in.SeriesID <= 0 || in.Name == "" {
		return false, fmt.Errorf("年款 ID、车系 ID 或名称无效")
	}
	if rows, err := s.store.Query(ctx, "SELECT brand_id, firm_id FROM series WHERE id = ?", in.SeriesID); err != nil {
		return false, err
	} else if len(rows.Values) == 0 || len(rows.Values[0]) < 2 {
		return false, fmt.Errorf("车系不存在: %d", in.SeriesID)
	} else {
		in.BrandID = valueInt64(rows.Values[0][0])
		in.FirmID = valueInt64(rows.Values[0][1])
	}

	if rows, err := s.store.Query(ctx, "SELECT brand_id, firm_id, series_id FROM spec WHERE id = ?", in.ID); err != nil {
		return false, err
	} else if len(rows.Values) > 0 && len(rows.Values[0]) >= 3 {
		if valueInt64(rows.Values[0][0]) != in.BrandID ||
			valueInt64(rows.Values[0][1]) != in.FirmID ||
			valueInt64(rows.Values[0][2]) != in.SeriesID {
			return false, fmt.Errorf("%w: spec id=%d", ErrIdentityConflict, in.ID)
		}
	}

	result, err := s.store.Execute(ctx, `
INSERT OR IGNORE INTO spec(
    id, brand_id, firm_id, series_id, group_name, name, series_name,
    model_year, body_form, price_min, price_max, news_price, news_id,
    dealer_id, sale_type, image_url, white_image_url, collected_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, in.ID, in.BrandID, in.FirmID, in.SeriesID, in.GroupName, in.Name, in.SeriesName,
		in.ModelYear, in.BodyForm, in.PriceMin, in.PriceMax, in.NewsPrice, in.NewsID,
		in.DealerID, in.SaleType, in.ImageURL, in.WhiteImageURL, nowString())
	if err != nil {
		return false, fmt.Errorf("插入年款失败: %w", err)
	}
	return result.Changes > 0, nil
}
