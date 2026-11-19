package media

import (
	"autocarsystem/internal/i18n"
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type ErrorRecord struct {
	Time    string `json:"time"`
	Action  string `json:"action"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

func AppendError(path string, action string, target string, err error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return i18n.Errorf("创建媒体日志目录失败: %w", err)
	}

	record := ErrorRecord{
		Time:    time.Now().Format(time.RFC3339),
		Action:  action,
		Path:    target,
		Message: err.Error(),
	}
	line, err := json.Marshal(record)
	if err != nil {
		return i18n.Errorf("序列化媒体错误失败: %w", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return i18n.Errorf("打开媒体错误日志失败: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	if _, err := w.Write(append(line, '\n')); err != nil {
		return i18n.Errorf("写入媒体错误日志失败: %w", err)
	}
	return w.Flush()
}
