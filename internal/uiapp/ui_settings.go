package uiapp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"autocarsystem/internal/config"
	"autocarsystem/internal/i18n"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/glass"
	myui "github.com/egoist/mygo/ui"
)

// OpenSettingsPage 打开设置窗口并定位到指定页面，供托盘菜单等外部入口调用。
// 未知页面回退到“通用”；窗口已打开时切换页面并刷新。
func (a *App) OpenSettingsPage(page string) {
	switch page {
	case "general", "category", "sync", "backup", "system", "license", "about":
	case "updates":
		if !a.updates {
			page = "general"
		}
	default:
		page = "general"
	}
	a.settingsPage = page
	if a.settingsWin != nil {
		a.settingsWin.Invalidate()
		a.settingsWin.Show()
		a.settingsWin.Focus()
		return
	}
	a.openSettings()
}

// TrayCheckUpdates 从托盘触发检查，由 updater 插件弹出原生更新窗口。
func (a *App) TrayCheckUpdates() {
	if !a.updates {
		return
	}
	checkForUpdates()
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

	c.OnShortcut(myui.Cmd, myui.KeyF, func() { a.settingsSearchBox.Focus(c) })
	c.OnShortcut(0, myui.KeyEscape, func() {
		if strings.TrimSpace(a.settingsQuery) != "" {
			a.settingsQuery = ""
		}
	})

	if c.Vibrancy() {
		c.Root().Background(myui.Transparent)
	}
	root := myui.Box(c).Fill()
	if !c.Vibrancy() {
		root.Background(p.bg)
	}
	query := strings.TrimSpace(a.settingsQuery)
	root.Children(func() {
		page := myui.Row(c).Fill().AlignItems(myui.Stretch)
		page.Children(func() {
			side := myui.Column(c).Width(176)
			// 设置页左侧栏使用方案适配的纯色，不跟随窗口 Vibrancy 的渐变材质。
			side.Background(p.sidebar)
			side.Children(func() {
				myui.Column(c).Fill().Padding(18, 12, 14, 12).Gap(2).DragWindow().Children(func() {
					myui.Text(c, i18n.T("设置")).FontSize(18).FontWeight(700).TextColor(p.text).Margin(0, 0, 14, 8)
					a.settingsSearch(c, p)
					if settingsPageMatches("general", query) {
						a.settingsNavItem(c, iconSettings, i18n.T("通用"), "general")
					}
					if settingsPageMatches("category", query) {
						a.settingsNavItem(c, iconCategory, i18n.T("分类"), "category")
					}
					if settingsPageMatches("sync", query) {
						a.settingsNavItem(c, iconTransfer, i18n.T("同步"), "sync")
					}
					if settingsPageMatches("backup", query) {
						a.settingsNavItem(c, iconBackup, i18n.T("备份"), "backup")
					}
					if settingsPageMatches("system", query) {
						a.settingsNavItem(c, iconCatalog, i18n.T("系统"), "system")
					}
					if a.updates && settingsPageMatches("updates", query) {
						a.settingsNavItem(c, iconSync, i18n.T("更新"), "updates")
					}
					if settingsPageMatches("license", query) {
						a.settingsNavItem(c, iconLicense, i18n.T("授权"), "license")
					}
					if settingsPageMatches("about", query) {
						a.settingsNavItem(c, iconAbout, i18n.T("关于"), "about")
					}
				})
			})

			main := myui.Box(c.Key("settings-main")).Grow(1).Background(p.bg)
			main.Children(func() {
				pad := float32(46)
				scrollKey := "settings-page-" + a.settingsPage
				if query != "" {
					scrollKey = "settings-search-" + query
				}
				myui.Scroll(c.Key(scrollKey)).
					TrackScroll(&a.settingsScroll).
					Grow(1).
					Padding(pad+4, 28, 28, 28).
					Gap(10).
					Children(func() {
						mark := myui.Box(c.Key("settings-start-mark")).Height(0).Bounds()
						if mark.W > 0 {
							a.settingsUnderTop = mark.Y < pad-0.5
						}
						if query != "" {
							a.settingsSearchResultsView(c, p, query)
						} else {
							a.settingsContent(c)
						}
					})
				edge := myui.Box(c.Key("settings-top-edge")).Absolute().Top(0).Left(0).Right(0).Height(pad).DragWindow().
					Material(glass.ScrollEdge{Hard: true, Background: p.bg})
				header := myui.Row(c.Key("settings-header")).Absolute().Top(0).Left(0).Right(0).Height(pad).
					Padding(0, 0, 0, 28).AlignItems(myui.Center).DragWindow()
				header.Children(func() {
					title := a.settingsTitle()
					if query != "" {
						title = i18n.T("搜索结果")
					}
					myui.Text(c, title).FontSize(15).FontWeight(700).TextColor(p.text)
				})
				edge.Opacity(edgeShown(edge, a.settingsUnderTop || header.Hovered()))
			})
		})
	})

	if a.toast.Text != "" {
		c.AddToast(myui.Toast{
			ID:      a.toast.Text,
			Title:   a.toast.Text,
			Type:    a.toast.Kind,
			Timeout: 3 * time.Second,
		})
		a.toast = appToast{}
	}
	a.toastViewport(c)
}

