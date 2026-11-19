package uiapp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"autocarsystem/internal/i18n"
	autosync "autocarsystem/internal/sync"
	myui "github.com/egoist/mygo/ui"
)

var (
	statusSuccessColor = myui.Oklch(0.62, 0.14, 150)
	statusDangerColor  = myui.Oklch(0.62, 0.19, 25)
)

func (a *App) renderSync(c *myui.Context) {
	p := paletteFor(c.Theme().Dark)
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
	statusLabel, statusColor := a.syncStatus(stage, p)
	progress := strings.TrimSpace(a.syncMessage)
	if progress == "" {
		progress = i18n.T("暂无同步进度")
	}
	modeText := syncModeLabel(mode)
	if !a.prefs.SyncSpecs {
		modeText += i18n.T("（不含年款）")
	}

	myui.Column(c).Grow(1).Gap(12).Children(func() {
		myui.Box(c).Radius(14).Background(p.bg).Border(1, p.hairline).Padding(18).Children(func() {
			myui.Column(c).Gap(14).Children(func() {
				myui.Row(c).Gap(12).AlignItems(myui.Center).Children(func() {
					myui.Box(c).Size(42, 42).Radius(12).Center().Background(p.accent.Alpha(0.12)).Children(func() {
						myui.Icon(c, iconTransfer).Size(20, 20).TextColor(p.accent)
					})
					myui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(4).Children(func() {
						myui.Text(c, i18n.T("同步中心")).FontSize(17).FontWeight(700).SingleLine()
						myui.Text(c, i18n.T("把远程目录数据同步到本地数据库")).FontSize(12).TextColor(p.muted).LineHeight(1.35)
					})
					a.statusPill(c, statusLabel, statusColor)
				})

				myui.Box(c).Height(1).Background(p.hairline)

				myui.Row(c).Gap(8).AlignItems(myui.Stretch).Children(func() {
					a.infoTile(c, i18n.T("同步模式"), modeText, p)
					a.infoTile(c, i18n.T("当前阶段"), syncStageLabel(stage), p)
					a.infoTile(c, i18n.T("上次同步"), syncTimeLabel(state.LastSyncAt), p)
				})

				myui.Box(c).Radius(10).Background(p.hover).Border(1, p.hairline).Padding(10, 12).Children(func() {
					myui.Row(c).Gap(9).AlignItems(myui.Center).Children(func() {
						myui.Icon(c, iconSync).Size(15, 15).TextColor(statusColor)
						myui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(2).Children(func() {
							myui.Text(c, i18n.T("同步进度")).FontSize(10.5).TextColor(p.muted)
							myui.Text(c, progress).FontSize(12.5).TextColor(p.text).SingleLine()
						})
					})
				})

				myui.Row(c).Gap(8).AlignItems(myui.Center).Children(func() {
					start := myui.ButtonBase(c.Key("sync-now")).
						Height(34).PaddingX(14).Radius(9).Center().FocusRing(false).
						Disabled(a.syncRunning).Label(i18n.T("开始同步"))
					startBg := p.accent
					if a.syncRunning {
						startBg = p.press
					} else if start.Hovered() {
						startBg = p.accent.Alpha(0.86)
					}
					start.Background(startBg)
					start.Children(func() {
						myui.Text(c, i18n.T("开始同步")).FontSize(12.5).FontWeight(600).TextColor(p.onAccent)
					})
					if start.Clicked() {
						go a.runSync()
					}

					refresh := myui.ButtonBase(c.Key("sync-refresh")).
						Height(34).PaddingX(14).Radius(9).Center().FocusRing(false).
						Label(i18n.T("刷新状态"))
					refreshBg := p.hover
					if refresh.Hovered() {
						refreshBg = p.press
					}
					refresh.Background(refreshBg).Border(1, p.hairline)
					refresh.Children(func() {
						myui.Text(c, i18n.T("刷新状态")).FontSize(12.5).FontWeight(500).TextColor(p.text)
					})
					if refresh.Clicked() {
						a.Refresh()
					}
				})
			})
		})

		a.settingsGroup(c, func() {
			a.settingsRow(c, i18n.T("同步年款"), i18n.T("默认不同步年款，开启后同步车系时一并抓取年款数据。"), func() {
				on := a.prefs.SyncSpecs
				if myui.Switch(c.Key("sync-specs"), &on).Label(i18n.T("同步年款")).Changed() {
					a.prefs.SyncSpecs = on
					a.savePrefsWithToast(i18n.T("同步年款"))
				}
			})
		})

		myui.Text(c, i18n.T("同步统计")).FontSize(14).FontWeight(700).TextColor(p.text)
		myui.Row(c).Gap(10).AlignItems(myui.Stretch).Children(func() {
			myui.Column(c).Grow(1).Basis(0).Gap(10).Children(func() {
				a.syncStatCard(c, syncStatItem{"brand", i18n.T("品牌"), iconCatalog}, state.Stats["brand"], p)
				a.syncStatCard(c, syncStatItem{"series", i18n.T("车系"), iconLayers}, state.Stats["series"], p)
			})
			myui.Column(c).Grow(1).Basis(0).Gap(10).Children(func() {
				a.syncStatCard(c, syncStatItem{"firm", i18n.T("厂商"), iconCategory}, state.Stats["firm"], p)
				specLabel := i18n.T("年款")
				if !a.prefs.SyncSpecs {
					specLabel += i18n.T("（未启用）")
				}
				a.syncStatCard(c, syncStatItem{"spec", specLabel, iconSettings}, state.Stats["spec"], p)
			})
		})

		a.settingsGroup(c, func() {
			a.settingsRow(c, i18n.T("同步模式"), modeText, func() {})
			a.settingsRow(c, i18n.T("当前阶段"), syncStageLabel(stage), func() {})
			a.settingsRow(c, i18n.T("上次同步"), syncTimeLabel(state.LastSyncAt), func() {})
		})

		a.settingsGroup(c, func() {
			a.settingsRow(c, i18n.T("同步说明"), i18n.T("同步采用全量遍历、只新增、不更新、不删除。"), func() {})
			a.settingsRow(c, i18n.T("原始响应"), a.paths.RawDir, func() {})
			a.settingsRow(c, i18n.T("冲突日志"), a.paths.Conflicts, func() {})
		})
	})
}

