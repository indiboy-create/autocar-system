package uiapp

import (
	"autocarsystem/internal/config"
	"autocarsystem/internal/i18n"
	"fmt"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/glass"
	myui "github.com/egoist/mygo/ui"
)

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
		Vibrancy:             mygo.VibrancySidebar,
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
			side := myui.Column(c).Width(176)
			if !c.Vibrancy() {
				side.Background(p.sidebar)
			}
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
					TrackScroll(&a.settingsScroll).
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

func (a *App) savePrefs() {
	if err := config.SavePrefs(a.paths.PrefsPath, a.prefs); err != nil {
		a.settingsMessage = err.Error()
		return
	}
	a.settingsMessage = ""
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
