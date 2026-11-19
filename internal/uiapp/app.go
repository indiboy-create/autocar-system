package uiapp

import (
	"autocarsystem/internal/config"
	"autocarsystem/internal/domain"
	"autocarsystem/internal/i18n"
	"autocarsystem/internal/service"
	"context"
	"runtime"
	"strconv"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/glass"
	myui "github.com/egoist/mygo/ui"
	"strings"
	"sync/atomic"
)

type palette struct {
	dark     bool
	bg       myui.Color
	sidebar  myui.Color
	text     myui.Color
	muted    myui.Color
	faint    myui.Color
	hairline myui.Color
	hover    myui.Color
	press    myui.Color
	accent   myui.Color
	onAccent myui.Color
}

var (
	iconCategory  = strokeIcon(`<path d="M4 6h16M4 12h16M4 18h16"/><circle cx="7" cy="6" r="1"/><circle cx="7" cy="12" r="1"/><circle cx="7" cy="18" r="1"/>`)
	iconHome      = strokeIcon(`<path d="m3 10.5 9-7.5 9 7.5"/><path d="M5 9.5V20h5v-6h4v6h5V9.5"/>`)
	iconUser      = strokeIcon(`<circle cx="12" cy="8" r="4"/><path d="M4.5 21a7.5 7.5 0 0 1 15 0"/>`)
	iconCatalog   = strokeIcon(`<ellipse cx="12" cy="5" rx="8" ry="3"/><path d="M4 5v14c0 1.7 3.6 3 8 3s8-1.3 8-3V5"/><path d="M4 12c0 1.7 3.6 3 8 3s8-1.3 8-3"/>`)
	iconSync      = strokeIcon(`<path d="M20 12a8 8 0 1 1-2.34-5.66L20 8"/><path d="M20 3v5h-5"/>`)
	iconTransfer  = strokeIcon(`<path d="M4 12a8 8 0 0 1 13.66-5.66L20 8M20 3v5h-5"/><path d="M20 12a8 8 0 0 1-13.66 5.66L4 16M4 21v-5h5"/>`)
	iconLayers    = strokeIcon(`<path d="m12 2 9 5-9 5-9-5 9-5z"/><path d="m3 12 9 5 9-5M3 17l9 5 9-5"/>`)
	iconGallery   = strokeIcon(`<rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="m21 15-5-5L5 21"/>`)
	iconSettings  = strokeIcon(`<path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/>`)
	iconBack      = strokeIcon(`<path d="m15 18-6-6 6-6"/>`)
	iconForward   = strokeIcon(`<path d="m9 18 6-6-6-6"/>`)
	iconAbout     = strokeIcon(`<circle cx="12" cy="12" r="9"/><path d="M12 11v5"/><circle cx="12" cy="8" r="0.6" fill="currentColor" stroke="none"/>`)
	iconLogo      = strokeIcon(`<rect x="3" y="4" width="18" height="16" rx="3"/><circle cx="8.5" cy="9.5" r="1.5"/><path d="m21 15-5-5L6 20"/>`)
	iconBackup    = strokeIcon(`<path d="M12 3v12M7 10l5 5 5-5"/><path d="M4 19h16"/>`)
	iconTrash     = strokeIcon(`<path d="M3 6h18"/><path d="M8 6V4h8v2"/><path d="M6 6l1 14h10l1-14"/><path d="M10 10v7M14 10v7"/>`)
	iconLicense   = strokeIcon(`<path d="M15 2 21 8l-9 9-6-6 9-9z"/><path d="m3 13 6 6-3 3H3v-6z"/><path d="m14 5 3 3"/><path d="m7 14 3 3"/>`)
	iconSearch    = strokeIcon(`<circle cx="11" cy="11" r="7"/><path d="m20 20-4-4"/>`)
	iconClose     = strokeIcon(`<path d="M6 6l12 12M18 6 6 18"/>`)
	iconArrowUp   = strokeIcon(`<path d="M12 19V5M6 11l6-6 6 6"/>`)
	iconArrowDown = strokeIcon(`<path d="M12 5v14M6 13l6 6 6-6"/>`)
)

type navRoute struct {
	page              string
	catalogEntity     string
	brandID           int64
	firmID            int64
	seriesID          int64
	specID            int64
	yuvFilterCategory int64
	filterCategory    string
}

