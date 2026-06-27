//go:build !linux

package process

import (
	"context"
	"time"
)

func ProcessGroupExists(_ int) bool {
	return false
}

func WaitProcessGroupGone(_ context.Context, _ int, _ time.Duration) error {
	return nil
}
