package uiapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"autocarsystem/internal/config"
	"autocarsystem/internal/domain"
	"autocarsystem/internal/i18n"
	"autocarsystem/internal/media"
	"autocarsystem/internal/service"
	autosync "autocarsystem/internal/sync"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/glass"
	myui "github.com/egoist/mygo/ui"
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

func paletteFor(dark bool) palette {
	if dark {
		return palette{
			dark:     true,
			bg:       myui.Hex("#141416"),
			sidebar:  myui.Hex("#1c1c1f"),
			text:     myui.Hex("#f5f5f7"),
			muted:    myui.RGBA(235, 235, 245, 0.56),
			faint:    myui.RGBA(235, 235, 245, 0.3),
			hairline: myui.RGBA(255, 255, 255, 0.07),
			hover:    myui.RGBA(255, 255, 255, 0.06),
			press:    myui.RGBA(255, 255, 255, 0.1),
			accent:   myui.Oklch(0.66, 0.2, 22),
			onAccent: myui.Hex("#ffffff"),
		}
	}
	return palette{
		bg:       myui.Hex("#ffffff"),
		sidebar:  myui.Hex("#f4f4f6"),
		text:     myui.Hex("#1d1d1f"),
		muted:    myui.RGBA(60, 60, 67, 0.62),
		faint:    myui.RGBA(60, 60, 67, 0.32),
		hairline: myui.RGBA(0, 0, 0, 0.07),
		hover:    myui.RGBA(0, 0, 0, 0.04),
		press:    myui.RGBA(0, 0, 0, 0.08),
		accent:   myui.Oklch(0.6, 0.21, 24),
		onAccent: myui.Hex("#ffffff"),
	}
}

func sidebarGradient(dark bool) myui.LinearGradient {
	if dark {
		return myui.LinearGradient{Angle: 180, From: myui.Hex("#1c1c1f"), To: myui.Hex("#26262b")}
	}
	return myui.LinearGradient{Angle: 180, From: myui.Hex("#f4f4f6"), To: myui.Hex("#e9e9ef")}
}

func navIconButton(c *myui.Context, key string, icon *myui.SVG, label string) myui.Element {
	p := paletteFor(c.Theme().Dark)
	b := myui.ButtonBase(c.Key(key)).Size(34, 34).Center().Radius(999).Label(label).
		TextColor(p.muted).FocusRing(false)
	bg := myui.Transparent
	if b.Hovered() {
		bg = myui.RGBA(128, 128, 128, 0.14)
	}
	b.Background(b.AnimateColor("bg", bg, myui.Animation{Duration: 120 * time.Millisecond}))
	b.Children(func() {
		myui.Icon(c, icon).Size(18, 18).TextColor(p.muted)
	})
	return b
}

func edgeShown(edge myui.Element, under bool) float32 {
	target := float32(0)
	if under {
		target = 1
	}
	return edge.AnimateTo("shown", target, myui.Animation{Duration: 180 * time.Millisecond})
}