type appCore struct {
	win        *mygo.Window
	tray       any
	svc        *service.Service
	paths      config.Paths
	status     string
	quitting   bool
	refreshSeq uint64       // 递增序号，丢弃过期的异步 Refresh 结果
	runOnMain  func(func()) // 主线程执行器；生产为 nil（用 mygo.RunOnMain），测试注入同步执行
}

type appNav struct {
	page             string
	history          []navRoute
	forward          []navRoute
	underTop         bool // 内容已滚到顶栏下方
	underBottom      bool // 内容延伸到底部栏下方
	settingsUnderTop bool
}

// appToast 是一次性提示：Text 为正文，Kind 为 "success" 或 "error"，
// 由设置窗口渲染时交给自定义 toast 外观显示后清空。
type appToast struct {
	Text string
	Kind string
}

type appSettings struct {
	prefs             config.Prefs
	settingsWin       *mygo.Window
	settingsPage      string
	settingsMessage   string
	settingsQuery     string      // 设置页搜索关键字
	settingsSearchBox myui.Handle // Cmd+F 聚焦的设置搜索框
	toast             appToast
	updates           bool // updater 插件已注册；决定更新页是否可用
}

// catalogPagedState 描述目录列表分页加载的进度，参照 music 的 paged。
type catalogPagedState struct {
	loaded  bool
	loading bool
	more    bool
	err     string
}

type appCatalogData struct {
	newName    string
	categories []domain.Category
	yuvisions  []domain.Yuvision
	galleries  []domain.Gallery
	counts     map[string]int64
	backups    []domain.BackupInfo
	// 各实体落地页当前展示的 ID，0 表示未选中。
	brandDetailID  int64
	firmDetailID   int64
	seriesDetailID int64
	specDetailID   int64
	// 各目录页分页加载状态（参照 mygo-music 的收藏专辑页）。
	brandPaged  catalogPagedState
	firmPaged   catalogPagedState
	seriesPaged catalogPagedState
	specPaged   catalogPagedState
	// 各实体服务端搜索：关键字变化时重新查询，结果独立分页。
	brandSearchKeyword  string            // 已发起查询的关键字
	brandSearchResults  []domain.Brand    // 服务端返回的搜索结果
	brandSearchPaged    catalogPagedState // 搜索结果分页状态
	brandSearchSeq      uint64            // 丢弃过期搜索结果
	firmSearchKeyword   string
	firmSearchResults   []domain.Firm
	firmSearchPaged     catalogPagedState
	firmSearchSeq       uint64
	seriesSearchKeyword string
	seriesSearchResults []domain.Series
	seriesSearchPaged   catalogPagedState
	seriesSearchSeq     uint64
	specSearchKeyword   string
	specSearchResults   []domain.Spec
	specSearchPaged     catalogPagedState
	specSearchSeq       uint64
	// 各实体卡片的搜索关键字。
	brandSearch       string
	brandList         myui.ListState // 品牌页虚拟列表状态
	brandListKeyword  string         // 搜索词变化时把列表滚回顶部
	firmSearch        string
	firmList          myui.ListState
	firmListKeyword   string
	seriesSearch      string
	seriesList        myui.ListState
	seriesListKeyword string
	specSearch        string
	specList          myui.ListState
	specListKeyword   string
}

type appVersionDraft struct {
	// 品牌 → 车型联动选中项（名称 #ID）；厂商由选中车系推导，不再单独选择。
	yvBrand     string
	yvSeries    string
	yvCategory  string
	yvName      string
	yvDesc      string
	yvUploadID  string
	yvUploadSrc string
}

type appGalleryDraft struct {
	galYuvisionID string
	galTitle      string
	galUploadID   string
	galUploadSrc  string
}

type appGalleryFilter struct {
	gallerySearch  string
	searchBox      myui.Handle // Cmd+F 聚焦的搜索框
	filterCategory string
	filterYuvision string
	gallerySort    string
	orphanDirs     []string
	// 品牌 → 车型联动选中项（名称 #ID），用于图集筛选。
	filterBrand       string
	filterSeries      string
	yuvFilterCategory int64
}

