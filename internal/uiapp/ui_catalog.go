package uiapp

import (
	"autocarsystem/internal/domain"
	"autocarsystem/internal/i18n"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	myui "github.com/egoist/mygo/ui"
)

func (a *App) renderCatalog(c *myui.Context) {
	switch a.catalogEntity {
	case "firm":
		a.renderFirmList(c)
	case "series":
		a.renderSeriesList(c)
	case "spec":
		a.renderSpecList(c)
	default:
		a.renderBrandCatalog(c)
	}
}

// entityButton 是车型数据实体的切换标签，选中项用主色高亮。
func (a *App) entityButton(c *myui.Context, label string, entity string) {
	p := paletteFor(c.Theme().Dark)
	selected := a.catalogEntity == entity
	bg := p.bg
	color := p.text
	if selected {
		bg = p.accent.Alpha(0.14)
		color = p.accent
	}
	b := myui.ButtonBase(c.Key("entity-"+entity)).
		Padding(8, 14).Radius(8).Background(bg).Label(label)
	if b.Hovered() && !selected {
		b.Background(p.hover)
	}
	b.Children(func() {
		myui.Text(c, label).FontSize(13).FontWeight(600).TextColor(color)
	})
	if b.Clicked() {
		a.catalogEntity = entity
	}
}

// catalogGridConfig 描述一个目录实体页共用的虚拟网格参数。
type catalogGridConfig struct {
	key               string
	list              *myui.ListState
	search            *string
	listKeyword       *string
	heading           string
	subtitle          string
	searchLabel       string
	searchHint        string
	searchPlaceholder string
	cardTitle         string
	groups            []catalogGroup
	emptyText         string
	noMatchText       string
	more              func(*myui.Context)
	// fixedCols > 0 时固定每行卡片数，否则按窗口宽度自适应。
	fixedCols int
	// renderCard 非 nil 时替换默认专辑式卡片渲染。
	renderCard func(*myui.Context, catalogGridCard, palette)
	// skipLocalFilter 为 true 时不做本地关键字过滤（服务端已过滤）。
	skipLocalFilter bool
}

type catalogGroup struct {
	label  string
	entity string
}

// catalogGridCard 是目录网格里的一张专辑式卡片数据。
type catalogGridCard struct {
	entity    string
	id        int64
	name      string
	meta      string
	search    string // 额外的搜索文本（如拼音），不展示在卡片上
	badges    []string
	logoPath  string
	fallbacks []string
}

// catalogGridRowKind 是虚拟列表的行类型。
type catalogGridRowKind int

const (
	catalogRowHeader catalogGridRowKind = iota
	catalogRowTabs
	catalogRowSearch
	catalogRowTitle
	catalogRowEmpty
	catalogRowCards
	catalogRowMore
)

type catalogGridRow struct {
	kind     catalogGridRowKind
	from, to int
}

const (
	catalogCardMinWidth = float32(160)
	catalogCardGap      = float32(20)
	catalogPagePad      = float32(22)
)

// catalogColumns 返回内容区能容纳的卡片列数，随窗口宽度自适应（下限 2）。
func (a *App) catalogColumns(c *myui.Context, min float32) int {
	w, _ := c.Size()
	inner := w - 2*catalogPagePad
	return max(2, int((inner+catalogCardGap)/(min+catalogCardGap)))
}

// renderCatalogGrid 仿 mygo-music 的 ui_browse.go：把目录页内容扁平成行，
// 用 myui.List 虚拟渲染，视图外的行不构建、图片也不加载。四个实体共用。
func (a *App) renderCatalogGrid(c *myui.Context, cfg catalogGridConfig, cards []catalogGridCard) {
	p := paletteFor(c.Theme().Dark)
	cols := a.catalogColumns(c, catalogCardMinWidth)
	if cfg.fixedCols > 0 {
		cols = cfg.fixedCols
	}
	renderCard := cfg.renderCard
	if renderCard == nil {
		renderCard = a.catalogAlbumCard
	}

	keyword := strings.ToLower(strings.TrimSpace(*cfg.search))
	if keyword != *cfg.listKeyword {
		*cfg.listKeyword = keyword
		cfg.list.ScrollTo(0, myui.Start)
	}
	items := make([]catalogGridCard, 0, len(cards))
	for _, item := range cards {
		if keyword != "" && !cfg.skipLocalFilter {
			haystack := strings.ToLower(item.name + " " + item.search + " " + item.meta + " " + strings.Join(item.badges, " ") + fmt.Sprintf("#%d", item.id))
			if !strings.Contains(haystack, keyword) {
				continue
			}
		}
		items = append(items, item)
	}

	rows := []catalogGridRow{
		{kind: catalogRowHeader},
		{kind: catalogRowTabs},
		{kind: catalogRowSearch},
		{kind: catalogRowTitle},
	}
	if len(items) == 0 {
		rows = append(rows, catalogGridRow{kind: catalogRowEmpty})
	} else {
		for from := 0; from < len(items); from += cols {
			rows = append(rows, catalogGridRow{kind: catalogRowCards, from: from, to: min(from+cols, len(items))})
		}
	}
	if cfg.more != nil {
		rows = append(rows, catalogGridRow{kind: catalogRowMore})
	}

	myui.List(c.Key(cfg.key), cfg.list, len(rows), func(i int) {
		row := rows[i]
		switch row.kind {
		case catalogRowHeader:
			myui.Column(c).Gap(6).Children(func() {
				myui.Text(c, cfg.heading).FontSize(30).FontWeight(700).TextColor(p.text)
				myui.Text(c, cfg.subtitle).FontSize(13).TextColor(p.muted)
			})
		case catalogRowTabs:
			myui.Row(c).Gap(8).Children(func() {
				for _, group := range cfg.groups {
					a.entityButton(c, group.label, group.entity)
				}
			})
		case catalogRowSearch:
			a.settingsGroup(c, func() {
				a.settingsRow(c, cfg.searchLabel, cfg.searchHint, func() {
					myui.TextInput(c.Key(cfg.key+"-search"), cfg.search).
						Placeholder(cfg.searchPlaceholder).Label(cfg.searchPlaceholder).Width(240)
				})
			})
		case catalogRowTitle:
			myui.Text(c, cfg.cardTitle).FontSize(19).FontWeight(700).TextColor(p.text)
		case catalogRowEmpty:
			empty := cfg.emptyText
			if keyword != "" {
				empty = cfg.noMatchText
			}
			myui.Row(c).Height(140).Center().Background(p.hover).Radius(12).Children(func() {
				myui.Text(c, empty).FontSize(13).TextColor(p.muted)
			})
		case catalogRowCards:
			myui.Grid(c).Columns(cols).GapX(catalogCardGap).GapY(catalogCardGap).Children(func() {
				for i := row.from; i < row.to; i++ {
					renderCard(c, items[i], p)
				}
			})
		case catalogRowMore:
			cfg.more(c)
		}
	}).ItemKey(func(i int) any { return i }).
		Gap(12).
		Grow(1).
		Margin(0, -22, 0, 0).Padding(0, 22, 0, 0).
		ScrollbarInsets(0)
}

// catalogAlbumCard 参考 mygo-music 的专辑卡：方形封面 + 标题 + 元信息，悬停轻微上浮。
func (a *App) catalogAlbumCard(c *myui.Context, item catalogGridCard, p palette) {
	card := myui.ButtonBase(c.Key(fmt.Sprintf("catalog-card-%s-%d", item.entity, item.id))).
		Column().Gap(10).AlignItems(myui.Stretch).FocusRing(false).
		Cursor(myui.CursorPointer).
		Label(fmt.Sprintf("%s #%d", item.name, item.id))

	card.Children(func() {
		art := myui.Box(c).AspectRatio(1).Radius(10)
		lift := myui.Pose{}
		if card.Hovered() {
			lift = myui.Pose{Y: -3}
		}
		art.AnimatePose(lift, myui.Snappy)
		art.Children(func() {
			a.catalogCover(c, item.logoPath, item.name, p, item.fallbacks...)
		})

		myui.Text(c, item.name).FontSize(13).FontWeight(500).MaxLines(2).LineHeight(1.35).TextColor(p.text)
		meta := fmt.Sprintf("#%d", item.id)
		if item.meta != "" {
			meta += "  ·  " + item.meta
		}
		if len(item.badges) > 0 {
			meta += "  ·  " + strings.Join(item.badges, "  ·  ")
		}
		myui.Text(c, meta).FontSize(11.5).TextColor(p.muted).SingleLine().Margin(-4, 0, 0, 0)
	})
	if card.Clicked() {
		switch item.entity {
		case "brand":
			a.open(navRoute{page: "brandDetail", brandID: item.id})
		case "firm":
			a.open(navRoute{page: "firmDetail", firmID: item.id})
		case "series":
			a.open(navRoute{page: "seriesDetail", seriesID: item.id})
		case "spec":
			a.open(navRoute{page: "specDetail", specID: item.id})
		}
	}
}

