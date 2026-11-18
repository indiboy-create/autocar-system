package uiapp

import (
	"autocarsystem/internal/domain"
	"autocarsystem/internal/i18n"
	"context"
	"fmt"
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
