package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	"autocarsystem/internal/config"
	"autocarsystem/internal/i18n"
	"autocarsystem/internal/service"
	"autocarsystem/internal/store"
	"autocarsystem/internal/uiapp"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

//go:embed resources/tray.png
var trayIcon []byte

func main() {
	if !mygo.App.RequestSingleInstanceLock() {
		return
	}

	paths, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if err := paths.EnsureDirs(); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	st, err := store.OpenAndMigrate(ctx, paths.DBPath, paths.BackupDir)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	svc := service.New(st, paths)
	if err := svc.Bootstrap(ctx); err != nil {
		log.Fatal(err)
	}

	app := uiapp.New(svc, paths)

	mygo.App.WhenReady(func() {
		win := mygo.NewWindow(mygo.WindowOptions{
			Title:                "autoCarSystem",
			Width:                1120,
			Height:               760,
			MinWidth:             880,
			MinHeight:            600,
			StateKey:             "autocarsystem-main",
			TitleBarStyle:        mygo.TitleBarHidden,
			TitleBarHeight:       46,
			TrafficLightPosition: &mygo.Point{X: 20, Y: 20},
			AutoHideMenuBar:      true,
			Vibrancy:             mygo.VibrancySidebar,
			Content:              ui.View(app.View),
		})
		app.AttachWindow(win)
		win.SetWindowButtonsTheme(mygo.ThemeSystem)

		showMain := func() {
			if !win.IsVisible() {
				win.Show()
			}
			win.Restore()
			win.Focus()
		}

		mygo.App.OnBeforeQuit(func(*mygo.QuitEvent) {
			app.SetQuitting(true)
		})
		win.OnClose(func(e *mygo.CloseEvent) {
			if !app.Quitting() {
				e.PreventDefault()
				win.Hide()
			}
		})
		// 切回窗口时重新读取数据，让在其它程序里改动的文件/数据库反映到界面。
		win.OnFocus(app.Refresh)

		tray, err := mygo.NewTray(mygo.TrayOptions{
			Icon:    trayIcon,
			ToolTip: "autoCarSystem",
			Menu: mygo.NewMenu([]*mygo.MenuItem{
				{
					Label: i18n.T("显示 / 隐藏主窗口"),
					Click: func(*mygo.MenuItem, *mygo.Window) {
						if win.IsVisible() {
							win.Hide()
							return
						}
						win.Show()
						win.Focus()
					},
				},
				mygo.Separator(),
				{Role: mygo.RoleQuit},
			}),
		})
		if err != nil {
			log.Printf("创建托盘失败: %v", err)
		} else {
			app.AttachTray(tray)
			tray.OnClick(showMain)
			// Windows 退出后通知区会残留图标直到鼠标划过，主动销毁。
			mygo.App.OnWillQuit(func(*mygo.QuitEvent) { tray.Destroy() })
		}

		mygo.App.OnActivate(func(bool) { showMain() })
		mygo.App.OnSecondInstance(func([]string, string) { showMain() })
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(fmt.Errorf("应用运行失败: %w", err))
	}
}
