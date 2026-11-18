package uiapp

import (
	"autocarsystem/internal/config"
	"autocarsystem/internal/domain"
	"autocarsystem/internal/i18n"
	"autocarsystem/internal/service"
	"context"
	"fmt"
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
	iconCategory = strokeIcon(`<path d="M4 6h16M4 12h16M4 18h16"/><circle cx="7" cy="6" r="1"/><circle cx="7" cy="12" r="1"/><circle cx="7" cy="18" r="1"/>`)
	iconCatalog  = strokeIcon(`<ellipse cx="12" cy="5" rx="8" ry="3"/><path d="M4 5v14c0 1.7 3.6 3 8 3s8-1.3 8-3V5"/><path d="M4 12c0 1.7 3.6 3 8 3s8-1.3 8-3"/>`)
	iconSync     = strokeIcon(`<path d="M20 12a8 8 0 1 1-2.34-5.66L20 8"/><path d="M20 3v5h-5"/>`)
	iconTransfer = strokeIcon(`<path d="M4 12a8 8 0 0 1 13.66-5.66L20 8M20 3v5h-5"/><path d="M20 12a8 8 0 0 1-13.66 5.66L4 16M4 21v-5h5"/>`)
	iconLayers   = strokeIcon(`<path d="m12 2 9 5-9 5-9-5 9-5z"/><path d="m3 12 9 5 9-5M3 17l9 5 9-5"/>`)
	iconGallery  = strokeIcon(`<rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="m21 15-5-5L5 21"/>`)
	iconSettings = strokeIcon(`<path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/>`)
	iconBack     = strokeIcon(`<path d="m15 18-6-6 6-6"/>`)
	iconForward  = strokeIcon(`<path d="m9 18 6-6-6-6"/>`)
	iconAbout    = strokeIcon(`<circle cx="12" cy="12" r="9"/><path d="M12 11v5"/><circle cx="12" cy="8" r="0.6" fill="currentColor" stroke="none"/>`)
	iconLogo     = strokeIcon(`<rect x="3" y="4" width="18" height="16" rx="3"/><circle cx="8.5" cy="9.5" r="1.5"/><path d="m21 15-5-5L6 20"/>`)
)

type navRoute struct {
	page              string
	catalogEntity     string
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
	refreshSeq uint64 // 递增序号，丢弃过期的异步 Refresh 结果
}

type appNav struct {
	page             string
	history          []navRoute
	forward          []navRoute
	underTop         bool
	settingsUnderTop bool
}

type appSettings struct {
	prefs           config.Prefs
	settingsWin     *mygo.Window
	settingsPage    string
	settingsMessage string
}

type appCatalogData struct {
	newName    string
	categories []domain.Category
	yuvisions  []domain.Yuvision
	galleries  []domain.Gallery
	counts     map[string]int64
	backups    []domain.BackupInfo
	gridState  myui.GridState
}

type appVersionDraft struct {
	yvBrandID   string
	yvFirmID    string
	yvSeriesID  string
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
	gallerySearch     string
	filterBrandID     string
	filterFirmID      string
	filterSeriesID    string
	filterCategory    string
	filterYuvision    string
	gallerySort       string
	orphanDirs        []string
	yuvFilterCategory int64
}

type appSeriesQuery struct {
	seriesQueryKeyword  string
	seriesQueryResults  []domain.Series
	seriesQueryLoaded   bool
	seriesFilterBrandID string
	seriesFilterFirmID  string
	seriesFilterLevel   string
	seriesFilterEnergy  string
	seriesFilterResults []domain.Series
	seriesFilterLoaded  bool
}

type appSync struct {
	syncRunning bool
	syncMessage string
	syncGate    atomic.Bool // 后台 goroutine 用的启动闸门，防止重复启动同步
}

type appLightbox struct {
	lightboxWin   *mygo.Window
	lightboxImage *myui.Bitmap
	lightboxTitle string
}