// catalogFirmCard 是厂商的横向列表卡片：左侧小 Logo + 右侧名称 / 品牌 / 拼音，
// 一行一个，区别于品牌的专辑式方形封面网格。点击进入厂商落地页。
func (a *App) catalogFirmCard(c *myui.Context, item catalogGridCard, p palette) {
	card := myui.ButtonBase(c.Key(fmt.Sprintf("catalog-card-%s-%d", item.entity, item.id))).
		Height(72).Padding(10, 12).Gap(14).AlignItems(myui.Center).
		Radius(12).Background(p.bg).Border(1, p.hairline).
		FocusRing(false).Cursor(myui.CursorPointer).
		Label(fmt.Sprintf("%s #%d", item.name, item.id))
	if card.Hovered() {
		card.Background(p.hover)
	}

	card.Children(func() {
		// 左侧小 Logo，四边贴合方形容器；无图时用首字占位。
		art := myui.Box(c).Size(48, 48).Radius(10)
		art.Children(func() {
			box := myui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Radius(10).Background(p.hover)
			box.Children(func() {
				bitmap := a.seriesLogo(item.logoPath)
				for _, fallback := range item.fallbacks {
					if bitmap != nil {
						break
					}
					bitmap = a.seriesLogo(fallback)
				}
				if bitmap != nil {
					myui.Image(c, bitmap).Absolute().Top(0).Left(0).Right(0).Bottom(0).Fit(myui.Cover).Radius(10)
					return
				}
				box.Center()
				letter := "?"
				for _, r := range item.name {
					letter = string(r)
					break
				}
				myui.Text(c, letter).FontSize(22).FontWeight(700).TextColor(p.accent.Alpha(0.7))
			})
		})
		// 右侧信息：名称 + 品牌 / 拼音 / ID 一行元信息。
		myui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(4).Children(func() {
			myui.Text(c, item.name).FontSize(13).FontWeight(600).SingleLine().TextColor(p.text)
			meta := fmt.Sprintf("#%d", item.id)
			if item.meta != "" {
				meta += "  ·  " + item.meta
			}
			if len(item.badges) > 0 {
				meta += "  ·  " + strings.Join(item.badges, "  ·  ")
			}
			myui.Text(c, meta).FontSize(11).TextColor(p.muted).SingleLine()
		})
	})

	if card.Clicked() {
		a.open(navRoute{page: "firmDetail", firmID: item.id})
	}
}

// catalogCover 显示目录图片（logoPath → 回退图），无图时用大号首字占位。
// 四边贴合父级方形容器 + Cover 裁切，图片始终居中铺满。
func (a *App) catalogCover(c *myui.Context, logoPath, name string, p palette, fallbacks ...string) {
	bitmap := a.seriesLogo(logoPath)
	for _, fallback := range fallbacks {
		if bitmap != nil {
			break
		}
		bitmap = a.seriesLogo(fallback)
	}
	box := myui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Radius(10)
	box.Background(p.hover)
	box.Children(func() {
		if bitmap != nil {
			myui.Image(c, bitmap).Absolute().Top(0).Left(0).Right(0).Bottom(0).Fit(myui.Cover).Radius(10)
			return
		}
		box.Center()
		letter := "?"
		for _, r := range name {
			letter = string(r)
			break
		}
		myui.Text(c, letter).FontSize(48).FontWeight(700).TextColor(p.accent.Alpha(0.7))
	})
}

func catalogGroups() []catalogGroup {
	return []catalogGroup{
		{label: i18n.T("品牌"), entity: "brand"},
		{label: i18n.T("厂商"), entity: "firm"},
		{label: i18n.T("车系"), entity: "series"},
		{label: i18n.T("年款"), entity: "spec"},
	}
}

func (a *App) renderBrandCatalog(c *myui.Context) {
	fallbacks := []string{filepath.Join(a.paths.Media, "autocar", "brand", "titlepic.png")}
	// go run 启动时程序目录是临时目录，再补一次工作目录下的媒体路径。
	if wd, err := os.Getwd(); err == nil {
		fallbacks = append(fallbacks, filepath.Join(wd, "media", "autocar", "brand", "titlepic.png"))
	}
	// 搜索词变化时重置搜索状态并发起服务端首页查询；为空则回到浏览列表。
	keyword := strings.TrimSpace(a.brandSearch)
	if keyword != a.brandSearchKeyword {
		a.brandSearchKeyword = keyword
		a.brandSearchSeq++
		a.brandSearchResults = nil
		a.brandSearchPaged = catalogPagedState{}
		if keyword != "" {
			a.loadBrandSearchPage()
		}
	}

	searching := keyword != ""
	source := a.brands
	paged := &a.brandPaged
	if searching {
		source = a.brandSearchResults
		paged = &a.brandSearchPaged
	}

	cards := make([]catalogGridCard, 0, len(source))
	for _, item := range source {
		meta := ""
		if item.FirstLetter != "" {
			meta = fmt.Sprintf(i18n.T("首字母 %s"), item.FirstLetter)
		}
		cards = append(cards, catalogGridCard{
			entity: "brand", id: item.ID, name: item.Name, meta: meta, search: item.FirstLetter,
			logoPath: item.LogoLocalPath, fallbacks: fallbacks,
		})
	}
	emptyText := a.catalogEmpty(i18n.T("暂无品牌数据。"), paged.loaded)
	if searching {
		emptyText = a.catalogEmpty(i18n.T("没有匹配的品牌。"), paged.loaded)
	}
	subtitle := fmt.Sprintf(i18n.T("车型数据 · 共 %d 个品牌"), a.countInt("brand"))
	if searching {
		subtitle = i18n.T("搜索结果")
	}
	cfg := catalogGridConfig{
		key: "catalog-brand-list", list: &a.brandList,
		search: &a.brandSearch, listKeyword: &a.brandListKeyword,
		heading: i18n.T("品牌"), subtitle: subtitle,
		searchLabel: i18n.T("品牌库"), searchHint: i18n.T("支持名称 / 首字母 / ID 搜索"),
		searchPlaceholder: i18n.T("搜索品牌"), cardTitle: i18n.T("品牌卡片"),
		groups: catalogGroups(), emptyText: emptyText, noMatchText: i18n.T("没有匹配的品牌。"),
		skipLocalFilter: searching,
	}
	if paged.more {
		if searching {
			cfg.more = a.brandSearchPageFoot
		} else {
			cfg.more = a.brandPageFoot
		}
	}
	a.renderCatalogGrid(c, cfg, cards)
}

// loadBrandSearchPage 加载下一页品牌搜索结果：offset 取已加载结果数，
// 关键字或刷新序号变化时丢弃过期结果；more 还需本次确实返回了数据。
func (a *App) loadBrandSearchPage() {
	pg := &a.brandSearchPaged
	if pg.loading || (pg.loaded && !pg.more) {
		return
	}
	keyword := a.brandSearchKeyword
	if keyword == "" {
		return
	}
	pg.loading = true
	pg.err = ""
	offset := len(a.brandSearchResults)
	seq := a.brandSearchSeq
	win := a.win

	go func() {
		items, more, err := a.svc.SearchBrandsPage(context.Background(), keyword, offset, brandPageSize)
		apply := func() {
			pg.loading = false
			if seq != a.brandSearchSeq || keyword != a.brandSearchKeyword {
				return
			}
			if err != nil {
				pg.err = err.Error()
				return
			}
			a.brandSearchResults = append(a.brandSearchResults, items...)
			pg.more = more && len(items) > 0
			pg.loaded = true
		}
		if win != nil {
			win.Update(apply)
		} else {
			a.postMain(apply)
		}
	}()
}