type syncStatItem struct {
	key   string
	label string
	icon  *myui.SVG
}

func (a *App) syncStatus(stage string, p palette) (string, myui.Color) {
	if a.syncRunning {
		return i18n.T("同步中"), p.accent
	}
	message := strings.TrimSpace(a.syncMessage)
	if stage == "failed" || strings.HasPrefix(message, i18n.T("同步失败：")) {
		return i18n.T("失败"), statusDangerColor
	}
	if stage == "done" || message == i18n.T("同步完成") {
		return i18n.T("已完成"), statusSuccessColor
	}
	return i18n.T("空闲"), p.muted
}

func (a *App) statusPill(c *myui.Context, label string, color myui.Color) {
	myui.Box(c).Padding(4, 10).Radius(999).Background(color.Alpha(0.14)).Children(func() {
		myui.Text(c, label).FontSize(11.5).FontWeight(600).TextColor(color)
	})
}

func (a *App) infoTile(c *myui.Context, label, value string, p palette) {
	myui.Box(c).Grow(1).Basis(0).Radius(10).Background(p.hover).Border(1, p.hairline).Padding(10, 12).Children(func() {
		myui.Column(c).Gap(4).Children(func() {
			myui.Text(c, label).FontSize(10.5).TextColor(p.muted)
			myui.Text(c, value).FontSize(12.5).FontWeight(600).TextColor(p.text).SingleLine()
		})
	})
}

func (a *App) syncStatCard(c *myui.Context, item syncStatItem, value autosync.StageStats, p palette) {
	failColor := p.muted
	if value.Failed > 0 {
		failColor = statusDangerColor
	}
	myui.Box(c.Key("sync-stat-"+item.key)).Height(112).Radius(12).Background(p.bg).
		Border(1, p.hairline).Padding(14).Children(func() {
		myui.Column(c).Gap(10).Children(func() {
			myui.Row(c).Gap(9).AlignItems(myui.Center).Children(func() {
				myui.Box(c).Size(26, 26).Radius(8).Center().Background(p.accent.Alpha(0.12)).Children(func() {
					myui.Icon(c, item.icon).Size(14, 14).TextColor(p.accent)
				})
				myui.Column(c).Grow(1).Basis(0).MinWidth(0).Children(func() {
					myui.Text(c, item.label).FontSize(12.5).FontWeight(600).SingleLine()
				})
				myui.Text(c, fmt.Sprintf("%d", value.Processed)).FontSize(17).FontWeight(700).TextColor(p.text)
			})
			myui.Box(c).Height(1).Background(p.hairline)
			myui.Row(c).Gap(8).AlignItems(myui.Stretch).Children(func() {
				metrics := []struct {
					label string
					value int64
					color myui.Color
				}{
					{i18n.T("处理"), value.Processed, p.text},
					{i18n.T("新增"), value.Inserted, p.accent},
					{i18n.T("跳过"), value.Skipped, p.muted},
					{i18n.T("失败"), value.Failed, failColor},
				}
				for _, metric := range metrics {
					current := metric
					myui.Column(c).Grow(1).Basis(0).Gap(3).Children(func() {
						myui.Text(c, current.label).FontSize(10.5).TextColor(p.muted)
						myui.Text(c, fmt.Sprintf("%d", current.value)).FontSize(13.5).FontWeight(600).TextColor(current.color)
					})
				}
			})
		})
	})
}

func syncModeLabel(mode string) string {
	if mode == "full_scan" {
		return i18n.T("全量扫描")
	}
	return mode
}

func syncStageLabel(stage string) string {
	switch stage {
	case "", "idle":
		return i18n.T("空闲")
	case "done":
		return i18n.T("已完成")
	case "failed":
		return i18n.T("失败")
	case "brand":
		return i18n.T("品牌")
	case "firm":
		return i18n.T("厂商")
	case "series":
		return i18n.T("车系")
	case "spec":
		return i18n.T("年款")
	default:
		return stage
	}
}

func syncTimeLabel(value string) string {
	if strings.TrimSpace(value) == "" {
		return i18n.T("尚未同步")
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.Local().Format("2006-01-02 15:04:05")
	}
	return value
}

func (a *App) runSync() {
	if !a.syncGate.CompareAndSwap(false, true) {
		return
	}
	defer a.syncGate.Store(false)
	a.updateSyncState(true, i18n.T("准备同步"))
	cfg := autosync.DefaultConfig()
	cfg.SyncSpecs = a.prefs.SyncSpecs
	err := autosync.RunWithConfig(context.Background(), a.svc, a.paths, cfg, func(progress autosync.Progress) {
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
