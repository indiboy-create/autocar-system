package uiapp

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"autocarsystem/internal/domain"

	myui "github.com/egoist/mygo/ui"
)

func testBitmap() *myui.Bitmap {
	return myui.NewBitmap(image.NewRGBA(image.Rect(0, 0, 1, 1)))
}

func TestRefreshIncrementsGeneration(t *testing.T) {
	ta := newTestApp(t)
	if got := ta.refreshSeq; got != 1 {
		t.Fatalf("new app refreshSeq = %d, want 1", got)
	}
	ta.gallerySearch = "宝马"
	ta.Refresh()
	if got := ta.refreshSeq; got != 2 {
		t.Fatalf("after Refresh refreshSeq = %d, want 2", got)
	}
	if ta.galleryFilter().Keyword != "宝马" {
		t.Fatalf("filter keyword not snapshotted from state")
	}
}

// 过期的异步结果必须整批丢弃，不能覆盖更新的状态。
func TestApplyAsyncRefreshDropsStaleGeneration(t *testing.T) {
	ta := newTestApp(t)
	old := []domain.Gallery{{ID: 1, Title: "旧数据"}}
	fresh := []domain.Gallery{{ID: 2, Title: "新数据"}}
	ta.galleries = old
	ta.refreshSeq = 7

	result := refreshResult{galleries: fresh}

	// 世代落后：什么都不应用。
	ta.applyAsyncRefresh(result, 6)
	if len(ta.galleries) != 1 || ta.galleries[0].Title != "旧数据" {
		t.Fatalf("stale result overwrote galleries: %+v", ta.galleries)
	}

	// 世代一致：正常应用。
	ta.applyAsyncRefresh(result, 7)
	if len(ta.galleries) != 1 || ta.galleries[0].ID != 2 {
		t.Fatalf("current result not applied: %+v", ta.galleries)
	}
}

// 查询失败（字段为 nil）时保留旧数据，只上报状态；成功字段照常覆盖。
func TestApplyRefreshKeepsOldDataOnFailure(t *testing.T) {
	ta := newTestApp(t)
	ta.categories = []domain.Category{{ID: 1, Name: "旧分类"}}
	ta.galleries = []domain.Gallery{{ID: 9, Title: "旧图集"}}

	r := refreshResult{
		categories: nil,                                     // 失败
		galleries:  []domain.Gallery{{ID: 3, Title: "新图集"}}, // 成功
		errors:     []string{"读取分类失败"},
	}
	ta.applyRefresh(r, ta.refreshSeq, false)

	if len(ta.categories) != 1 || ta.categories[0].Name != "旧分类" {
		t.Fatalf("nil field cleared old categories: %+v", ta.categories)
	}
	if len(ta.galleries) != 1 || ta.galleries[0].ID != 3 {
		t.Fatalf("successful field not applied: %+v", ta.galleries)
	}
	if ta.status != "读取分类失败" {
		t.Fatalf("status = %q, want error message", ta.status)
	}
}

// 封面改为按需解码：命中后复用缓存，读不到的文件只记一次 nil，避免每帧重试。
func TestBitmapAtDecodesOnDemandAndCachesMisses(t *testing.T) {
	ta := newTestApp(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "cover.png")
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write cover: %v", err)
	}

	first := ta.bitmapAt(path)
	if first == nil {
		t.Fatal("cover not decoded")
	}
	if again := ta.bitmapAt(path); again != first {
		t.Fatal("decoded cover not cached")
	}

	missing := filepath.Join(dir, "missing.png")
	if ta.bitmapAt(missing) != nil {
		t.Fatal("missing cover should be nil")
	}
	if _, ok := ta.coverBitmaps[missing]; !ok {
		t.Fatal("missing cover should be cached as nil")
	}
	if ta.bitmapAt("") != nil {
		t.Fatal("empty path should be nil")
	}
}

// 列表不再截断：超过 200 条图集也必须完整返回，计数与列表长度一致。
func TestFetchRefreshHasNoRowCap(t *testing.T) {
	ta := newTestApp(t)
	ta.exec(t, "INSERT INTO brand(id, name) VALUES (1, '测试品牌')")
	ta.exec(t, "INSERT INTO firm(id, brand_id, name) VALUES (1, 1, '测试厂商')")
	ta.exec(t, "INSERT INTO series(id, brand_id, firm_id, name) VALUES (1, 1, 1, '测试车系')")
	ta.exec(t, "INSERT INTO youvision(id, brand_id, firm_id, series_id, category_id, name) VALUES (1, 1, 1, 1, 1, '测试版型')")
	for i := 1; i <= 205; i++ {
		ta.exec(t, "INSERT INTO gallery(id, youvision_id, brand_id, firm_id, series_id, category_id, title) VALUES (?, 1, 1, 1, 1, 1, ?)",
			i, fmt.Sprintf("图集%d", i))
	}

	r := ta.fetchRefresh(context.Background(), domain.GalleryFilter{}, catalogBatch, catalogBatch)
	if len(r.galleries) != 205 {
		t.Fatalf("galleries = %d, want 205 (no LIMIT cap)", len(r.galleries))
	}
	ta.applyRefresh(r, ta.refreshSeq, false)
	if len(ta.galleries) != 205 || ta.countInt("gallery") != 205 {
		t.Fatalf("list %d vs count %d mismatch", len(ta.galleries), ta.countInt("gallery"))
	}
}