type appSeriesQuery struct {
	seriesQueryKeyword string
	seriesQueryResults []domain.Series
	seriesQueryLoaded  bool
	seriesQueryLoading bool
	seriesQueryErr     string
	seriesQuerySeq     int
	queryVisible       int // 车型查询页当前展示的结果条数，按批增加
	filterVisible      int // 筛选查询页当前展示的结果条数，按批增加
	// seriesBrowse 是空关键词时的全量车系，首次进入页面时后台加载一次；
	// 列表只渲染可视行，不依赖目录页的分批上限。
	seriesBrowse        []domain.Series
	seriesBrowseLoading bool
	seriesBrowseLoaded  bool
	seriesBrowseErr     string

	// 品牌 → 车型两级联动下拉的选中项，存的是选项文本（名称 #ID），
	// i18n 的“全部”表示不限；ID 在 runSeriesFilter 中解析。
	seriesFilterBrand   string
	seriesFilterSeries  string
	seriesFilterLevel   string // 级别下拉的显示文案，在 runSeriesFilter 中映射为匹配关键词
	seriesFilterEnergy  string
	seriesFilterKeyword string // 关键字框：匹配车系名称 / 拼音 / 车组 / ID
	seriesFilterResults []domain.Series
	seriesFilterLoaded  bool
	seriesFilterLoading bool
	seriesFilterErr     string
	seriesFilterSeq     int
}

type appSync struct {
	syncRunning bool
	syncMessage string
	syncGate    atomic.Bool // 后台 goroutine 用的启动闸门，防止重复启动同步
}

type appLightbox struct {
	lightboxWin     *mygo.Window
	lightboxImage   *myui.Bitmap
	lightboxTitle   string
	closeLightboxFn func() // 测试可注入，验证 Esc/按钮关闭路径；生产为 nil（用 lightboxWin.Hide）
}

type appCatalogEdit struct {
	catalogEntity string
	brands        []domain.Brand
	firms         []domain.Firm
	series        []domain.Series
	specs         []domain.Spec
}

type appCovers struct {
	coverBitmaps map[string]*myui.Bitmap
}

type appScroll struct {
	contentScroll  myui.ScrollState
	settingsScroll myui.ScrollState
	navScroll      myui.ScrollState
}

// App 按职责内嵌状态分组；字段名提升后原有 a.field 访问保持不变。
type App struct {
	appCore
	appNav
	appSettings
	appCatalogData
	appVersionDraft
	appGalleryDraft
	appGalleryFilter
	appSeriesQuery
	appSync
	appLightbox
	appCatalogEdit
	appCovers
	appScroll
}

func New(svc *service.Service, paths config.Paths) *App {
	prefs, err := config.LoadPrefs(paths.PrefsPath)
	if err != nil {
		prefs = config.DefaultPrefs()
	}
	i18n.Set(prefs.Language)
	applyAppearance(prefs.Appearance)
	setColorScheme(prefs.ColorScheme)
	app := &App{
		svc:           svc,
		paths:         paths,
		page:          "home",
		gallerySort:   "updated_desc",
		coverBitmaps:  make(map[string]*myui.Bitmap),
		catalogEntity: "brand",
		prefs:         prefs,
		settingsPage:  "general",
	}
	app.Refresh()
	return app
}

func (a *App) invalidateWindows() {
	if a.win != nil {
		a.win.Invalidate()
	}
	if a.settingsWin != nil {
		a.settingsWin.Invalidate()
	}
}

func (a *App) AttachWindow(win *mygo.Window) { a.win = win }

func (a *App) AttachTray(tray any) { a.tray = tray }

func (a *App) SetQuitting(v bool) { a.quitting = v }

func (a *App) Quitting() bool { return a.quitting }

// postMain 在主线程同步执行 fn。生产环境走 mygo 事件循环；测试注入 runOnMain
// 直接执行，从而无需窗口与事件循环即可验证状态逻辑。
func (a *App) postMain(fn func()) {
	if a.runOnMain != nil {
		a.runOnMain(fn)
		return
	}
	mygo.RunOnMain(fn)
}

// Refresh 可从任意 goroutine 调用：序号自增与筛选条件快照都在主线程完成，
// 数据库查询在后台 goroutine 执行，结果再经 Window.Update 回投主线程。
func (a *App) Refresh() {
	var filter domain.GalleryFilter
	var seq uint64
	var win *mygo.Window

	// postMain 在主线程上直接执行；从后台调用时会切到主线程并等待，
	// 保证 refreshSeq 与筛选字段的读写都在同一线程，消除竞态。
	a.postMain(func() {
		seq = a.refreshSeq + 1
		a.refreshSeq = seq
		filter = a.galleryFilter()
		win = a.win
	})

	if win == nil {
		// 启动阶段（AttachWindow 之前）无窗口可回投，同步加载一次。
		a.applyRefresh(a.fetchRefresh(context.Background(), filter), seq, false)
		return
	}

	go func() {
		result := a.fetchRefresh(context.Background(), filter)
		win.Update(func() {
			a.applyAsyncRefresh(result, seq)
		})
	}()
}

