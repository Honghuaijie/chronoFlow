package conf

import (
	"fmt"
	"strings"
)

func ValidateExec(c *Bootstrap) error {
	if c == nil {
		return fmt.Errorf("bootstrap config is nil")
	}
	if c.Executor == nil {
		return fmt.Errorf("executor config is required")
	}
	if strings.TrimSpace(c.Executor.Token) == "" {
		return fmt.Errorf("executor.token is required")
	}
	if strings.TrimSpace(c.Executor.DataDir) == "" {
		return fmt.Errorf("executor.data_dir is required")
	}
	if strings.TrimSpace(c.Executor.TempDir) == "" {
		return fmt.Errorf("executor.temp_dir is required")
	}
	if strings.TrimSpace(c.Executor.ShellPath) == "" {
		return fmt.Errorf("executor.shell_path is required")
	}
	if c.Executor.MaxLogBytes <= 0 {
		return fmt.Errorf("executor.max_log_bytes must be greater than 0")
	}
	return nil
}
