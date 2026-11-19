package uiapp

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"
)

// 图片预览：Esc 与关闭按钮走同一关闭路径；关闭按钮自动获得焦点。
func TestLightboxCloseByEscAndButton(t *testing.T) {
	ta := newTestApp(t)
	closed := 0
	ta.closeLightboxFn = func() { closed++ }
	ta.lightboxImage = testBitmap()
	ta.lightboxTitle = "测试图集"

	tt := ui.NewTester(ta.renderLightbox, 1000, 760)
	tt.Frame()

	// 关闭按钮自动聚焦。
	if !tt.Focused("关闭") {
		t.Fatal("close button did not take focus on open")
	}

	// Esc 关闭。
	tt.Key(0, ui.KeyEscape)
	if closed != 1 {
		t.Fatalf("Esc closed %d times, want 1", closed)
	}

	// 关闭按钮同样触发。
	if err := tt.Click("关闭"); err != nil {
		t.Fatalf("click 关闭: %v", err)
	}
	if closed != 2 {
		t.Fatalf("button closed %d times total, want 2", closed)
	}
}

// 键盘后退快捷键与顶部后退按钮等效。
func TestBackShortcut(t *testing.T) {
	ta := newTestApp(t)
	tt := ui.NewTester(ta.View, 1200, 800)
	settled(tt)
	if err := tt.Click("车型查询"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if ta.page != "seriesQuery" {
		t.Fatalf("page = %q, want seriesQuery", ta.page)
	}

	tt.Key(0, ui.KeyBack)
	tt.Frame()
	if ta.page != "home" {
		t.Fatalf("page after KeyBack = %q, want home", ta.page)
	}
}

// Cmd+F 在图集页聚焦搜索框。
func TestFocusSearchShortcut(t *testing.T) {
	ta := newTestApp(t)
	tt := ui.NewTester(ta.View, 1200, 800)
	settled(tt)
	if err := tt.Click("图集列表"); err != nil {
		t.Fatalf("click 图集列表: %v", err)
	}
	tt.Frame()

	tt.Key(ui.Cmd, ui.KeyF)
	tt.Frame()
	if !tt.Focused("搜索标题 / 图集 ID / 版型 ID") {
		t.Fatal("Cmd+F did not focus the gallery search box")
	}
}

// 底部状态栏不渲染操作反馈，setStatus 只保留在内存中供设置页使用。
func TestStatusBarHidesFeedback(t *testing.T) {
	ta := newTestApp(t)
	ta.setStatus("图集 #3 已创建，请选择图片上传")
	tt := ui.NewTester(ta.View, 1200, 800)
	settled(tt)
	if err := tt.Click("图集列表"); err != nil {
		t.Fatalf("click 图集列表: %v", err)
	}
	tt.Frame()

	if tt.HasText("状态：图集 #3 已创建，请选择图片上传") {
		t.Fatalf("status feedback should not be shown in the bottom bar; texts: %v", tt.Texts())
	}
}

// 底部栏只保留框架：空闲时也不再显示根目录概览。
func TestStatusBarHidesSummaryWhenIdle(t *testing.T) {
	ta := newTestApp(t)
	ta.status = ""
	tt := ui.NewTester(ta.View, 1200, 800)
	settled(tt)
	if err := tt.Click("图集列表"); err != nil {
		t.Fatalf("click 图集列表: %v", err)
	}
	tt.Frame()

	for _, text := range tt.Texts() {
		if len(text) >= len("根目录：") && text[:len("根目录：")] == "根目录：" {
			t.Fatalf("idle summary should not be shown in the bottom bar; texts: %v", tt.Texts())
		}
	}
}

// 点击侧边栏导航会切换页面并写入前进/后退历史，渲染内容随之改变。
func TestSidebarNavigationChangesPage(t *testing.T) {
	ta := newTestApp(t)
	tt := ui.NewTester(ta.View, 1200, 800)
	settled(tt)

	if ta.page != "home" {
		t.Fatalf("initial page = %q, want home", ta.page)
	}
	if err := tt.Click("车型查询"); err != nil {
		t.Fatalf("click 车型查询: %v", err)
	}
	tt.Frame()
	if ta.page != "seriesQuery" {
		t.Fatalf("page after click = %q, want seriesQuery", ta.page)
	}
	if len(ta.history) == 0 {
		t.Fatal("navigation did not record history")
	}
	if !tt.HasText("按车系名称、拼音或 ID 查询") {
		t.Fatalf("query page content not rendered: %v", tt.Texts())
	}
}

// 前进/后退按钮按顺序在历史栈上移动，不改变数据状态。
func TestNavigationBackAndForward(t *testing.T) {
	ta := newTestApp(t)
	tt := ui.NewTester(ta.View, 1200, 800)
	settled(tt)

	if err := tt.Click("车型查询"); err != nil {
		t.Fatalf("click 车型查询: %v", err)
	}
	tt.Frame()
	if ta.page != "seriesQuery" {
		t.Fatalf("page = %q, want seriesQuery", ta.page)
	}

	// 后退回到首页。
	if err := tt.Click("后退"); err != nil {
		t.Fatalf("click 后退: %v", err)
	}
	tt.Frame()
	if ta.page != "home" {
		t.Fatalf("page after back = %q, want home", ta.page)
	}
	if len(ta.forward) == 0 {
		t.Fatal("back did not record forward stack")
	}

	// 前进回到查询页（首页不渲染悬浮导航按钮，用快捷键前进）。
	tt.Key(0, ui.KeyForward)
	tt.Frame()
	if ta.page != "seriesQuery" {
		t.Fatalf("page after forward = %q, want seriesQuery", ta.page)
	}
}

// 车型查询整页滚动：标题、查询卡片和结果分行都随同一个滚动区移动。
func TestSeriesQueryUsesPageScroll(t *testing.T) {
	ta := newTestApp(t)
	seedSeries(t, ta, 60)
	ta.page = "seriesQuery"
	tt := ui.NewTester(ta.View, 1000, 700)
	settled(tt)
	if ta.contentScroll.MaxY <= 0 {
		t.Fatalf("series query content did not overflow: MaxY = %f", ta.contentScroll.MaxY)
	}
}

// 品牌页用框架的虚拟列表：只构建可见行，滚到底部才构建末尾卡片行。
func TestBrandCatalogVirtualList(t *testing.T) {
	ta := newTestApp(t)
	ta.exec(t, "DELETE FROM brand")
	for i := 1; i <= 60; i++ {
		ta.exec(t, fmt.Sprintf("INSERT INTO brand(id, name) VALUES (%d, '品牌%03d')", i, i))
	}
	ta.Refresh()
	ta.page = "catalog"
	ta.catalogEntity = "brand"

	tt := ui.NewTester(ta.View, 1200, 900)
	tt.Frame()
	_, last := ta.brandList.Visible()
	if last <= 4 {
		t.Fatalf("unexpected initial visible rows: first=%d last=%d", 0, last)
	}
	if !tt.HasText("品牌001") || tt.HasText("品牌060") {
		t.Fatalf("initial virtual rows are wrong; texts: %v", tt.Texts())
	}

	ta.brandList.ScrollToEnd()
	tt.Frame()
	_, last = ta.brandList.Visible()
	if last <= 4 || !tt.HasText("品牌060") {
		t.Fatalf("last rows not built after ScrollToEnd: last=%d; texts: %v", last, tt.Texts())
	}
}

// 首页内容足够长，常规窗口下中间区域就会出现滚动条。
func TestHomeContentOverflows(t *testing.T) {
	ta := newTestApp(t)
	tt := ui.NewTester(ta.View, 1200, 800)
	settled(tt)
	if ta.contentScroll.MaxY <= 0 {
		t.Fatalf("home content did not overflow: MaxY = %f", ta.contentScroll.MaxY)
	}
}

// 滚动位置由 TrackScroll 记录在 App 状态中：滚动内容区后，
// contentScroll.Y 更新且被限制在框架计算的可滚动范围内。
func TestScrollStateIsTracked(t *testing.T) {
	ta := newTestApp(t)
	// 极矮窗口确保内容区溢出，从而产生可滚动范围。
	tt := ui.NewTester(ta.View, 900, 180)
	tt.Frame()

	if ta.contentScroll.MaxY <= 0 {
		t.Fatalf("content did not overflow: MaxY = %f", ta.contentScroll.MaxY)
	}
	// 应用恢复滚动位置的做法：写回状态后重新渲染，位置应保留而非被清零。
	want := ta.contentScroll.MaxY / 2
	ta.contentScroll.Y = want
	tt.Frame()
	if ta.contentScroll.Y != want {
		t.Fatalf("scroll Y = %f, want preserved %f", ta.contentScroll.Y, want)
	}

	// 滚轮滚动（带动画推进）同样写回状态。
	tt.Wheel(500, 100, 0, 100)
	tt.Advance(1e9)
	tt.Frame()
	if ta.contentScroll.Y <= 0 {
		t.Fatalf("after wheel, scroll Y = %f; want Y > 0", ta.contentScroll.Y)
	}
	if ta.contentScroll.Y > ta.contentScroll.MaxY+1 {
		t.Fatalf("scroll Y %f exceeds MaxY %f", ta.contentScroll.Y, ta.contentScroll.MaxY)
	}
}