func (a *App) brandSearchPageFoot(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	pg := &a.brandSearchPaged
	foot := myui.Row(c.Key("brand-search-page-foot")).Height(56).Gap(12).Center()
	foot.Children(func() {
		switch {
		case pg.loading:
			myui.Text(c, i18n.T("加载中…")).FontSize(12.5).TextColor(p.muted)
		case pg.err != "":
			myui.Text(c, pg.err).FontSize(12.5).TextColor(p.muted)
			a.cardActionButton(c, "brand-search-page-retry", i18n.T("重试"), iconSync, false, p, a.loadBrandSearchPage)
		default:
			a.cardActionButton(c, "brand-search-page-more", i18n.T("加载更多"), iconArrowDown, true, p, a.loadBrandSearchPage)
		}
	})
}

// loadBrandPage 加载下一页品牌，参照 music 的 loadSavedAlbums：
// offset 取已加载条数，结果回投主线程后追加；刷新序号变化时丢弃过期结果。
func (a *App) loadBrandPage() {
	if a.brandPaged.loading || !a.brandPaged.more {
		return
	}
	a.brandPaged.loading = true
	a.brandPaged.err = ""
	offset := len(a.brands)
	seq := a.refreshSeq
	win := a.win

	go func() {
		items, more, err := a.svc.ListBrandsPage(context.Background(), offset, brandPageSize)
		apply := func() {
			a.brandPaged.loading = false
			if seq != a.refreshSeq {
				return
			}
			if err != nil {
				a.brandPaged.err = err.Error()
				return
			}
			a.brands = append(a.brands, items...)
			a.brandPaged.more = more
			a.brandPaged.loaded = true
		}
		if win != nil {
			win.Update(apply)
		} else {
			a.postMain(apply)
		}
	}()
}

// brandPageFoot 是品牌列表末尾的加载页脚：空闲时显示“加载更多”按钮，
// 加载中/失败分别显示状态与重试；只有点击按钮才请求下一页。
func (a *App) brandPageFoot(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	pg := &a.brandPaged
	foot := myui.Row(c.Key("brand-page-foot")).Height(56).Gap(12).Center()
	foot.Children(func() {
		switch {
		case pg.loading:
			myui.Text(c, i18n.T("加载中…")).FontSize(12.5).TextColor(p.muted)
		case pg.err != "":
			myui.Text(c, pg.err).FontSize(12.5).TextColor(p.muted)
			a.cardActionButton(c, "brand-page-retry", i18n.T("重试"), iconSync, false, p, a.loadBrandPage)
		default:
			a.cardActionButton(c, "brand-page-more", i18n.T("加载更多"), iconArrowDown, true, p, a.loadBrandPage)
		}
	})
}

func (a *App) renderFirmList(c *myui.Context) {
	keyword := strings.TrimSpace(a.firmSearch)
	if keyword != a.firmSearchKeyword {
		a.firmSearchKeyword = keyword
		a.firmSearchSeq++
		a.firmSearchResults = nil
		a.firmSearchPaged = catalogPagedState{}
		if keyword != "" {
			a.loadFirmSearchPage()
		}
	}
	searching := keyword != ""
	source := a.firms
	paged := &a.firmPaged
	if searching {
		source = a.firmSearchResults
		paged = &a.firmSearchPaged
	}

	brandNames := make(map[int64]string, len(a.brands))
	for _, b := range a.brands {
		brandNames[b.ID] = b.Name
	}
	cards := make([]catalogGridCard, 0, len(source))
	for _, item := range source {
		meta := fmt.Sprintf(i18n.T("品牌 %s"), nameOrID(brandNames, item.BrandID))
		var badges []string
		if item.PY != "" {
			badges = append(badges, item.PY)
		}
		cards = append(cards, catalogGridCard{
			entity: "firm", id: item.ID, name: item.Name, meta: meta, badges: badges, search: item.PY,
			logoPath: item.LogoLocalPath, fallbacks: a.catalogFallbacks("firm"),
		})
	}
	emptyText := a.catalogEmpty(i18n.T("暂无厂商数据。"), paged.loaded)
	if searching {
		emptyText = a.catalogEmpty(i18n.T("没有匹配的厂商。"), paged.loaded)
	}
	subtitle := fmt.Sprintf(i18n.T("共 %d 个厂商"), a.countInt("firm"))
	if searching {
		subtitle = i18n.T("搜索结果")
	}
	cfg := catalogGridConfig{
		key: "catalog-firm-list", list: &a.firmList,
		search: &a.firmSearch, listKeyword: &a.firmListKeyword,
		heading: i18n.T("厂商"), subtitle: subtitle,
		searchLabel: i18n.T("厂商库"), searchHint: i18n.T("支持名称 / 拼音 / ID 搜索"),
		searchPlaceholder: i18n.T("搜索厂商"), cardTitle: i18n.T("厂商卡片"),
		groups: catalogGroups(), emptyText: emptyText, noMatchText: i18n.T("没有匹配的厂商。"),
		fixedCols: 1, renderCard: a.catalogFirmCard, skipLocalFilter: searching,
	}
	if paged.more {
		if searching {
			cfg.more = func(c *myui.Context) {
				a.catalogSearchPageFoot(c, "firm-search", &a.firmSearchPaged, a.loadFirmSearchPage)
			}
		} else {
			cfg.more = a.firmPageFoot
		}
	}
	a.renderCatalogGrid(c, cfg, cards)
}

// loadFirmSearchPage 加载下一页厂商搜索结果；关键字或序号变化时丢弃过期结果。
func (a *App) loadFirmSearchPage() {
	pg := &a.firmSearchPaged
	if pg.loading || (pg.loaded && !pg.more) {
		return
	}
	keyword := a.firmSearchKeyword
	if keyword == "" {
		return
	}
	pg.loading = true
	pg.err = ""
	offset := len(a.firmSearchResults)
	seq := a.firmSearchSeq
	win := a.win

	go func() {
		items, more, err := a.svc.SearchFirmsPage(context.Background(), keyword, offset, firmPageSize)
		apply := func() {
			pg.loading = false
			if seq != a.firmSearchSeq || keyword != a.firmSearchKeyword {
				return
			}
			if err != nil {
				pg.err = err.Error()
				return
			}
			a.firmSearchResults = append(a.firmSearchResults, items...)
			pg.more = more && len(items) > 0
			pg.loaded = true
		}
		if win != nil {
			win.Update(apply)
		} else {
			a.postMain(apply)
		}
	}()
}

// loadFirmPage 加载下一页厂商，与 loadBrandPage 相同：offset 取已加载条数，
// 回投主线程后追加，刷新序号变化时丢弃过期结果。
func (a *App) loadFirmPage() {
	if a.firmPaged.loading || !a.firmPaged.more {
		return
	}
	a.firmPaged.loading = true
	a.firmPaged.err = ""
	offset := len(a.firms)
	seq := a.refreshSeq
	win := a.win

	go func() {
		items, more, err := a.svc.ListFirmsPage(context.Background(), offset, firmPageSize)
		apply := func() {
			a.firmPaged.loading = false
			if seq != a.refreshSeq {
				return
			}
			if err != nil {
				a.firmPaged.err = err.Error()
				return
			}
			a.firms = append(a.firms, items...)
			a.firmPaged.more = more
			a.firmPaged.loaded = true
		}
		if win != nil {
			win.Update(apply)
		} else {
			a.postMain(apply)
		}
	}()
}

func (a *App) firmPageFoot(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	pg := &a.firmPaged
	foot := myui.Row(c.Key("firm-page-foot")).Height(56).Gap(12).Center()
	foot.Children(func() {
		switch {
		case pg.loading:
			myui.Text(c, i18n.T("加载中…")).FontSize(12.5).TextColor(p.muted)
		case pg.err != "":
			myui.Text(c, pg.err).FontSize(12.5).TextColor(p.muted)
			a.cardActionButton(c, "firm-page-retry", i18n.T("重试"), iconSync, false, p, a.loadFirmPage)
		default:
			a.cardActionButton(c, "firm-page-more", i18n.T("加载更多"), iconArrowDown, true, p, a.loadFirmPage)
		}
	})
}

