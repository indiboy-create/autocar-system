package i18n

import (
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
