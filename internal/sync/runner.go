package sync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"autocarsystem/internal/config"
	"autocarsystem/internal/service"
)

type Config struct {
	BrandURL        string
	SeriesURLFormat string
	SpecURLFormat   string
	Timeout         time.Duration
	Retry           int
	Delay           time.Duration
	SaveRaw         bool
	UserAgent       string
}

type Progress struct {
	Stage     string
	Message   string
	Processed int
	Total     int
}

type ProgressFunc func(Progress)

func DefaultConfig() Config {
	return Config{
		BrandURL:        "https://car-web-m.autohome.com.cn/car/getbrandmenu?_appid=m",
		SeriesURLFormat: "https://car-web-m.autohome.com.cn/car/series/getserieslistbybrandid?_appid=m&brandid=%d",
		SpecURLFormat:   "https://autoapi.autohome.com.cn/autohome-www-api/inquirylayer/getSpecGroupedInfoListBySeriesId?_appId=cms&cityId=1&seriesId=%d",
		Timeout:         20 * time.Second,
		Retry:           3,
		Delay:           500 * time.Millisecond,
		SaveRaw:         true,
		UserAgent:       "Mozilla/5.0 autoCarSystem",
	}
}

type seriesIdentity struct {
	ID      int64
	BrandID int64
	FirmID  int64
}

type runner struct {
	cfg      Config
	svc      *service.Service
	paths    config.Paths
	state    State
	client   *http.Client
	progress ProgressFunc
}

func Run(ctx context.Context, svc *service.Service, paths config.Paths, progress ProgressFunc) error {
	return RunWithConfig(ctx, svc, paths, DefaultConfig(), progress)
}

func RunWithConfig(ctx context.Context, svc *service.Service, paths config.Paths, cfg Config, progress ProgressFunc) error {
	r := &runner{
		cfg:      cfg,
		svc:      svc,
		paths:    paths,
		client:   &http.Client{Timeout: cfg.Timeout},
		progress: progress,
	}
	state, err := LoadState(paths.StatePath)
	if err != nil {
		return err
	}
	r.state = state
	r.state.Stage = "brand"
	r.state.LastSyncAt = ""
	r.state.Stats = NewState().Stats
	r.saveState()

	if err := r.syncBrands(ctx); err != nil {
		r.fail(err)
		return err
	}
	if err := r.syncFirmsAndSeries(ctx); err != nil {
		r.fail(err)
		return err
	}
	series, err := r.svc.ListSeries(ctx, 100000)
	if err != nil {
		r.fail(err)
		return err
	}
	identities := make([]seriesIdentity, 0, len(series))
	for _, item := range series {
		identities = append(identities, seriesIdentity{ID: item.ID, BrandID: item.BrandID, FirmID: item.FirmID})
	}
	if err := r.syncSpecs(ctx, identities); err != nil {
		r.fail(err)
		return err
	}

	r.state.Stage = "done"
	r.state.LastSyncAt = time.Now().Format(time.RFC3339)
	r.saveState()
	r.report("done", "同步完成", 0, 0)
	return nil
}

func (r *runner) syncBrands(ctx context.Context) error {
	r.report("brand", "同步品牌", 0, 0)
	root, err := r.fetch(ctx, "brand", "all", r.cfg.BrandURL)
	if err != nil {
		return fmt.Errorf("获取品牌失败: %w", err)
	}
	result := asMap(root["result"])
	groups := asArray(result["brand"])
	total := 0
	for _, group := range groups {
		total += len(asArray(asMap(group)["brandlist"]))
	}
	processed := 0

	for _, groupValue := range groups {
		group := asMap(groupValue)
		firstLetter := toText(group["firstletter"])
		for _, itemValue := range asArray(group["brandlist"]) {
			item := asMap(itemValue)
			processed++
			id := toInt64(item["id"])
			name := toText(item["name"])
			if id <= 0 || name == "" {
				r.conflict("brand", id, "invalid_brand", "品牌 ID 或名称无效")
				r.bump("brand", false, true)
				continue
			}
			inserted, err := r.svc.InsertBrand(ctx, service.BrandInput{
				ID:            id,
				Name:          name,
				FirstLetter:   firstLetter,
				LogoURL:       toText(item["logo"]),
				State:         toInt64(item["state"]),
				HaveNewEnergy: toInt64(item["havenewenergy"]),
			})
			if err != nil {
				r.handleInsertError("brand", id, err)
				r.bump("brand", false, true)
				continue
			}
			r.bump("brand", inserted, false)
		}
	}
	r.report("brand", "品牌同步完成", processed, total)
	r.saveState()
	return nil
}