// applyAsyncRefresh 回投主线程后只应用最后一次刷新的结果：序号过期则整批丢弃，
// 避免快速输入或连续操作产生的旧结果覆盖新状态。
func (a *App) applyAsyncRefresh(result refreshResult, seq uint64) {
	if seq != a.refreshSeq {
		return
	}
	a.applyRefresh(result, seq, true)
}

func (a *App) currentRoute() navRoute {
	return navRoute{
		page:              a.page,
		catalogEntity:     a.catalogEntity,
		brandID:           a.brandDetailID,
		firmID:            a.firmDetailID,
		seriesID:          a.seriesDetailID,
		specID:            a.specDetailID,
		yuvFilterCategory: a.yuvFilterCategory,
		filterCategory:    a.filterCategory,
	}
}

func (a *App) open(r navRoute) {
	current := a.currentRoute()
	if r == current {
		return
	}
	a.history = append(a.history, current)
	a.forward = nil
	a.show(r)
}

func (a *App) show(r navRoute) {
	a.page = r.page
	a.catalogEntity = r.catalogEntity
	a.brandDetailID = r.brandID
	a.firmDetailID = r.firmID
	a.seriesDetailID = r.seriesID
	a.specDetailID = r.specID
	a.yuvFilterCategory = r.yuvFilterCategory
	a.filterCategory = r.filterCategory
	a.Refresh()
}

func (a *App) back() {
	if len(a.history) == 0 {
		return
	}
	a.forward = append(a.forward, a.currentRoute())
	r := a.history[len(a.history)-1]
	a.history = a.history[:len(a.history)-1]
	a.show(r)
}

func (a *App) goForward() {
	if len(a.forward) == 0 {
		return
	}
	a.history = append(a.history, a.currentRoute())
	r := a.forward[len(a.forward)-1]
	a.forward = a.forward[:len(a.forward)-1]
	a.show(r)
}

// shortcuts 注册全局键盘快捷键：前进/后退、以及聚焦图集搜索框。
// 与侧栏/顶部的按钮共用同一状态方法，保证鼠标与键盘行为一致。
func (a *App) shortcuts(c *myui.Context) {
	c.OnShortcut(0, myui.KeyBack, a.back)
	c.OnShortcut(0, myui.KeyForward, a.goForward)
	c.OnShortcut(myui.Cmd, myui.KeyBracketLeft, a.back)
	c.OnShortcut(myui.Cmd, myui.KeyBracketRight, a.goForward)
	c.OnShortcut(myui.Cmd, myui.KeyF, func() {
		if a.page == "gallery" {
			a.searchBox.Focus(c)
		}
	})
}

// usesVirtualList 报告页面是否由 ui.List 自己管理中间区域的滚动。
// 品牌卡片数量可达数百，交给框架的虚拟列表只构建可见行。
func (a *App) usesVirtualList() bool {
	return a.page == "catalog"
}

