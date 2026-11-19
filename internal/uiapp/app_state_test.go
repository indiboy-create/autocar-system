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
	"autocarsystem/internal/i18n"
	"autocarsystem/internal/service"

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

func TestColorSchemeSelection(t *testing.T) {
	original := currentColorSchemeKey
	defer setColorScheme(original)

	setColorScheme("blue")
	if currentColorSchemeKey != "blue" {
		t.Fatalf("scheme key = %q, want blue", currentColorSchemeKey)
	}
	if activeColorScheme().key != "blue" {
		t.Fatal("activeColorScheme did not follow selection")
	}
	// 明暗两套强调色应随方案变化且与方案自带预览色一致。
	light, dark := paletteFor(false), paletteFor(true)
	scheme := activeColorScheme()
	if !colorEqual(light.accent, scheme.lightAccent) || !colorEqual(dark.accent, scheme.darkAccent) {
		t.Fatal("palette accent did not follow selected scheme")
	}
	// 左侧栏背景与渐变也应跟随方案色调。
	lightFrom, lightTo := scheme.sidebarColors(false)
	darkFrom, _ := scheme.sidebarColors(true)
	if !colorEqual(light.sidebar, lightFrom) || !colorEqual(dark.sidebar, darkFrom) {
		t.Fatal("palette sidebar did not follow selected scheme")
	}
	if g := sidebarGradient(false); !colorEqual(g.From, lightFrom) || !colorEqual(g.To, lightTo) {
		t.Fatal("sidebar gradient did not follow selected scheme")
	}
	// 未知 key 回退到默认方案。
	setColorScheme("does-not-exist")
	if activeColorScheme().key != colorSchemes[0].key {
		t.Fatal("unknown scheme did not fall back")
	}
}

func colorEqual(a, b myui.Color) bool {
	return a.R == b.R && a.G == b.G && a.B == b.B && a.A == b.A
}

func TestSeriesByIDDerivesOwnership(t *testing.T) {
	ta := newTestApp(t)
	ta.exec(t, "INSERT INTO brand(id, name) VALUES (1, '品牌甲')")
	ta.exec(t, "INSERT INTO firm(id, brand_id, name) VALUES (7, 1, '厂商甲')")
	ta.exec(t, "INSERT INTO series(id, brand_id, firm_id, name) VALUES (5, 1, 7, '车型甲')")
	ta.ensureSeriesBrowse()
	waitBrowse(t, ta.App)

	s, ok := ta.seriesByID(5)
	if !ok || s.BrandID != 1 || s.FirmID != 7 {
		t.Fatalf("seriesByID(5) = %+v, ok=%v", s, ok)
	}
	if _, ok := ta.seriesByID(999); ok {
		t.Fatal("seriesByID should miss for unknown id")
	}
}

func TestFilterOptionRoundTrip(t *testing.T) {
	opt := filterOption("奥迪", 42)
	if got := filterOptionID(opt); got != 42 {
		t.Fatalf("filterOptionID(%q) = %d, want 42", opt, got)
	}
	if got := filterOptionID(i18n.T("全部")); got != 0 {
		t.Fatalf("filterOptionID(all) = %d, want 0", got)
	}
}

func TestSeriesFilterCascade(t *testing.T) {
	ta := newTestApp(t)
	// 两个品牌，各一个厂商和两个车系。
	ta.exec(t, "INSERT INTO brand(id, name) VALUES (1, '品牌甲'), (2, '品牌乙')")
	ta.exec(t, "INSERT INTO firm(id, brand_id, name) VALUES (1, 1, '厂商甲'), (2, 2, '厂商乙')")
	ta.exec(t, `INSERT INTO series(id, brand_id, firm_id, name) VALUES
		(1, 1, 1, '甲车一'), (2, 1, 1, '甲车二'), (3, 2, 2, '乙车一'), (4, 2, 2, '乙车二')`)

	// 重新加载品牌/厂商到内存，并加载全量车系供联动下拉使用。
	ta.Refresh()
	ta.ensureSeriesBrowse()
	waitBrowse(t, ta.App)

	brandOpts := ta.linkedOptions("brand", 0)
	seriesOpts := ta.linkedOptions("series", 1)
	if len(seriesOpts) != 3 { // “全部” + 品牌甲下的 2 个车系
		t.Fatalf("series options = %d, want 3", len(seriesOpts))
	}
	// 品牌乙的车系不能出现在品牌甲下的车型选项中。
	for _, o := range seriesOpts {
		if strings.Contains(o, "乙车") {
			t.Fatalf("series options leaked other brand: %q", o)
		}
	}
	if len(brandOpts) != 3 {
		t.Fatalf("brand options = %d, want 3", len(brandOpts))
	}

	// 选中品牌甲 + 车系“甲车二”后，查询只返回该车系。
	ta.seriesFilterBrand = filterOption("品牌甲", 1)
	ta.seriesFilterSeries = filterOption("甲车二", 2)
	filters := service.SeriesFilter{
		BrandID:  filterOptionID(ta.seriesFilterBrand),
		SeriesID: filterOptionID(ta.seriesFilterSeries),
		Limit:    -1,
	}
	items, err := ta.svc.FilterSeries(context.Background(), filters)
	if err != nil {
		t.Fatalf("FilterSeries: %v", err)
	}
	if len(items) != 1 || items[0].ID != 2 {
		t.Fatalf("filtered series = %+v, want only series 2", items)
	}

	// 联动失效：切到品牌乙后，品牌甲的车型选项应被重置为“全部”。
	stale := normalizeFilterChoice(filterOption("甲车二", 2), ta.linkedOptions("series", 2))
	if stale != i18n.T("全部") {
		t.Fatalf("normalizeFilterChoice = %q, want all", stale)
	}
}