func (a *App) settingsTitle() string {
	switch a.settingsPage {
	case "category":
		return i18n.T("分类")
	case "sync":
		return i18n.T("同步")
	case "backup":
		return i18n.T("备份")
	case "system":
		return i18n.T("系统")
	case "updates":
		return i18n.T("更新")
	case "license":
		return i18n.T("授权")
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
		a.settingsQuery = ""
	}
}

func (a *App) settingsContent(c *myui.Context) {
	switch a.settingsPage {
	case "category":
		a.categorySettings(c)
	case "sync":
		a.syncSettings(c)
	case "backup":
		a.backupSettings(c)
	case "system":
		a.systemSettings(c)
	case "updates":
		if a.updates {
			a.updateSettings(c)
		} else {
			a.generalSettings(c)
		}
	case "license":
		a.licenseSettings(c)
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

// savePrefsWithToast 保存偏好后安排一条一次性提示，由设置窗口渲染时弹出。
// 启动时的偏好迁移不经过这里，避免打开设置窗口时弹出无关的“已保存”。
func (a *App) savePrefsWithToast(label string) {
	a.savePrefs()
	if message := strings.TrimSpace(a.settingsMessage); message != "" {
		a.toast = appToast{Text: fmt.Sprintf(i18n.T("设置保存失败：%s"), message), Kind: "error"}
		return
	}
	a.notifySettingSaved(label)
}

// notifySettingSaved 安排一条成功提示，文案形如“语言设置已保存”。
func (a *App) notifySettingSaved(label string) {
	a.toast = appToast{
		Text: fmt.Sprintf(i18n.T("%s设置已保存"), label),
		Kind: "success",
	}
}

// runBackupNow 立即备份一次数据库；结果会显示在备份页并刷新备份列表。
func (a *App) runBackupNow() {
	a.postMain(func() {
		a.settingsMessage = i18n.T("正在备份数据库…")
		if a.settingsWin != nil {
			a.settingsWin.Invalidate()
		}
	})
	a.createBackup(context.Background(), false)
}

// StartAutoBackup 启动后台定时备份循环。ctx 取消或应用退出后停止。
// 必须在设置页之外调用一次；循环本身每分钟检查一次是否到期。
func (a *App) StartAutoBackup(ctx context.Context) {
	go func() {
		a.runAutoBackupIfDue(ctx, time.Now())
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				var quitting bool
				a.postMain(func() { quitting = a.quitting })
				if quitting {
					return
				}
				a.runAutoBackupIfDue(ctx, now)
			}
		}
	}()
}

// runAutoBackupIfDue 在开启定时备份且距上次备份超过设定周期时执行一次备份。
// 从未备份过（LastBackupAt 为空）时立即备份一次，避免默认开启后迟迟没有首份备份。
func (a *App) runAutoBackupIfDue(ctx context.Context, now time.Time) {
	var due bool
	a.postMain(func() {
		interval := time.Duration(a.prefs.BackupIntervalHours) * time.Hour
		due = a.prefs.AutoBackup && !a.quitting && interval > 0
		if !due {
			return
		}
		lastText := strings.TrimSpace(a.prefs.LastBackupAt)
		if lastText == "" {
			due = true
			return
		}
		last, err := time.Parse(time.RFC3339, lastText)
		due = err != nil || now.Sub(last) >= interval
	})
	if due {
		a.createBackup(ctx, true)
	}
}

