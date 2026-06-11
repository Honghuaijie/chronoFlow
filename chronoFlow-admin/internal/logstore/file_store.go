package logstore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"chronoFlow-admin/internal/biz"
	"chronoFlow-admin/internal/conf"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewFileStoreFromConf,
	wire.Bind(new(biz.LogReader), new(*FileStore)),
	wire.Bind(new(biz.LogWriter), new(*FileStore)),
)

type FileStore struct {
	dataDir string
}

func NewFileStoreFromConf(c *conf.Logs) *FileStore {
	return NewFileStore(c.DataDir)
}

func NewFileStore(dataDir string) *FileStore {
	return &FileStore{dataDir: filepath.Clean(dataDir)}
}

func (s *FileStore) Write(_ context.Context, logID int64, jobID int64, content string) (string, int64, error) {
	now := time.Now()
	relPath := filepath.ToSlash(filepath.Join(
		"logs",
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", int(now.Month())),
		fmt.Sprintf("%02d", now.Day()),
		fmt.Sprintf("job-%d", jobID),
		fmt.Sprintf("log-%d.log", logID),
	))
	fullPath, err := s.fullPath(relPath)
	if err != nil {
		return "", 0, err
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", 0, err
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		return "", 0, err
	}
	return relPath, int64(len(content)), nil
}

func (s *FileStore) Read(_ context.Context, relPath string) (string, error) {
	fullPath, err := s.fullPath(relPath)
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (s *FileStore) Delete(_ context.Context, relPath string) error {
	fullPath, err := s.fullPath(relPath)
	if err != nil {
		return err
	}
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *FileStore) fullPath(relPath string) (string, error) {
	if relPath == "" {
		return "", fmt.Errorf("log path is empty")
	}
	cleanRel := filepath.Clean(filepath.FromSlash(relPath))
	if filepath.IsAbs(cleanRel) || cleanRel == "." || strings.HasPrefix(cleanRel, ".."+string(filepath.Separator)) || cleanRel == ".." {
		return "", fmt.Errorf("invalid log path")
	}
	fullPath := filepath.Join(s.dataDir, cleanRel)
	cleanRoot := s.dataDir + string(filepath.Separator)
	if fullPath != s.dataDir && !strings.HasPrefix(fullPath, cleanRoot) {
		return "", fmt.Errorf("invalid log path")
	}
	return fullPath, nil
}