func strokeIcon(paths string) *myui.SVG {
	return myui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">` + paths + `</svg>`))
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

type App struct {
	win  *mygo.Window
	tray any
	svc  *service.Service

	paths       config.Paths
	page        string
	history     []navRoute
	forward     []navRoute
	status      string
	quitting    bool
	syncRunning bool
	syncMessage string

	prefs           config.Prefs
	settingsWin     *mygo.Window
	settingsPage    string
	settingsMessage string

	newName    string
	categories []domain.Category
	yuvisions  []domain.Yuvision
	galleries  []domain.Gallery
	counts     map[string]int64
	backups    []domain.BackupInfo

	yvBrandID   string
	yvFirmID    string
	yvSeriesID  string
	yvCategory  string
	yvName      string
	yvDesc      string
	yvUploadID  string
	yvUploadSrc string

	galYuvisionID string
	galTitle      string
	galUploadID   string
	galUploadSrc  string

	gallerySearch  string
	filterBrandID  string
	filterFirmID   string
	filterSeriesID string
	filterCategory string
	filterYuvision string
	gallerySort    string
	orphanDirs     []string

	seriesQueryKeyword  string
	seriesFilterBrandID string
	seriesFilterFirmID  string
	seriesFilterLevel   string
	seriesFilterEnergy  string
	yuvFilterCategory   int64
	underTop            bool
	settingsUnderTop    bool

	gridState     myui.GridState
	coverBitmaps  map[string]*myui.Bitmap
	lightboxWin   *mygo.Window
	lightboxImage *myui.Bitmap
	lightboxTitle string

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
func (a *App) AttachTray(tray any)           { a.tray = tray }
func (a *App) SetQuitting(v bool)            { a.quitting = v }
func (a *App) Quitting() bool                { return a.quitting }

func (a *App) Refresh() {
	ctx := context.Background()

	categories, err := a.svc.ListCategories(ctx)
	if err != nil {
		a.setStatus(err.Error())
	} else {
		a.categories = categories
	}

	yuvisions, err := a.svc.ListYuvisions(ctx, 200)
	if err != nil {
		a.setStatus(err.Error())
	} else {
		a.yuvisions = yuvisions
	}

	galleries, err := a.svc.SearchGalleries(ctx, a.galleryFilter(), 200)
	if err != nil {
		a.setStatus(err.Error())
	} else {
		a.galleries = galleries
		a.loadCoverBitmaps()
	}

	counts, err := a.svc.Counts(ctx)
	if err != nil {
		a.setStatus(err.Error())
	} else {
		a.counts = counts
	}

	if brands, err := a.svc.ListBrands(ctx, 200); err == nil {
		a.brands = brands
	}
	if firms, err := a.svc.ListFirms(ctx, 200); err == nil {
		a.firms = firms
	}
	if series, err := a.svc.ListSeries(ctx, 200); err == nil {
		a.series = series
	}
	if specs, err := a.svc.ListSpecs(ctx, 200); err == nil {
		a.specs = specs
	}
	if backups, err := a.svc.ListBackups(); err == nil {
		a.backups = backups
	}
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

	root := myui.Box(c).Fill().Background(p.bg)
	root.Children(func() {
		page := myui.Row(c).Fill().AlignItems(myui.Stretch)
		page.Children(func() {
			side := myui.Column(c).Width(232).Background(p.sidebar).LinearGradient(sidebarGradient(c.Theme().Dark))
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

func (a *App) sidebar(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)

	myui.Column(c).Fill().Padding(14, 14, 14, 14).Gap(4).DragWindow().Children(func() {
		myui.Row(c).AlignItems(myui.Center).Gap(10).Margin(0, 0, 12, 6).Children(func() {
			myui.Box(c).Size(36, 36).Radius(10).Background(p.accent).Center().Children(func() {
				myui.Icon(c, iconLogo).Size(20, 20).TextColor(p.onAccent)
			})
			myui.Column(c).Gap(2).Children(func() {
				myui.Text(c, "新方圆").FontSize(18).FontWeight(700).TextColor(p.text)
				myui.Text(c, i18n.T("版型图集管理")).FontSize(12).TextColor(p.muted)
			})
		})

		myui.Scroll(c.Key("nav")).Grow(1).Margin(0, -6).Padding(0, 6).Gap(4).Children(func() {
			a.navItemKey(c, "nav-series-query", iconCatalog, i18n.T("车型查询"), "seriesQuery")
			a.navItemKey(c, "nav-series-filter", iconSync, i18n.T("筛选查询"), "seriesFilter")

			a.navSection(c, i18n.T("版型管理"))
			for _, item := range a.categories {
				a.yuvisionCategoryNavItem(c, item.Name, item.ID, iconLayers)
			}

			a.navSection(c, i18n.T("图集管理"))
			for _, item := range a.categories {
				a.galleryCategoryNavItem(c, item.Name, strconv.FormatInt(item.ID, 10), iconGallery)
			}
		})

		a.settingsEntry(c)
	})
}

func (a *App) settingsEntry(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	b := myui.ButtonBase(c.Key("settings-entry")).
		Height(40).
		PaddingX(8).
		Gap(10).
		Radius(10).
		AlignItems(myui.Center).
		Justify(myui.Start).
		FocusRing(false).
		Label(i18n.T("设置"))

	if b.Hovered() {
		b.Background(p.hover)
	}
	b.Children(func() {
		myui.Box(c).Size(26, 26).Center().Children(func() {
			myui.Icon(c, iconSettings).Size(15, 15).TextColor(p.muted)
		})
		myui.Text(c, i18n.T("设置")).FontSize(13).FontWeight(500).TextColor(p.muted)
	})
	if b.Clicked() {
		a.openSettings()
	}
}

func (a *App) navSection(c *myui.Context, label string) {
	p := paletteFor(c.Theme().Dark)
	myui.Text(c, label).FontSize(11.5).FontWeight(600).TextColor(p.faint).Margin(18, 4, 6, 6)
}

func (a *App) navItemKey(c *myui.Context, key string, icon *myui.SVG, label string, page string) {
	p := paletteFor(c.Theme().Dark)
	selected := a.page == page

	b := myui.ButtonBase(c.Key(key)).
		Height(34).
		PaddingX(10).
		Gap(10).
		Radius(8).
		AlignItems(myui.Center).
		Justify(myui.Start).
		FocusRing(false).
		Label(label)

	bg := myui.Transparent
	if b.Hovered() {
		bg = p.hover
	}
	if selected {
		bg = p.press
	}
	b.Background(bg)

	color := p.text
	if selected {
		color = p.accent
	}
	b.Children(func() {
		myui.Icon(c, icon).Size(17, 17).TextColor(color)
		myui.Text(c, label).FontSize(13.5).FontWeight(500).TextColor(p.text)
	})

	if b.Clicked() {
		a.open(navRoute{page: page})
	}
}

func (a *App) navLink(c *myui.Context, key string, icon *myui.SVG, label string, selected bool, clicked func()) {
	p := paletteFor(c.Theme().Dark)
	b := myui.ButtonBase(c.Key(key)).
		Height(34).
		PaddingX(8).
		Gap(10).
		Radius(8).
		AlignItems(myui.Center).
		Justify(myui.Start).
		FocusRing(false).
		Label(label)

	bg := myui.Transparent
	if b.Hovered() {
		bg = p.hover
	}
	if selected {
		bg = p.press
	}
	b.Background(bg)

	color := p.text
	if selected {
		color = p.accent
	}
	b.Children(func() {
		myui.Icon(c, icon).Size(17, 17).TextColor(color)
		myui.Text(c, label).FontSize(13).TextColor(p.text)
	})
	if b.Clicked() {
		clicked()
	}
}

func (a *App) catalogNavItem(c *myui.Context, entity string, label string, icon *myui.SVG) {
	selected := a.page == "catalog" && a.catalogEntity == entity
	a.navLink(c, "nav-catalog-"+entity, icon, label, selected, func() {
		a.open(navRoute{page: "catalog", catalogEntity: entity})
	})
}

func (a *App) yuvisionCategoryNavItem(c *myui.Context, label string, categoryID int64, icon *myui.SVG) {
	selected := a.page == "youvision" && a.yuvFilterCategory == categoryID
	a.navLink(c, fmt.Sprintf("nav-yuv-%d", categoryID), icon, label, selected, func() {
		a.open(navRoute{page: "youvision", yuvFilterCategory: categoryID})
	})
}

func (a *App) galleryCategoryNavItem(c *myui.Context, label string, category string, icon *myui.SVG) {
	selected := a.page == "gallery" && a.filterCategory == category
	a.navLink(c, "nav-gal-"+category, icon, label, selected, func() {
		a.open(navRoute{page: "gallery", filterCategory: category})
	})
}

func (a *App) openSettings() {
	if a.settingsWin != nil {
		a.settingsWin.Show()
		a.settingsWin.Focus()
		return
	}

	win := mygo.NewWindow(mygo.WindowOptions{
		Title:                i18n.T("设置"),
		Width:                780,
		Height:               640,
		MinWidth:             640,
		MinHeight:            420,
		DisableMaximize:      true,
		DisableMinimize:      true,
		TitleBarStyle:        mygo.TitleBarHidden,
		TitleBarHeight:       46,
		TrafficLightPosition: &mygo.Point{X: 20, Y: 20},
		AutoHideMenuBar:      true,
		StateKey:             "autocarsystem-settings",
		BackgroundColor:      "light-dark(#f4f4f6, #1c1c1f)",
		Content:              myui.View(a.settingsView),
	})
	win.OnClosed(func() {
		a.settingsWin = nil
	})
	a.settingsWin = win
	win.SetWindowButtonsTheme(mygo.ThemeSystem)
	win.Show()
}

func (a *App) settingsView(c *myui.Context) {
	a.applyTheme(c)
	p := paletteFor(c.Theme().Dark)

	root := myui.Box(c).Fill().Background(p.bg)
	root.Children(func() {
		page := myui.Row(c).Fill().AlignItems(myui.Stretch)
		page.Children(func() {
			side := myui.Column(c).Width(176).Background(p.sidebar)
			side.Children(func() {
				myui.Column(c).Fill().Padding(18, 12, 14, 12).Gap(2).DragWindow().Children(func() {
					myui.Text(c, i18n.T("设置")).FontSize(18).FontWeight(700).TextColor(p.text).Margin(0, 0, 16, 8)
					a.settingsNavItem(c, iconSettings, i18n.T("通用"), "general")
					a.settingsNavItem(c, iconCategory, i18n.T("分类"), "category")
					a.settingsNavItem(c, iconTransfer, i18n.T("同步"), "sync")
					a.settingsNavItem(c, iconCatalog, i18n.T("系统"), "system")
					a.settingsNavItem(c, iconSync, i18n.T("更新"), "updates")
					a.settingsNavItem(c, iconAbout, i18n.T("关于"), "about")
				})
			})

			main := myui.Box(c.Key("settings-main")).Grow(1).Background(p.bg)
			main.Children(func() {
				pad := float32(46)
				myui.Scroll(c.Key("settings-page-"+a.settingsPage)).
					Grow(1).
					Padding(pad+4, 28, 28, 28).
					Gap(10).
					Children(func() {
						mark := myui.Box(c.Key("settings-start-mark")).Height(0).Bounds()
						if mark.W > 0 {
							a.settingsUnderTop = mark.Y < pad-0.5
						}
						a.settingsContent(c)
					})
				edge := myui.Box(c.Key("settings-top-edge")).Absolute().Top(0).Left(0).Right(0).Height(pad).DragWindow().
					Material(glass.ScrollEdge{Hard: true, Background: p.bg})
				header := myui.Row(c.Key("settings-header")).Absolute().Top(0).Left(0).Right(0).Height(pad).
					Padding(0, 0, 0, 28).AlignItems(myui.Center).DragWindow()
				header.Children(func() {
					myui.Text(c, a.settingsTitle()).FontSize(15).FontWeight(700).TextColor(p.text)
				})
				edge.Opacity(edgeShown(edge, a.settingsUnderTop || header.Hovered()))
			})
		})
	})
}

func (a *App) settingsTitle() string {
	switch a.settingsPage {
	case "category":
		return i18n.T("分类")
	case "sync":
		return i18n.T("同步")
	case "system":
		return i18n.T("系统")
	case "updates":
		return i18n.T("更新")
	case "about":
		return i18n.T("关于")
	default:
		return i18n.T("通用")
	}
}

func (a *App) settingsNavItem(c *myui.Context, icon *myui.SVG, label string, page string) {
	p := paletteFor(c.Theme().Dark)
	selected := a.settingsPage == page

	b := myui.ButtonBase(c.Key("settings-" + page)).
		Height(32).
		PaddingX(9).
		Gap(9).
		Radius(8).
		AlignItems(myui.Center).
		Justify(myui.Start).
		FocusRing(false).
		Label(label)

	bg := myui.Transparent
	if b.Hovered() {
		bg = p.hover
	}
	if selected {
		bg = p.press
	}
	b.Background(bg)

	color := p.text
	if selected {
		color = p.accent
	}
	b.Children(func() {
		myui.Icon(c, icon).Size(16, 16).TextColor(color)
		myui.Text(c, label).FontSize(13).FontWeight(500).TextColor(color)
	})
	if b.Clicked() {
		a.settingsPage = page
	}
}

func (a *App) settingsContent(c *myui.Context) {
	switch a.settingsPage {
	case "category":
		a.categorySettings(c)
	case "sync":
		a.syncSettings(c)
	case "system":
		a.systemSettings(c)
	case "updates":
		a.updateSettings(c)
	case "about":
		a.aboutSettings(c)
	default:
		a.generalSettings(c)
	}
}

// settingsGroup 按 mygo-music 的样式绘制设置卡片：圆角、描边、行间分隔线。
func (a *App) settingsGroup(c *myui.Context, build func()) {
	p := paletteFor(c.Theme().Dark)
	myui.Column(c).Radius(12).Background(p.bg).Border(1, p.hairline).Dividers(1, p.hairline).Children(build)
}

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

func (a *App) settingsChoice(c *myui.Context, values []string, labels []string, selected string, onChanged func(string)) {
	p := paletteFor(c.Theme().Dark)
	myui.Row(c).Gap(6).Children(func() {
		for i, value := range values {
			label := labels[i]
			active := selected == value
			b := myui.ButtonBase(c.Key("choice-" + label)).
				Height(30).
				PaddingX(10).
				Radius(7).
				AlignItems(myui.Center).
				FocusRing(false).
				Label(label)

			color := p.text
			switch {
			case active:
				b.Background(p.press)
				color = p.accent
			case b.Hovered():
				b.Background(p.hover)
			}
			b.Children(func() {
				myui.Text(c, label).FontSize(12.5).TextColor(color)
			})
			if b.Clicked() {
				onChanged(value)
			}
		}
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

func (a *App) savePrefs() {
	if err := config.SavePrefs(a.paths.PrefsPath, a.prefs); err != nil {
		a.settingsMessage = err.Error()
		return
	}
	a.settingsMessage = ""
}

func applyAppearance(appearance string) {
	src := mygo.ThemeSystem
	switch appearance {
	case "light":
		src = mygo.ThemeLight
	case "dark":
		src = mygo.ThemeDark
	}
	mygo.Theme.SetSource(src)
}

func (a *App) applyTheme(c *myui.Context) {
	t := *c.Theme()
	p := paletteFor(t.Dark)
	black, white := myui.Hex("#000000"), myui.Hex("#ffffff")
	t.Accent = p.accent
	t.AccentText = p.onAccent
	t.AccentHover = p.accent.MixOklab(black, 0.1)
	t.AccentPressed = p.accent.MixOklab(black, 0.2)
	if p.dark {
		t.AccentHover = p.accent.MixOklab(white, 0.15)
		t.AccentPressed = p.accent.MixOklab(black, 0.15)
	}
	c.SetTheme(&t)
}

func (a *App) statusLine() string {
	if strings.TrimSpace(a.status) != "" {
		return i18n.T("状态：") + a.status
	}
	return fmt.Sprintf(
		i18n.T("根目录：%s    图集：%d    版型：%d    分类：%d"),
		a.paths.Root,
		a.countInt("gallery"),
		a.countInt("youvision"),
		a.countInt("category"),
	)
}

func (a *App) setStatus(message string) {
	a.status = strings.TrimSpace(i18n.T(message))
}

func (a *App) uiStatus(message string) {
	if a.win != nil {
		a.win.Update(func() {
			a.setStatus(message)
			a.Refresh()
		})
		return
	}
	a.setStatus(message)
	a.Refresh()
}

func (a *App) renderPage(c *myui.Context) {
	switch a.page {
	case "seriesQuery":
		a.renderSeriesQuery(c)
	case "seriesFilter":
		a.renderSeriesFilter(c)
	case "youvision":
		a.renderYuvision(c)
	case "catalog":
		a.renderCatalog(c)
	default:
		a.renderGallery(c)
	}
}

func (a *App) renderSeriesQuery(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	myui.Text(c, i18n.T("车型查询")).FontSize(22).FontWeight(700).TextColor(p.text)

	a.settingsGroup(c, func() {
		a.settingsRow(c, i18n.T("关键词"), i18n.T("按车系名称、拼音或 ID 查询"), func() {
			myui.TextInput(c.Key("series-query-keyword"), &a.seriesQueryKeyword).Placeholder(i18n.T("输入关键词")).Width(240)
			if myui.Button(c.Key("series-query-run"), i18n.T("查询")).Clicked() {
				a.Refresh()
			}
		})
	})

	keyword := strings.ToLower(strings.TrimSpace(a.seriesQueryKeyword))
	matched := make([]domain.Series, 0, len(a.series))
	for _, item := range a.series {
		if keyword == "" {
			matched = append(matched, item)
			continue
		}
		text := strings.ToLower(item.Name + " " + item.PY + " " + item.GroupName + " " + strconv.FormatInt(item.ID, 10))
		if strings.Contains(text, keyword) {
			matched = append(matched, item)
		}
	}
	a.renderSeriesResults(c, matched)
}

func (a *App) renderSeriesFilter(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	myui.Text(c, i18n.T("筛选查询")).FontSize(22).FontWeight(700).TextColor(p.text)

	a.settingsGroup(c, func() {
		a.settingsRow(c, i18n.T("品牌 / 厂商 ID"), i18n.T("只查询指定品牌或厂商的车系"), func() {
			myui.TextInput(c.Key("series-filter-brand"), &a.seriesFilterBrandID).Placeholder(i18n.T("品牌 ID")).Width(110)
			myui.TextInput(c.Key("series-filter-firm"), &a.seriesFilterFirmID).Placeholder(i18n.T("厂商 ID")).Width(110)
		})
		a.settingsRow(c, i18n.T("级别 / 能源"), i18n.T("级别支持模糊匹配"), func() {
			myui.TextInput(c.Key("series-filter-level"), &a.seriesFilterLevel).Placeholder(i18n.T("级别，如 SUV")).Width(140)
			myui.Select(c.Key("series-filter-energy"), &a.seriesFilterEnergy, []string{i18n.T("全部"), i18n.T("新能源"), i18n.T("燃油")}).Width(110)
		})
		a.settingsRow(c, "", "", func() {
			if myui.Button(c.Key("series-filter-run"), i18n.T("查询")).Clicked() {
				a.Refresh()
			}
		})
	})

	brandID, err := parseOptionalID(a.seriesFilterBrandID)
	if err != nil {
		a.setStatus(i18n.T("品牌 ID 无效"))
	}
	firmID, err := parseOptionalID(a.seriesFilterFirmID)
	if err != nil {
		a.setStatus(i18n.T("厂商 ID 无效"))
	}
	level := strings.ToLower(strings.TrimSpace(a.seriesFilterLevel))

	matched := make([]domain.Series, 0, len(a.series))
	for _, item := range a.series {
		if brandID != 0 && item.BrandID != brandID {
			continue
		}
		if firmID != 0 && item.FirmID != firmID {
			continue
		}
		if level != "" && !strings.Contains(strings.ToLower(item.LevelName), level) {
			continue
		}
		switch a.seriesFilterEnergy {
		case i18n.T("新能源"):
			if item.NewEnergy == 0 {
				continue
			}
		case i18n.T("燃油"):
			if item.NewEnergy != 0 {
				continue
			}
		}
		matched = append(matched, item)
	}
	a.renderSeriesResults(c, matched)
}

func (a *App) renderSeriesResults(c *myui.Context, items []domain.Series) {
	p := paletteFor(c.Theme().Dark)
	myui.Text(c, fmt.Sprintf(i18n.T("共 %d 个车系"), len(items))).FontSize(13).TextColor(p.muted).Margin(0, 10, 8, 0)
	if len(items) == 0 {
		myui.Text(c, i18n.T("没有匹配的车系。")).FontSize(13).TextColor(p.muted)
		return
	}

	brandNames := make(map[int64]string, len(a.brands))
	for _, item := range a.brands {
		brandNames[item.ID] = item.Name
	}
	firmNames := make(map[int64]string, len(a.firms))
	for _, item := range a.firms {
		firmNames[item.ID] = item.Name
	}

	const limit = 200
	a.settingsGroup(c, func() {
		for i, item := range items {
			if i >= limit {
				break
			}
			current := item
			name := current.Name
			if current.GroupName != "" && current.GroupName != current.Name {
				name += " · " + current.GroupName
			}
			desc := fmt.Sprintf(i18n.T("品牌 %s  ·  厂商 %s  ·  级别 %s  ·  指导价 %.2f-%.2f"),
				brandNames[current.BrandID], firmNames[current.FirmID], current.LevelName, current.PriceMin, current.PriceMax)
			if current.NewEnergy != 0 {
				desc += i18n.T("  ·  新能源")
			}
			a.settingsRow(c, fmt.Sprintf("%s  #%d", name, current.ID), desc, func() {})
		}
	})
	if len(items) > limit {
		myui.Text(c, fmt.Sprintf(i18n.T("仅显示前 %d 条，共 %d 条。"), limit, len(items))).FontSize(12).TextColor(p.muted)
	}
}

func (a *App) renderGallery(c *myui.Context) {
	myui.Text(c, i18n.T("图集列表")).FontSize(18).Bold()
	myui.Row(c).Gap(8).Children(func() {
		if myui.TextInput(c.Key("gal-yuvision"), &a.galYuvisionID).Placeholder(i18n.T("版型 ID")).Submitted() {
			a.createGallery()
		}
		myui.TextInput(c.Key("gal-title"), &a.galTitle).Placeholder(i18n.T("图集标题"))
		if myui.Button(c, i18n.T("新增图集")).Clicked() {
			a.createGallery()
		}
	})
	myui.Text(c, i18n.T("新增后请选择图片上传；封面自动取第一张图片。"))

	myui.Row(c).Gap(8).Children(func() {
		if myui.TextInput(c.Key("gal-search"), &a.gallerySearch).Placeholder(i18n.T("搜索标题 / 图集 ID / 版型 ID")).Changed() {
			a.Refresh()
		}
		if myui.Button(c, i18n.T("搜索")).Clicked() {
			a.Refresh()
		}
		if myui.Button(c, i18n.T("清除筛选")).Clicked() {
			a.gallerySearch = ""
			a.filterBrandID = ""
			a.filterFirmID = ""
			a.filterSeriesID = ""
			a.filterCategory = ""
			a.filterYuvision = ""
			a.Refresh()
		}
	})
	myui.Row(c).Gap(8).Children(func() {
		if myui.TextInput(c.Key("filter-brand"), &a.filterBrandID).Placeholder(i18n.T("品牌 ID")).Changed() {
			a.Refresh()
		}
		if myui.TextInput(c.Key("filter-firm"), &a.filterFirmID).Placeholder(i18n.T("厂商 ID")).Changed() {
			a.Refresh()
		}
		if myui.TextInput(c.Key("filter-series"), &a.filterSeriesID).Placeholder(i18n.T("车系 ID")).Changed() {
			a.Refresh()
		}
	})
	myui.Row(c).Gap(8).Children(func() {
		if myui.TextInput(c.Key("filter-category"), &a.filterCategory).Placeholder(i18n.T("分类 ID")).Changed() {
			a.Refresh()
		}
		if myui.TextInput(c.Key("filter-yuvision"), &a.filterYuvision).Placeholder(i18n.T("版型 ID")).Changed() {
			a.Refresh()
		}
	})
	myui.Row(c).Gap(8).Children(func() {
		if myui.Button(c, i18n.T("更新时间")).Clicked() {
			a.gallerySort = "updated_desc"
			a.Refresh()
		}
		if myui.Button(c, i18n.T("创建时间")).Clicked() {
			a.gallerySort = "created_desc"
			a.Refresh()
		}
		if myui.Button(c, i18n.T("标题")).Clicked() {
			a.gallerySort = "title_asc"
			a.Refresh()
		}
		if myui.Button(c, i18n.T("图片数")).Clicked() {
			a.gallerySort = "image_count_desc"
			a.Refresh()
		}
		if myui.Button(c, "ID").Clicked() {
			a.gallerySort = "id_desc"
			a.Refresh()
		}
	})

	if len(a.galleries) > 0 {
		myui.GridView(c, &a.gridState, len(a.galleries), 180, 150, func(i int) {
			item := a.galleries[i]
			myui.Column(c.Key(fmt.Sprintf("gallery-card-%d", item.ID))).Gap(6).Children(func() {
				path := a.galleryCoverPath(item)
				if bitmap := a.bitmapAt(path); bitmap != nil {
					myui.Image(c, bitmap).
						Size(168, 118).
						Fit(myui.Cover).
						Radius(8).
						OnClick(func() { a.openLightbox(item) })
				} else {
					myui.Text(c, i18n.T("暂无图片"))
				}
				myui.Text(c, fmt.Sprintf("#%d %s", item.ID, item.Title))
				if myui.Button(c, i18n.T("删除")).Clicked() {
					a.deleteGallery(item.ID)
				}
			})
		})
	} else {
		myui.Text(c, i18n.T("暂无图集。请先创建版型，再创建图集。"))
	}

	myui.Text(c, i18n.T("图集图片上传")).FontSize(16).Bold()
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("gal-upload-id"), &a.galUploadID).Placeholder(i18n.T("图集 ID"))
		myui.TextInput(c.Key("gal-upload-src"), &a.galUploadSrc).Placeholder(i18n.T("本地图片完整路径"))
		if myui.Button(c, i18n.T("选择图片")).Clicked() {
			idText := a.galUploadID
			go a.pickGalleryImages(idText)
		}
		if myui.Button(c, i18n.T("上传路径")).Clicked() {
			a.uploadGalleryImage()
		}
	})
}

func (a *App) renderYuvision(c *myui.Context) {
	myui.Text(c, i18n.T("版型管理")).FontSize(18).Bold()
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("yv-brand"), &a.yvBrandID).Placeholder(i18n.T("品牌 ID"))
		myui.TextInput(c.Key("yv-firm"), &a.yvFirmID).Placeholder(i18n.T("厂商 ID"))
		myui.TextInput(c.Key("yv-series"), &a.yvSeriesID).Placeholder(i18n.T("车系 ID"))
		myui.TextInput(c.Key("yv-category"), &a.yvCategory).Placeholder(i18n.T("分类 ID"))
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("yv-name"), &a.yvName).Placeholder(i18n.T("版型名称"))
		myui.TextInput(c.Key("yv-desc"), &a.yvDesc).Placeholder(i18n.T("描述"))
		if myui.Button(c, i18n.T("新增版型")).Clicked() {
			a.createYuvision()
		}
	})
	myui.Text(c, i18n.T("新增版型时不允许选择禁用分类；车系、品牌、厂商必须匹配。"))

	yuvisions := make([]domain.Yuvision, 0, len(a.yuvisions))
	for _, item := range a.yuvisions {
		if a.yuvFilterCategory == 0 || item.CategoryID == a.yuvFilterCategory {
			yuvisions = append(yuvisions, item)
		}
	}
	for _, item := range yuvisions {
		current := item
		myui.Row(c.Key(fmt.Sprintf("youvision-%d", current.ID))).Gap(8).Children(func() {
			myui.Text(c, fmt.Sprintf(i18n.T("版型 #%d  %s  分类 #%d"), current.ID, current.Name, current.CategoryID)).Grow(1)
			if myui.Button(c, i18n.T("删除")).Clicked() {
				a.deleteYuvision(current.ID)
			}
		})
	}

	myui.Text(c, i18n.T("版型图片上传")).FontSize(16).Bold()
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("yv-upload-id"), &a.yvUploadID).Placeholder(i18n.T("版型 ID"))
		myui.TextInput(c.Key("yv-upload-src"), &a.yvUploadSrc).Placeholder(i18n.T("本地图片完整路径"))
		if myui.Button(c, i18n.T("选择图片")).Clicked() {
			idText := a.yvUploadID
			go a.pickYuvisionImages(idText)
		}
		if myui.Button(c, i18n.T("上传路径")).Clicked() {
			a.uploadYuvisionImage()
		}
	})
	if a.countInt("youvision") == 0 {
		myui.Text(c, i18n.T("暂无版型。请先同步目录数据或手工准备品牌、厂商、车系记录。"))
	}
}