// createBackup 在后台创建备份，并把结果与更新后的 LastBackupAt 回投主线程。
func (a *App) createBackup(ctx context.Context, automatic bool) {
	go func() {
		info, err := a.svc.BackupDatabase(ctx)
		if err == nil {
			items, listErr := a.svc.ListBackups()
			if listErr == nil {
				a.postMain(func() {
					a.backups = items
					a.invalidateWindows()
				})
			}
		}
		a.postMain(func() {
			if err != nil {
				prefix := i18n.T("备份失败")
				if automatic {
					prefix = i18n.T("自动备份失败")
				}
				a.settingsMessage = fmt.Sprintf("%s：%v", prefix, err)
				a.invalidateWindows()
				return
			}
			a.prefs.LastBackupAt = time.Now().UTC().Format(time.RFC3339)
			if saveErr := config.SavePrefs(a.paths.PrefsPath, a.prefs); saveErr != nil {
				a.settingsMessage = saveErr.Error()
				a.invalidateWindows()
				return
			}
			if automatic {
				a.settingsMessage = fmt.Sprintf(i18n.T("自动备份完成：%s"), info.Name)
			} else {
				a.settingsMessage = fmt.Sprintf(i18n.T("备份完成：%s"), info.Name)
			}
			a.invalidateWindows()
		})
	}()
}

// backupSettings 数据库备份页：立即备份、备份列表与定时备份设置。
func (a *App) backupSettings(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)

	a.settingsGroup(c, func() {
		a.settingsRow(c, i18n.T("立即备份"), i18n.T("把当前数据库保存一份到备份目录"), func() {
			if myui.Button(c.Key("backup-now"), i18n.T("立即备份")).Clicked() {
				a.runBackupNow()
			}
		})
	})

	a.settingsGroup(c, func() {
		a.settingsRow(c, i18n.T("定时备份"), i18n.T("按设定周期自动备份数据库"), func() {
			on := a.prefs.AutoBackup
			if myui.Switch(c.Key("auto-backup"), &on).Label(i18n.T("定时备份")).Changed() {
				a.prefs.AutoBackup = on
				a.savePrefsWithToast(i18n.T("定时备份"))
				if on {
					go a.runAutoBackupIfDue(context.Background(), time.Now())
				}
			}
		})
		a.settingsRow(c, i18n.T("备份周期"), i18n.T("每隔指定小时自动备份一次"), func() {
			hours := []int{1, 6, 12, 24, 72, 168}
			labels := []string{i18n.T("每 1 小时"), i18n.T("每 6 小时"), i18n.T("每 12 小时"), i18n.T("每 1 天"), i18n.T("每 3 天"), i18n.T("每 7 天")}
			chosen := labels[len(labels)-1]
			for i, h := range hours {
				if a.prefs.BackupIntervalHours == h {
					chosen = labels[i]
					break
				}
			}
			if myui.Select(c.Key("backup-interval"), &chosen, labels).Label(i18n.T("备份周期")).Width(140).Changed() {
				for i, label := range labels {
					if chosen == label {
						a.prefs.BackupIntervalHours = hours[i]
						break
					}
				}
				a.savePrefsWithToast(i18n.T("备份周期"))
			}
		})
	})

	a.settingsGroup(c, func() {
		label := fmt.Sprintf(i18n.T("备份列表（%d）"), len(a.backups))
		a.settingsRow(c, label, i18n.T("数据库备份保存在数据目录的 backup 文件夹"), func() {})
		if len(a.backups) == 0 {
			a.settingsRow(c, i18n.T("暂无备份。"), "", func() {})
		} else {
			for _, item := range a.backups {
				current := item
				desc := fmt.Sprintf("%s  ·  %s", current.ModTime.Format("2006-01-02 15:04:05"), formatBytes(current.Size))
				a.settingsRow(c, current.Name, desc, func() {})
			}
		}
	})

	if strings.TrimSpace(a.settingsMessage) != "" {
		myui.Text(c, a.settingsMessage).FontSize(12.5).TextColor(p.muted)
	}
}

// formatBytes 把字节数格式化为可读大小。
func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	value := float64(n)
	for _, suffix := range []string{"KB", "MB", "GB"} {
		value /= unit
		if value < unit {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}
	return fmt.Sprintf("%.1f TB", value/unit)
}

// cardActionButton 是系统页的带图标操作按钮，主次按钮共用一致的圆角与间距。
func (a *App) cardActionButton(c *myui.Context, key, label string, icon *myui.SVG, primary bool, p palette, onClick func()) {
	b := myui.ButtonBase(c.Key(key)).
		Height(34).
		PaddingX(12).
		Gap(8).
		Radius(9).
		AlignItems(myui.Center).
		FocusRing(false).
		Label(label)

	color := p.text
	if primary {
		b.Background(p.accent)
		color = p.onAccent
	} else {
		b.Background(p.hover)
	}
	b.Children(func() {
		myui.Icon(c, icon).Size(15, 15).TextColor(color)
		myui.Text(c, label).FontSize(12.5).FontWeight(600).TextColor(color)
	})
	if b.Clicked() {
		onClick()
	}
}