// 端到端：store 中的数据经 fetchRefresh 查询、applyRefresh 应用后进入 UI 状态。
func TestRefreshLoadsDataFromStore(t *testing.T) {
	ta := newTestApp(t)
	ta.exec(t, "INSERT INTO brand(id, name) VALUES (1, '测试品牌')")
	ta.exec(t, "INSERT INTO firm(id, brand_id, name) VALUES (1, 1, '测试厂商')")
	ta.exec(t, "INSERT INTO series(id, brand_id, firm_id, name) VALUES (1, 1, 1, '测试车系')")
	ta.exec(t, "INSERT INTO youvision(id, brand_id, firm_id, series_id, category_id, name) VALUES (1, 1, 1, 1, 1, '测试版型')")
	ta.exec(t, "INSERT INTO gallery(id, youvision_id, brand_id, firm_id, series_id, category_id, title) VALUES (1, 1, 1, 1, 1, 1, '测试图集')")

	ta.Refresh()

	if len(ta.brands) != 1 || ta.brands[0].Name != "测试品牌" {
		t.Fatalf("brands not loaded: %+v", ta.brands)
	}
	if len(ta.galleries) != 1 || ta.galleries[0].Title != "测试图集" {
		t.Fatalf("galleries not loaded: %+v", ta.galleries)
	}
	if ta.countInt("gallery") != 1 {
		t.Fatalf("gallery count = %d, want 1", ta.countInt("gallery"))
	}
	tt := myui.NewTester(ta.View, 1200, 800)
	tt.Frame()
	if !tt.HasText("测试图集") {
		t.Fatal("loaded gallery title not shown in the UI")
	}
}

// 空关键词浏览走全量加载，不再像目录页那样只保留前 200 条。
func TestSeriesBrowseLoadsAllRows(t *testing.T) {
	ta := newTestApp(t)
	seedSeries(t, ta, 205)

	ta.page = "seriesQuery"
	ta.seriesQueryKeyword = ""
	tt := myui.NewTester(ta.View, 1000, 700)
	tt.Frame()
	waitBrowse(t, ta.App)
	tt.Frame()

	if len(ta.seriesBrowse) != 205 {
		t.Fatalf("browse rows = %d, want 205", len(ta.seriesBrowse))
	}
	if !tt.HasText("车系001") {
		t.Fatal("first series row not rendered")
	}
	rows := 0
	for _, info := range tt.AccessAll() {
		if strings.HasPrefix(info.Name, "车系") && strings.Contains(info.Name, " #") {
			rows++
		}
	}
	if rows == 0 || rows > 30 {
		t.Fatalf("rendered %d series rows; the list is not virtualized", rows)
	}
	if tt.HasText("仅显示前 200 条") {
		t.Fatal("browse list still reports the 200-row cap")
	}
	if err := tt.Click("车系001 #1"); err != nil {
		t.Fatalf("click first row: %v", err)
	}
	if ta.status != "已选择车系 #1 车系001" {
		t.Fatalf("status = %q", ta.status)
	}
	if info, ok := tt.Access("车系001 #1"); !ok || info.Bounds.W < 400 || info.Bounds.H < 50 {
		t.Fatalf("series row bounds = %+v, want a full-width 58 DIP row", info.Bounds)
	}
}

// 关键词查询同样不限条数，且结果在后台回投后才标记完成。
func TestSeriesSearchLoadsAllMatches(t *testing.T) {
	ta := newTestApp(t)
	seedSeries(t, ta, 205)

	ta.seriesQueryKeyword = "车系"
	ta.runSeriesSearch()
	deadline := time.Now().Add(2 * time.Second)
	for !ta.seriesQueryLoaded && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if !ta.seriesQueryLoaded {
		t.Fatal("series search did not finish")
	}
	if len(ta.seriesQueryResults) != 205 {
		t.Fatalf("search rows = %d, want 205", len(ta.seriesQueryResults))
	}
}

// 筛选查询默认不限条数，并能在后台完成、写入结果。
func TestSeriesFilterLoadsAllRows(t *testing.T) {
	ta := newTestApp(t)
	seedSeries(t, ta, 205)

	ta.runSeriesFilter()
	deadline := time.Now().Add(2 * time.Second)
	for !ta.seriesFilterLoaded && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if !ta.seriesFilterLoaded {
		t.Fatal("series filter did not finish")
	}
	if len(ta.seriesFilterResults) != 205 {
		t.Fatalf("filter rows = %d, want 205", len(ta.seriesFilterResults))
	}
}

// seedSeries 插入一个品牌、一个厂商和 n 个车系，满足外键约束。
func seedSeries(t *testing.T, ta *testApp, n int) {
	t.Helper()
	ta.exec(t, "INSERT INTO brand(id, name) VALUES (1, '测试品牌')")
	ta.exec(t, "INSERT INTO firm(id, brand_id, name) VALUES (1, 1, '测试厂商')")
	for i := 1; i <= n; i++ {
		ta.exec(t, `INSERT INTO series(id, brand_id, firm_id, name, level_name)
			VALUES (?, 1, 1, ?, 'SUV')`, i, fmt.Sprintf("车系%03d", i))
	}
}

func waitBrowse(t *testing.T, a *App) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !a.seriesBrowseLoaded && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if !a.seriesBrowseLoaded {
		t.Fatal("series browse load did not finish")
	}
}

func TestGalleryFilterValidAndInvalidIDs(t *testing.T) {
	ta := newTestApp(t)
	ta.gallerySearch = "  奥迪  "
	ta.filterBrandID = "12"
	ta.filterCategory = "not-a-number"
	ta.gallerySort = "updated_desc"

	f := ta.galleryFilter()
	if f.Keyword != "  奥迪  " || f.BrandID != 12 {
		t.Fatalf("unexpected filter: %+v", f)
	}
	if f.CategoryID != 0 {
		t.Fatalf("invalid ID should stay zero, got %d", f.CategoryID)
	}
	if ta.status == "" {
		t.Fatal("invalid ID did not report a status")
	}
}