func (a *App) renderSeriesList(c *myui.Context) {
	keyword := strings.TrimSpace(a.seriesSearch)
	if keyword != a.seriesSearchKeyword {
		a.seriesSearchKeyword = keyword
		a.seriesSearchSeq++
		a.seriesSearchResults = nil
		a.seriesSearchPaged = catalogPagedState{}
		if keyword != "" {
			a.loadSeriesSearchPage()
		}
	}
	searching := keyword != ""
	source := a.series
	paged := &a.seriesPaged
	if searching {
		source = a.seriesSearchResults
		paged = &a.seriesSearchPaged
	}

	brandNames := make(map[int64]string, len(a.brands))
	for _, b := range a.brands {
		brandNames[b.ID] = b.Name
	}
	firmNames := make(map[int64]string, len(a.firms))
	for _, f := range a.firms {
		firmNames[f.ID] = f.Name
	}
	cards := make([]catalogGridCard, 0, len(source))
	for _, item := range source {
		meta := nameOrID(firmNames, item.FirmID)
		if item.LevelName != "" {
			meta += "  ·  " + item.LevelName
		}
		var badges []string
		if item.NewEnergy != 0 {
			badges = append(badges, i18n.T("新能源"))
		}
		if b := nameOrID(brandNames, item.BrandID); b != "" {
			badges = append(badges, b)
		}
		search := item.PY
		cards = append(cards, catalogGridCard{
			entity: "series", id: item.ID, name: item.Name, meta: meta, badges: badges, search: search,
			logoPath: item.LogoLocalPath, fallbacks: a.catalogFallbacks("series"),
		})
	}
	emptyText := a.catalogEmpty(i18n.T("暂无车系数据。"), paged.loaded)
	if searching {
		emptyText = a.catalogEmpty(i18n.T("没有匹配的车型。"), paged.loaded)
	}
	subtitle := fmt.Sprintf(i18n.T("共 %d 个车型"), a.countInt("series"))
	if searching {
		subtitle = i18n.T("搜索结果")
	}
	cfg := catalogGridConfig{
		key: "catalog-series-list", list: &a.seriesList,
		search: &a.seriesSearch, listKeyword: &a.seriesListKeyword,
		heading: i18n.T("车型"), subtitle: subtitle,
		searchLabel: i18n.T("车型库"), searchHint: i18n.T("支持名称 / 拼音 / ID 搜索"),
		searchPlaceholder: i18n.T("搜索车型"), cardTitle: i18n.T("车型卡片"),
		groups: catalogGroups(), emptyText: emptyText, noMatchText: i18n.T("没有匹配的车型。"),
		skipLocalFilter: searching,
	}
	if paged.more {
		if searching {
			cfg.more = func(c *myui.Context) {
				a.catalogSearchPageFoot(c, "series-search", &a.seriesSearchPaged, a.loadSeriesSearchPage)
			}
		} else {
			cfg.more = a.seriesPageFoot
		}
	}
	a.renderCatalogGrid(c, cfg, cards)
}

// loadSeriesSearchPage 加载下一页车系搜索结果；关键字或序号变化时丢弃过期结果。
func (a *App) loadSeriesSearchPage() {
	pg := &a.seriesSearchPaged
	if pg.loading || (pg.loaded && !pg.more) {
		return
	}
	keyword := a.seriesSearchKeyword
	if keyword == "" {
		return
	}
	pg.loading = true
	pg.err = ""
	offset := len(a.seriesSearchResults)
	seq := a.seriesSearchSeq
	win := a.win

	go func() {
		items, more, err := a.svc.SearchSeriesPage(context.Background(), keyword, offset, seriesPageSize)
		apply := func() {
			pg.loading = false
			if seq != a.seriesSearchSeq || keyword != a.seriesSearchKeyword {
				return
			}
			if err != nil {
				pg.err = err.Error()
				return
			}
			a.seriesSearchResults = append(a.seriesSearchResults, items...)
			pg.more = more && len(items) > 0
			pg.loaded = true
		}
		if win != nil {
			win.Update(apply)
		} else {
			a.postMain(apply)
		}
	}()
}

// loadSeriesPage 加载下一页车系，与 loadBrandPage 相同：offset 取已加载条数，
// 回投主线程后追加，刷新序号变化时丢弃过期结果。
func (a *App) loadSeriesPage() {
	if a.seriesPaged.loading || !a.seriesPaged.more {
		return
	}
	a.seriesPaged.loading = true
	a.seriesPaged.err = ""
	offset := len(a.series)
	seq := a.refreshSeq
	win := a.win

	go func() {
		items, more, err := a.svc.ListSeriesPage(context.Background(), offset, seriesPageSize)
		apply := func() {
			a.seriesPaged.loading = false
			if seq != a.refreshSeq {
				return
			}
			if err != nil {
				a.seriesPaged.err = err.Error()
				return
			}
			a.series = append(a.series, items...)
			a.seriesPaged.more = more
			a.seriesPaged.loaded = true
		}
		if win != nil {
			win.Update(apply)
		} else {
			a.postMain(apply)
		}
	}()
}

func (a *App) seriesPageFoot(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	pg := &a.seriesPaged
	foot := myui.Row(c.Key("series-page-foot")).Height(56).Gap(12).Center()
	foot.Children(func() {
		switch {
		case pg.loading:
			myui.Text(c, i18n.T("加载中…")).FontSize(12.5).TextColor(p.muted)
		case pg.err != "":
			myui.Text(c, pg.err).FontSize(12.5).TextColor(p.muted)
			a.cardActionButton(c, "series-page-retry", i18n.T("重试"), iconSync, false, p, a.loadSeriesPage)
		default:
			a.cardActionButton(c, "series-page-more", i18n.T("加载更多"), iconArrowDown, true, p, a.loadSeriesPage)
		}
	})
}

func (a *App) renderSpecList(c *myui.Context) {
	keyword := strings.TrimSpace(a.specSearch)
	if keyword != a.specSearchKeyword {
		a.specSearchKeyword = keyword
		a.specSearchSeq++
		a.specSearchResults = nil
		a.specSearchPaged = catalogPagedState{}
		if keyword != "" {
			a.loadSpecSearchPage()
		}
	}
	searching := keyword != ""
	source := a.specs
	paged := &a.specPaged
	if searching {
		source = a.specSearchResults
		paged = &a.specSearchPaged
	}

	seriesNames := make(map[int64]string, len(a.series))
	for _, s := range a.series {
		seriesNames[s.ID] = s.Name
	}
	cards := make([]catalogGridCard, 0, len(source))
	for _, item := range source {
		meta := nameWithFallback(item.SeriesName, seriesNames, item.SeriesID)
		var badges []string
		if item.ModelYear != "" {
			badges = append(badges, item.ModelYear)
		}
		if item.BodyForm != "" {
			badges = append(badges, item.BodyForm)
		}
		cards = append(cards, catalogGridCard{
			entity: "spec", id: item.ID, name: item.Name, meta: meta, badges: badges,
			logoPath: item.ImageLocalPath, fallbacks: a.catalogFallbacks("spec"),
		})
	}
	emptyText := a.catalogEmpty(i18n.T("暂无年款数据。"), paged.loaded)
	if searching {
		emptyText = a.catalogEmpty(i18n.T("没有匹配的年款。"), paged.loaded)
	}
	subtitle := fmt.Sprintf(i18n.T("共 %d 个年款"), a.countInt("spec"))
	if searching {
		subtitle = i18n.T("搜索结果")
	}
	cfg := catalogGridConfig{
		key: "catalog-spec-list", list: &a.specList,
		search: &a.specSearch, listKeyword: &a.specListKeyword,
		heading: i18n.T("年款"), subtitle: subtitle,
		searchLabel: i18n.T("年款库"), searchHint: i18n.T("支持名称 / 车系 / 年款 / ID 搜索"),
		searchPlaceholder: i18n.T("搜索年款"), cardTitle: i18n.T("年款卡片"),
		groups: catalogGroups(), emptyText: emptyText, noMatchText: i18n.T("没有匹配的年款。"),
		skipLocalFilter: searching,
	}
	if paged.more {
		if searching {
			cfg.more = func(c *myui.Context) {
				a.catalogSearchPageFoot(c, "spec-search", &a.specSearchPaged, a.loadSpecSearchPage)
			}
		} else {
			cfg.more = a.specPageFoot
		}
	}
	a.renderCatalogGrid(c, cfg, cards)
}

