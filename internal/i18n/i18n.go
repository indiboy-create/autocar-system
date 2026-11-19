package i18n

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/egoist/mygo"
)

type Lang int32

const (
	System Lang = iota
	Chinese
	English
)

var current atomic.Int32

// Set 根据偏好设置选择界面语言；system 时跟随系统首选语言。
func Set(code string) {
	lang := System
	switch strings.ToLower(strings.TrimSpace(code)) {
	case "zh", "zh-cn", "chinese":
		lang = Chinese
	case "en", "en-us", "english":
		lang = English
	case "system", "":
		for _, preferred := range mygo.App.PreferredLanguages() {
			tag := strings.ToLower(preferred)
			switch {
			case strings.HasPrefix(tag, "zh"):
				lang = Chinese
			case strings.HasPrefix(tag, "en"):
				lang = English
			}
			if lang != System {
				break
			}
		}
	}
	current.Store(int32(lang))
}

// Current 返回当前选定的语言。
func Current() Lang { return Lang(current.Load()) }

// T 返回当前语言的文案。英文模式下查字典，未收录的文案保持原文。
func T(s string) string {
	if Lang(current.Load()) != English {
		return s
	}
	if en, ok := dict[s]; ok {
		return en
	}
	return s
}

// Tf 翻译格式串后按参数格式化。占位符顺序须与原文一致。
func Tf(format string, args ...any) string {
	return fmt.Sprintf(T(format), args...)
}

// Errorf 翻译格式串后生成错误，供各层错误文案统一走字典。
func Errorf(format string, args ...any) error {
	return fmt.Errorf(T(format), args...)
}

// N 标记“留存后展示”的文案（如配色名、车型标签），本身按原文返回，
// 供门禁扫描；这些值存入结构体后，展示时由 T 按当前语言翻译。
func N(s string) string { return s }

// Data 标记不面向用户的原文数据（如数据库中的中文取值、匹配关键字）。
// 它按原文返回，无需译文；门禁据此允许这些中文字面量存在。
func Data(s string) string { return s }
