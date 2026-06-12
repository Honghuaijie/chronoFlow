package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"chronoFlow-exec/internal/conf"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewManagerFromConf)

const (
	StatusSuccess = "success"
	StatusFailed  = "failed"
	StatusTimeout = "timeout"
	StatusKilled  = "killed"
)

type Config struct {
	ShellPath        string
	TempDir          string
	MaxLogBytes      int64
	KillGraceSeconds int32
}

type RunRequest struct {
	JobID          int64
	LogID          int64
	Script         string
	TimeoutSeconds int32
	CallbackURL    string
	CallbackToken  string
}

type Result struct {
	JobID         int64
	LogID         int64
	Status        string
	ExitCode      int32
	LogContent    string
	LogTruncated  bool
	StartTime     time.Time
	EndTime       time.Time
	DurationMS    int64
	ErrorMessage  string
	CallbackURL   string
	CallbackToken string
}

type RunState struct {
	JobID  int64
	LogID  int64
	cancel context.CancelFunc
	cmd    *exec.Cmd
}

type Manager struct {
	config  Config
	mu      sync.Mutex
	running map[int64]*RunState
}

func NewManager(config Config) *Manager {
	if config.ShellPath == "" {
		config.ShellPath = "/bin/bash"
	}
	if config.MaxLogBytes <= 0 {
		config.MaxLogBytes = 5 * 1024 * 1024
	}
	if config.KillGraceSeconds <= 0 {
		config.KillGraceSeconds = 5
	}
	return &Manager{config: config, running: make(map[int64]*RunState)}
}

func NewManagerFromConf(c *conf.Executor) *Manager {
	cfg := Config{}
	if c != nil {
		cfg.ShellPath = c.ShellPath
		cfg.TempDir = c.TempDir
		cfg.MaxLogBytes = c.MaxLogBytes
		cfg.KillGraceSeconds = c.KillGraceSeconds
	}
	return NewManager(cfg)
}

func (m *Manager) Run(ctx context.Context, req RunRequest, onDone func(*Result)) error {
	m.mu.Lock()
	if _, ok := m.running[req.JobID]; ok {
		m.mu.Unlock()
		return fmt.Errorf("job is running")
	}
	runCtx, cancel := context.WithCancel(ctx)
	state := &RunState{JobID: req.JobID, LogID: req.LogID, cancel: cancel}
	m.running[req.JobID] = state
	m.mu.Unlock()

	go m.run(runCtx, req, state, onDone)
	return nil
}

func (m *Manager) Kill(jobID int64, logID int64) error {
	m.mu.Lock()
	state := m.running[jobID]
	m.mu.Unlock()
	if state == nil || state.LogID != logID {
		return fmt.Errorf("job is not running")
	}
	state.cancel()
	if state.cmd != nil && state.cmd.Process != nil {
		_ = syscall.Kill(-state.cmd.Process.Pid, syscall.SIGTERM)
		time.Sleep(time.Duration(m.config.KillGraceSeconds) * time.Second)
		_ = syscall.Kill(-state.cmd.Process.Pid, syscall.SIGKILL)
	}
	return nil
}

func (m *Manager) IsRunning(jobID int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.running[jobID]
	return ok
}

func (m *Manager) run(ctx context.Context, req RunRequest, state *RunState, onDone func(*Result)) {
	start := time.Now()
	result := &Result{
		JobID:         req.JobID,
		LogID:         req.LogID,
		Status:        StatusFailed,
		StartTime:     start,
		CallbackURL:   req.CallbackURL,
		CallbackToken: req.CallbackToken,
	}
	defer func() {
		result.EndTime = time.Now()
		result.DurationMS = result.EndTime.Sub(result.StartTime).Milliseconds()
		m.mu.Lock()
		delete(m.running, req.JobID)
		m.mu.Unlock()
		onDone(result)
	}()

	if err := os.MkdirAll(m.config.TempDir, 0o755); err != nil {
		result.ErrorMessage = err.Error()
		return
	}
	scriptPath := filepath.Join(m.config.TempDir, fmt.Sprintf("job-%d-log-%d.sh", req.JobID, req.LogID))
	if err := os.WriteFile(scriptPath, []byte(req.Script), 0o700); err != nil {
		result.ErrorMessage = err.Error()
		return
	}
	defer os.Remove(scriptPath)

	timeout := time.Duration(req.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 24 * time.Hour
	}
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, m.config.ShellPath, scriptPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	buf := NewLogBuffer(m.config.MaxLogBytes)
	cmd.Stdout = buf
	cmd.Stderr = buf
	state.cmd = cmd

	err := cmd.Run()
	result.LogContent = buf.String()
	result.LogTruncated = buf.Truncated()
	if cmdCtx.Err() == context.DeadlineExceeded {
		result.Status = StatusTimeout
		result.ErrorMessage = "任务执行超时"
		return
	}
	if ctx.Err() == context.Canceled {
		result.Status = StatusKilled
		result.ErrorMessage = "任务被终止"
		return
	}
	if err != nil {
		result.Status = StatusFailed
		result.ErrorMessage = err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = int32(exitErr.ExitCode())
		} else {
			result.ExitCode = -1
		}
		return
	}
	result.Status = StatusSuccess
	result.ExitCode = 0
}