func (a *App) renderCategories(c *myui.Context) {
	myui.Row(c).Gap(8).Children(func() {
		if myui.TextInput(c, &a.newName).Placeholder(i18n.T("新分类名称")).Submitted() {
			a.createCategory()
		}
		if myui.Button(c, i18n.T("新增分类")).Clicked() {
			a.createCategory()
		}
	})
	for _, category := range a.categories {
		item := category
		myui.Row(c.Key(item.ID)).Gap(8).Children(func() {
			label := item.Name
			if !item.Enabled {
				label += i18n.T("（已禁用）")
			}
			myui.Text(c, label).Grow(1)
			action := i18n.T("禁用")
			if !item.Enabled {
				action = i18n.T("启用")
			}
			if myui.Button(c, action).Clicked() {
				err := a.svc.SetCategoryEnabled(context.Background(), item.ID, !item.Enabled)
				if err != nil {
					a.setStatus(err.Error())
				} else {
					a.setStatus("")
					a.Refresh()
				}
			}
			if myui.Button(c, i18n.T("删除")).Clicked() {
				if err := a.svc.DeleteCategory(context.Background(), item.ID); err != nil {
					a.setStatus(err.Error())
				} else {
					a.setStatus("")
					a.Refresh()
				}
			}
		})
	}
}