// loadSpecSearchPage 加载下一页年款搜索结果；关键字或序号变化时丢弃过期结果。
func (a *App) loadSpecSearchPage() {
	pg := &a.specSearchPaged
	if pg.loading || (pg.loaded && !pg.more) {
		return
	}
	keyword := a.specSearchKeyword
	if keyword == "" {
		return
	}
	pg.loading = true
	pg.err = ""
	offset := len(a.specSearchResults)
	seq := a.specSearchSeq
	win := a.win

	go func() {
		items, more, err := a.svc.SearchSpecsPage(context.Background(), keyword, offset, specPageSize)
		apply := func() {
			pg.loading = false
			if seq != a.specSearchSeq || keyword != a.specSearchKeyword {
				return
			}
			if err != nil {
				pg.err = err.Error()
				return
			}
			a.specSearchResults = append(a.specSearchResults, items...)
			pg.more = more && len(items) > 0
			pg.loaded = true
		}
		if win != nil {
			win.Update(apply)
		} else {
			a.postMain(apply)
		}
	}()
}

// loadSpecPage 加载下一页年款，与 loadBrandPage 相同：offset 取已加载条数，
// 回投主线程后追加，刷新序号变化时丢弃过期结果。
func (a *App) loadSpecPage() {
	if a.specPaged.loading || !a.specPaged.more {
		return
	}
	a.specPaged.loading = true
	a.specPaged.err = ""
	offset := len(a.specs)
	seq := a.refreshSeq
	win := a.win

	go func() {
		items, more, err := a.svc.ListSpecsPage(context.Background(), offset, specPageSize)
		apply := func() {
			a.specPaged.loading = false
			if seq != a.refreshSeq {
				return
			}
			if err != nil {
				a.specPaged.err = err.Error()
				return
			}
			a.specs = append(a.specs, items...)
			a.specPaged.more = more
			a.specPaged.loaded = true
		}
		if win != nil {
			win.Update(apply)
		} else {
			a.postMain(apply)
		}
	}()
}

func (a *App) specPageFoot(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	pg := &a.specPaged
	foot := myui.Row(c.Key("spec-page-foot")).Height(56).Gap(12).Center()
	foot.Children(func() {
		switch {
		case pg.loading:
			myui.Text(c, i18n.T("加载中…")).FontSize(12.5).TextColor(p.muted)
		case pg.err != "":
			myui.Text(c, pg.err).FontSize(12.5).TextColor(p.muted)
			a.cardActionButton(c, "spec-page-retry", i18n.T("重试"), iconSync, false, p, a.loadSpecPage)
		default:
			a.cardActionButton(c, "spec-page-more", i18n.T("加载更多"), iconArrowDown, true, p, a.loadSpecPage)
		}
	})
}

// catalogFallbacks 返回各实体缺失图片时的回退路径。
func (a *App) catalogFallbacks(entity string) []string {
	var paths []string
	switch entity {
	case "brand":
		paths = append(paths, filepath.Join(a.paths.Media, "autocar", "brand", "titlepic.png"))
	case "firm":
		paths = append(paths, filepath.Join(a.paths.Media, "autocar", "firm", "titlepic.png"))
	case "series":
		paths = append(paths, filepath.Join(a.paths.Media, "autocar", "series", "titlepic.png"))
	case "spec":
		paths = append(paths, filepath.Join(a.paths.Media, "autocar", "spec", "titlepic.png"))
	}
	// go run 启动时程序目录是临时目录，再补一次工作目录下的媒体路径。
	if wd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(wd, "media", "autocar", entity, "titlepic.png"))
	}
	return paths
}

// renderBrandDetail 是品牌落地页：展示品牌资料与该品牌下的厂商、车系。
func (a *App) renderBrandDetail(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	brand := a.brandDetailSelected()
	if brand == nil {
		myui.Row(c).Height(180).Center().Background(p.hover).Radius(12).Children(func() {
			myui.Text(c, i18n.T("品牌不存在")).FontSize(13).TextColor(p.muted)
		})
		return
	}

	myui.Column(c).Gap(22).Children(func() {
		a.brandDetailHero(c, *brand, p)
		a.brandDetailStats(c, brand.ID, p)
		a.brandDetailProfile(c, *brand, p)
		a.brandDetailRelatedFirms(c, brand.ID, p)
		a.brandDetailRelatedSeries(c, brand.ID, p)
	})
}

func (a *App) brandDetailSelected() *domain.Brand {
	if a.brandDetailID <= 0 {
		return nil
	}
	for i := range a.brands {
		if a.brands[i].ID == a.brandDetailID {
			return &a.brands[i]
		}
	}
	return nil
}

// brandDetailHero 是落地页头部：方形 Logo、品牌名与关键标识。
func (a *App) brandDetailHero(c *myui.Context, brand domain.Brand, p palette) {
	bitmap := a.seriesLogo(brand.LogoLocalPath)
	for _, fallback := range a.catalogFallbacks("brand") {
		if bitmap != nil {
			break
		}
		bitmap = a.seriesLogo(fallback)
	}
	myui.Row(c).Gap(24).AlignItems(myui.Center).Children(func() {
		art := myui.Box(c).Size(148, 148).Radius(16)
		art.Children(func() {
			bg := myui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Radius(16).Background(p.hover)
			bg.Children(func() {
				if bitmap != nil {
					myui.Image(c, bitmap).Absolute().Top(0).Left(0).Right(0).Bottom(0).Fit(myui.Cover).Radius(16)
					return
				}
				bg.Center()
				letter := "?"
				for _, r := range brand.Name {
					letter = string(r)
					break
				}
				myui.Text(c, letter).FontSize(64).FontWeight(700).TextColor(p.accent.Alpha(0.7))
			})
		})
		myui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(8).Children(func() {
			myui.Text(c, i18n.T("品牌")).FontSize(12).FontWeight(600).TextColor(p.accent)
			myui.Text(c, brand.Name).FontSize(30).FontWeight(700).MaxLines(2).LineHeight(1.2).TextColor(p.text)
			var meta []string
			meta = append(meta, fmt.Sprintf("#%d", brand.ID))
			if brand.FirstLetter != "" {
				meta = append(meta, fmt.Sprintf(i18n.T("首字母 %s"), brand.FirstLetter))
			}
			if brand.HaveNewEnergy != 0 {
				meta = append(meta, i18n.T("新能源"))
			}
			if brand.State != 0 {
				meta = append(meta, fmt.Sprintf(i18n.T("状态 %d"), brand.State))
			}
			myui.Text(c, strings.Join(meta, "  ·  ")).FontSize(13).TextColor(p.muted)
		})
	})
}

// brandDetailStats 展示该品牌关联数据的条目数，全部来自已加载的内存数据。
func (a *App) brandDetailStats(c *myui.Context, brandID int64, p palette) {
	firms, series, specs := 0, 0, 0
	for _, item := range a.firms {
		if item.BrandID == brandID {
			firms++
		}
	}
	for _, item := range a.series {
		if item.BrandID == brandID {
			series++
		}
	}
	for _, item := range a.specs {
		if item.BrandID == brandID {
			specs++
		}
	}
	myui.Grid(c).Columns(3).GapX(12).GapY(12).Children(func() {
		for _, item := range []struct {
			label string
			value int
		}{
			{i18n.T("厂商"), firms},
			{i18n.T("车系"), series},
			{i18n.T("年款"), specs},
		} {
			myui.Column(c).Gap(4).Padding(16).Radius(12).Background(p.hover).Children(func() {
				myui.Text(c, fmt.Sprintf("%d", item.value)).FontSize(24).FontWeight(700).TextColor(p.text)
				myui.Text(c, item.label).FontSize(12).TextColor(p.muted)
			})
		}
	})
}

// brandDetailProfile 展示品牌自身的全部字段，只读。
func (a *App) brandDetailProfile(c *myui.Context, brand domain.Brand, p palette) {
	myui.Text(c, i18n.T("品牌信息")).FontSize(19).FontWeight(700).TextColor(p.text)
	a.settingsGroup(c, func() {
		a.settingsRow(c, i18n.T("品牌 ID"), fmt.Sprintf("#%d", brand.ID), func() {})
		a.settingsRow(c, i18n.T("名称"), brand.Name, func() {})
		a.settingsRow(c, i18n.T("首字母"), brandDash(brand.FirstLetter), func() {})
		a.settingsRow(c, i18n.T("状态"), formatInt(brand.State), func() {})
		a.settingsRow(c, i18n.T("新能源"), formatInt(brand.HaveNewEnergy), func() {})
		a.settingsRow(c, i18n.T("接口排序"), formatInt(brand.APISort), func() {})
		a.settingsRow(c, i18n.T("采集时间"), brandDash(brand.CollectedAt), func() {})
		a.settingsRow(c, i18n.T("用户更新时间"), brandDash(brand.UserUpdatedAt), func() {})
		a.settingsRow(c, i18n.T("媒体更新时间"), brandDash(brand.MediaUpdatedAt), func() {})
	})
}

