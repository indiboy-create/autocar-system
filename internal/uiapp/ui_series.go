package uiapp

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"autocarsystem/internal/domain"
	"autocarsystem/internal/i18n"
	"autocarsystem/internal/service"

	myui "github.com/egoist/mygo/ui"
)

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

func (a *App) renderSeriesFilter(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	myui.Column(c).Grow(1).Gap(12).Children(func() {
		myui.Text(c, i18n.T("筛选查询")).FontSize(22).FontWeight(700).TextColor(p.text)

		a.settingsGroup(c, func() {
			a.settingsRow(c, i18n.T("品牌 / 厂商 ID"), i18n.T("只查询指定品牌或厂商的车系"), func() {
				myui.TextInput(c.Key("series-filter-brand"), &a.seriesFilterBrandID).Placeholder(i18n.T("品牌 ID")).Label(i18n.T("品牌 ID")).Width(110)
				myui.TextInput(c.Key("series-filter-firm"), &a.seriesFilterFirmID).Placeholder(i18n.T("厂商 ID")).Label(i18n.T("厂商 ID")).Width(110)
			})
			a.settingsRow(c, i18n.T("级别 / 能源"), i18n.T("级别支持模糊匹配"), func() {
				myui.TextInput(c.Key("series-filter-level"), &a.seriesFilterLevel).Placeholder(i18n.T("级别，如 SUV")).Label(i18n.T("级别，如 SUV")).Width(140)
				myui.Select(c.Key("series-filter-energy"), &a.seriesFilterEnergy, []string{i18n.T("全部"), i18n.T("新能源"), i18n.T("燃油")}).Label(i18n.T("能源")).Width(110)
			})
			a.settingsRow(c, "", "", func() {
				if myui.Button(c.Key("series-filter-run"), i18n.T("查询")).Clicked() {
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

	myui.List(c, state, len(items), func(i int) {
		a.seriesRow(c, items[i], brandNames, firmNames, p)
	}).ItemKey(func(i int) any { return items[i].ID }).Grow(1)
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

// runSeriesFilter 在后台执行筛选查询；校验与能源值的语言映射都在这里完成。
func (a *App) runSeriesFilter() {
	brandID, err := parseOptionalID(a.seriesFilterBrandID)
	if err != nil {
		a.seriesFilterErr = i18n.T("品牌 ID 无效")
		a.seriesFilterLoaded = true
		return
	}
	firmID, err := parseOptionalID(a.seriesFilterFirmID)
	if err != nil {
		a.seriesFilterErr = i18n.T("厂商 ID 无效")
		a.seriesFilterLoaded = true
		return
	}
	energy := a.seriesFilterEnergy
	switch energy {
	case i18n.T("全部"):
		energy = ""
	case i18n.T("新能源"):
		energy = "新能源"
	case i18n.T("燃油"):
		energy = "燃油"
	}

	a.seriesFilterSeq++
	seq := a.seriesFilterSeq
	a.seriesFilterLoading = true
	a.seriesFilterLoaded = false
	a.seriesFilterErr = ""
	a.seriesFilterResults = nil
	a.seriesFilterList = myui.ListState{}
	filter := service.SeriesFilter{
		BrandID: brandID,
		FirmID:  firmID,
		Level:   strings.ToLower(strings.TrimSpace(a.seriesFilterLevel)),
		Energy:  energy,
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