func (a *App) createCategory() {
	name := strings.TrimSpace(a.newName)
	if name == "" {
		a.setStatus(i18n.T("分类名称不能为空"))
		return
	}
	if err := a.svc.CreateCategory(context.Background(), name, len(a.categories)); err != nil {
		a.setStatus(err.Error())
		return
	}
	a.newName = ""
	a.setStatus("")
	a.Refresh()
}

func (a *App) createYuvision() {
	ctx := context.Background()
	brandID, err := parseID(a.yvBrandID)
	if err != nil {
		a.setStatus(i18n.T("品牌 ID 无效"))
		return
	}
	firmID, err := parseID(a.yvFirmID)
	if err != nil {
		a.setStatus(i18n.T("厂商 ID 无效"))
		return
	}
	seriesID, err := parseID(a.yvSeriesID)
	if err != nil {
		a.setStatus(i18n.T("车系 ID 无效"))
		return
	}
	categoryID, err := parseID(a.yvCategory)
	if err != nil {
		a.setStatus(i18n.T("分类 ID 无效"))
		return
	}

	id, err := a.svc.CreateYuvision(ctx, brandID, firmID, seriesID, categoryID, a.yvName, a.yvDesc)
	if err != nil {
		a.setStatus(err.Error())
		return
	}
	a.yvName, a.yvDesc = "", ""
	a.yvUploadID = strconv.FormatInt(id, 10)
	a.setStatus(fmt.Sprintf(i18n.T("版型 #%d 已创建，请选择图片上传"), id))
	a.Refresh()
}