func (r *runner) syncFirmsAndSeries(ctx context.Context) error {
	for _, brandID := range r.brandIDs(ctx) {
		url := fmt.Sprintf(r.cfg.SeriesURLFormat, brandID)
		root, err := r.fetch(ctx, "series", fmt.Sprintf("brand-%d", brandID), url)
		if err != nil {
			r.conflict("series", brandID, "fetch_failed", err.Error())
			continue
		}
		result := asMap(root["result"])
		if err := r.insertFirms(ctx, brandID, asArray(result["fcttagdto"])); err != nil {
			return err
		}
		if err := r.insertSeries(ctx, brandID, asArray(result["groupList"])); err != nil {
			return err
		}
	}
	r.saveState()
	return nil
}

func (r *runner) insertFirms(ctx context.Context, brandID int64, values []any) error {
	for _, value := range values {
		item := asMap(value)
		id := toInt64(item["id"])
		name := toText(item["name"])
		if id <= 0 || name == "" {
			r.conflict("firm", id, "invalid_firm", "厂商 ID 或名称无效")
			r.bump("firm", false, true)
			continue
		}
		inserted, err := r.svc.InsertFirm(ctx, service.FirmInput{
			ID:    id,
			Brand: brandID,
			Name:  name,
			Link:  toText(item["link"]),
		})
		if err != nil {
			r.handleInsertError("firm", id, err)
			r.bump("firm", false, true)
			continue
		}
		r.bump("firm", inserted, false)
	}
	return nil
}

func (r *runner) insertSeries(ctx context.Context, brandID int64, groups []any) error {
	for _, groupValue := range groups {
		group := asMap(groupValue)
		groupID := toInt64(group["id"])
		groupName := toText(group["name"])
		for _, value := range asArray(group["serieslist"]) {
			item := asMap(value)
			id := toInt64(item["seriesid"])
			name := toText(item["seriesName"])
			firmID := toInt64(item["fctid"])
			if id <= 0 || name == "" {
				r.conflict("series", id, "invalid_series", "车系 ID 或名称无效")
				r.bump("series", false, true)
				continue
			}
			if firmID <= 0 {
				r.conflict("series", id, "missing_firm", "车系缺少有效 fctid")
				r.bump("series", false, true)
				continue
			}
			r.ensureFirm(ctx, brandID, firmID, toText(item["fctname"]))

			inserted, err := r.svc.InsertSeries(ctx, service.SeriesInput{
				ID:           id,
				BrandID:      brandID,
				FirmID:       firmID,
				GroupID:      groupID,
				GroupName:    groupName,
				Name:         name,
				State:        toInt64(item["seriesState"]),
				PriceMin:     toFloat(item["seriesPriceMin"]),
				PriceMax:     toFloat(item["seriesPriceMax"]),
				Place:        toText(item["seriesplace"]),
				PY:           toText(item["seriespy"]),
				LogoURL:      toText(item["pnglogo"]),
				LevelID:      toInt64(item["levelId"]),
				LevelName:    toText(item["levelName"]),
				NewEnergy:    toInt64(item["newenergy"]),
				IsForeignCar: toInt64(item["isForeignCar"]),
				Rank:         toInt64(item["rank"]),
				UVRank:       toInt64(item["uvRank"]),
				HotRank:      toInt64(item["hotRank"]),
				FuelTypes:    joinTextArray(item["fuelnameList"]),
				CartypeName:  toText(item["cartypename"]),
				Tag1:         toText(item["seriesTag1"]),
			})
			if err != nil {
				r.handleInsertError("series", id, err)
				r.bump("series", false, true)
				continue
			}
			r.bump("series", inserted, false)
		}
	}
	return nil
}

