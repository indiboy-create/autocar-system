package uiapp

import (
	"strings"

	"autocarsystem/internal/i18n"
	myui "github.com/egoist/mygo/ui"
)

// settingsSearchEntry 描述一个可被设置页搜索命中的设置项。
// page 是跳转目标，title 是展示文案，keywords 是额外的搜索关键词（如别名、英文）。
type settingsSearchEntry struct {
	page     string
	title    string
	keywords string
}

// settingsSearchIndex 是所有可搜索设置项的集中索引。新增设置项时在这里登记，
// 即可自动出现在设置页搜索结果中，无需改动搜索与绘制逻辑。
func settingsSearchIndex() []settingsSearchEntry {
	// keywords 是不展示的搜索别名，用 Data 标记：英文模式下仍保留中文别名，
	// 以便用中文关键字也能检索到设置项。
	return []settingsSearchEntry{
		// 通用
		{"general", i18n.T("界面语言"), i18n.Data("language 语言 english 中文")},
		{"general", i18n.T("主题"), i18n.Data("theme 外观 深色 浅色 dark light system 跟随系统")},
		{"general", i18n.T("配色方案"), i18n.Data("color scheme 颜色 accent 强调色")},
		// 分类
		{"category", i18n.T("分类"), i18n.Data("category 新增分类 分类列表 启用 禁用")},
		// 同步
		{"sync", i18n.T("同步"), i18n.Data("sync 同步年款 同步模式 全量 抓取")},
		{"sync", i18n.T("同步年款"), i18n.Data("spec year 年款 车型")},
		// 备份
		{"backup", i18n.T("立即备份"), i18n.Data("backup 备份 database 数据库")},
		{"backup", i18n.T("定时备份"), i18n.Data("auto backup 自动备份 定时")},
		{"backup", i18n.T("备份周期"), i18n.Data("interval 周期 小时 hours")},
		// 系统
		{"system", i18n.T("根目录"), i18n.Data("root path 路径 目录")},
		{"system", i18n.T("数据目录"), i18n.Data("data path 路径 目录")},
		{"system", i18n.T("媒体目录"), i18n.Data("media path 路径 目录")},
		{"system", i18n.T("数据库"), i18n.Data("database sqlite db 路径")},
		{"system", i18n.T("配置文件"), i18n.Data("prefs config 偏好 配置 路径")},
		{"system", i18n.T("孤儿目录"), i18n.Data("orphan 目录 扫描 清理")},
		// 更新
		{"updates", i18n.T("自动检查更新"), i18n.Data("update check 检查 更新")},
		{"updates", i18n.T("自动下载并安装"), i18n.Data("update download install 下载 安装 更新")},
		// 授权
		{"license", i18n.T("授权"), i18n.Data("license 许可 授权")},
		// 关于
		{"about", i18n.T("应用名称"), i18n.Data("about 关于 版本 version")},
		{"about", i18n.T("版本"), i18n.Data("version 版本 about 关于")},
	}

}

// settingsSearchMatches 判断搜索词是否命中设置项。标题与关键词均转小写后匹配。
func settingsSearchMatches(e settingsSearchEntry, query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return true
	}
	return strings.Contains(strings.ToLower(e.title), q) ||
		strings.Contains(strings.ToLower(e.keywords), q)
}

// settingsSearchResults 返回按页面分组的命中项，保持索引顺序。
func settingsSearchResults(query string) []settingsSearchEntry {
	var out []settingsSearchEntry
	for _, e := range settingsSearchIndex() {
		if settingsSearchMatches(e, query) {
			out = append(out, e)
		}
	}
	return out
}

// settingsPageMatches 判断某个页面在搜索中是否有命中项，用于过滤侧栏导航。
func settingsPageMatches(page, query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return true
	}
	for _, e := range settingsSearchIndex() {
		if e.page == page && settingsSearchMatches(e, q) {
			return true
		}
	}
	// 页面标题本身命中时，展示整页。
	return strings.Contains(strings.ToLower(settingsPageTitle(page)), q)
}

// settingsPageTitle 返回页面的本地化标题。
func settingsPageTitle(page string) string {
	switch page {
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

// settingsSearch 绘制侧栏顶部的搜索框，聚焦时高亮，输入时实时过滤。
func (a *App) settingsSearch(c *myui.Context, p palette) {
	field := myui.Row(c).Height(32).Radius(8).AlignItems(myui.Stretch).Margin(0, 0, 10, 0)
	if a.settingsSearchBox.Focused(c) {
		field.Background(p.bg).Border(1, p.accent)
	} else {
		field.Background(p.hover).Border(1, p.hairline)
	}
	field.Children(func() {
		myui.TextInputBase(c.Key("settings-search"), &a.settingsQuery).
			Grow(1).MinWidth(0).PaddingStart(28).PaddingEnd(24).
			Placeholder(i18n.T("搜索设置")).Label(i18n.T("搜索设置")).
			Bind(&a.settingsSearchBox)
		myui.Icon(c, iconSearch).Size(13, 13).TextColor(p.muted).Absolute().Left(8).Top(9).IgnoreInput(true)
		if strings.TrimSpace(a.settingsQuery) != "" {
			clear := myui.ButtonBase(c.Key("settings-search-clear")).
				Size(22, 22).Radius(6).Center().FocusRing(false).Label(i18n.T("清除搜索"))
			clear.Children(func() {
				myui.Icon(c, iconClose).Size(11, 11).TextColor(p.muted)
			})
			if clear.Clicked() {
				a.settingsQuery = ""
			}
		}
	})
}

// settingsSearchResultsView 绘制跨页面的搜索结果，按页面分组，点击跳转到对应设置页。
func (a *App) settingsSearchResultsView(c *myui.Context, p palette, query string) {
	results := settingsSearchResults(query)
	if len(results) == 0 {
		myui.Column(c).Gap(8).AlignItems(myui.Center).Padding(48, 0, 0, 0).Children(func() {
			myui.Icon(c, iconSearch).Size(28, 28).TextColor(p.muted)
			myui.Text(c, i18n.T("未找到相关设置")).FontSize(13).TextColor(p.muted)
		})
		return
	}

	lastPage := ""
	for _, item := range results {
		if item.page != lastPage {
			myui.Text(c, settingsPageTitle(item.page)).FontSize(12).FontWeight(600).
				TextColor(p.muted).Margin(12, 4, 4, 4)
			lastPage = item.page
		}
		entry := item
		b := myui.ButtonBase(c.Key("search-result-" + entry.page + "-" + entry.title)).
			Height(46).PaddingX(16).Gap(12).Radius(10).AlignItems(myui.Center).
			Justify(myui.Start).FocusRing(false).Label(entry.title)
		if b.Hovered() {
			b.Background(p.hover)
		} else {
			b.Background(p.bg)
		}
		b.Children(func() {
			myui.Icon(c, iconSettings).Size(15, 15).TextColor(p.accent)
			myui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(2).Children(func() {
				myui.Text(c, entry.title).FontSize(13.5).FontWeight(500).TextColor(p.text).SingleLine()
				myui.Text(c, settingsPageTitle(entry.page)).FontSize(11).TextColor(p.muted)
			})
			myui.Icon(c, iconForward).Size(14, 14).TextColor(p.muted)
		})
		if b.Clicked() {
			a.settingsPage = entry.page
			a.settingsQuery = ""
			// 更新页在未注册插件时不可达，回退到通用。
			if entry.page == "updates" && !a.updates {
				a.settingsPage = "general"
			}
		}
	}
}