func TestLevelChoicesAndKeywordFilter(t *testing.T) {
	// 级别文案到关键词的映射：全部为空，未收录文案也视为全部。
	if got := levelKeyword(i18n.T("全部")); got != "" {
		t.Fatalf("levelKeyword(all) = %q, want empty", got)
	}
	if got := levelKeyword("nonexistent"); got != "" {
		t.Fatalf("levelKeyword(unknown) = %q, want empty", got)
	}
	foundSUV, foundSedan := false, false
	for _, c := range levelChoices() {
		if c.keyword == "suv" {
			foundSUV = true
		}
		if c.keyword == "型车" {
			foundSedan = true
		}
	}
	if !foundSUV || !foundSedan {
		t.Fatalf("level choices missing suv/sedan: %+v", levelChoices())
	}
	if len(levelLabels()) != len(levelChoices()) {
		t.Fatal("levelLabels length mismatch")
	}

	ta := newTestApp(t)
	ta.exec(t, "INSERT INTO brand(id, name) VALUES (1, '品牌甲')")
	ta.exec(t, "INSERT INTO firm(id, brand_id, name) VALUES (1, 1, '厂商甲')")
	ta.exec(t, `INSERT INTO series(id, brand_id, firm_id, name, py, group_name, level_name) VALUES
		(1, 1, 1, '奥迪 A4L', 'aodia4l', '奥迪', '中型车'),
		(2, 1, 1, '宝马 X3',  'baomax3', '宝马', '中型SUV'),
		(3, 1, 1, '奔驰 C级', 'benzC',   '奔驰', '中型车')`)

	// 关键字按名称模糊匹配。
	items, err := ta.svc.FilterSeries(context.Background(), service.SeriesFilter{Keyword: "奥迪", Limit: -1})
	if err != nil || len(items) != 1 || items[0].ID != 1 {
		t.Fatalf("keyword by name = %+v, err=%v", items, err)
	}
	// 纯数字按 ID 精确匹配。
	items, err = ta.svc.FilterSeries(context.Background(), service.SeriesFilter{Keyword: "2", Limit: -1})
	if err != nil || len(items) != 1 || items[0].ID != 2 {
		t.Fatalf("keyword by ID = %+v, err=%v", items, err)
	}
	// 级别关键词(“型车”)命中两辆轿车；再叠加关键字只留一辆。
	items, err = ta.svc.FilterSeries(context.Background(), service.SeriesFilter{Level: "型车", Limit: -1})
	if err != nil || len(items) != 2 {
		t.Fatalf("level sedan = %+v, err=%v", items, err)
	}
	items, err = ta.svc.FilterSeries(context.Background(), service.SeriesFilter{Level: "suv", Keyword: "宝马", Limit: -1})
	if err != nil || len(items) != 1 || items[0].ID != 2 {
		t.Fatalf("level suv + keyword = %+v, err=%v", items, err)
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
	ta.filterBrand = filterOption("品牌甲", 12)
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

func waitMessagePrefix(t *testing.T, a *App, prefix string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if strings.HasPrefix(a.settingsMessage, prefix) {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("message = %q, want prefix %q", a.settingsMessage, prefix)
}

func TestRunBackupNowCreatesBackup(t *testing.T) {
	ta := newTestApp(t)
	ta.settingsMessage = ""
	before := len(ta.backups)

	ta.runBackupNow()
	waitMessagePrefix(t, ta.App, i18n.T("备份完成："))

	if len(ta.backups) <= before {
		t.Fatal("backup list was not refreshed")
	}
	if ta.prefs.LastBackupAt == "" {
		t.Fatal("last backup time was not recorded")
	}
	if _, err := os.Stat(filepath.Join(ta.paths.BackupDir, ta.backups[0].Name)); err != nil {
		t.Fatalf("backup file missing: %v", err)
	}
}

func TestAutoBackupRunsWhenDue(t *testing.T) {
	ta := newTestApp(t)
	ta.prefs.AutoBackup = true
	ta.prefs.BackupIntervalHours = 1
	ta.prefs.LastBackupAt = time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339)
	ta.settingsMessage = ""
	before := len(ta.backups)

	ta.runAutoBackupIfDue(context.Background(), time.Now())
	waitMessagePrefix(t, ta.App, i18n.T("自动备份完成："))

	if len(ta.backups) <= before {
		t.Fatal("scheduled backup did not create a backup")
	}
}

func TestAutoBackupSkipsWhenNotDue(t *testing.T) {
	ta := newTestApp(t)
	ta.prefs.AutoBackup = true
	ta.prefs.BackupIntervalHours = 24
	ta.prefs.LastBackupAt = time.Now().UTC().Format(time.RFC3339)
	ta.settingsMessage = ""
	before := len(ta.backups)

	ta.runAutoBackupIfDue(context.Background(), time.Now())
	time.Sleep(50 * time.Millisecond)

	if ta.settingsMessage != "" {
		t.Fatalf("unexpected message: %q", ta.settingsMessage)
	}
	if len(ta.backups) != before {
		t.Fatalf("backup should be skipped, count changed from %d to %d", before, len(ta.backups))
	}
}

func TestAutoBackupRunsWhenNeverBackedUp(t *testing.T) {
	ta := newTestApp(t)
	ta.prefs.AutoBackup = true
	ta.prefs.BackupIntervalHours = 72
	ta.prefs.LastBackupAt = ""
	ta.settingsMessage = ""
	before := len(ta.backups)

	ta.runAutoBackupIfDue(context.Background(), time.Now())
	waitMessagePrefix(t, ta.App, i18n.T("自动备份完成："))

	if len(ta.backups) <= before {
		t.Fatal("first backup was not created")
	}
}

func TestCategorySettingsPageRenders(t *testing.T) {
	ta := newTestApp(t)
	i18n.Set("zh-CN")
	ta.settingsPage = "category"
	tt := myui.NewTester(ta.settingsView, 900, 700)
	tt.Frame()

	for _, want := range []string{
		"新增分类", "创建用于版型与图集的分类", "新分类名称",
		"分类列表", "启用或禁用分类，删除不再使用的分类",
		fmt.Sprintf("共 %d 个分类", len(ta.categories)),
		"手缝把套", "通用把套",
	} {
		if !tt.HasText(want) {
			t.Errorf("category page missing %q; texts: %v", want, tt.Texts())
		}
	}
}

func TestSystemSettingsPageRenders(t *testing.T) {
	ta := newTestApp(t)
	i18n.Set("zh-CN")
	ta.settingsPage = "system"
	tt := myui.NewTester(ta.settingsView, 900, 700)
	tt.Frame()

	for _, want := range []string{
		"系统信息", "本地运行", "根目录", "数据目录", "媒体目录", "数据库", "配置文件",
		ta.paths.Root, ta.paths.DBPath,
		"孤儿图集目录", "未发现", "扫描孤儿目录", "清理孤儿目录",
		"未发现孤儿图集目录，或尚未扫描。",
	} {
		if !tt.HasText(want) {
			t.Errorf("system page missing %q; texts: %v", want, tt.Texts())
		}
	}
}

func TestLicenseSettingsPageRenders(t *testing.T) {
	ta := newTestApp(t)
	i18n.Set("zh-CN")
	ta.settingsPage = "license"
	tt := myui.NewTester(ta.settingsView, 900, 700)
	tt.Frame()

	for _, want := range []string{
		"授权", "软件授权", "查看与管理本软件的授权状态", "待配置",
		"授权功能正在建设中，后续将在此提供授权状态与激活管理。",
	} {
		if !tt.HasText(want) {
			t.Errorf("license page missing %q; texts: %v", want, tt.Texts())
		}
	}
}