func (a *App) deleteYuvision(id int64) {
	if err := a.svc.DeleteYuvision(context.Background(), id); err != nil {
		a.setStatus(err.Error())
		return
	}
	a.setStatus("")
	a.Refresh()
}

func (a *App) uploadYuvisionImage() {
	id, err := parseID(a.yvUploadID)
	if err != nil {
		a.setStatus(i18n.T("版型 ID 无效"))
		return
	}
	rel, err := a.svc.AddYuvisionImage(context.Background(), id, strings.TrimSpace(a.yvUploadSrc))
	if err != nil {
		a.setStatus(err.Error())
		return
	}
	a.setStatus(i18n.T("已上传图片：") + rel)
	a.Refresh()
}

func (a *App) createGallery() {
	yuvisionID, err := parseID(a.galYuvisionID)
	if err != nil {
		a.setStatus(i18n.T("版型 ID 无效"))
		return
	}
	id, err := a.svc.CreateGallery(context.Background(), yuvisionID, a.galTitle)
	if err != nil {
		a.setStatus(err.Error())
		return
	}
	a.galTitle = ""
	a.galUploadID = strconv.FormatInt(id, 10)
	a.setStatus(fmt.Sprintf(i18n.T("图集 #%d 已创建，请选择图片上传"), id))
	a.Refresh()
}

func (a *App) deleteGallery(id int64) {
	if err := a.svc.DeleteGallery(context.Background(), id); err != nil {
		a.setStatus(err.Error())
		return
	}
	a.setStatus("")
	a.Refresh()
}

func (a *App) uploadGalleryImage() {
	id, err := parseID(a.galUploadID)
	if err != nil {
		a.setStatus(i18n.T("图集 ID 无效"))
		return
	}
	rel, err := a.svc.AddGalleryImage(context.Background(), id, strings.TrimSpace(a.galUploadSrc))
	if err != nil {
		a.setStatus(err.Error())
		return
	}
	a.setStatus(i18n.T("已上传图片：") + rel)
	a.Refresh()
}

func (a *App) pickYuvisionImages(idText string) {
	paths, err := a.openImageDialog(i18n.T("选择版型图片"))
	if err != nil {
		a.uiStatus(err.Error())
		return
	}
	if paths == nil {
		return
	}
	id, err := parseID(idText)
	if err != nil {
		a.uiStatus(i18n.T("版型 ID 无效"))
		return
	}
	added := 0
	for _, path := range paths {
		if _, err := a.svc.AddYuvisionImage(context.Background(), id, path); err != nil {
			a.uiStatus(err.Error())
			return
		}
		added++
	}
	a.uiStatus(fmt.Sprintf(i18n.T("已上传 %d 张版型图片"), added))
}

func (a *App) pickGalleryImages(idText string) {
	paths, err := a.openImageDialog(i18n.T("选择图集图片"))
	if err != nil {
		a.uiStatus(err.Error())
		return
	}
	if paths == nil {
		return
	}
	id, err := parseID(idText)
	if err != nil {
		a.uiStatus(i18n.T("图集 ID 无效"))
		return
	}
	added := 0
	for _, path := range paths {
		if _, err := a.svc.AddGalleryImage(context.Background(), id, path); err != nil {
			a.uiStatus(err.Error())
			return
		}
		added++
	}
	a.uiStatus(fmt.Sprintf(i18n.T("已上传 %d 张图集图片"), added))
}

func (a *App) openImageDialog(title string) ([]string, error) {
	return mygo.Dialog.Open(mygo.OpenDialogOptions{
		Parent: a.win,
		Title:  title,
		Filters: []mygo.FileFilter{{
			Name:       i18n.T("图片"),
			Extensions: []string{"jpg", "jpeg", "png", "gif", "webp"},
		}},
		Multiple: true,
	})
}

func (a *App) renderCatalog(c *myui.Context) {
	myui.Text(c, i18n.T("目录数据")).FontSize(18).Bold()
	myui.Text(c, i18n.T("采集只新增；这里只编辑非身份字段，不提供新增、删除、隐藏。"))
	myui.Row(c).Gap(8).Children(func() {
		a.entityButton(c, i18n.T("品牌"), "brand")
		a.entityButton(c, i18n.T("厂商"), "firm")
		a.entityButton(c, i18n.T("车系"), "series")
		a.entityButton(c, i18n.T("年款"), "spec")
	})

	switch a.catalogEntity {
	case "firm":
		a.renderFirmList(c)
	case "series":
		a.renderSeriesList(c)
	case "spec":
		a.renderSpecList(c)
	default:
		a.renderBrandList(c)
	}
	a.renderCatalogEdit(c)
}

func (a *App) entityButton(c *myui.Context, label string, entity string) {
	if myui.Button(c, label).Clicked() {
		a.catalogEntity = entity
		a.catalogEditID = 0
	}
}

func (a *App) renderBrandList(c *myui.Context) {
	if len(a.brands) == 0 {
		myui.Text(c, i18n.T("暂无品牌数据。"))
		return
	}
	for _, item := range a.brands {
		current := item
		myui.Row(c.Key(fmt.Sprintf("brand-%d", current.ID))).Gap(8).Children(func() {
			myui.Text(c, fmt.Sprintf("#%d %s", current.ID, current.Name)).Grow(1)
			if myui.Button(c, i18n.T("编辑")).Clicked() {
				a.loadCatalogEdit("brand", current.ID)
			}
		})
	}
}

func (a *App) renderFirmList(c *myui.Context) {
	if len(a.firms) == 0 {
		myui.Text(c, i18n.T("暂无厂商数据。"))
		return
	}
	for _, item := range a.firms {
		current := item
		myui.Row(c.Key(fmt.Sprintf("firm-%d", current.ID))).Gap(8).Children(func() {
			myui.Text(c, fmt.Sprintf(i18n.T("#%d 品牌#%d %s"), current.ID, current.BrandID, current.Name)).Grow(1)
			if myui.Button(c, i18n.T("编辑")).Clicked() {
				a.loadCatalogEdit("firm", current.ID)
			}
		})
	}
}

func (a *App) renderSeriesList(c *myui.Context) {
	if len(a.series) == 0 {
		myui.Text(c, i18n.T("暂无车系数据。"))
		return
	}
	for _, item := range a.series {
		current := item
		myui.Row(c.Key(fmt.Sprintf("series-%d", current.ID))).Gap(8).Children(func() {
			myui.Text(c, fmt.Sprintf(i18n.T("#%d 厂商#%d %s"), current.ID, current.FirmID, current.Name)).Grow(1)
			if myui.Button(c, i18n.T("编辑")).Clicked() {
				a.loadCatalogEdit("series", current.ID)
			}
		})
	}
}

func (a *App) renderSpecList(c *myui.Context) {
	if len(a.specs) == 0 {
		myui.Text(c, i18n.T("暂无年款数据。"))
		return
	}
	for _, item := range a.specs {
		current := item
		myui.Row(c.Key(fmt.Sprintf("spec-%d", current.ID))).Gap(8).Children(func() {
			myui.Text(c, fmt.Sprintf(i18n.T("#%d 车系#%d %s %s"), current.ID, current.SeriesID, current.Name, current.ModelYear)).Grow(1)
			if myui.Button(c, i18n.T("编辑")).Clicked() {
				a.loadCatalogEdit("spec", current.ID)
			}
		})
	}
}

