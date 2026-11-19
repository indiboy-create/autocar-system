package uiapp

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"autocarsystem/internal/domain"
	"autocarsystem/internal/i18n"
	"autocarsystem/internal/service"

	myui "github.com/egoist/mygo/ui"
)

func (a *App) renderPage(c *myui.Context) {
	switch a.page {
	case "home":
		a.renderHome(c)
	case "userCenter":
		a.renderUserCenter(c)
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

// renderHome 是首页：欢迎信息与应用概况卡片。
func (a *App) renderHome(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	myui.Column(c).Gap(12).Children(func() {
		myui.Text(c, i18n.T("首页")).FontSize(22).FontWeight(700).TextColor(p.text)
		myui.Text(c, i18n.T("欢迎使用版型图集管理")).FontSize(13).TextColor(p.muted)
		a.settingsGroup(c, func() {
			a.settingsRow(c, i18n.T("应用名称"), "autoCarSystem", func() {})
			a.settingsRow(c, i18n.T("描述"), i18n.T("本地图集与版型数据管理工具"), func() {})
		})
	})
}

// renderUserCenter 是用户中心：当前环境与数据目录信息。
func (a *App) renderUserCenter(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	myui.Column(c).Gap(12).Children(func() {
		myui.Text(c, i18n.T("用户中心")).FontSize(22).FontWeight(700).TextColor(p.text)
		a.settingsGroup(c, func() {
			a.settingsRow(c, i18n.T("数据目录"), a.paths.Root, func() {})
			a.settingsRow(c, i18n.T("数据库"), a.paths.DBPath, func() {})
		})
	})
}

func (a *App) renderSeriesQuery(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	myui.Column(c).Grow(1).Gap(12).Children(func() {
		myui.Text(c, i18n.T("车型查询")).FontSize(22).FontWeight(700).TextColor(p.text)

		a.settingsGroup(c, func() {
			a.settingsRow(c, i18n.T("关键词"), i18n.T("按车系名称、拼音或 ID 查询"), func() {
				submitted := myui.TextInput(c.Key("series-query-keyword"), &a.seriesQueryKeyword).
					Placeholder(i18n.T("输入关键词")).Label(i18n.T("输入关键词")).Width(240).Submitted()
				if myui.Button(c.Key("series-query-run"), i18n.T("查询")).Clicked() {
					submitted = true
				}
				if submitted {
					a.runSeriesSearch()
				}
				if strings.TrimSpace(a.seriesQueryKeyword) != "" &&
					myui.Button(c.Key("series-query-clear"), i18n.T("清除")).Clicked() {
					a.clearSeriesSearch()
				}
			})
		})

		if strings.TrimSpace(a.seriesQueryKeyword) == "" {
			a.ensureSeriesBrowse()
			a.renderSeriesResults(c, a.seriesBrowse, a.seriesBrowseLoading, a.seriesBrowseErr, a.retrySeriesBrowse, &a.seriesList)
			return
		}
		a.renderSeriesResults(c, a.seriesQueryResults, a.seriesQueryLoading, a.seriesQueryErr, a.runSeriesSearch, &a.seriesList)
	})
}

// filterOption 把名称与 ID 编码成下拉选项文本，ID 从末尾的 "  #ID" 解析。
// 选项文本既作为 Select 的显示值，也承载筛选所需的 ID，避免额外维护映射表。
func filterOption(name string, id int64) string {
	return fmt.Sprintf("%s  #%d", name, id)
}

func filterOptionID(option string) int64 {
	i := strings.LastIndex(option, "  #")
	if i < 0 {
		return 0
	}
	id, err := strconv.ParseInt(option[i+3:], 10, 64)
	if err != nil {
		return 0
	}
	return id
}

// linkedOptions 生成某一级下拉的选项。brandID 为品牌下拉选中的 ID：
// 为 0（全部）时车系列表不做过滤。
func (a *App) linkedOptions(level string, brandID int64) []string {
	all := i18n.T("全部")
	opts := []string{all}
	switch level {
	case "brand":
		for _, b := range a.brands {
			opts = append(opts, filterOption(b.Name, b.ID))
		}
	case "series":
		for _, s := range a.seriesBrowse {
			if brandID != 0 && s.BrandID != brandID {
				continue
			}
			opts = append(opts, filterOption(s.Name, s.ID))
		}
	}
	return opts
}

// seriesByID 从已加载的全量车系中按 ID 查找，用于由车型推导其品牌与厂商。
func (a *App) seriesByID(id int64) (domain.Series, bool) {
	for _, s := range a.seriesBrowse {
		if s.ID == id {
			return s, true
		}
	}
	return domain.Series{}, false
}

// brandSeriesSelect 绘制“品牌 → 车型”两级联动下拉，返回选中的品牌 ID、
// 车型 ID 以及本帧是否有选择变化。选中项存的是选项文本（名称 #ID）。
// 品牌变化时车型重置为“全部”；车型下拉按当前品牌过滤。
// 供筛选查询、版型列表、图集列表复用，保证联动行为一致。
func (a *App) brandSeriesSelect(c *myui.Context, brandKey, seriesKey string, brand, series *string, width float32) (int64, int64, bool) {
	a.ensureSeriesBrowse()
	all := i18n.T("全部")

	brandOpts := a.linkedOptions("brand", 0)
	*brand = normalizeFilterChoice(*brand, brandOpts)
	seriesOpts := a.linkedOptions("series", filterOptionID(*brand))
	*series = normalizeFilterChoice(*series, seriesOpts)

	changed := false
	if myui.Select(c.Key(brandKey), brand, brandOpts).Label(i18n.T("品牌")).Width(width).Changed() {
		*series = all
		seriesOpts = a.linkedOptions("series", filterOptionID(*brand))
		changed = true
	}
	if myui.Select(c.Key(seriesKey), series, seriesOpts).Label(i18n.T("车型")).Width(width).Changed() {
		changed = true
	}
	return filterOptionID(*brand), filterOptionID(*series), changed
}

// levelChoice 把一个级别大类映射到库内 level_name 的模糊匹配关键词。
// 数据里共 23 种级别，按车型归类为下列选项，关键词均为其稳定子串。
type levelChoice struct {
	label   string
	keyword string
}

// levelChoices 返回“全部”+ 按车型整理的级别大类。
// 覆盖库内全部取值：轿车(6)、SUV(5)、MPV(4)、跑车、皮卡(2)、
// 微面、微卡、轻客、轻卡、飞行汽车。
func levelChoices() []levelChoice {
	return []levelChoice{
		{i18n.T("全部"), ""},
		{i18n.T("轿车"), i18n.Data("型车")},
		{i18n.T("SUV"), "suv"},
		{i18n.T("MPV"), "mpv"},
		{i18n.T("跑车"), i18n.Data("跑车")},
		{i18n.T("皮卡"), i18n.Data("皮卡")},
		{i18n.T("微面"), i18n.Data("微面")},
		{i18n.T("微卡"), i18n.Data("微卡")},
		{i18n.T("轻客"), i18n.Data("轻客")},
		{i18n.T("轻卡"), i18n.Data("轻卡")},
		{i18n.T("飞行汽车"), i18n.Data("飞行汽车")},
	}
}

func levelLabels() []string {
	choices := levelChoices()
	labels := make([]string, len(choices))
	for i, c := range choices {
		labels[i] = c.label
	}
	return labels
}

// levelKeyword 把级别下拉的显示文案映射为匹配关键词；找不到时视为“全部”。
func levelKeyword(label string) string {
	if label == i18n.T("全部") {
		return ""
	}
	for _, c := range levelChoices() {
		if c.label == label {
			return c.keyword
		}
	}
	return ""
}

// normalizeFilterChoice 在联动后把已失效的选中项（不在新选项里）重置为“全部”。
func normalizeFilterChoice(choice string, options []string) string {
	if choice == i18n.T("全部") {
		return choice
	}
	for _, o := range options {
		if o == choice {
			return choice
		}
	}
	return i18n.T("全部")
}

func (a *App) renderSeriesFilter(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)

	// 语言切换后级别文案会变化，旧文案不在新选项里时也重置为“全部”。
	a.seriesFilterLevel = normalizeFilterChoice(a.seriesFilterLevel, levelLabels())

	myui.Column(c).Grow(1).Gap(12).Children(func() {
		myui.Text(c, i18n.T("筛选查询")).FontSize(22).FontWeight(700).TextColor(p.text)

		a.settingsGroup(c, func() {
			a.settingsRow(c, i18n.T("品牌 / 车型"), i18n.T("先选品牌，车型选项随品牌联动"), func() {
				a.brandSeriesSelect(c, "series-filter-brand", "series-filter-series",
					&a.seriesFilterBrand, &a.seriesFilterSeries, 190)
			})
			a.settingsRow(c, i18n.T("级别 / 能源"), i18n.T("按车型归类选择级别"), func() {
				myui.Select(c.Key("series-filter-level"), &a.seriesFilterLevel, levelLabels()).Label(i18n.T("级别")).Width(160)
				myui.Select(c.Key("series-filter-energy"), &a.seriesFilterEnergy, []string{i18n.T("全部"), i18n.T("新能源"), i18n.T("燃油")}).Label(i18n.T("能源")).Width(110)
			})
			a.settingsRow(c, i18n.T("关键字"), i18n.T("匹配车系名称、拼音或车组；纯数字按 ID 查询"), func() {
				submitted := myui.TextInput(c.Key("series-filter-keyword"), &a.seriesFilterKeyword).
					Placeholder(i18n.T("输入关键词")).Label(i18n.T("输入关键词")).Width(240).Submitted()
				if myui.Button(c.Key("series-filter-run"), i18n.T("查询")).Clicked() {
					submitted = true
				}
				if submitted {
					a.runSeriesFilter()
				}
			})
		})

		// 首次进入按空条件查一次，保持原来的默认行为；之后只在点查询时刷新。
		if !a.seriesFilterLoaded && !a.seriesFilterLoading && a.seriesFilterErr == "" {
			a.runSeriesFilter()
		}
		a.renderSeriesResults(c, a.seriesFilterResults, a.seriesFilterLoading, a.seriesFilterErr, a.runSeriesFilter, &a.seriesFilterList)
	})
}

// renderSeriesResults 绘制结果计数、加载/错误状态和虚拟列表本身。
// 列表只构建可视行，几万条车系也不会一次性排版。
func (a *App) renderSeriesResults(c *myui.Context, items []domain.Series, loading bool, errText string, retry func(), state *myui.ListState) {
	p := paletteFor(c.Theme().Dark)
	myui.Row(c).AlignItems(myui.Center).Gap(10).Children(func() {
		switch {
		case loading:
			myui.Text(c, i18n.T("查询中…")).FontSize(13).TextColor(p.muted)
		case errText != "":
			myui.Text(c, fmt.Sprintf(i18n.T("查询失败：%s"), errText)).FontSize(13).TextColor(p.muted).Grow(1).SingleLine()
			if myui.Button(c, i18n.T("重试")).Clicked() {
				retry()
			}
		default:
			myui.Text(c, fmt.Sprintf(i18n.T("共 %d 个车系"), len(items))).FontSize(13).TextColor(p.muted)
		}
	})

	if len(items) == 0 {
		if loading {
			a.seriesSkeleton(c)
			return
		}
		if errText == "" {
			myui.Text(c, i18n.T("没有匹配的车系。")).FontSize(13).TextColor(p.muted)
		}
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

	// 负 margin 让列表延伸到窗口右缘（抵消外层 content 列的右 padding），
	// 滚动条贴住最右侧，与普通页 ScrollbarInsets 右 0 的效果一致；
	// 行内容再用内层右 padding 缩进，与页面上方元素对齐。
	myui.List(c, state, len(items), func(i int) {
		a.seriesRow(c, items[i], brandNames, firmNames, p)
	}).ItemKey(func(i int) any { return items[i].ID }).Grow(1).
		Margin(0, -22, 0, 0).Padding(0, 22, 0, 0)
}

// seriesRowHeight 是车系行的固定高度，与虚拟列表的估算一致。
const seriesRowHeight = 58

// seriesRow 是 music 歌曲行的车系版本：封面、名称/元信息两行、
// 新能源标记和右对齐的 ID，悬停时整行高亮。
func (a *App) seriesRow(c *myui.Context, item domain.Series, brandNames, firmNames map[int64]string, p palette) {
	row := myui.ButtonBase(c.Key(fmt.Sprintf("series-row-%d", item.ID))).
		Height(seriesRowHeight).PaddingX(12).Gap(12).Radius(10).
		AlignItems(myui.Center).FocusRing(false).
		Label(fmt.Sprintf("%s #%d", item.Name, item.ID))
	bg := myui.Transparent
	if row.Hovered() {
		bg = p.hover
	}
	row.Background(row.AnimateColor("bg", bg, myui.Animation{Duration: 120 * time.Millisecond}))
	row.Children(func() {
		a.seriesCover(c, item, brandNames[item.BrandID], p)

		myui.Column(c).Grow(3).Basis(0).MinWidth(0).Gap(3).Children(func() {
			name := item.Name
			if item.GroupName != "" && item.GroupName != item.Name {
				name += " · " + item.GroupName
			}
			myui.Text(c, name).FontSize(14).FontWeight(500).SingleLine()
			desc := fmt.Sprintf(i18n.T("品牌 %s  ·  厂商 %s  ·  级别 %s  ·  指导价 %.2f-%.2f"),
				nameOrID(brandNames, item.BrandID), nameOrID(firmNames, item.FirmID),
				item.LevelName, item.PriceMin, item.PriceMax)
			myui.Text(c, desc).FontSize(12.5).TextColor(p.muted).SingleLine()
		})

		if item.NewEnergy != 0 {
			myui.Box(c).Padding(2, 7).Radius(999).Background(p.accent.Alpha(0.12)).Children(func() {
				myui.Text(c, i18n.T("新能源")).FontSize(11).TextColor(p.accent)
			})
		}
		myui.Text(c, fmt.Sprintf("#%d", item.ID)).FontSize(12).FontFeatures("tnum").
			TextColor(p.faint).Width(70).TextAlign(myui.End)
	})
	if row.Clicked() {
		a.setStatus(fmt.Sprintf(i18n.T("已选择车系 #%d %s"), item.ID, item.Name))
	}
}

// seriesCover 显示车系 Logo；本地文件缺失时退化为品牌首字占位，
// 与 music 的封面占位一样保持固定尺寸，不因缺图改变行高。
func (a *App) seriesCover(c *myui.Context, item domain.Series, brand string, p palette) {
	box := myui.Box(c).Size(38, 38).Radius(6).Center()
	if bitmap := a.seriesLogo(item.LogoLocalPath); bitmap != nil {
		box.Children(func() {
			myui.Image(c, bitmap).Size(38, 38).Fit(myui.Cover).Radius(6)
		})
		return
	}
	box.Background(p.hover)
	label := "?"
	for _, r := range brand {
		label = string(r)
		break
	}
	if label == "?" {
		for _, r := range item.Name {
			label = string(r)
			break
		}
	}
	box.Children(func() {
		myui.Text(c, label).FontSize(15).FontWeight(600).TextColor(p.faint)
	})
}

// seriesLogo 解析车系 Logo 的本地路径并走图集相同的位图缓存；
// 只有可视行会调用，缺图时缓存 nil，不反复读盘。
func (a *App) seriesLogo(path string) *myui.Bitmap {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(a.paths.Root, filepath.FromSlash(path))
	}
	return a.bitmapAt(path)
}

func nameOrID(names map[int64]string, id int64) string {
	if name := names[id]; name != "" {
		return name
	}
	return fmt.Sprintf("#%d", id)
}

// seriesSkeleton 在首次加载时按行高铺占位块，避免空白页。
func (a *App) seriesSkeleton(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	for i := 0; i < 8; i++ {
		myui.Row(c.Key(fmt.Sprintf("series-skeleton-%d", i))).
			Height(seriesRowHeight).PaddingX(12).Gap(12).AlignItems(myui.Center).Children(func() {
			myui.Box(c).Size(38, 38).Radius(6).Background(p.hover)
			myui.Column(c).Grow(1).Gap(6).Children(func() {
				myui.Box(c).Height(10).Width(160).Radius(5).Background(p.hover)
				myui.Box(c).Height(9).Width(280).Radius(4).Background(p.hover)
			})
		})
	}
}

// ensureSeriesBrowse 在空关键词时后台加载全部车系，只做一次。
func (a *App) ensureSeriesBrowse() {
	if a.seriesBrowseLoaded || a.seriesBrowseLoading || a.seriesBrowseErr != "" {
		return
	}
	a.seriesBrowseLoading = true
	go func() {
		items, err := a.svc.ListSeries(context.Background(), -1)
		a.postMain(func() {
			a.seriesBrowseLoading = false
			if err != nil {
				a.seriesBrowseErr = err.Error()
			} else {
				a.seriesBrowse = items
				a.seriesBrowseLoaded = true
			}
			a.invalidateWindows()
		})
	}()
}

func (a *App) retrySeriesBrowse() {
	a.seriesBrowseErr = ""
	a.seriesList = myui.ListState{}
	a.ensureSeriesBrowse()
}

// runSeriesSearch 在后台执行关键词查询，结果回投主线程；序号丢弃过期结果。
func (a *App) runSeriesSearch() {
	keyword := strings.TrimSpace(a.seriesQueryKeyword)
	a.seriesQuerySeq++
	seq := a.seriesQuerySeq
	a.seriesQueryLoaded = false
	a.seriesQueryErr = ""
	a.seriesQueryResults = nil
	a.seriesList = myui.ListState{}
	if keyword == "" {
		a.seriesQueryLoading = false
		a.invalidateWindows()
		return
	}
	a.seriesQueryLoading = true
	go func() {
		items, err := a.svc.SearchSeries(context.Background(), keyword, -1)
		a.postMain(func() {
			if seq != a.seriesQuerySeq {
				return
			}
			a.seriesQueryLoading = false
			if err != nil {
				a.seriesQueryErr = err.Error()
			} else {
				a.seriesQueryResults = items
				a.seriesQueryLoaded = true
			}
			a.invalidateWindows()
		})
	}()
}

func (a *App) clearSeriesSearch() {
	a.seriesQueryKeyword = ""
	a.seriesQuerySeq++
	a.seriesQueryLoaded = false
	a.seriesQueryLoading = false
	a.seriesQueryErr = ""
	a.seriesQueryResults = nil
	a.seriesList = myui.ListState{}
	a.invalidateWindows()
}

// runSeriesFilter 在后台执行筛选查询；ID 由联动下拉的选中项解析，能源值的语言映射在这里完成。
func (a *App) runSeriesFilter() {
	brandID := filterOptionID(a.seriesFilterBrand)
	seriesID := filterOptionID(a.seriesFilterSeries)
	energy := a.seriesFilterEnergy
	switch energy {
	case i18n.T("全部"):
		energy = ""
	case i18n.T("新能源"):
		energy = i18n.Data("新能源")
	case i18n.T("燃油"):
		energy = i18n.Data("燃油")
	}

	a.seriesFilterSeq++
	seq := a.seriesFilterSeq
	a.seriesFilterLoading = true
	a.seriesFilterLoaded = false
	a.seriesFilterErr = ""
	a.seriesFilterResults = nil
	a.seriesFilterList = myui.ListState{}
	filter := service.SeriesFilter{
		BrandID:  brandID,
		SeriesID: seriesID,
		Level:    levelKeyword(a.seriesFilterLevel),
		Energy:   energy,
		Keyword:  strings.TrimSpace(a.seriesFilterKeyword),
		// SQLite 的负 LIMIT 表示不限条数；0 会一行都不返回。
		Limit: -1,
	}
	go func() {
		items, err := a.svc.FilterSeries(context.Background(), filter)
		a.postMain(func() {
			if seq != a.seriesFilterSeq {
				return
			}
			a.seriesFilterLoading = false
			a.seriesFilterLoaded = true
			if err != nil {
				a.seriesFilterErr = err.Error()
			} else {
				a.seriesFilterResults = items
			}
			a.invalidateWindows()
		})
	}()
}
