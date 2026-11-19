package uiapp

import (
	"context"
	"fmt"
	"time"

	"autocarsystem/internal/i18n"

	"github.com/egoist/mygo"
)

// 更新能力由构建时的 mygo.json updates 配置决定。这两个接缝便于测试覆盖
// “已配置/未配置”两条路径；默认走 mygo 的真实实现。
var (
	updatesEnabled = func() bool { return mygo.Updater.Enabled() }
	checkForUpdate = func(ctx context.Context) (*mygo.Update, error) { return mygo.Updater.Check(ctx) }
	installUpdate  = func(ctx context.Context, u *mygo.Update) error { return u.Install(ctx, nil) }
)

// updateCheckInterval 是自动检查更新的最小间隔。
const updateCheckInterval = 24 * time.Hour

// postSettingsUpdate 把更新结果送回设置窗口线程；无窗口（测试）时直接执行。
func (a *App) postSettingsUpdate(fn func()) {
	if a.settingsWin != nil {
		a.settingsWin.Update(fn)
		return
	}
	fn()
}

// AutoCheckUpdates 在启动时按偏好做一次每日更新检查：最近检查过就跳过，
// 检查成功后把结果显示在更新设置页；开启自动下载时在后台安装新版本。
// 必须在主线程调用（WhenReady 内）。
func (a *App) AutoCheckUpdates() {
	if !a.prefs.AutoCheckUpdates || !updatesEnabled() {
		return
	}
	now := time.Now()
	if last, err := time.Parse(time.RFC3339, a.prefs.LastUpdateCheck); err == nil &&
		now.Sub(last) < updateCheckInterval {
		return
	}
	// 先记录本次尝试：离线或失败时同样按每日一次节流，不在每次启动重试。
	a.prefs.LastUpdateCheck = now.UTC().Format(time.RFC3339)
	a.savePrefs()

	autoInstall := a.prefs.AutoInstallUpdates
	go func() {
		u, err := checkForUpdate(context.Background())
		if err != nil || u == nil {
			return
		}
		if autoInstall {
			if err := installUpdate(context.Background(), u); err != nil {
				a.postSettingsUpdate(func() {
					a.settingsMessage = fmt.Sprintf("%s: %s", i18n.T("检查更新失败"), err.Error())
				})
				return
			}
			a.postSettingsUpdate(func() {
				a.settingsMessage = fmt.Sprintf(i18n.T("新版本 %s 已下载，下次启动时安装"), u.Version)
			})
			return
		}
		a.postSettingsUpdate(func() {
			a.settingsMessage = fmt.Sprintf(i18n.T("发现新版本 %s"), u.Version)
		})
	}()
}

// checkUpdateNow 真正执行一次版本检查：未配置更新时给出明确说明，绝不假装可用。
func (a *App) checkUpdateNow() {
	if !updatesEnabled() {
		a.settingsMessage = i18n.T("当前构建未启用自动更新")
		return
	}
	a.settingsMessage = i18n.T("正在检查更新")
	go func() {
		u, err := checkForUpdate(context.Background())
		a.postSettingsUpdate(func() {
			switch {
			case err != nil:
				a.settingsMessage = fmt.Sprintf("%s: %s", i18n.T("检查更新失败"), err.Error())
			case u == nil:
				a.settingsMessage = i18n.T("已是最新版本")
			default:
				a.settingsMessage = fmt.Sprintf(i18n.T("发现新版本 %s"), u.Version)
			}
		})
	}()
}