type appCatalogEdit struct {
	catalogEntity string
	catalogEditID int64
	brands        []domain.Brand
	firms         []domain.Firm
	series        []domain.Series
	specs         []domain.Spec

	brandEdit struct {
		Name          string
		FirstLetter   string
		LogoURL       string
		LogoLocalPath string
		State         string
		HaveNewEnergy string
		APISort       string
	}

	firmEdit struct {
		Name          string
		PY            string
		Link          string
		LogoURL       string
		LogoLocalPath string
	}

	seriesEdit struct {
		Name          string
		GroupName     string
		State         string
		PriceMin      string
		PriceMax      string
		Place         string
		PY            string
		LogoURL       string
		LogoLocalPath string
		LevelName     string
		NewEnergy     string
		IsForeignCar  string
		IsTop         string
		FuelTypes     string
		CartypeName   string
		Tag1          string
		Rank          string
		UVRank        string
		HotRank       string
	}

	specEdit struct {
		GroupName           string
		Name                string
		SeriesName          string
		ModelYear           string
		BodyForm            string
		PriceMin            string
		PriceMax            string
		NewsPrice           string
		NewsID              string
		DealerID            string
		SaleType            string
		ImageURL            string
		ImageLocalPath      string
		ImageSHA256         string
		WhiteImageURL       string
		WhiteImageLocalPath string
		WhiteImageSHA256    string
	}
}

type appCovers struct {
	coverBitmaps map[string]*myui.Bitmap
}

// App 按职责内嵌状态分组；字段名提升后原有 a.field 访问保持不变。

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
}

