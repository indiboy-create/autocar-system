package uiapp

import (
	"fmt"
	"time"

	"autocarsystem/internal/i18n"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater"
	"github.com/egoist/mygo/plugins/updater/native"
	myui "github.com/egoist/mygo/ui"
)

// 更新的检查、下载与安装全部交给 MyGo 内置的 updater 插件（原生 UI 更新窗口）。
// 这些接缝指向插件函数；插件未注册时这些函数会 panic，测试需替换为替身。
var (
	checkForUpdates    = updater.CheckForUpdates
	updateChecks       = updater.AutomaticChecks
	setUpdateChecks    = updater.SetAutomaticChecks
	updateDownloads    = updater.AutomaticDownloads
	setUpdateDownloads = updater.SetAutomaticDownloads
	lastUpdateCheck    = updater.LastCheck
)

// pluginLanguage 把应用语言映射为更新窗口使用的语言标签。
func pluginLanguage() string {
	if i18n.Current() == i18n.English {
		return "en"
	}
	return "zh-Hans"
}

// UseUpdater 注册 updater 原生插件：启动约 10 秒后首次检查、之后每天一次，
// 发现新版本时按偏好在独立更新窗口中提示或直接后台安装。必须在 App.Run 之前调用。
func (a *App) UseUpdater() {
	mygo.Use(native.New(updater.Options{Language: pluginLanguage()}))
	a.updates = true
	// A check or an answer in the update window changes what the settings show.
	updater.OnChange(func() {
		if a.settingsWin != nil {
			a.settingsWin.Invalidate()
		}
	})
	a.migrateUpdatePrefs()
}

// migrateUpdatePrefs 把旧 prefs 中的更新偏好一次性迁移到插件状态（updater.json），
// 之后以插件状态为准，不再由 prefs 覆盖用户在更新窗口里的选择。
func (a *App) migrateUpdatePrefs() {
	if a.prefs.UpdatePrefsMigrated {
		return
	}
	setUpdateChecks(a.prefs.AutoCheckUpdates)
	setUpdateDownloads(a.prefs.AutoInstallUpdates)
	a.prefs.UpdatePrefsMigrated = true
	a.savePrefs()
}

// renderUpdateSettings 按 mygo-music 的设置行风格展示更新设置：
// 自动检查、自动下载、当前版本与手动检查依次排列在一个卡片中。
func (a *App) renderUpdateSettings(c *myui.Context) {
	checkDesc := i18n.T("每天检查一次是否有新版本")
	if t := lastUpdateCheck(); !t.IsZero() {
		checkDesc = fmt.Sprintf(i18n.T("每天检查一次是否有新版本。上次检查：%s"), updateCheckAgo(t))
	}

	a.settingsGroup(c, func() {
		a.settingsRow(c, i18n.T("自动检查更新"), checkDesc, func() {
			on := updateChecks()
			if myui.Switch(c.Key("update-checks"), &on).Label(i18n.T("自动检查更新")).Changed() {
				setUpdateChecks(on)
				a.notifySettingSaved(i18n.T("自动检查更新"))
			}
		})
		a.settingsRow(c, i18n.T("自动下载并安装"), i18n.T("在后台下载新版本，下次打开时生效"), func() {
			on := updateDownloads()
			if myui.Switch(c.Key("update-downloads"), &on).Label(i18n.T("自动下载并安装")).
				Disabled(!updateChecks()).Changed() {
				setUpdateDownloads(on)
				a.notifySettingSaved(i18n.T("自动下载并安装"))
			}
		})
		a.settingsRow(c, fmt.Sprintf(i18n.T("当前版本 %s"), mygo.App.Version()), "", func() {
			if myui.Button(c.Key("check-updates"), i18n.T("检查更新")).Clicked() {
				checkForUpdates()
			}
		})
	})
}

// updateCheckAgo 把上次检查时间说成相对时间，与 mygo-music 的 Ago 一致。
func updateCheckAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return i18n.T("刚刚")
	case d < 2*time.Minute:
		return i18n.T("一分钟前")
	case d < time.Hour:
		return fmt.Sprintf(i18n.T("%d 分钟前"), int(d/time.Minute))
	case d < 2*time.Hour:
		return i18n.T("一小时前")
	case d < 24*time.Hour:
		return fmt.Sprintf(i18n.T("%d 小时前"), int(d/time.Hour))
	case d < 48*time.Hour:
		return i18n.T("昨天")
	case d < 30*24*time.Hour:
		return fmt.Sprintf(i18n.T("%d 天前"), int(d/(24*time.Hour)))
	}
	return t.Format("2006-01-02")
}
