package conf

import (
	"fmt"
	"strings"
)

const tokenEncryptKeyBytes = 32

// ValidateChronoFlow 校验调度器启动所需的核心配置。
// 这里不区分 dev/prod，避免本地启动成功但部署后才暴露关键安全配置缺失。
func ValidateChronoFlow(c *Bootstrap) error {
	if c == nil {
		return fmt.Errorf("bootstrap config is nil")
	}
	if c.Security == nil {
		return fmt.Errorf("security config is required")
	}
	if len([]byte(c.Security.TokenEncryptKey)) != tokenEncryptKeyBytes {
		return fmt.Errorf("security.token_encrypt_key must be %d bytes", tokenEncryptKeyBytes)
	}
	if strings.TrimSpace(c.Security.CallbackToken) == "" {
		return fmt.Errorf("security.callback_token is required")
	}
	if c.Logs == nil {
		return fmt.Errorf("logs config is required")
	}
	if strings.TrimSpace(c.Logs.DataDir) == "" {
		return fmt.Errorf("logs.data_dir is required")
	}
	if c.Logs.MaxLogBytes <= 0 {
		return fmt.Errorf("logs.max_log_bytes must be greater than 0")
	}
	if c.Logs.RetentionDays <= 0 {
		return fmt.Errorf("logs.retention_days must be greater than 0")
	}
	if c.Executor != nil {
		if c.Executor.HealthCheckIntervalSeconds <= 0 {
			return fmt.Errorf("executor.health_check_interval_seconds must be greater than 0")
		}
		if c.Executor.HealthCheckFailThreshold <= 0 {
			return fmt.Errorf("executor.health_check_fail_threshold must be greater than 0")
		}
		if c.Executor.RequestTimeoutSeconds <= 0 {
			return fmt.Errorf("executor.request_timeout_seconds must be greater than 0")
		}
	}
	if c.Recovery != nil {
		if c.Recovery.StartupRunningGraceSeconds <= 0 {
			return fmt.Errorf("recovery.startup_running_grace_seconds must be greater than 0")
		}
		if c.Recovery.KillingTimeoutSeconds <= 0 {
			return fmt.Errorf("recovery.killing_timeout_seconds must be greater than 0")
		}
	}
	return nil
}