func (a *App) renderSystemSettings(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)

	// 概览卡片：与更新页一致的图标、标题、状态徽标和分隔线布局。
	myui.Box(c).Radius(14).Background(p.bg).Border(1, p.hairline).Padding(18).Children(func() {
		myui.Column(c).Gap(14).Children(func() {
			myui.Row(c).Gap(12).AlignItems(myui.Center).Children(func() {
				myui.Box(c).Size(42, 42).Radius(12).Center().Background(p.accent.Alpha(0.12)).Children(func() {
					myui.Icon(c, iconSettings).Size(20, 20).TextColor(p.accent)
				})
				myui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(4).Children(func() {
					myui.Text(c, i18n.T("系统信息")).FontSize(17).FontWeight(700).SingleLine()
					myui.Text(c, i18n.T("本地存储路径与运行状态")).FontSize(12).TextColor(p.muted).LineHeight(1.35)
				})
				a.statusPill(c, i18n.T("本地运行"), statusSuccessColor)
			})
			myui.Box(c).Height(1).Background(p.hairline)
			myui.Text(c, i18n.T("数据与图片保存在主程序同级目录，卸载时会保留。")).FontSize(12.5).TextColor(p.muted).LineHeight(1.4)
		})
	})

	// 路径信息：完整显示，便于定位与复制。
	a.settingsGroup(c, func() {
		a.settingsRow(c, i18n.T("根目录"), a.paths.Root, func() {})
		a.settingsRow(c, i18n.T("数据目录"), a.paths.Data, func() {})
		a.settingsRow(c, i18n.T("媒体目录"), a.paths.Media, func() {})
		a.settingsRow(c, i18n.T("数据库"), a.paths.DBPath, func() {})
		a.settingsRow(c, i18n.T("配置文件"), a.paths.PrefsPath, func() {})
	})

	// 孤儿图集目录：数量徽标随扫描结果变化，操作反馈显示在同一张卡片内。
	count := len(a.orphanDirs)
	badge := i18n.T("未发现")
	badgeColor := p.muted
	if count > 0 {
		badge = fmt.Sprintf(i18n.T("发现 %d 个"), count)
		badgeColor = statusDangerColor
	}
	myui.Box(c).Radius(14).Background(p.bg).Border(1, p.hairline).Padding(18).Children(func() {
		myui.Column(c).Gap(14).Children(func() {
			myui.Row(c).Gap(12).AlignItems(myui.Center).Children(func() {
				myui.Box(c).Size(42, 42).Radius(12).Center().Background(p.accent.Alpha(0.12)).Children(func() {
					myui.Icon(c, iconGallery).Size(20, 20).TextColor(p.accent)
				})
				myui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(4).Children(func() {
					myui.Text(c, i18n.T("孤儿图集目录")).FontSize(17).FontWeight(700).SingleLine()
					myui.Text(c, i18n.T("扫描并清理不再被引用的图集目录")).FontSize(12).TextColor(p.muted).LineHeight(1.35)
				})
				a.statusPill(c, badge, badgeColor)
			})
			myui.Box(c).Height(1).Background(p.hairline)

			myui.Row(c).Gap(8).Children(func() {
				a.cardActionButton(c, "scan-orphans", i18n.T("扫描孤儿目录"), iconSync, true, p, a.scanOrphanDirs)
				a.cardActionButton(c, "clean-orphans", i18n.T("清理孤儿目录"), iconTrash, false, p, a.cleanupOrphanDirs)
			})

			if status := strings.TrimSpace(a.status); status != "" {
				myui.Box(c).Radius(10).Background(p.hover).Border(1, p.hairline).Padding(10, 12).Children(func() {
					myui.Row(c).Gap(9).AlignItems(myui.Center).Children(func() {
						myui.Icon(c, iconAbout).Size(15, 15).TextColor(p.accent)
						myui.Text(c, status).FontSize(12.5).TextColor(p.text).SingleLine()
					})
				})
			}

			if count == 0 {
				myui.Text(c, i18n.T("未发现孤儿图集目录，或尚未扫描。")).FontSize(12.5).TextColor(p.muted)
			} else {
				for i, rel := range a.orphanDirs {
					if i >= 30 {
						break
					}
					a.settingsRow(c, "media/gallery/"+rel, "", func() {})
				}
				if count > 30 {
					myui.Text(c, fmt.Sprintf(i18n.T("仅显示前 30 个，共 %d 个。"), count)).FontSize(12).TextColor(p.muted)
				}
			}
		})
	})
}

