package sync

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"

	"autocarsystem/internal/fileutil"
	"autocarsystem/internal/i18n"
)

type StageStats struct {
	Processed int64 `json:"processed"`
	Inserted  int64 `json:"inserted"`
	Skipped   int64 `json:"skipped"`
	Failed    int64 `json:"failed"`
}

type State struct {
	LastSyncAt string                `json:"last_sync_at"`
	Stage      string                `json:"stage"`
	Mode       string                `json:"mode"`
	Stats      map[string]StageStats `json:"stats"`
}

type Conflict struct {
	Entity    string `json:"entity"`
	ID        string `json:"id"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

func NewState() State {
	return State{
		Stage: "idle",
		Mode:  "full_scan",
		Stats: map[string]StageStats{
			"brand":  {},
			"firm":   {},
			"series": {},
			"spec":   {},
		},
	}
}

func LoadState(path string) (State, error) {
	state := NewState()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}
		return state, i18n.Errorf("读取同步状态失败: %w", err)
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return state, i18n.Errorf("解析同步状态失败: %w", err)
	}
	if state.Stats == nil {
		state.Stats = NewState().Stats
	}
	if state.Mode == "" {
		state.Mode = "full_scan"
	}
	return state, nil
}

func SaveState(path string, state State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return i18n.Errorf("创建同步状态目录失败: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return i18n.Errorf("序列化同步状态失败: %w", err)
	}
	if err := writeFileAtomic(path, data, 0o644); err != nil {
		return i18n.Errorf("写入同步状态失败: %w", err)
	}
	return nil
}

// writeFileAtomic 是 fileutil.WriteFileAtomic 的薄封装，保留本包既有调用与测试入口。
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	return fileutil.WriteFileAtomic(path, data, perm)
}

func AppendConflict(path string, conflict Conflict) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return i18n.Errorf("创建冲突日志目录失败: %w", err)
	}
	line, err := json.Marshal(conflict)
	if err != nil {
		return i18n.Errorf("序列化冲突日志失败: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return i18n.Errorf("打开冲突日志失败: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	if _, err := w.Write(append(line, '\n')); err != nil {
		return i18n.Errorf("写入冲突日志失败: %w", err)
	}
	if err := w.Flush(); err != nil {
		return i18n.Errorf("刷新冲突日志失败: %w", err)
	}
	return nil
}
