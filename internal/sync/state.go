package sync

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
		return state, fmt.Errorf("读取同步状态失败: %w", err)
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return state, fmt.Errorf("解析同步状态失败: %w", err)
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
		return fmt.Errorf("创建同步状态目录失败: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化同步状态失败: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("写入同步状态失败: %w", err)
	}
	return nil
}

func AppendConflict(path string, conflict Conflict) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建冲突日志目录失败: %w", err)
	}
	line, err := json.Marshal(conflict)
	if err != nil {
		return fmt.Errorf("序列化冲突日志失败: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("打开冲突日志失败: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	if _, err := w.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("写入冲突日志失败: %w", err)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("刷新冲突日志失败: %w", err)
	}
	return nil
}