func (r *runner) ensureFirm(ctx context.Context, brandID int64, firmID int64, name string) {
	if firmID <= 0 || name == "" {
		return
	}
	if _, err := r.svc.InsertFirm(ctx, service.FirmInput{ID: firmID, Brand: brandID, Name: name}); err != nil {
		r.conflict("firm", firmID, "insert_failed", err.Error())
	}
}

func (r *runner) syncSpecs(ctx context.Context, series []seriesIdentity) error {
	for index, item := range series {
		url := fmt.Sprintf(r.cfg.SpecURLFormat, item.ID)
		root, err := r.fetch(ctx, "spec", fmt.Sprintf("series-%d", item.ID), url)
		if err != nil {
			r.conflict("spec", item.ID, "fetch_failed", err.Error())
			continue
		}
		for _, groupValue := range asArray(root["result"]) {
			group := asMap(groupValue)
			groupName := toText(group["volumeGroupName"])
			for _, specValue := range asArray(group["specList"]) {
				spec := asMap(specValue)
				id := toInt64(spec["specId"])
				name := toText(spec["specName"])
				if id <= 0 || name == "" {
					r.conflict("spec", id, "invalid_spec", "年款 ID 或名称无效")
					r.bump("spec", false, true)
					continue
				}
				inserted, err := r.svc.InsertSpec(ctx, service.SpecInput{
					ID:            id,
					SeriesID:      item.ID,
					GroupName:     groupName,
					Name:          name,
					SeriesName:    toText(spec["seriesName"]),
					ModelYear:     extractModelYear(name),
					PriceMin:      toFloat(spec["minOriginalPrice"]),
					PriceMax:      toFloat(spec["maxOriginalPrice"]),
					NewsPrice:     toFloat(spec["newsPrice"]),
					NewsID:        toInt64(spec["newsId"]),
					DealerID:      toInt64(spec["dealerId"]),
					SaleType:      toText(spec["saleType"]),
					ImageURL:      toText(spec["imageUrl"]),
					WhiteImageURL: toText(spec["specWhiteImgUrl"]),
				})
				if err != nil {
					r.handleInsertError("spec", id, err)
					r.bump("spec", false, true)
					continue
				}
				r.bump("spec", inserted, false)
			}
		}
		r.report("spec", fmt.Sprintf("年款同步 %d/%d", index+1, len(series)), index+1, len(series))
	}
	r.saveState()
	return nil
}

func (r *runner) brandIDs(ctx context.Context) []int64 {
	rows, err := r.svc.ListBrands(ctx, 100000)
	if err != nil {
		r.conflict("brand", 0, "list_failed", err.Error())
		return nil
	}
	ids := make([]int64, 0, len(rows))
	for _, item := range rows {
		ids = append(ids, item.ID)
	}
	return ids
}

func (r *runner) fetch(ctx context.Context, endpoint string, key string, url string) (map[string]any, error) {
	var lastErr error
	for attempt := 0; attempt < r.cfg.Retry; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", r.cfg.UserAgent)
		resp, err := r.client.Do(req)
		if err != nil {
			lastErr = err
			r.sleep(ctx)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = err
			r.sleep(ctx)
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			r.sleep(ctx)
			continue
		}
		if r.cfg.SaveRaw {
			_ = saveRaw(r.paths.RawDir, endpoint, key, body)
		}
		var root map[string]any
		if err := json.Unmarshal(body, &root); err != nil {
			lastErr = fmt.Errorf("解析 JSON 失败: %w", err)
			r.sleep(ctx)
			continue
		}
		if code := toInt64(root["returncode"]); code != 0 {
			return nil, fmt.Errorf("接口返回错误码 %d: %s", code, toText(root["message"]))
		}
		return root, nil
	}
	if lastErr == nil {
		lastErr = errors.New("请求失败")
	}
	return nil, lastErr
}

