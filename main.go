package main

import (
	"context"
	"embed"
	"log"

	"autocarsystem/internal/config"
	"autocarsystem/internal/i18n"
	"autocarsystem/internal/media"
	"autocarsystem/internal/service"
	"autocarsystem/internal/store"
	"autocarsystem/internal/uiapp"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

//go:embed resources/tray.png
var trayIcon []byte

//go:embed resources/data/car.db
var seedDB []byte

//go:embed all:resources/media
var seedMedia embed.FS

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
	// 预置媒体随程序编译发布：只补建缺失的目录/文件，已有内容不删除、不覆盖。
	if err := media.EnsureSeeded(paths.Media, seedMedia); err != nil {
		log.Fatal(err)
	}
	// 预置库随程序编译发布：已有库直接沿用、绝不覆盖；缺失时才释放一份。
	if err := store.EnsureSeeded(paths.DBPath, seedDB); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	st, err := store.Open(ctx, paths.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	svc := service.New(st, paths)
	if err := svc.Bootstrap(ctx); err != nil {
		log.Fatal(err)
	}

	app := uiapp.New(svc, paths)

	// 必须在 App.Run 之前注册更新插件；自动检查由插件自行调度。
	app.UseUpdater()

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

		app.StartAutoBackup(ctx)

		showMain := func() {
			if !win.IsVisible() {
				win.Show()
			}
			win.Restore()
			win.Focus()
		}

		mygo.App.OnBeforeQuit(func(*mygo.QuitEvent) {
			app.SetQuitting(true)
			cancel()
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
					Label: i18n.T("显示软件"),
					Click: func(*mygo.MenuItem, *mygo.Window) { showMain() },
				},
				{
					Label: i18n.T("系统设置"),
					Click: func(*mygo.MenuItem, *mygo.Window) { app.OpenSettingsPage("system") },
				},
				{
					Label: i18n.T("检查更新"),
					Click: func(*mygo.MenuItem, *mygo.Window) { app.TrayCheckUpdates() },
				},
				{
					Label: i18n.T("关于我们"),
					Click: func(*mygo.MenuItem, *mygo.Window) { app.OpenSettingsPage("about") },
				},
				mygo.Separator(),
				{
					Label: i18n.T("退出软件"),
					Click: func(*mygo.MenuItem, *mygo.Window) { mygo.App.Quit() },
				},
			}),
		})
		if err != nil {
			log.Print(i18n.Tf("创建托盘失败: %v", err))
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
		log.Fatal(i18n.Errorf("应用运行失败: %w", err))
	}
}
