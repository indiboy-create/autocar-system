package uiapp

import (
	"autocarsystem/internal/domain"
	"autocarsystem/internal/i18n"
	"autocarsystem/internal/service"
	"context"
	"fmt"
	myui "github.com/egoist/mygo/ui"
	"strings"
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
	myui.Text(c, i18n.T("车型查询")).FontSize(22).FontWeight(700).TextColor(p.text)

	a.settingsGroup(c, func() {
		a.settingsRow(c, i18n.T("关键词"), i18n.T("按车系名称、拼音或 ID 查询"), func() {
			if myui.TextInput(c.Key("series-query-keyword"), &a.seriesQueryKeyword).Placeholder(i18n.T("输入关键词")).Label(i18n.T("输入关键词")).Width(240).Submitted() {
				a.seriesQueryLoaded = false
				a.Refresh()
			}
			if myui.Button(c.Key("series-query-run"), i18n.T("查询")).Clicked() {
				a.seriesQueryLoaded = false
				a.Refresh()
			}
		})
	})

	keyword := strings.TrimSpace(a.seriesQueryKeyword)
	if keyword != "" && !a.seriesQueryLoaded {
		ctx := context.Background()
		if results, err := a.svc.SearchSeries(ctx, keyword, 200); err == nil {
			a.seriesQueryResults = results
		} else {
			a.setStatus(err.Error())
			a.seriesQueryResults = nil
		}
		a.seriesQueryLoaded = true
	}

	var matched []domain.Series
	if keyword == "" {
		matched = a.series
	} else if a.seriesQueryLoaded {
		matched = a.seriesQueryResults
	}
	a.renderSeriesResults(c, matched)
}

func (a *App) renderSeriesFilter(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
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
				a.seriesFilterLoaded = false
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

	if !a.seriesFilterLoaded {
		ctx := context.Background()
		if results, err := a.svc.FilterSeries(ctx, service.SeriesFilter{
			BrandID: brandID,
			FirmID:  firmID,
			Level:   level,
			Energy:  a.seriesFilterEnergy,
			Limit:   200,
		}); err == nil {
			a.seriesFilterResults = results
		} else {
			a.setStatus(err.Error())
			a.seriesFilterResults = nil
		}
		a.seriesFilterLoaded = true
	}

	a.renderSeriesResults(c, a.seriesFilterResults)
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
