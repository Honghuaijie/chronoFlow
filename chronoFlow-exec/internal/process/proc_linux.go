//go:build linux

package process

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func ProcessGroupExists(pgid int) bool {
	if pgid <= 0 {
		return false
	}
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 {
			continue
		}
		info, err := readProcStatInfo(pid)
		if err != nil {
			continue
		}
		if info.Zombie {
			continue
		}
		if info.ProcessGroupID == pgid {
			return true
		}
	}
	return false
}

func WaitProcessGroupGone(ctx context.Context, pgid int, interval time.Duration) error {
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if !ProcessGroupExists(pgid) {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("process group %d still exists: %w", pgid, ctx.Err())
		case <-ticker.C:
		}
	}
}

func readProcProcessGroupID(pid int) (int, error) {
	info, err := readProcStatInfo(pid)
	if err != nil {
		return 0, err
	}
	return info.ProcessGroupID, nil
}

func readProcStatInfo(pid int) (procStatInfo, error) {
	b, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
	if err != nil {
		return procStatInfo{}, err
	}
	return parseProcStat(string(b))
}