func (a *App) loadCatalogEdit(entity string, id int64) {
	a.catalogEntity = entity
	a.catalogEditID = id

	switch entity {
	case "brand":
		for _, item := range a.brands {
			if item.ID != id {
				continue
			}
			a.brandEdit.Name = item.Name
			a.brandEdit.FirstLetter = item.FirstLetter
			a.brandEdit.LogoURL = item.LogoURL
			a.brandEdit.LogoLocalPath = item.LogoLocalPath
			a.brandEdit.State = formatInt(item.State)
			a.brandEdit.HaveNewEnergy = formatInt(item.HaveNewEnergy)
			a.brandEdit.APISort = formatInt(item.APISort)
			return
		}
	case "firm":
		for _, item := range a.firms {
			if item.ID != id {
				continue
			}
			a.firmEdit.Name = item.Name
			a.firmEdit.PY = item.PY
			a.firmEdit.Link = item.Link
			a.firmEdit.LogoURL = item.LogoURL
			a.firmEdit.LogoLocalPath = item.LogoLocalPath
			return
		}
	case "series":
		for _, item := range a.series {
			if item.ID != id {
				continue
			}
			a.seriesEdit.Name = item.Name
			a.seriesEdit.GroupName = item.GroupName
			a.seriesEdit.State = formatInt(item.State)
			a.seriesEdit.PriceMin = formatFloat(item.PriceMin)
			a.seriesEdit.PriceMax = formatFloat(item.PriceMax)
			a.seriesEdit.Place = item.Place
			a.seriesEdit.PY = item.PY
			a.seriesEdit.LogoURL = item.LogoURL
			a.seriesEdit.LogoLocalPath = item.LogoLocalPath
			a.seriesEdit.LevelName = item.LevelName
			a.seriesEdit.NewEnergy = formatInt(item.NewEnergy)
			a.seriesEdit.IsForeignCar = formatInt(item.IsForeignCar)
			a.seriesEdit.IsTop = formatInt(item.IsTop)
			a.seriesEdit.FuelTypes = item.FuelTypes
			a.seriesEdit.CartypeName = item.CartypeName
			a.seriesEdit.Tag1 = item.Tag1
			a.seriesEdit.Rank = formatInt(item.Rank)
			a.seriesEdit.UVRank = formatInt(item.UVRank)
			a.seriesEdit.HotRank = formatInt(item.HotRank)
			return
		}
	case "spec":
		for _, item := range a.specs {
			if item.ID != id {
				continue
			}
			a.specEdit.GroupName = item.GroupName
			a.specEdit.Name = item.Name
			a.specEdit.SeriesName = item.SeriesName
			a.specEdit.ModelYear = item.ModelYear
			a.specEdit.BodyForm = item.BodyForm
			a.specEdit.PriceMin = formatFloat(item.PriceMin)
			a.specEdit.PriceMax = formatFloat(item.PriceMax)
			a.specEdit.NewsPrice = formatFloat(item.NewsPrice)
			a.specEdit.NewsID = formatInt(item.NewsID)
			a.specEdit.DealerID = formatInt(item.DealerID)
			a.specEdit.SaleType = item.SaleType
			a.specEdit.ImageURL = item.ImageURL
			a.specEdit.ImageLocalPath = item.ImageLocalPath
			a.specEdit.ImageSHA256 = item.ImageSHA256
			a.specEdit.WhiteImageURL = item.WhiteImageURL
			a.specEdit.WhiteImageLocalPath = item.WhiteImageLocalPath
			a.specEdit.WhiteImageSHA256 = item.WhiteImageSHA256
			return
		}
	}
}

func (a *App) renderCatalogEdit(c *myui.Context) {
	if a.catalogEditID == 0 {
		return
	}
	myui.Text(c, fmt.Sprintf(i18n.T("编辑 %s #%d"), a.catalogEntity, a.catalogEditID)).FontSize(16).Bold()

	switch a.catalogEntity {
	case "brand":
		myui.Row(c).Gap(8).Children(func() {
			myui.TextInput(c.Key("edit-brand-name"), &a.brandEdit.Name).Placeholder(i18n.T("名称"))
			myui.TextInput(c.Key("edit-brand-letter"), &a.brandEdit.FirstLetter).Placeholder(i18n.T("首字母"))
		})
		myui.Row(c).Gap(8).Children(func() {
			myui.TextInput(c.Key("edit-brand-logo-url"), &a.brandEdit.LogoURL).Placeholder("Logo URL")
			myui.TextInput(c.Key("edit-brand-logo-local"), &a.brandEdit.LogoLocalPath).Placeholder(i18n.T("Logo 本地路径"))
		})
		myui.Row(c).Gap(8).Children(func() {
			myui.TextInput(c.Key("edit-brand-state"), &a.brandEdit.State).Placeholder("state")
			myui.TextInput(c.Key("edit-brand-energy"), &a.brandEdit.HaveNewEnergy).Placeholder("havenewenergy")
			myui.TextInput(c.Key("edit-brand-sort"), &a.brandEdit.APISort).Placeholder("api_sort")
		})
	case "firm":
		myui.Row(c).Gap(8).Children(func() {
			myui.TextInput(c.Key("edit-firm-name"), &a.firmEdit.Name).Placeholder(i18n.T("名称"))
			myui.TextInput(c.Key("edit-firm-py"), &a.firmEdit.PY).Placeholder("py")
			myui.TextInput(c.Key("edit-firm-link"), &a.firmEdit.Link).Placeholder("link")
		})
		myui.Row(c).Gap(8).Children(func() {
			myui.TextInput(c.Key("edit-firm-logo-url"), &a.firmEdit.LogoURL).Placeholder("Logo URL")
			myui.TextInput(c.Key("edit-firm-logo-local"), &a.firmEdit.LogoLocalPath).Placeholder(i18n.T("Logo 本地路径"))
		})
	case "series":
		a.renderSeriesEdit(c)
	case "spec":
		a.renderSpecEdit(c)
	}

	if myui.Button(c, i18n.T("保存修改")).Clicked() {
		a.saveCatalogEdit()
	}
	if myui.Button(c, i18n.T("取消")).Clicked() {
		a.catalogEditID = 0
	}
}

func (a *App) renderSeriesEdit(c *myui.Context) {
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("se-name"), &a.seriesEdit.Name).Placeholder(i18n.T("名称"))
		myui.TextInput(c.Key("se-group"), &a.seriesEdit.GroupName).Placeholder("group_name")
		myui.TextInput(c.Key("se-place"), &a.seriesEdit.Place).Placeholder("place")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("se-price-min"), &a.seriesEdit.PriceMin).Placeholder("price_min")
		myui.TextInput(c.Key("se-price-max"), &a.seriesEdit.PriceMax).Placeholder("price_max")
		myui.TextInput(c.Key("se-state"), &a.seriesEdit.State).Placeholder("state")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("se-py"), &a.seriesEdit.PY).Placeholder("py")
		myui.TextInput(c.Key("se-level"), &a.seriesEdit.LevelName).Placeholder("level_name")
		myui.TextInput(c.Key("se-fuel"), &a.seriesEdit.FuelTypes).Placeholder("fuel_types")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("se-cartype"), &a.seriesEdit.CartypeName).Placeholder("cartypename")
		myui.TextInput(c.Key("se-tag"), &a.seriesEdit.Tag1).Placeholder("tag1")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("se-logo-url"), &a.seriesEdit.LogoURL).Placeholder("logo_url")
		myui.TextInput(c.Key("se-logo-local"), &a.seriesEdit.LogoLocalPath).Placeholder("logo_local_path")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("se-energy"), &a.seriesEdit.NewEnergy).Placeholder("newenergy")
		myui.TextInput(c.Key("se-foreign"), &a.seriesEdit.IsForeignCar).Placeholder("is_foreign_car")
		myui.TextInput(c.Key("se-top"), &a.seriesEdit.IsTop).Placeholder("is_top")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("se-rank"), &a.seriesEdit.Rank).Placeholder("rank")
		myui.TextInput(c.Key("se-uv"), &a.seriesEdit.UVRank).Placeholder("uv_rank")
		myui.TextInput(c.Key("se-hot"), &a.seriesEdit.HotRank).Placeholder("hot_rank")
	})
}