func (a *App) brandDetailRelatedFirms(c *myui.Context, brandID int64, p palette) {
	firms := make([]domain.Firm, 0)
	for _, item := range a.firms {
		if item.BrandID == brandID {
			firms = append(firms, item)
		}
	}
	myui.Text(c, i18n.T("该品牌的厂商")).FontSize(19).FontWeight(700).TextColor(p.text)
	if len(firms) == 0 {
		myui.Row(c).Height(80).Center().Background(p.hover).Radius(12).Children(func() {
			myui.Text(c, i18n.T("暂无关联厂商。")).FontSize(12).TextColor(p.muted)
		})
		return
	}
	myui.Grid(c).Columns(a.catalogColumns(c, 220)).GapX(12).GapY(12).Children(func() {
		for _, item := range firms {
			a.brandDetailTile(c, item.Name, item.PY, item.LogoLocalPath, a.catalogFallbacks("firm"), p)
		}
	})
}

func (a *App) brandDetailRelatedSeries(c *myui.Context, brandID int64, p palette) {
	series := make([]domain.Series, 0)
	for _, item := range a.series {
		if item.BrandID == brandID {
			series = append(series, item)
		}
	}
	myui.Text(c, i18n.T("该品牌的车系")).FontSize(19).FontWeight(700).TextColor(p.text)
	if len(series) == 0 {
		myui.Row(c).Height(80).Center().Background(p.hover).Radius(12).Children(func() {
			myui.Text(c, i18n.T("暂无关联车系。")).FontSize(12).TextColor(p.muted)
		})
		return
	}
	myui.Grid(c).Columns(a.catalogColumns(c, 220)).GapX(12).GapY(12).Children(func() {
		for _, item := range series {
			a.brandDetailTile(c, item.Name, item.LevelName, item.LogoLocalPath, a.catalogFallbacks("series"), p)
		}
	})
}

// brandDetailTile 是落地页里的只读关联条目：小图 + 名称 + 一行说明。
func (a *App) brandDetailTile(c *myui.Context, name, meta, logoPath string, fallbacks []string, p palette) {
	myui.Row(c).Gap(10).Padding(10).Radius(10).Background(p.hover).AlignItems(myui.Center).Children(func() {
		art := myui.Box(c).Size(40, 40).Radius(8)
		art.Children(func() {
			bitmap := a.seriesLogo(logoPath)
			for _, fallback := range fallbacks {
				if bitmap != nil {
					break
				}
				bitmap = a.seriesLogo(fallback)
			}
			bg := myui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Radius(8).Background(p.bg)
			bg.Children(func() {
				if bitmap != nil {
					myui.Image(c, bitmap).Absolute().Top(0).Left(0).Right(0).Bottom(0).Fit(myui.Cover).Radius(8)
					return
				}
				bg.Center()
				letter := "?"
				for _, r := range name {
					letter = string(r)
					break
				}
				myui.Text(c, letter).FontSize(18).FontWeight(700).TextColor(p.accent)
			})
		})
		myui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(2).Children(func() {
			myui.Text(c, name).FontSize(13).FontWeight(600).SingleLine().TextColor(p.text)
			if meta != "" {
				myui.Text(c, meta).FontSize(11).TextColor(p.muted).SingleLine()
			}
		})
	})
}

func brandDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}

// catalogEmpty 在首屏数据尚未加载完成时显示加载中，而不是空白态。
func (a *App) catalogEmpty(text string, loaded bool) string {
	if !loaded {
		return i18n.T("加载中…")
	}
	return text
}

// catalogSearchPageFoot 是各实体服务端搜索结果的通用页脚：加载中/失败重试/
// 点击加载更多。key 前缀区分不同实体，避免组件状态串用。
func (a *App) catalogSearchPageFoot(c *myui.Context, key string, pg *catalogPagedState, load func()) {
	p := paletteFor(c.Theme().Dark)
	foot := myui.Row(c.Key(key + "-foot")).Height(56).Gap(12).Center()
	foot.Children(func() {
		switch {
		case pg.loading:
			myui.Text(c, i18n.T("加载中…")).FontSize(12.5).TextColor(p.muted)
		case pg.err != "":
			myui.Text(c, pg.err).FontSize(12.5).TextColor(p.muted)
			a.cardActionButton(c, key+"-retry", i18n.T("重试"), iconSync, false, p, load)
		default:
			a.cardActionButton(c, key+"-more", i18n.T("加载更多"), iconArrowDown, true, p, load)
		}
	})
}

// renderFirmDetail 是厂商落地页：展示厂商资料与该厂商下的车系。
func (a *App) renderFirmDetail(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	firm := a.firmDetailSelected()
	if firm == nil {
		myui.Row(c).Height(180).Center().Background(p.hover).Radius(12).Children(func() {
			myui.Text(c, i18n.T("厂商不存在")).FontSize(13).TextColor(p.muted)
		})
		return
	}

	myui.Column(c).Gap(22).Children(func() {
		a.firmDetailHero(c, *firm, p)
		a.firmDetailStats(c, firm.ID, p)
		a.firmDetailProfile(c, *firm, p)
		a.firmDetailRelatedSeries(c, firm.ID, p)
	})
}

func (a *App) firmDetailSelected() *domain.Firm {
	if a.firmDetailID <= 0 {
		return nil
	}
	for i := range a.firms {
		if a.firms[i].ID == a.firmDetailID {
			return &a.firms[i]
		}
	}
	return nil
}

// firmDetailHero 是落地页头部：Logo、厂商名与关键标识。
func (a *App) firmDetailHero(c *myui.Context, firm domain.Firm, p palette) {
	bitmap := a.seriesLogo(firm.LogoLocalPath)
	for _, fallback := range a.catalogFallbacks("firm") {
		if bitmap != nil {
			break
		}
		bitmap = a.seriesLogo(fallback)
	}
	brandNames := make(map[int64]string, len(a.brands))
	for _, b := range a.brands {
		brandNames[b.ID] = b.Name
	}
	myui.Row(c).Gap(24).AlignItems(myui.Center).Children(func() {
		art := myui.Box(c).Size(148, 148).Radius(16)
		art.Children(func() {
			bg := myui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Radius(16).Background(p.hover)
			bg.Children(func() {
				if bitmap != nil {
					myui.Image(c, bitmap).Absolute().Top(0).Left(0).Right(0).Bottom(0).Fit(myui.Cover).Radius(16)
					return
				}
				bg.Center()
				letter := "?"
				for _, r := range firm.Name {
					letter = string(r)
					break
				}
				myui.Text(c, letter).FontSize(64).FontWeight(700).TextColor(p.accent.Alpha(0.7))
			})
		})
		myui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(8).Children(func() {
			myui.Text(c, i18n.T("厂商")).FontSize(12).FontWeight(600).TextColor(p.accent)
			myui.Text(c, firm.Name).FontSize(30).FontWeight(700).MaxLines(2).LineHeight(1.2).TextColor(p.text)
			var meta []string
			meta = append(meta, fmt.Sprintf("#%d", firm.ID))
			if b := nameOrID(brandNames, firm.BrandID); b != "" {
				meta = append(meta, fmt.Sprintf(i18n.T("所属品牌 %s"), b))
			}
			if firm.PY != "" {
				meta = append(meta, firm.PY)
			}
			myui.Text(c, strings.Join(meta, "  ·  ")).FontSize(13).TextColor(p.muted)
		})
	})
}