func (a *App) View(c *myui.Context) {
	a.applyTheme(c)
	a.shortcuts(c)
	p := paletteFor(c.Theme().Dark)

	if c.Vibrancy() {
		c.Root().Background(myui.Transparent)
	}
	root := myui.Box(c).Fill()
	if !c.Vibrancy() {
		root.Background(p.bg)
	}
	root.Children(func() {
		page := myui.Row(c).Fill().AlignItems(myui.Stretch)
		page.Children(func() {
			side := myui.Column(c).Width(232)
			// 侧栏渐变随配色方案变化，不受窗口 Vibrancy 的材质影响。
			side.Background(p.sidebar).LinearGradient(sidebarGradient(c.Theme().Dark))
			side.Children(func() {
				a.sidebar(c)
			})

			main := myui.Box(c.Key("main")).Grow(1).Background(p.bg)
			main.Children(func() {
				// 三段式布局（参考 mygo-music）：顶部固定滚动边缘与导航按钮、
				// 底部固定状态栏，中间内容滚动。
				pad := float32(46)
				// 底部给悬浮状态栏留出空间，内容从栏下方滚过而不被遮挡。
				bottomPad := bottomBarHeight + 16
				if a.usesVirtualList() {
					// 虚拟列表自己滚动，作为中间唯一滚动容器；顶/底边缘由覆盖层绘制。
					a.underTop, a.underBottom = false, false
					myui.Column(c.Key("content-"+a.page)).Grow(1).
						Padding(pad, 22, bottomPad, 22).Children(func() {
						a.renderPage(c)
					})
				} else {
					myui.Scroll(c.Key("content-"+a.page)).
						TrackScroll(&a.contentScroll).
						Grow(1).
						Padding(pad, 22, bottomPad, 22).
						ScrollbarInsets(pad, 0, bottomPad, 0).
						Children(func() {
							mark := myui.Box(c.Key("content-start-mark")).Height(0).Bounds()
							if mark.W > 0 {
								a.underTop = mark.Y < pad-0.5
							}
							a.renderPage(c)
							end := myui.Box(c.Key("content-end-mark")).Height(0).Bounds()
							if end.W > 0 {
								_, h := c.Size()
								a.underBottom = end.Y > h-bottomBarHeight+0.5
							}
						})
				}
				a.statusBar(c, p)
				edge := myui.Box(c.Key("top-edge")).Absolute().Top(0).Left(0).Right(0).Height(pad).DragWindow().
					Material(glass.ScrollEdge{Hard: true, Background: p.bg})
				edge.Opacity(edgeShown(edge, a.underTop || edge.Hovered()))
				if len(a.history) > 0 || len(a.forward) > 0 {
					myui.Row(c).Absolute().Top(0).Left(20).Height(pad).AlignItems(myui.Center).Children(func() {
						back := navIconButton(c, "nav-back", iconBack, i18n.T("后退"))
						if len(a.history) == 0 {
							back.Disabled(true)
							back.Opacity(0.35)
						}
						if back.Clicked() {
							a.back()
						}
						if len(a.forward) > 0 {
							forward := navIconButton(c, "nav-forward", iconForward, i18n.T("前进"))
							if forward.Clicked() {
								a.goForward()
							}
						}
					})
				}
				a.backToTopButton(c, p, bottomPad)
			})
		})
	})
}

// settingsGroup 按 mygo-music 的样式绘制设置卡片：圆角、描边、行间分隔线。
func (a *App) settingsGroup(c *myui.Context, build func()) {
	p := paletteFor(c.Theme().Dark)
	myui.Column(c).Radius(12).Background(p.bg).Border(1, p.hairline).Dividers(1, p.hairline).Children(build)
}

// settingsRow 左侧为标题与说明，右侧为控件，与 mygo-music 的设置行一致。

// settingsRow 左侧为标题与说明，右侧为控件，与 mygo-music 的设置行一致。
func (a *App) settingsRow(c *myui.Context, label, desc string, build func()) {
	p := paletteFor(c.Theme().Dark)
	myui.Row(c).Padding(12, 16).Gap(16).AlignItems(myui.Center).MinHeight(48).Children(func() {
		myui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(3).Children(func() {
			myui.Text(c, label).FontSize(13.5).FontWeight(500)
			if desc != "" {
				myui.Text(c, desc).FontSize(12).TextColor(p.muted).LineHeight(1.4)
			}
		})
		build()
	})
}