func (a *App) renderSpecEdit(c *myui.Context) {
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("sp-name"), &a.specEdit.Name).Placeholder(i18n.T("名称"))
		myui.TextInput(c.Key("sp-group"), &a.specEdit.GroupName).Placeholder("group_name")
		myui.TextInput(c.Key("sp-series-name"), &a.specEdit.SeriesName).Placeholder("series_name")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("sp-year"), &a.specEdit.ModelYear).Placeholder("model_year")
		myui.TextInput(c.Key("sp-body"), &a.specEdit.BodyForm).Placeholder("body_form")
		myui.TextInput(c.Key("sp-sale"), &a.specEdit.SaleType).Placeholder("sale_type")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("sp-price-min"), &a.specEdit.PriceMin).Placeholder("price_min")
		myui.TextInput(c.Key("sp-price-max"), &a.specEdit.PriceMax).Placeholder("price_max")
		myui.TextInput(c.Key("sp-news-price"), &a.specEdit.NewsPrice).Placeholder("news_price")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("sp-news-id"), &a.specEdit.NewsID).Placeholder("news_id")
		myui.TextInput(c.Key("sp-dealer-id"), &a.specEdit.DealerID).Placeholder("dealer_id")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("sp-image-url"), &a.specEdit.ImageURL).Placeholder("image_url")
		myui.TextInput(c.Key("sp-image-local"), &a.specEdit.ImageLocalPath).Placeholder("image_local_path")
		myui.TextInput(c.Key("sp-image-sha"), &a.specEdit.ImageSHA256).Placeholder("image_sha256")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("sp-white-url"), &a.specEdit.WhiteImageURL).Placeholder("white_image_url")
		myui.TextInput(c.Key("sp-white-local"), &a.specEdit.WhiteImageLocalPath).Placeholder("white_image_local_path")
		myui.TextInput(c.Key("sp-white-sha"), &a.specEdit.WhiteImageSHA256).Placeholder("white_image_sha256")
	})
}

func (a *App) saveCatalogEdit() {
	ctx := context.Background()
	id := a.catalogEditID
	var err error

	switch a.catalogEntity {
	case "brand":
		var state, energy, sort int64
		if state, err = parseInt64Value(a.brandEdit.State); err != nil {
			a.setStatus(i18n.T("state 无效"))
			return
		}
		if energy, err = parseInt64Value(a.brandEdit.HaveNewEnergy); err != nil {
			a.setStatus(i18n.T("havenewenergy 无效"))
			return
		}
		if sort, err = parseInt64Value(a.brandEdit.APISort); err != nil {
			a.setStatus(i18n.T("api_sort 无效"))
			return
		}
		err = a.svc.UpdateBrand(ctx, id, domain.BrandEdit{
			Name: a.brandEdit.Name, FirstLetter: a.brandEdit.FirstLetter,
			LogoURL: a.brandEdit.LogoURL, LogoLocalPath: a.brandEdit.LogoLocalPath,
			State: state, HaveNewEnergy: energy, APISort: sort,
		})
	case "firm":
		err = a.svc.UpdateFirm(ctx, id, domain.FirmEdit{
			Name: a.firmEdit.Name, PY: a.firmEdit.PY, Link: a.firmEdit.Link,
			LogoURL: a.firmEdit.LogoURL, LogoLocalPath: a.firmEdit.LogoLocalPath,
		})
	case "series":
		edit, ok := a.seriesEditInput()
		if !ok {
			return
		}
		err = a.svc.UpdateSeries(ctx, id, edit)
	case "spec":
		edit, ok := a.specEditInput()
		if !ok {
			return
		}
		err = a.svc.UpdateSpec(ctx, id, edit)
	}

	if err != nil {
		a.setStatus(err.Error())
		return
	}
	a.setStatus("")
	a.catalogEditID = 0
	a.Refresh()
}

func (a *App) seriesEditInput() (domain.SeriesEdit, bool) {
	var edit domain.SeriesEdit
	var err error
	edit.Name = a.seriesEdit.Name
	edit.GroupName = a.seriesEdit.GroupName
	edit.Place = a.seriesEdit.Place
	edit.PY = a.seriesEdit.PY
	edit.LogoURL = a.seriesEdit.LogoURL
	edit.LogoLocalPath = a.seriesEdit.LogoLocalPath
	edit.LevelName = a.seriesEdit.LevelName
	edit.FuelTypes = a.seriesEdit.FuelTypes
	edit.CartypeName = a.seriesEdit.CartypeName
	edit.Tag1 = a.seriesEdit.Tag1

	fields := []struct {
		name  string
		value string
		out   *int64
	}{
		{"state", a.seriesEdit.State, &edit.State},
		{"newenergy", a.seriesEdit.NewEnergy, &edit.NewEnergy},
		{"is_foreign_car", a.seriesEdit.IsForeignCar, &edit.IsForeignCar},
		{"is_top", a.seriesEdit.IsTop, &edit.IsTop},
		{"rank", a.seriesEdit.Rank, &edit.Rank},
		{"uv_rank", a.seriesEdit.UVRank, &edit.UVRank},
		{"hot_rank", a.seriesEdit.HotRank, &edit.HotRank},
	}
	for _, field := range fields {
		value, err := parseInt64Value(field.value)
		if err != nil {
			a.setStatus(field.name + i18n.T(" 无效"))
			return edit, false
		}
		*field.out = value
	}

	if edit.PriceMin, err = parseFloatValue(a.seriesEdit.PriceMin); err != nil {
		a.setStatus(i18n.T("price_min 无效"))
		return edit, false
	}
	if edit.PriceMax, err = parseFloatValue(a.seriesEdit.PriceMax); err != nil {
		a.setStatus(i18n.T("price_max 无效"))
		return edit, false
	}
	return edit, true
}

func (a *App) specEditInput() (domain.SpecEdit, bool) {
	var edit domain.SpecEdit
	var err error
	edit.GroupName = a.specEdit.GroupName
	edit.Name = a.specEdit.Name
	edit.SeriesName = a.specEdit.SeriesName
	edit.ModelYear = a.specEdit.ModelYear
	edit.BodyForm = a.specEdit.BodyForm
	edit.SaleType = a.specEdit.SaleType
	edit.ImageURL = a.specEdit.ImageURL
	edit.ImageLocalPath = a.specEdit.ImageLocalPath
	edit.ImageSHA256 = a.specEdit.ImageSHA256
	edit.WhiteImageURL = a.specEdit.WhiteImageURL
	edit.WhiteImageLocalPath = a.specEdit.WhiteImageLocalPath
	edit.WhiteImageSHA256 = a.specEdit.WhiteImageSHA256

	if edit.PriceMin, err = parseFloatValue(a.specEdit.PriceMin); err != nil {
		a.setStatus(i18n.T("price_min 无效"))
		return edit, false
	}
	if edit.PriceMax, err = parseFloatValue(a.specEdit.PriceMax); err != nil {
		a.setStatus(i18n.T("price_max 无效"))
		return edit, false
	}
	if edit.NewsPrice, err = parseFloatValue(a.specEdit.NewsPrice); err != nil {
		a.setStatus(i18n.T("news_price 无效"))
		return edit, false
	}
	if edit.NewsID, err = parseInt64Value(a.specEdit.NewsID); err != nil {
		a.setStatus(i18n.T("news_id 无效"))
		return edit, false
	}
	if edit.DealerID, err = parseInt64Value(a.specEdit.DealerID); err != nil {
		a.setStatus(i18n.T("dealer_id 无效"))
		return edit, false
	}
	return edit, true
}

func (a *App) renderSync(c *myui.Context) {
	state, err := autosync.LoadState(a.paths.StatePath)
	if err != nil {
		a.setStatus(err.Error())
	}
	stage := state.Stage
	if stage == "" {
		stage = "idle"
	}
	mode := state.Mode
	if mode == "" {
		mode = "full_scan"
	}

	myui.Row(c).Gap(8).Children(func() {
		if myui.Button(c, i18n.T("开始同步")).Disabled(a.syncRunning).Clicked() {
			go a.runSync()
		}
		if myui.Button(c, i18n.T("刷新状态")).Clicked() {
			a.Refresh()
		}
	})
	myui.Text(c, fmt.Sprintf(i18n.T("同步模式：%s，当前阶段：%s"), mode, stage))
	myui.Text(c, fmt.Sprintf(i18n.T("上次同步：%s"), state.LastSyncAt))
	if strings.TrimSpace(a.syncMessage) != "" {
		myui.Text(c, i18n.T("同步进度：")+a.syncMessage)
	}
	for _, entity := range []string{"brand", "firm", "series", "spec"} {
		stat := state.Stats[entity]
		myui.Text(c, fmt.Sprintf(i18n.T("%s：处理 %d，新增 %d，跳过 %d，失败 %d"),
			entity, stat.Processed, stat.Inserted, stat.Skipped, stat.Failed))
	}
	myui.Text(c, i18n.T("同步采用全量遍历、只新增、不更新、不删除。"))
	myui.Text(c, i18n.T("原始响应： ")+a.paths.RawDir)
	myui.Text(c, i18n.T("冲突日志： ")+a.paths.Conflicts)
}