// firmDetailStats 展示该厂商关联数据的条目数，全部来自已加载的内存数据。
func (a *App) firmDetailStats(c *myui.Context, firmID int64, p palette) {
	series, specs := 0, 0
	for _, item := range a.series {
		if item.FirmID == firmID {
			series++
		}
	}
	for _, item := range a.specs {
		if item.FirmID == firmID {
			specs++
		}
	}
	myui.Grid(c).Columns(2).GapX(12).GapY(12).Children(func() {
		for _, item := range []struct {
			label string
			value int
		}{
			{i18n.T("车系"), series},
			{i18n.T("年款"), specs},
		} {
			myui.Column(c).Gap(4).Padding(16).Radius(12).Background(p.hover).Children(func() {
				myui.Text(c, fmt.Sprintf("%d", item.value)).FontSize(24).FontWeight(700).TextColor(p.text)
				myui.Text(c, item.label).FontSize(12).TextColor(p.muted)
			})
		}
	})
}

// firmDetailProfile 展示厂商自身的全部字段，只读。
func (a *App) firmDetailProfile(c *myui.Context, firm domain.Firm, p palette) {
	brandNames := make(map[int64]string, len(a.brands))
	for _, b := range a.brands {
		brandNames[b.ID] = b.Name
	}
	myui.Text(c, i18n.T("厂商信息")).FontSize(19).FontWeight(700).TextColor(p.text)
	a.settingsGroup(c, func() {
		a.settingsRow(c, i18n.T("厂商 ID"), fmt.Sprintf("#%d", firm.ID), func() {})
		a.settingsRow(c, i18n.T("名称"), firm.Name, func() {})
		a.settingsRow(c, i18n.T("所属品牌"), nameOrID(brandNames, firm.BrandID), func() {})
		a.settingsRow(c, i18n.T("拼音"), brandDash(firm.PY), func() {})
		a.settingsRow(c, i18n.T("链接"), brandDash(firm.Link), func() {})
		a.settingsRow(c, "Logo URL", brandDash(firm.LogoURL), func() {})
		a.settingsRow(c, i18n.T("Logo 本地路径"), brandDash(firm.LogoLocalPath), func() {})
		a.settingsRow(c, i18n.T("采集时间"), brandDash(firm.CollectedAt), func() {})
		a.settingsRow(c, i18n.T("用户更新时间"), brandDash(firm.UserUpdatedAt), func() {})
		a.settingsRow(c, i18n.T("媒体更新时间"), brandDash(firm.MediaUpdatedAt), func() {})
	})
}

func (a *App) firmDetailRelatedSeries(c *myui.Context, firmID int64, p palette) {
	series := make([]domain.Series, 0)
	for _, item := range a.series {
		if item.FirmID == firmID {
			series = append(series, item)
		}
	}
	myui.Text(c, i18n.T("该厂商的车系")).FontSize(19).FontWeight(700).TextColor(p.text)
	if len(series) == 0 {
		myui.Row(c).Height(80).Center().Background(p.hover).Radius(12).Children(func() {
			myui.Text(c, i18n.T("暂无关联车系。")).FontSize(12).TextColor(p.muted)
		})
		return
	}
	myui.Grid(c).Columns(a.catalogColumns(c, 220)).GapX(12).GapY(12).Children(func() {
		for _, item := range series {
			a.brandDetailTile(c, item.Name, item.LevelName, item.LogoLocalPath, a.catalogFallbacks("series"), p)
		}
	})
}

// nameWithFallback 优先使用记录自带名称（Spec.SeriesName），缺失时再按 ID 在内存映射里查找。
func nameWithFallback(own string, names map[int64]string, id int64) string {
	if strings.TrimSpace(own) != "" {
		return own
	}
	return nameOrID(names, id)
}

// catalogNames 返回指定实体的 ID→名称映射，供落地页展示关联名称。
func (a *App) catalogNames(entity string) map[int64]string {
	names := make(map[int64]string)
	switch entity {
	case "brand":
		for _, item := range a.brands {
			names[item.ID] = item.Name
		}
	case "firm":
		for _, item := range a.firms {
			names[item.ID] = item.Name
		}
	case "series":
		for _, item := range a.series {
			names[item.ID] = item.Name
		}
	}
	return names
}

// renderSeriesDetail 是车型落地页：展示车型资料与该车型下的年款。
func (a *App) renderSeriesDetail(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	item := a.seriesDetailSelected()
	if item == nil {
		myui.Row(c).Height(180).Center().Background(p.hover).Radius(12).Children(func() {
			myui.Text(c, i18n.T("车型不存在")).FontSize(13).TextColor(p.muted)
		})
		return
	}
	myui.Column(c).Gap(22).Children(func() {
		a.seriesDetailHero(c, *item, p)
		a.seriesDetailStats(c, item.ID, p)
		a.seriesDetailProfile(c, *item, p)
		a.seriesDetailRelatedSpecs(c, item.ID, p)
	})
}

func (a *App) seriesDetailSelected() *domain.Series {
	if a.seriesDetailID <= 0 {
		return nil
	}
	for i := range a.series {
		if a.series[i].ID == a.seriesDetailID {
			return &a.series[i]
		}
	}
	return nil
}

// seriesDetailHero 是落地页头部：Logo、车型名与关键标识。
func (a *App) seriesDetailHero(c *myui.Context, item domain.Series, p palette) {
	bitmap := a.seriesLogo(item.LogoLocalPath)
	for _, fallback := range a.catalogFallbacks("series") {
		if bitmap != nil {
			break
		}
		bitmap = a.seriesLogo(fallback)
	}
	brandNames := a.catalogNames("brand")
	firmNames := a.catalogNames("firm")
	myui.Row(c).Gap(24).AlignItems(myui.Center).Children(func() {
		art := myui.Box(c).Size(148, 148).Radius(16)
		art.Children(func() {
			bg := myui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Radius(16).Background(p.hover)
			bg.Children(func() {
				if bitmap != nil {
					myui.Image(c, bitmap).Absolute().Top(0).Left(0).Right(0).Bottom(0).Fit(myui.Cover).Radius(16)
					return
				}
				bg.Center()
				letter := "?"
				for _, r := range item.Name {
					letter = string(r)
					break
				}
				myui.Text(c, letter).FontSize(64).FontWeight(700).TextColor(p.accent.Alpha(0.7))
			})
		})
		myui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(8).Children(func() {
			myui.Text(c, i18n.T("车型")).FontSize(12).FontWeight(600).TextColor(p.accent)
			myui.Text(c, item.Name).FontSize(30).FontWeight(700).MaxLines(2).LineHeight(1.2).TextColor(p.text)
			var meta []string
			meta = append(meta, fmt.Sprintf("#%d", item.ID))
			if b := nameOrID(brandNames, item.BrandID); b != "" {
				meta = append(meta, fmt.Sprintf(i18n.T("所属品牌 %s"), b))
			}
			if f := nameOrID(firmNames, item.FirmID); f != "" {
				meta = append(meta, fmt.Sprintf(i18n.T("所属厂商 %s"), f))
			}
			if item.LevelName != "" {
				meta = append(meta, item.LevelName)
			}
			myui.Text(c, strings.Join(meta, "  ·  ")).FontSize(13).TextColor(p.muted)
		})
	})
}

// seriesDetailStats 展示车型关联数据条目数，全部来自已加载的内存数据。
func (a *App) seriesDetailStats(c *myui.Context, seriesID int64, p palette) {
	specs := 0
	for _, item := range a.specs {
		if item.SeriesID == seriesID {
			specs++
		}
	}
	myui.Grid(c).Columns(1).GapX(12).GapY(12).Children(func() {
		myui.Column(c).Gap(4).Padding(16).Radius(12).Background(p.hover).Children(func() {
			myui.Text(c, fmt.Sprintf("%d", specs)).FontSize(24).FontWeight(700).TextColor(p.text)
			myui.Text(c, i18n.T("年款")).FontSize(12).TextColor(p.muted)
		})
	})
}

