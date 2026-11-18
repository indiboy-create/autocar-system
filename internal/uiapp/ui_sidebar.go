package uiapp

import (
	"autocarsystem/internal/i18n"
	"fmt"
	myui "github.com/egoist/mygo/ui"
	"strconv"
)

func (a *App) sidebar(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)

	myui.Column(c).Fill().Padding(14, 14, 14, 14).Gap(4).DragWindow().Children(func() {
		myui.Row(c).AlignItems(myui.Center).Gap(10).Margin(0, 0, 12, 6).Children(func() {
			myui.Box(c).Size(36, 36).Radius(10).Background(p.accent).Center().Children(func() {
				myui.Icon(c, iconLogo).Size(20, 20).TextColor(p.onAccent)
			})
			myui.Column(c).Gap(2).Children(func() {
				myui.Text(c, i18n.T("新方圆")).FontSize(18).FontWeight(700).TextColor(p.text)
				myui.Text(c, i18n.T("版型图集管理")).FontSize(12).TextColor(p.muted)
			})
		})

		myui.Scroll(c.Key("nav")).TrackScroll(&a.navScroll).Grow(1).Margin(0, -6).Padding(0, 6).Gap(4).Children(func() {
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
