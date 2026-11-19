package uiapp

import (
	"autocarsystem/internal/domain"
	"autocarsystem/internal/i18n"
	"context"
	"fmt"
	"strings"
	"time"

	myui "github.com/egoist/mygo/ui"
)

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

// catalogCardHeight 是品牌/厂商卡片的固定高度，与 GridView 的行高一致。
const catalogCardHeight = 96

// catalogGridHeight 限制网格高度，让上千张卡片只构建可视区域。
const catalogGridHeight = 480

// logoThumb 显示品牌或厂商 Logo；本地文件缺失时退化为名称首字占位。
// 走与车系相同的位图缓存，缺图记 nil，不反复读盘。
func (a *App) logoThumb(c *myui.Context, path, name string, p palette) {
	const size = 40
	box := myui.Box(c).Size(size, size).Radius(9).Center()
	if bitmap := a.seriesLogo(path); bitmap != nil {
		box.Children(func() {
			myui.Image(c, bitmap).Size(size, size).Fit(myui.Cover).Radius(9)
		})
		return
	}
	box.Background(p.hover)
	label := "?"
	for _, r := range name {
		label = string(r)
		break
	}
	box.Children(func() {
		myui.Text(c, label).FontSize(17).FontWeight(600).TextColor(p.faint)
	})
}

// catalogCard 绘制一张可点击编辑的品牌/厂商卡片：Logo、名称、ID 与元信息。
func (a *App) catalogCard(c *myui.Context, key string, logoPath, name, meta string, badges []string, id int64, entity string, p palette) {
	card := myui.ButtonBase(c.Key(key)).
		Height(catalogCardHeight).Padding(12).Gap(12).Radius(12).
		AlignItems(myui.Center).FocusRing(false).
		Label(fmt.Sprintf("%s #%d", name, id))
	bg := p.bg
	if card.Hovered() {
		bg = p.hover
	}
	if a.catalogEditID == id && a.catalogEntity == entity {
		bg = p.press
	}
	card.Background(card.AnimateColor("bg", bg, myui.Animation{Duration: 120 * time.Millisecond}))
	card.Border(1, p.hairline)
	card.Children(func() {
		a.logoThumb(c, logoPath, name, p)
		myui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(5).Children(func() {
			myui.Text(c, name).FontSize(14).FontWeight(600).SingleLine()
			myui.Text(c, meta).FontSize(11.5).TextColor(p.muted).SingleLine()
			if len(badges) > 0 {
				myui.Row(c).Gap(6).Children(func() {
					for _, badge := range badges {
						myui.Box(c).Padding(2, 7).Radius(999).Background(p.accent.Alpha(0.12)).Children(func() {
							myui.Text(c, badge).FontSize(10.5).TextColor(p.accent)
						})
					}
				})
			}
		})
	})
	if card.Clicked() {
		a.loadCatalogEdit(entity, id)
	}
}