func New(svc *service.Service, paths config.Paths) *App {
	prefs, err := config.LoadPrefs(paths.PrefsPath)
	if err != nil {
		prefs = config.DefaultPrefs()
	}
	i18n.Set(prefs.Language)
	applyAppearance(prefs.Appearance)
	page := prefs.DefaultPage
	switch page {
	case "gallery", "youvision", "catalog", "seriesQuery", "seriesFilter":
	default:
		page = "gallery"
	}

	app := &App{
		svc:           svc,
		paths:         paths,
		page:          page,
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

func (a *App) Refresh() {
	seq := a.refreshSeq + 1
	a.refreshSeq = seq

	// 在主线程快照筛选条件，避免后台读到正在被用户输入修改的字段。
	filter := a.galleryFilter()

	if a.win == nil {
		// 启动阶段（AttachWindow 之前）无窗口可回投，同步加载一次。
		a.applyRefresh(a.fetchRefresh(context.Background(), filter), seq, false)
		return
	}

	go func() {
		result := a.fetchRefresh(context.Background(), filter)
		a.win.Update(func() {
			// 只应用最后一次刷新的结果，丢弃快速输入/连续操作产生的旧结果。
			if seq != a.refreshSeq {
				return
			}
			a.applyRefresh(result, seq, true)
		})
	}()
}

func (a *App) currentRoute() navRoute {
	return navRoute{
		page:              a.page,
		catalogEntity:     a.catalogEntity,
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
	a.yuvFilterCategory = r.yuvFilterCategory
	a.filterCategory = r.filterCategory
	a.catalogEditID = 0
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

func (a *App) View(c *myui.Context) {
	a.applyTheme(c)
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
			if !c.Vibrancy() {
				side.Background(p.sidebar).LinearGradient(sidebarGradient(c.Theme().Dark))
			}
			side.Children(func() {
				a.sidebar(c)
			})

			main := myui.Box(c.Key("main")).Grow(1).Background(p.bg)
			main.Children(func() {
				pad := float32(46)
				myui.Scroll(c.Key("content-"+a.page)).
					Grow(1).
					Padding(pad, 22, 24, 22).
					Children(func() {
						mark := myui.Box(c.Key("content-start-mark")).Height(0).Bounds()
						if mark.W > 0 {
							a.underTop = mark.Y < pad-0.5
						}
						a.renderPage(c)
					})
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
			names := []string{follow, "English", "简体中文"}
			chosen := follow
			switch a.prefs.Language {
			case "en", "en-US":
				chosen = "English"
			case "zh-CN":
				chosen = "简体中文"
			}
			if myui.Select(c.Key("language"), &chosen, names).Label(i18n.T("界面语言")).Width(170).Changed() {
				code := "system"
				switch chosen {
				case "English":
					code = "en"
				case "简体中文":
					code = "zh-CN"
				}
				a.prefs.Language = code
				i18n.Set(code)
				a.seriesFilterEnergy = ""
				a.savePrefs()
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
					a.savePrefs()
					applyAppearance(value)
					a.invalidateWindows()
				},
			)
		})
		a.settingsRow(c, i18n.T("默认页面"), "", func() {
			codes := []string{"seriesQuery", "seriesFilter", "catalog", "gallery", "youvision"}
			names := []string{
				i18n.T("车型查询"), i18n.T("筛选查询"), i18n.T("数据管理"),
				i18n.T("图集管理"), i18n.T("版型管理"),
			}
			chosen := names[0]
			for i, code := range codes {
				if a.prefs.DefaultPage == code {
					chosen = names[i]
					break
				}
			}
			if myui.Select(c.Key("default-page"), &chosen, names).Label(i18n.T("默认页面")).Width(170).Changed() {
				for i, name := range names {
					if chosen == name {
						a.prefs.DefaultPage = codes[i]
						break
					}
				}
				a.savePrefs()
			}
		})
	})

	if strings.TrimSpace(a.settingsMessage) != "" {
		myui.Text(c, a.settingsMessage).FontSize(12.5).TextColor(p.muted)
	}
}

func (a *App) categorySettings(c *myui.Context) {
	a.renderCategories(c)
}

func (a *App) syncSettings(c *myui.Context) {
	a.renderSync(c)
}

func (a *App) systemSettings(c *myui.Context) {
	a.renderSystemSettings(c)
}

func (a *App) updateSettings(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)

	a.settingsGroup(c, func() {
		a.settingsRow(c, i18n.T("自动检查更新"), i18n.T("每天检查一次是否有新版本"), func() {
			on := a.prefs.AutoCheckUpdates
			if myui.Switch(c.Key("update-checks"), &on).Label(i18n.T("自动检查更新")).Changed() {
				a.prefs.AutoCheckUpdates = on
				a.savePrefs()
			}
		})
		a.settingsRow(c, i18n.T("自动下载并安装"), i18n.T("在后台下载新版本，下次打开时生效"), func() {
			on := a.prefs.AutoInstallUpdates
			if myui.Switch(c.Key("update-downloads"), &on).Label(i18n.T("自动下载并安装")).
				Disabled(!a.prefs.AutoCheckUpdates).Changed() {
				a.prefs.AutoInstallUpdates = on
				a.savePrefs()
			}
		})
		a.settingsRow(c, fmt.Sprintf(i18n.T("当前版本 %s"), mygo.App.Version()), "", func() {
			if myui.Button(c.Key("check-updates"), i18n.T("检查更新")).Clicked() {
				a.settingsMessage = i18n.T("当前尚未配置更新服务器")
			}
		})
	})

	if strings.TrimSpace(a.settingsMessage) != "" {
		myui.Text(c, a.settingsMessage).FontSize(12.5).TextColor(p.muted)
	}
}

func (a *App) aboutSettings(c *myui.Context) {
	a.settingsGroup(c, func() {
		a.settingsRow(c, i18n.T("应用名称"), "autoCarSystem", func() {})
		a.settingsRow(c, i18n.T("版本"), mygo.App.Version(), func() {})
		a.settingsRow(c, i18n.T("描述"), i18n.T("本地图集与版型数据管理工具"), func() {})
	})
}