// seriesDetailProfile 展示车型自身全部字段，只读。
func (a *App) seriesDetailProfile(c *myui.Context, item domain.Series, p palette) {
	brandNames := a.catalogNames("brand")
	firmNames := a.catalogNames("firm")
	myui.Text(c, i18n.T("车型信息")).FontSize(19).FontWeight(700).TextColor(p.text)
	a.settingsGroup(c, func() {
		a.settingsRow(c, i18n.T("车型 ID"), fmt.Sprintf("#%d", item.ID), func() {})
		a.settingsRow(c, i18n.T("名称"), item.Name, func() {})
		a.settingsRow(c, i18n.T("所属品牌"), nameOrID(brandNames, item.BrandID), func() {})
		a.settingsRow(c, i18n.T("所属厂商"), nameOrID(firmNames, item.FirmID), func() {})
		a.settingsRow(c, i18n.T("分组"), brandDash(item.GroupName), func() {})
		a.settingsRow(c, i18n.T("级别"), brandDash(item.LevelName), func() {})
		a.settingsRow(c, i18n.T("价格区间"), fmt.Sprintf("%.2f-%.2f", item.PriceMin, item.PriceMax), func() {})
		a.settingsRow(c, i18n.T("产地"), brandDash(item.Place), func() {})
		a.settingsRow(c, i18n.T("拼音"), brandDash(item.PY), func() {})
		a.settingsRow(c, i18n.T("新能源"), formatInt(item.NewEnergy), func() {})
		a.settingsRow(c, i18n.T("进口车"), formatInt(item.IsForeignCar), func() {})
		a.settingsRow(c, i18n.T("置顶"), formatInt(item.IsTop), func() {})
		a.settingsRow(c, i18n.T("燃料类型"), brandDash(item.FuelTypes), func() {})
		a.settingsRow(c, i18n.T("车型类别"), brandDash(item.CartypeName), func() {})
		a.settingsRow(c, i18n.T("标签"), brandDash(item.Tag1), func() {})
		a.settingsRow(c, i18n.T("排名"), formatInt(item.Rank), func() {})
		a.settingsRow(c, "UV rank", formatInt(item.UVRank), func() {})
		a.settingsRow(c, i18n.T("热度排名"), formatInt(item.HotRank), func() {})
		a.settingsRow(c, "Logo URL", brandDash(item.LogoURL), func() {})
		a.settingsRow(c, i18n.T("Logo 本地路径"), brandDash(item.LogoLocalPath), func() {})
		a.settingsRow(c, i18n.T("采集时间"), brandDash(item.CollectedAt), func() {})
		a.settingsRow(c, i18n.T("用户更新时间"), brandDash(item.UserUpdatedAt), func() {})
		a.settingsRow(c, i18n.T("媒体更新时间"), brandDash(item.MediaUpdatedAt), func() {})
	})
}

func (a *App) seriesDetailRelatedSpecs(c *myui.Context, seriesID int64, p palette) {
	specs := make([]domain.Spec, 0)
	for _, item := range a.specs {
		if item.SeriesID == seriesID {
			specs = append(specs, item)
		}
	}
	myui.Text(c, i18n.T("该车型的年款")).FontSize(19).FontWeight(700).TextColor(p.text)
	if len(specs) == 0 {
		myui.Row(c).Height(80).Center().Background(p.hover).Radius(12).Children(func() {
			myui.Text(c, i18n.T("暂无关联年款。")).FontSize(12).TextColor(p.muted)
		})
		return
	}
	myui.Grid(c).Columns(a.catalogColumns(c, 220)).GapX(12).GapY(12).Children(func() {
		for _, item := range specs {
			meta := item.ModelYear
			if item.BodyForm != "" {
				meta += "  ·  " + item.BodyForm
			}
			a.brandDetailTile(c, item.Name, meta, item.ImageLocalPath, a.catalogFallbacks("spec"), p)
		}
	})
}

// renderSpecDetail 是年款落地页：展示年款的完整资料。
func (a *App) renderSpecDetail(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	item := a.specDetailSelected()
	if item == nil {
		myui.Row(c).Height(180).Center().Background(p.hover).Radius(12).Children(func() {
			myui.Text(c, i18n.T("年款不存在")).FontSize(13).TextColor(p.muted)
		})
		return
	}
	myui.Column(c).Gap(22).Children(func() {
		a.specDetailHero(c, *item, p)
		a.specDetailProfile(c, *item, p)
	})
}

func (a *App) specDetailSelected() *domain.Spec {
	if a.specDetailID <= 0 {
		return nil
	}
	for i := range a.specs {
		if a.specs[i].ID == a.specDetailID {
			return &a.specs[i]
		}
	}
	return nil
}

// specDetailHero 是落地页头部：图片、年款名与关键标识。
func (a *App) specDetailHero(c *myui.Context, item domain.Spec, p palette) {
	bitmap := a.seriesLogo(item.ImageLocalPath)
	for _, fallback := range a.catalogFallbacks("spec") {
		if bitmap != nil {
			break
		}
		bitmap = a.seriesLogo(fallback)
	}
	seriesNames := a.catalogNames("series")
	myui.Row(c).Gap(24).AlignItems(myui.Center).Children(func() {
		art := myui.Box(c).Size(148, 148).Radius(16)
		art.Children(func() {
			bg := myui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Radius(16).Background(p.hover)
			bg.Children(func() {
				if bitmap != nil {
					myui.Image(c, bitmap).Absolute().Top(0).Left(0).Right(0).Bottom(0).Fit(myui.Cover).Radius(16)
					return
				}
				bg.Center()
				letter := "?"
				for _, r := range item.Name {
					letter = string(r)
					break
				}
				myui.Text(c, letter).FontSize(64).FontWeight(700).TextColor(p.accent.Alpha(0.7))
			})
		})
		myui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(8).Children(func() {
			myui.Text(c, i18n.T("年款")).FontSize(12).FontWeight(600).TextColor(p.accent)
			myui.Text(c, item.Name).FontSize(30).FontWeight(700).MaxLines(2).LineHeight(1.2).TextColor(p.text)
			var meta []string
			meta = append(meta, fmt.Sprintf("#%d", item.ID))
			if s := nameWithFallback(item.SeriesName, seriesNames, item.SeriesID); s != "" {
				meta = append(meta, fmt.Sprintf(i18n.T("所属车系 %s"), s))
			}
			if item.ModelYear != "" {
				meta = append(meta, item.ModelYear)
			}
			if item.BodyForm != "" {
				meta = append(meta, item.BodyForm)
			}
			myui.Text(c, strings.Join(meta, "  ·  ")).FontSize(13).TextColor(p.muted)
		})
	})
}

// specDetailProfile 展示年款自身全部字段，只读。
func (a *App) specDetailProfile(c *myui.Context, item domain.Spec, p palette) {
	seriesNames := a.catalogNames("series")
	myui.Text(c, i18n.T("年款信息")).FontSize(19).FontWeight(700).TextColor(p.text)
	a.settingsGroup(c, func() {
		a.settingsRow(c, i18n.T("年款 ID"), fmt.Sprintf("#%d", item.ID), func() {})
		a.settingsRow(c, i18n.T("名称"), item.Name, func() {})
		a.settingsRow(c, i18n.T("所属车系"), nameWithFallback(item.SeriesName, seriesNames, item.SeriesID), func() {})
		a.settingsRow(c, i18n.T("分组"), brandDash(item.GroupName), func() {})
		a.settingsRow(c, i18n.T("年款"), brandDash(item.ModelYear), func() {})
		a.settingsRow(c, i18n.T("车身形式"), brandDash(item.BodyForm), func() {})
		a.settingsRow(c, i18n.T("指导价"), fmt.Sprintf("%.2f-%.2f", item.PriceMin, item.PriceMax), func() {})
		a.settingsRow(c, i18n.T("新闻价"), formatFloat(item.NewsPrice), func() {})
		a.settingsRow(c, "News ID", formatInt(item.NewsID), func() {})
		a.settingsRow(c, "Dealer ID", formatInt(item.DealerID), func() {})
		a.settingsRow(c, i18n.T("销售类型"), brandDash(item.SaleType), func() {})
		a.settingsRow(c, "Image URL", brandDash(item.ImageURL), func() {})
		a.settingsRow(c, i18n.T("图片本地路径"), brandDash(item.ImageLocalPath), func() {})
		a.settingsRow(c, "Image SHA256", brandDash(item.ImageSHA256), func() {})
		a.settingsRow(c, "White Image URL", brandDash(item.WhiteImageURL), func() {})
		a.settingsRow(c, i18n.T("白色图片本地路径"), brandDash(item.WhiteImageLocalPath), func() {})
		a.settingsRow(c, "White Image SHA256", brandDash(item.WhiteImageSHA256), func() {})
		a.settingsRow(c, i18n.T("采集时间"), brandDash(item.CollectedAt), func() {})
		a.settingsRow(c, i18n.T("用户更新时间"), brandDash(item.UserUpdatedAt), func() {})
		a.settingsRow(c, i18n.T("媒体更新时间"), brandDash(item.MediaUpdatedAt), func() {})
	})
}