// renderCategorySettings 用与更新页一致的设置行风格展示分类的新增与管理。
func (a *App) renderCategorySettings(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)

	a.settingsGroup(c, func() {
		a.settingsRow(c, i18n.T("新增分类"), i18n.T("创建用于版型与图集的分类"), func() {
			myui.Row(c).Gap(8).AlignItems(myui.Center).Children(func() {
				input := myui.TextInput(c.Key("category-new-name"), &a.newName).
					Placeholder(i18n.T("新分类名称")).
					Label(i18n.T("新分类名称")).
					Width(200)
				if input.Submitted() {
					a.createCategory()
				}
				a.cardActionButton(c, "category-create", i18n.T("新增分类"), iconCategory, true, p, a.createCategory)
			})
		})
	})

	a.settingsGroup(c, func() {
		a.settingsRow(c, i18n.T("分类列表"), i18n.T("启用或禁用分类，删除不再使用的分类"), func() {
			a.statusPill(c, fmt.Sprintf(i18n.T("共 %d 个分类"), len(a.categories)), p.muted)
		})
		if len(a.categories) == 0 {
			a.settingsRow(c, i18n.T("暂无分类。"), "", func() {})
			return
		}
		for _, category := range a.categories {
			item := category
			label := item.Name
			if !item.Enabled {
				label += i18n.T("（已禁用）")
			}
			a.settingsRow(c, label, fmt.Sprintf(i18n.T("ID %d"), item.ID), func() {
				myui.Row(c).Gap(8).AlignItems(myui.Center).Children(func() {
					on := item.Enabled
					if myui.Switch(c.Key(fmt.Sprintf("category-enabled-%d", item.ID)), &on).
						Label(item.Name).Changed() {
						if err := a.svc.SetCategoryEnabled(context.Background(), item.ID, on); err != nil {
							a.setStatus(err.Error())
						} else {
							a.setStatus("")
							a.Refresh()
						}
					}
					a.cardActionButton(c, fmt.Sprintf("category-delete-%d", item.ID), i18n.T("删除"), iconTrash, false, p, func() {
						if err := a.svc.DeleteCategory(context.Background(), item.ID); err != nil {
							a.setStatus(err.Error())
						} else {
							a.setStatus("")
							a.Refresh()
						}
					})
				})
			})
		}
	})

	if status := strings.TrimSpace(a.status); status != "" {
		myui.Box(c).Radius(10).Background(p.hover).Border(1, p.hairline).Padding(10, 12).Children(func() {
			myui.Row(c).Gap(9).AlignItems(myui.Center).Children(func() {
				myui.Icon(c, iconAbout).Size(15, 15).TextColor(p.accent)
				myui.Text(c, status).FontSize(12.5).TextColor(p.text).SingleLine()
			})
		})
	}
}

// licenseSettings 是授权页的占位框架：仅展示状态与说明，尚未接入任何授权逻辑。
func (a *App) licenseSettings(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)

	myui.Box(c).Radius(14).Background(p.bg).Border(1, p.hairline).Padding(18).Children(func() {
		myui.Column(c).Gap(14).Children(func() {
			myui.Row(c).Gap(12).AlignItems(myui.Center).Children(func() {
				myui.Box(c).Size(42, 42).Radius(12).Center().Background(p.accent.Alpha(0.12)).Children(func() {
					myui.Icon(c, iconLicense).Size(20, 20).TextColor(p.accent)
				})
				myui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(4).Children(func() {
					myui.Text(c, i18n.T("软件授权")).FontSize(17).FontWeight(700).SingleLine()
					myui.Text(c, i18n.T("查看与管理本软件的授权状态")).FontSize(12).TextColor(p.muted).LineHeight(1.35)
				})
				a.statusPill(c, i18n.T("待配置"), p.muted)
			})
			myui.Box(c).Height(1).Background(p.hairline)
			myui.Text(c, i18n.T("授权功能正在建设中，后续将在此提供授权状态与激活管理。")).FontSize(12.5).TextColor(p.muted).LineHeight(1.4)
		})
	})
}
