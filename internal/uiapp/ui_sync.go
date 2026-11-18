package uiapp

import (
	"autocarsystem/internal/i18n"
	autosync "autocarsystem/internal/sync"
	"context"
	"fmt"
	myui "github.com/egoist/mygo/ui"
	"strings"
)

func (a *App) renderSync(c *myui.Context) {
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

	myui.Row(c).Gap(8).Children(func() {
		if myui.Button(c, i18n.T("开始同步")).Disabled(a.syncRunning).Clicked() {
			go a.runSync()
		}
		if myui.Button(c, i18n.T("刷新状态")).Clicked() {
			a.Refresh()
		}
	})
	myui.Text(c, fmt.Sprintf(i18n.T("同步模式：%s，当前阶段：%s"), mode, stage))
	myui.Text(c, fmt.Sprintf(i18n.T("上次同步：%s"), state.LastSyncAt))
	if strings.TrimSpace(a.syncMessage) != "" {
		myui.Text(c, i18n.T("同步进度：")+a.syncMessage)
	}
	for _, entity := range []string{"brand", "firm", "series", "spec"} {
		stat := state.Stats[entity]
		myui.Text(c, fmt.Sprintf(i18n.T("%s：处理 %d，新增 %d，跳过 %d，失败 %d"),
			entity, stat.Processed, stat.Inserted, stat.Skipped, stat.Failed))
	}
	myui.Text(c, i18n.T("同步采用全量遍历、只新增、不更新、不删除。"))
	myui.Text(c, i18n.T("原始响应： ")+a.paths.RawDir)
	myui.Text(c, i18n.T("冲突日志： ")+a.paths.Conflicts)
}

func (a *App) runSync() {
	if !a.syncGate.CompareAndSwap(false, true) {
		return
	}
	defer a.syncGate.Store(false)
	a.updateSyncState(true, i18n.T("准备同步"))
	err := autosync.Run(context.Background(), a.svc, a.paths, func(progress autosync.Progress) {
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
