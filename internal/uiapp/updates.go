package uiapp

import (
	"context"
	"fmt"

	"autocarsystem/internal/i18n"

	"github.com/egoist/mygo"
)

// 更新能力由构建时的 mygo.json updates 配置决定。这两个接缝便于测试覆盖
// “已配置/未配置”两条路径；默认走 mygo 的真实实现。
var (
	updatesEnabled = func() bool { return mygo.Updater.Enabled() }
	checkForUpdate = func(ctx context.Context) (*mygo.Update, error) { return mygo.Updater.Check(ctx) }
)

// postSettingsUpdate 把更新结果送回设置窗口线程；无窗口（测试）时直接执行。
func (a *App) postSettingsUpdate(fn func()) {
	if a.settingsWin != nil {
		a.settingsWin.Update(fn)
		return
	}
	fn()
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