func (a *App) renderBrandList(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	myui.Row(c).Gap(10).AlignItems(myui.Center).Children(func() {
		myui.Text(c, fmt.Sprintf(i18n.T("共 %d 个品牌"), len(a.brands))).FontSize(13).TextColor(p.muted).Grow(1)
		myui.TextInput(c.Key("brand-search"), &a.brandSearch).
			Placeholder(i18n.T("搜索名称 / 首字母 / ID")).Label(i18n.T("搜索品牌")).Width(220)
	})
	if len(a.brands) == 0 {
		myui.Text(c, i18n.T("暂无品牌数据。"))
		return
	}
	keyword := strings.ToLower(strings.TrimSpace(a.brandSearch))
	items := make([]domain.Brand, 0, len(a.brands))
	for _, item := range a.brands {
		if keyword != "" {
			id := fmt.Sprintf("#%d", item.ID)
			name := strings.ToLower(item.Name)
			letter := strings.ToLower(item.FirstLetter)
			if !strings.Contains(name, keyword) && !strings.Contains(letter, keyword) && !strings.Contains(id, keyword) {
				continue
			}
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		myui.Text(c, i18n.T("没有匹配的品牌。")).FontSize(13).TextColor(p.muted)
		return
	}
	myui.GridView(c, &a.brandGrid, len(items), 260, catalogCardHeight, func(i int) {
		item := items[i]
		meta := fmt.Sprintf(i18n.T("ID %d"), item.ID)
		if item.FirstLetter != "" {
			meta += "  ·  " + fmt.Sprintf(i18n.T("首字母 %s"), item.FirstLetter)
		}
		var badges []string
		if item.State != 0 {
			badges = append(badges, i18n.T("启用"))
		}
		if item.HaveNewEnergy != 0 {
			badges = append(badges, i18n.T("新能源"))
		}
		a.catalogCard(c, fmt.Sprintf("brand-card-%d", item.ID), item.LogoLocalPath,
			item.Name, meta, badges, item.ID, "brand", p)
	}).Height(catalogGridHeight).Gap(12)
}

func (a *App) renderFirmList(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	brandNames := make(map[int64]string, len(a.brands))
	for _, b := range a.brands {
		brandNames[b.ID] = b.Name
	}
	myui.Row(c).Gap(10).AlignItems(myui.Center).Children(func() {
		myui.Text(c, fmt.Sprintf(i18n.T("共 %d 个厂商"), len(a.firms))).FontSize(13).TextColor(p.muted).Grow(1)
		myui.TextInput(c.Key("firm-search"), &a.firmSearch).
			Placeholder(i18n.T("搜索名称 / 拼音 / ID")).Label(i18n.T("搜索厂商")).Width(220)
	})
	if len(a.firms) == 0 {
		myui.Text(c, i18n.T("暂无厂商数据。"))
		return
	}
	keyword := strings.ToLower(strings.TrimSpace(a.firmSearch))
	items := make([]domain.Firm, 0, len(a.firms))
	for _, item := range a.firms {
		if keyword != "" {
			id := fmt.Sprintf("#%d", item.ID)
			name := strings.ToLower(item.Name)
			py := strings.ToLower(item.PY)
			if !strings.Contains(name, keyword) && !strings.Contains(py, keyword) && !strings.Contains(id, keyword) {
				continue
			}
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		myui.Text(c, i18n.T("没有匹配的厂商。")).FontSize(13).TextColor(p.muted)
		return
	}
	myui.GridView(c, &a.firmGrid, len(items), 280, catalogCardHeight, func(i int) {
		item := items[i]
		meta := fmt.Sprintf(i18n.T("ID %d  ·  品牌 %s"), item.ID, nameOrID(brandNames, item.BrandID))
		var badges []string
		if item.PY != "" {
			badges = append(badges, item.PY)
		}
		a.catalogCard(c, fmt.Sprintf("firm-card-%d", item.ID), item.LogoLocalPath,
			item.Name, meta, badges, item.ID, "firm", p)
	}).Height(catalogGridHeight).Gap(12)
}

func (a *App) renderSeriesList(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	brandNames := make(map[int64]string, len(a.brands))
	for _, b := range a.brands {
		brandNames[b.ID] = b.Name
	}
	firmNames := make(map[int64]string, len(a.firms))
	for _, f := range a.firms {
		firmNames[f.ID] = f.Name
	}
	myui.Row(c).Gap(10).AlignItems(myui.Center).Children(func() {
		myui.Text(c, fmt.Sprintf(i18n.T("已加载 %d / %d 个车型"), len(a.series), a.countInt("series"))).FontSize(13).TextColor(p.muted).Grow(1)
		myui.TextInput(c.Key("series-search"), &a.seriesSearch).
			Placeholder(i18n.T("搜索名称 / 拼音 / ID")).Label(i18n.T("搜索车型")).Width(220)
	})
	if len(a.series) == 0 {
		myui.Text(c, i18n.T("暂无车系数据。"))
		return
	}
	keyword := strings.ToLower(strings.TrimSpace(a.seriesSearch))
	items := make([]domain.Series, 0, len(a.series))
	for _, item := range a.series {
		if keyword != "" {
			id := fmt.Sprintf("#%d", item.ID)
			if !strings.Contains(strings.ToLower(item.Name), keyword) &&
				!strings.Contains(strings.ToLower(item.PY), keyword) &&
				!strings.Contains(id, keyword) {
				continue
			}
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		myui.Text(c, i18n.T("没有匹配的车型。")).FontSize(13).TextColor(p.muted)
	} else {
		myui.GridView(c, &a.seriesGrid, len(items), 260, catalogCardHeight, func(i int) {
			item := items[i]
			meta := fmt.Sprintf(i18n.T("ID %d  ·  %s"), item.ID, nameOrID(firmNames, item.FirmID))
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
			a.catalogCard(c, fmt.Sprintf("series-card-%d", item.ID), item.LogoLocalPath,
				item.Name, meta, badges, item.ID, "series", p)
		}).Height(catalogGridHeight).Gap(12)
	}
	a.catalogMore(c, len(a.series), a.countInt("series"), func() {
		a.seriesLimit += catalogBatch
		a.Refresh()
	})
}

func (a *App) renderSpecList(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
	seriesNames := make(map[int64]string, len(a.series))
	for _, s := range a.series {
		seriesNames[s.ID] = s.Name
	}
	myui.Row(c).Gap(10).AlignItems(myui.Center).Children(func() {
		myui.Text(c, fmt.Sprintf(i18n.T("已加载 %d / %d 个年款"), len(a.specs), a.countInt("spec"))).FontSize(13).TextColor(p.muted).Grow(1)
		myui.TextInput(c.Key("spec-search"), &a.specSearch).
			Placeholder(i18n.T("搜索名称 / 车系 / ID")).Label(i18n.T("搜索年款")).Width(220)
	})
	if len(a.specs) == 0 {
		myui.Text(c, i18n.T("暂无年款数据。"))
		return
	}
	keyword := strings.ToLower(strings.TrimSpace(a.specSearch))
	items := make([]domain.Spec, 0, len(a.specs))
	for _, item := range a.specs {
		if keyword != "" {
			id := fmt.Sprintf("#%d", item.ID)
			if !strings.Contains(strings.ToLower(item.Name), keyword) &&
				!strings.Contains(strings.ToLower(item.SeriesName), keyword) &&
				!strings.Contains(strings.ToLower(item.ModelYear), keyword) &&
				!strings.Contains(id, keyword) {
				continue
			}
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		myui.Text(c, i18n.T("没有匹配的年款。")).FontSize(13).TextColor(p.muted)
	} else {
		myui.GridView(c, &a.specGrid, len(items), 260, catalogCardHeight, func(i int) {
			item := items[i]
			meta := fmt.Sprintf(i18n.T("ID %d  ·  %s"), item.ID,
				nameWithFallback(item.SeriesName, seriesNames, item.SeriesID))
			var badges []string
			if item.ModelYear != "" {
				badges = append(badges, item.ModelYear)
			}
			if item.BodyForm != "" {
				badges = append(badges, item.BodyForm)
			}
			a.catalogCard(c, fmt.Sprintf("spec-card-%d", item.ID), item.ImageLocalPath,
				item.Name, meta, badges, item.ID, "spec", p)
		}).Height(catalogGridHeight).Gap(12)
	}
	a.catalogMore(c, len(a.specs), a.countInt("spec"), func() {
		a.specsLimit += catalogBatch
		a.Refresh()
	})
}

// nameWithFallback 优先使用记录自带名称（Spec.SeriesName），缺失时再按 ID 在内存映射里查找。
func nameWithFallback(own string, names map[int64]string, id int64) string {
	if strings.TrimSpace(own) != "" {
		return own
	}
	return nameOrID(names, id)
}

// catalogBatch 是目录页每次多加载的车系/年款条数。
const catalogBatch = 200

// catalogMore 在列表被分批截断时显示“已显示 / 共”并给出“加载更多”：
// 车系与年款可达万级，一次全量查询和渲染都会拖慢目录页，分批加载让全部
// 数据都能到达，又不把首屏变重。
func (a *App) catalogMore(c *myui.Context, loaded int, total int64, loadMore func()) {
	if total <= 0 || int64(loaded) >= total {
		return
	}
	myui.Row(c).Gap(8).AlignItems(myui.Center).Children(func() {
		myui.Text(c, fmt.Sprintf(i18n.T("已显示 %d / 共 %d 条"), loaded, total)).FontSize(12.5)
		if myui.Button(c, i18n.T("加载更多")).Clicked() {
			loadMore()
		}
	})
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
			myui.TextInput(c.Key("edit-brand-name"), &a.brandEdit.Name).Placeholder(i18n.T("名称")).Label(i18n.T("名称"))
			myui.TextInput(c.Key("edit-brand-letter"), &a.brandEdit.FirstLetter).Placeholder(i18n.T("首字母")).Label(i18n.T("首字母"))
		})
		myui.Row(c).Gap(8).Children(func() {
			myui.TextInput(c.Key("edit-brand-logo-url"), &a.brandEdit.LogoURL).Placeholder("Logo URL").Label("Logo URL")
			myui.TextInput(c.Key("edit-brand-logo-local"), &a.brandEdit.LogoLocalPath).Placeholder(i18n.T("Logo 本地路径")).Label(i18n.T("Logo 本地路径"))
		})
		myui.Row(c).Gap(8).Children(func() {
			myui.TextInput(c.Key("edit-brand-state"), &a.brandEdit.State).Placeholder("state").Label("state")
			myui.TextInput(c.Key("edit-brand-energy"), &a.brandEdit.HaveNewEnergy).Placeholder("havenewenergy").Label("havenewenergy")
			myui.TextInput(c.Key("edit-brand-sort"), &a.brandEdit.APISort).Placeholder("api_sort").Label("api_sort")
		})
	case "firm":
		myui.Row(c).Gap(8).Children(func() {
			myui.TextInput(c.Key("edit-firm-name"), &a.firmEdit.Name).Placeholder(i18n.T("名称")).Label(i18n.T("名称"))
			myui.TextInput(c.Key("edit-firm-py"), &a.firmEdit.PY).Placeholder("py").Label("py")
			myui.TextInput(c.Key("edit-firm-link"), &a.firmEdit.Link).Placeholder("link").Label("link")
		})
		myui.Row(c).Gap(8).Children(func() {
			myui.TextInput(c.Key("edit-firm-logo-url"), &a.firmEdit.LogoURL).Placeholder("Logo URL").Label("Logo URL")
			myui.TextInput(c.Key("edit-firm-logo-local"), &a.firmEdit.LogoLocalPath).Placeholder(i18n.T("Logo 本地路径")).Label(i18n.T("Logo 本地路径"))
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
		myui.TextInput(c.Key("se-name"), &a.seriesEdit.Name).Placeholder(i18n.T("名称")).Label(i18n.T("名称"))
		myui.TextInput(c.Key("se-group"), &a.seriesEdit.GroupName).Placeholder("group_name").Label("group_name")
		myui.TextInput(c.Key("se-place"), &a.seriesEdit.Place).Placeholder("place").Label("place")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("se-price-min"), &a.seriesEdit.PriceMin).Placeholder("price_min").Label("price_min")
		myui.TextInput(c.Key("se-price-max"), &a.seriesEdit.PriceMax).Placeholder("price_max").Label("price_max")
		myui.TextInput(c.Key("se-state"), &a.seriesEdit.State).Placeholder("state").Label("state")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("se-py"), &a.seriesEdit.PY).Placeholder("py").Label("py")
		myui.TextInput(c.Key("se-level"), &a.seriesEdit.LevelName).Placeholder("level_name").Label("level_name")
		myui.TextInput(c.Key("se-fuel"), &a.seriesEdit.FuelTypes).Placeholder("fuel_types").Label("fuel_types")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("se-cartype"), &a.seriesEdit.CartypeName).Placeholder("cartypename").Label("cartypename")
		myui.TextInput(c.Key("se-tag"), &a.seriesEdit.Tag1).Placeholder("tag1").Label("tag1")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("se-logo-url"), &a.seriesEdit.LogoURL).Placeholder("logo_url").Label("logo_url")
		myui.TextInput(c.Key("se-logo-local"), &a.seriesEdit.LogoLocalPath).Placeholder("logo_local_path").Label("logo_local_path")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("se-energy"), &a.seriesEdit.NewEnergy).Placeholder("newenergy").Label("newenergy")
		myui.TextInput(c.Key("se-foreign"), &a.seriesEdit.IsForeignCar).Placeholder("is_foreign_car").Label("is_foreign_car")
		myui.TextInput(c.Key("se-top"), &a.seriesEdit.IsTop).Placeholder("is_top").Label("is_top")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("se-rank"), &a.seriesEdit.Rank).Placeholder("rank").Label("rank")
		myui.TextInput(c.Key("se-uv"), &a.seriesEdit.UVRank).Placeholder("uv_rank").Label("uv_rank")
		myui.TextInput(c.Key("se-hot"), &a.seriesEdit.HotRank).Placeholder("hot_rank").Label("hot_rank")
	})
}

