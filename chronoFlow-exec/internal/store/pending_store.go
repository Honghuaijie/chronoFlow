package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"chronoFlow-exec/internal/conf"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewPendingStoreFromConf)

type CallbackItem struct {
	LogID         int64     `json:"log_id"`
	JobID         int64     `json:"job_id"`
	CallbackURL   string    `json:"callback_url"`
	CallbackToken string    `json:"callback_token"`
	Status        string    `json:"status"`
	ExitCode      int32     `json:"exit_code"`
	LogContent    string    `json:"log_content"`
	LogTruncated  bool      `json:"log_truncated"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	DurationMS    int64     `json:"duration_ms"`
	ErrorMessage  string    `json:"error_message"`
	CreatedAt     time.Time `json:"created_at"`
}

type PendingStore struct {
	dataDir string
}

func NewPendingStore(dataDir string) *PendingStore {
	return &PendingStore{dataDir: dataDir}
}

func NewPendingStoreFromConf(c *conf.Executor) *PendingStore {
	dataDir := "./data"
	if c != nil && c.DataDir != "" {
		dataDir = c.DataDir
	}
	return NewPendingStore(dataDir)
}

func (s *PendingStore) Save(item *CallbackItem) error {
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}
	if err := os.MkdirAll(s.pendingDir(), 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.pendingPath(item.LogID), payload, 0o644)
}

func (s *PendingStore) ListPending() ([]*CallbackItem, error) {
	entries, err := os.ReadDir(s.pendingDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	items := make([]*CallbackItem, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		payload, err := os.ReadFile(filepath.Join(s.pendingDir(), entry.Name()))
		if err != nil {
			return nil, err
		}
		var item CallbackItem
		if err := json.Unmarshal(payload, &item); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}

func (s *PendingStore) DeletePending(logID int64) error {
	err := os.Remove(s.pendingPath(logID))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *PendingStore) pendingDir() string {
	return filepath.Join(s.dataDir, "callbacks", "pending")
}

func (s *PendingStore) pendingPath(logID int64) string {
	return filepath.Join(s.pendingDir(), fmt.Sprintf("log-%d.json", logID))
}