func (a *App) generalSettings(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)

	a.settingsGroup(c, func() {
		a.settingsRow(c, i18n.T("界面语言"), "", func() {
			follow := i18n.T("跟随系统")
			names := []string{follow, "English", i18n.T("简体中文")}
			chosen := follow
			switch a.prefs.Language {
			case "en", "en-US":
				chosen = "English"
			case "zh-CN":
				chosen = i18n.T("简体中文")
			}
			if myui.Select(c.Key("language"), &chosen, names).Label(i18n.T("界面语言")).Width(170).Changed() {
				code := "system"
				switch chosen {
				case "English":
					code = "en"
				case i18n.T("简体中文"):
					code = "zh-CN"
				}
				a.prefs.Language = code
				i18n.Set(code)
				a.seriesFilterEnergy = ""
				a.savePrefsWithToast(i18n.T("界面语言"))
				a.Refresh()
				if a.settingsWin != nil {
					a.settingsWin.SetTitle(i18n.T("设置"))
					a.settingsWin.Update(func() {})
				}
			}
		})
		a.settingsRow(c, i18n.T("主题"), i18n.T("选择浅色、深色或跟随系统"), func() {
			a.settingsChoice(c,
				[]string{"system", "light", "dark"},
				[]string{i18n.T("跟随系统"), i18n.T("浅色"), i18n.T("深色")},
				a.prefs.Appearance,
				func(value string) {
					a.prefs.Appearance = value
					a.savePrefsWithToast(i18n.T("主题"))
					applyAppearance(value)
					a.invalidateWindows()
				},
			)
		})
		a.settingsRow(c, i18n.T("配色方案"), i18n.T("同时改变强调色与左侧栏背景，立即应用"), func() {
			myui.Row(c).Gap(10).AlignItems(myui.Center).Children(func() {
				for _, scheme := range colorSchemes {
					current := scheme
					active := currentColorSchemeKey == current.key
					pick := myui.ButtonBase(c.Key("color-scheme-"+current.key)).
						Size(40, 40).Radius(999).Center().FocusRing(false).
						Label(i18n.T(current.label))
					bg := myui.Transparent
					if active {
						bg = p.press
					} else if pick.Hovered() {
						bg = p.hover
					}
					pick.Background(bg)
					pick.Children(func() {
						// 外圈是侧栏底色（深色主题），内点是强调色，直观预览方案效果。
						sideFrom, _ := current.sidebarColors(true)
						ring := myui.Box(c).Size(28, 28).Radius(999).Center().Background(sideFrom)
						if active {
							ring.Border(2, current.swatch)
						} else {
							ring.Border(1, p.hairline)
						}
						ring.Children(func() {
							myui.Box(c).Size(11, 11).Radius(999).Background(current.swatch)
						})
					})
					if pick.Clicked() {
						a.prefs.ColorScheme = current.key
						setColorScheme(current.key)
						a.savePrefsWithToast(i18n.T("配色方案"))
						a.invalidateWindows()
					}
				}
			})
		})
	})

	if strings.TrimSpace(a.settingsMessage) != "" {
		myui.Text(c, a.settingsMessage).FontSize(12.5).TextColor(p.muted)
	}
}

func (a *App) categorySettings(c *myui.Context) {
	a.renderCategorySettings(c)
}

func (a *App) syncSettings(c *myui.Context) {
	a.renderSync(c)
}

func (a *App) systemSettings(c *myui.Context) {
	a.renderSystemSettings(c)
}

func (a *App) updateSettings(c *myui.Context) {
	a.renderUpdateSettings(c)
}

func (a *App) aboutSettings(c *myui.Context) {
	// 应用信息：名称、版本、描述、技术栈与运行时环境。
	a.settingsGroup(c, func() {
		a.settingsRow(c, i18n.T("应用名称"), "autoCarSystem", func() {})
		a.settingsRow(c, i18n.T("版本"), mygo.App.Version(), func() {})
		a.settingsRow(c, i18n.T("描述"), i18n.T("本地图集与版型数据管理工具"), func() {})
		a.settingsRow(c, i18n.T("技术栈"), "Go  ·  MyGo  ·  SQLite", func() {})
		a.settingsRow(c, i18n.T("运行时"), runtime.Version()+"  ·  "+runtime.GOOS+"/"+runtime.GOARCH, func() {})
	})

	// 数据统计：复用已加载的计数，覆盖全部实体。
	a.settingsGroup(c, func() {
		for _, item := range []struct{ label, key string }{
			{i18n.T("品牌"), "brand"},
			{i18n.T("厂商"), "firm"},
			{i18n.T("车型"), "series"},
			{i18n.T("年款"), "spec"},
			{i18n.T("分类"), "category"},
			{i18n.T("版型"), "youvision"},
			{i18n.T("图集"), "gallery"},
		} {
			a.settingsRow(c, item.label, strconv.FormatInt(a.countInt(item.key), 10), func() {})
		}
	})

	// 存储位置：便于用户定位数据与配置文件。
	a.settingsGroup(c, func() {
		a.settingsRow(c, i18n.T("根目录"), a.paths.Root, func() {})
		a.settingsRow(c, i18n.T("数据目录"), a.paths.Data, func() {})
		a.settingsRow(c, i18n.T("媒体目录"), a.paths.Media, func() {})
		a.settingsRow(c, i18n.T("数据库"), a.paths.DBPath, func() {})
		a.settingsRow(c, i18n.T("配置文件"), a.paths.PrefsPath, func() {})
	})
}