func (a *App) runSync() {
	if a.syncRunning {
		return
	}
	a.updateSyncState(true, i18n.T("准备同步"))
	err := autosync.Run(context.Background(), a.svc, a.paths, func(progress autosync.Progress) {
		a.updateSyncState(true, progress.Message)
	})
	message := i18n.T("同步完成")
	if err != nil {
		message = i18n.T("同步失败：") + err.Error()
	}
	a.updateSyncState(false, message)
	a.uiStatus(message)
}

func (a *App) updateSyncState(running bool, message string) {
	if a.win != nil {
		a.win.Update(func() {
			a.syncRunning = running
			a.syncMessage = message
		})
		return
	}
	a.syncRunning = running
	a.syncMessage = message
}

func (a *App) renderSystemSettings(c *myui.Context) {
	myui.Text(c, i18n.T("设置")).FontSize(18).Bold()
	myui.Text(c, i18n.T("根目录： ")+a.paths.Root)
	myui.Text(c, i18n.T("数据目录： ")+a.paths.Data)
	myui.Text(c, i18n.T("媒体目录： ")+a.paths.Media)
	myui.Text(c, i18n.T("数据库： ")+a.paths.DBPath)

	myui.Text(c, i18n.T("数据库备份")).FontSize(16).Bold()
	if len(a.backups) == 0 {
		myui.Text(c, i18n.T("暂无迁移备份。"))
	} else {
		latest := a.backups[0]
		myui.Text(c, fmt.Sprintf(i18n.T("共 %d 个备份，最近一次：%s，大小 %d 字节"), len(a.backups), latest.Name, latest.Size))
	}

	myui.Text(c, i18n.T("孤儿图集目录")).FontSize(16).Bold()
	myui.Row(c).Gap(8).Children(func() {
		if myui.Button(c, i18n.T("扫描孤儿目录")).Clicked() {
			a.scanOrphanDirs()
		}
		if myui.Button(c, i18n.T("清理孤儿目录")).Clicked() {
			a.cleanupOrphanDirs()
		}
	})
	if len(a.orphanDirs) == 0 {
		myui.Text(c, i18n.T("未发现孤儿图集目录，或尚未扫描。"))
	} else {
		myui.Text(c, fmt.Sprintf(i18n.T("发现 %d 个孤儿目录："), len(a.orphanDirs)))
		for i, rel := range a.orphanDirs {
			if i >= 30 {
				break
			}
			myui.Text(c, "media/gallery/"+rel)
		}
		if len(a.orphanDirs) > 30 {
			myui.Text(c, fmt.Sprintf(i18n.T("仅显示前 30 个，共 %d 个。"), len(a.orphanDirs)))
		}
	}
}

func (a *App) renderLightbox(c *myui.Context) {
	myui.Column(c).Fill().Center().Padding(20).Gap(12).Children(func() {
		myui.Text(c, a.lightboxTitle).FontSize(18).Bold()
		if a.lightboxImage != nil {
			myui.Image(c, a.lightboxImage).Size(900, 620).Fit(myui.Contain).Radius(8)
		} else {
			myui.Text(c, i18n.T("图片无法预览"))
		}
		if myui.Button(c, i18n.T("关闭")).Clicked() {
			if a.lightboxWin != nil {
				a.lightboxWin.Hide()
			}
		}
	})
}

func (a *App) openLightbox(item domain.Gallery) {
	path := a.galleryOriginalPath(item)
	if path == "" {
		a.uiStatus(i18n.T("该图集暂无图片"))
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		a.uiStatus(err.Error())
		return
	}
	bitmap, err := myui.DecodeBitmap(data)
	if err != nil {
		a.uiStatus(err.Error())
		return
	}
	a.lightboxImage = bitmap
	a.lightboxTitle = fmt.Sprintf(i18n.T("图集 #%d %s"), item.ID, item.Title)

	if a.lightboxWin == nil {
		win := mygo.NewWindow(mygo.WindowOptions{
			Title:           i18n.T("图片预览"),
			Width:           960,
			Height:          720,
			AutoHideMenuBar: true,
			Content:         myui.View(a.renderLightbox),
		})
		win.OnClose(func(e *mygo.CloseEvent) {
			if !a.quitting {
				e.PreventDefault()
				win.Hide()
			}
		})
		a.lightboxWin = win
	}
	a.lightboxWin.Show()
	a.lightboxWin.Focus()
}

func (a *App) loadCoverBitmaps() {
	a.coverBitmaps = make(map[string]*myui.Bitmap)
	for _, item := range a.galleries {
		path := a.galleryCoverPath(item)
		if path == "" || a.coverBitmaps[path] != nil {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		bitmap, err := myui.DecodeBitmap(data)
		if err != nil {
			continue
		}
		a.coverBitmaps[path] = bitmap
	}
}

func (a *App) bitmapAt(path string) *myui.Bitmap {
	if path == "" {
		return nil
	}
	return a.coverBitmaps[path]
}

func (a *App) galleryCoverPath(item domain.Gallery) string {
	if item.CoverImage == "" {
		return ""
	}
	base := filepath.Join(a.paths.Gallery, fmt.Sprint(item.SeriesID), fmt.Sprint(item.ID))
	thumb := filepath.Join(base, filepath.FromSlash(media.ThumbnailRel(item.CoverImage)))
	if _, err := os.Stat(thumb); err == nil {
		return thumb
	}
	return filepath.Join(base, filepath.FromSlash(item.CoverImage))
}

func (a *App) galleryOriginalPath(item domain.Gallery) string {
	if item.CoverImage == "" {
		return ""
	}
	base := filepath.Join(a.paths.Gallery, fmt.Sprint(item.SeriesID), fmt.Sprint(item.ID))
	return filepath.Join(base, filepath.FromSlash(item.CoverImage))
}

func (a *App) countInt(name string) int64 {
	if a.counts == nil {
		return 0
	}
	return a.counts[name]
}

func (a *App) galleryFilter() domain.GalleryFilter {
	filter := domain.GalleryFilter{
		Keyword: a.gallerySearch,
		Sort:    a.gallerySort,
	}

	fields := []struct {
		name  string
		value string
		out   *int64
	}{
		{i18n.T("品牌 ID"), a.filterBrandID, &filter.BrandID},
		{i18n.T("厂商 ID"), a.filterFirmID, &filter.FirmID},
		{i18n.T("车系 ID"), a.filterSeriesID, &filter.SeriesID},
		{i18n.T("分类 ID"), a.filterCategory, &filter.CategoryID},
		{i18n.T("版型 ID"), a.filterYuvision, &filter.YouvisionID},
	}
	for _, field := range fields {
		value, err := parseOptionalID(field.value)
		if err != nil {
			a.setStatus(field.name + i18n.T(" 无效"))
			continue
		}
		*field.out = value
	}
	return filter
}

func (a *App) scanOrphanDirs() {
	orphans, err := a.svc.ScanOrphanGalleryDirs(context.Background())
	if err != nil {
		a.setStatus(err.Error())
		return
	}
	a.orphanDirs = orphans
	if len(orphans) == 0 {
		a.setStatus(i18n.T("未发现孤儿图集目录"))
		return
	}
	a.setStatus(fmt.Sprintf(i18n.T("发现 %d 个孤儿图集目录"), len(orphans)))
}

func (a *App) cleanupOrphanDirs() {
	removed, err := a.svc.CleanupOrphanGalleryDirs(context.Background())
	if err != nil {
		a.setStatus(err.Error())
	} else {
		a.setStatus(fmt.Sprintf(i18n.T("已清理 %d 个孤儿图集目录"), removed))
	}
	a.scanOrphanDirs()
}

func parseOptionalID(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id < 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}

func parseID(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty id")
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}

func parseInt64Value(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	return strconv.ParseInt(value, 10, 64)
}

func parseFloatValue(value string) (float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	return strconv.ParseFloat(value, 64)
}

func formatInt(value int64) string {
	return strconv.FormatInt(value, 10)
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