func (r *runner) sleep(ctx context.Context) {
	if r.cfg.Delay <= 0 {
		return
	}
	select {
	case <-ctx.Done():
	case <-time.After(r.cfg.Delay):
	}
}

func saveRaw(root string, endpoint string, key string, data []byte) error {
	dir := filepath.Join(root, endpoint)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	cleanKey := strings.NewReplacer("/", "_", "\\", "_", ":", "_").Replace(key)
	return os.WriteFile(filepath.Join(dir, cleanKey+".json"), data, 0o644)
}

func (r *runner) bump(stage string, inserted bool, failed bool) {
	stats := r.state.Stats[stage]
	stats.Processed++
	if failed {
		stats.Failed++
	} else if inserted {
		stats.Inserted++
	} else {
		stats.Skipped++
	}
	r.state.Stats[stage] = stats
}

func (r *runner) report(stage string, message string, processed int, total int) {
	if r.progress == nil {
		return
	}
	r.progress(Progress{Stage: stage, Message: message, Processed: processed, Total: total})
}

func (r *runner) saveState() {
	_ = SaveState(r.paths.StatePath, r.state)
}

func (r *runner) fail(err error) {
	r.state.Stage = "failed"
	r.saveState()
	r.report("failed", err.Error(), 0, 0)
}

func (r *runner) conflict(entity string, id int64, reason string, message string) {
	_ = AppendConflict(r.paths.Conflicts, Conflict{
		Entity:    entity,
		ID:        fmt.Sprint(id),
		Reason:    reason,
		Message:   message,
		CreatedAt: time.Now().Format(time.RFC3339),
	})
}

func (r *runner) handleInsertError(entity string, id int64, err error) {
	reason := "insert_failed"
	if errors.Is(err, service.ErrIdentityConflict) {
		reason = "identity_conflict"
	}
	r.conflict(entity, id, reason, err.Error())
}

func asMap(value any) map[string]any {
	if result, ok := value.(map[string]any); ok {
		return result
	}
	return map[string]any{}
}

func asArray(value any) []any {
	if result, ok := value.([]any); ok {
		return result
	}
	return []any{}
}

func toText(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case json.Number:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

func toInt64(value any) int64 {
	switch v := value.(type) {
	case nil:
		return 0
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case int64:
		return v
	case bool:
		if v {
			return 1
		}
		return 0
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err == nil {
			return n
		}
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err == nil {
			return int64(f)
		}
		return 0
	case json.Number:
		n, err := v.Int64()
		if err == nil {
			return n
		}
		f, err := v.Float64()
		if err == nil {
			return int64(f)
		}
		return 0
	default:
		return 0
	}
}

func toFloat(value any) float64 {
	switch v := value.(type) {
	case nil:
		return 0
	case float64:
		return v
	case int64:
		return float64(v)
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err == nil {
			return f
		}
		return 0
	case json.Number:
		f, err := v.Float64()
		if err == nil {
			return f
		}
		return 0
	default:
		return 0
	}
}

func joinTextArray(value any) string {
	items := asArray(value)
	parts := make([]string, 0, len(items))
	for _, item := range items {
		if text := toText(item); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, ",")
}

var yearPattern = regexp.MustCompile(`(19|20)\d{2}款`)

func extractModelYear(name string) string {
	match := yearPattern.FindString(name)
	return strings.TrimSuffix(match, "款")
}
