package uiapp

import (
	"context"
	"image"
	"testing"

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

// 封面缓存只在本次确实解码了封面时才覆盖，避免非图集页把缓存清空。
func TestApplyRefreshCoversOnlyWhenFetched(t *testing.T) {
	ta := newTestApp(t)
	sentinel := testBitmap()
	ta.coverBitmaps = map[string]*myui.Bitmap{"keep.jpg": sentinel}

	// withCovers=false：即使结果里带了空 map，也不覆盖。
	ta.applyRefresh(refreshResult{covers: map[string]*myui.Bitmap{}, withCovers: false}, ta.refreshSeq, false)
	if ta.coverBitmaps["keep.jpg"] != sentinel {
		t.Fatal("covers overwritten when withCovers is false")
	}

	// withCovers=true：用本次结果替换。
	replacement := map[string]*myui.Bitmap{"new.jpg": testBitmap()}
	ta.applyRefresh(refreshResult{covers: replacement, withCovers: true}, ta.refreshSeq, false)
	if _, ok := ta.coverBitmaps["keep.jpg"]; ok {
		t.Fatal("old cover survived a withCovers=true apply")
	}
	if ta.coverBitmaps["new.jpg"] != replacement["new.jpg"] {
		t.Fatal("new covers not applied")
	}
}

// 非图集页跳过封面解码：withCovers=false 时 covers 为 nil 且标志为 false。
func TestFetchRefreshSkipsCoversWhenNotRequested(t *testing.T) {
	ta := newTestApp(t)
	r := ta.fetchRefresh(context.Background(), domain.GalleryFilter{}, false)
	if r.withCovers {
		t.Fatal("withCovers should be false when not requested")
	}
	if r.covers != nil {
		t.Fatal("covers should be nil when not requested")
	}
	if r.categories == nil || r.counts == nil {
		t.Fatal("non-cover queries should still run")
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