func (a *App) renderSpecEdit(c *myui.Context) {
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("sp-name"), &a.specEdit.Name).Placeholder(i18n.T("名称")).Label(i18n.T("名称"))
		myui.TextInput(c.Key("sp-group"), &a.specEdit.GroupName).Placeholder("group_name").Label("group_name")
		myui.TextInput(c.Key("sp-series-name"), &a.specEdit.SeriesName).Placeholder("series_name").Label("series_name")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("sp-year"), &a.specEdit.ModelYear).Placeholder("model_year").Label("model_year")
		myui.TextInput(c.Key("sp-body"), &a.specEdit.BodyForm).Placeholder("body_form").Label("body_form")
		myui.TextInput(c.Key("sp-sale"), &a.specEdit.SaleType).Placeholder("sale_type").Label("sale_type")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("sp-price-min"), &a.specEdit.PriceMin).Placeholder("price_min").Label("price_min")
		myui.TextInput(c.Key("sp-price-max"), &a.specEdit.PriceMax).Placeholder("price_max").Label("price_max")
		myui.TextInput(c.Key("sp-news-price"), &a.specEdit.NewsPrice).Placeholder("news_price").Label("news_price")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("sp-news-id"), &a.specEdit.NewsID).Placeholder("news_id").Label("news_id")
		myui.TextInput(c.Key("sp-dealer-id"), &a.specEdit.DealerID).Placeholder("dealer_id").Label("dealer_id")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("sp-image-url"), &a.specEdit.ImageURL).Placeholder("image_url").Label("image_url")
		myui.TextInput(c.Key("sp-image-local"), &a.specEdit.ImageLocalPath).Placeholder("image_local_path").Label("image_local_path")
		myui.TextInput(c.Key("sp-image-sha"), &a.specEdit.ImageSHA256).Placeholder("image_sha256").Label("image_sha256")
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.TextInput(c.Key("sp-white-url"), &a.specEdit.WhiteImageURL).Placeholder("white_image_url").Label("white_image_url")
		myui.TextInput(c.Key("sp-white-local"), &a.specEdit.WhiteImageLocalPath).Placeholder("white_image_local_path").Label("white_image_local_path")
		myui.TextInput(c.Key("sp-white-sha"), &a.specEdit.WhiteImageSHA256).Placeholder("white_image_sha256").Label("white_image_sha256")
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
