package uiapp

import (
	"context"

	"autocarsystem/internal/domain"
)

// refreshResult 是一次后台刷新的全部结果。字段只在后台 goroutine 写入，
// 经 Window.Update 回投到主线程后才被应用到 App 状态。
type refreshResult struct {
	categories []domain.Category
	yuvisions  []domain.Yuvision
	galleries  []domain.Gallery
	counts     map[string]int64
	brands     []domain.Brand
	firms      []domain.Firm
	series     []domain.Series
	specs      []domain.Spec
	backups    []domain.BackupInfo
	errors     []string
}

// fetchRefresh 在后台执行所有数据库查询，不触碰 App 的可变状态。
// 品牌、厂商、版型、图集一次取全（名称映射需要完整数据，图集总数也要与列表
// 一致）；车系与年款可达万级/十万级，按 seriesLimit/specLimit 分批取，由目录
// 页的“加载更多”提高上限。封面不在这里解码，改由图集卡片按需读取并缓存。
func (a *App) fetchRefresh(ctx context.Context, filter domain.GalleryFilter, seriesLimit, specLimit int) refreshResult {
	var r refreshResult

	if v, err := a.svc.ListCategories(ctx); err != nil {
		r.errors = append(r.errors, err.Error())
	} else {
		r.categories = v
	}

	if v, err := a.svc.ListYuvisions(ctx, 0); err != nil {
		r.errors = append(r.errors, err.Error())
	} else {
		r.yuvisions = v
	}

	if v, err := a.svc.SearchGalleries(ctx, filter, 0); err != nil {
		r.errors = append(r.errors, err.Error())
	} else {
		r.galleries = v
	}

	if v, err := a.svc.Counts(ctx); err != nil {
		r.errors = append(r.errors, err.Error())
	} else {
		r.counts = v
	}

	if v, err := a.svc.ListBrands(ctx, 0); err == nil {
		r.brands = v
	}
	if v, err := a.svc.ListFirms(ctx, 0); err == nil {
		r.firms = v
	}
	if v, err := a.svc.ListSeries(ctx, seriesLimit); err == nil {
		r.series = v
	}
	if v, err := a.svc.ListSpecs(ctx, specLimit); err == nil {
		r.specs = v
	}
	if v, err := a.svc.ListBackups(); err == nil {
		r.backups = v
	}

	return r
}

// applyRefresh 在主线程应用后台结果。只有查询成功（对应字段非 nil）时才覆盖
// 现有数据，保持与同步版本一致：失败只上报状态，不清空旧数据。
func (a *App) applyRefresh(r refreshResult, seq uint64, async bool) {
	for _, message := range r.errors {
		a.setStatus(message)
	}
	if r.categories != nil {
		a.categories = r.categories
	}
	if r.yuvisions != nil {
		a.yuvisions = r.yuvisions
	}
	if r.galleries != nil {
		a.galleries = r.galleries
	}
	if r.counts != nil {
		a.counts = r.counts
	}
	if r.brands != nil {
		a.brands = r.brands
	}
	if r.firms != nil {
		a.firms = r.firms
	}
	if r.series != nil {
		a.series = r.series
	}
	if r.specs != nil {
		a.specs = r.specs
	}
	if r.backups != nil {
		a.backups = r.backups
	}
	if async {
		a.invalidateWindows()
	}
}
