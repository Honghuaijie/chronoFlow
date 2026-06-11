package conf

import "testing"

func TestValidateChronoFlowConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Bootstrap
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Bootstrap{
				Server: &Server{PublicBaseUrl: "http://127.0.0.1:10003"},
				Scheduler: &Scheduler{
					Timezone: "Asia/Shanghai",
				},
				Executor: &Executor{
					HealthCheckIntervalSeconds: 10,
					HealthCheckFailThreshold:   3,
					RequestTimeoutSeconds:       10,
				},
				Security: &Security{
					TokenEncryptKey: "12345678901234567890123456789012",
					CallbackToken:   "callback-token",
				},
				Logs: &Logs{
					DataDir:       "/tmp/chronoflow",
					MaxLogBytes:   5 * 1024 * 1024,
					RetentionDays: 30,
					CleanupCron:   "0 0 3 * * *",
				},
				Recovery: &Recovery{
					StartupRunningGraceSeconds: 120,
					KillingTimeoutSeconds:      60,
				},
			},
			wantErr: false,
		},
		{
			name: "token encrypt key must be 32 bytes",
			cfg: &Bootstrap{
				Security: &Security{
					TokenEncryptKey: "short",
					CallbackToken:   "callback-token",
				},
				Logs: &Logs{DataDir: "/tmp/chronoflow", MaxLogBytes: 1, RetentionDays: 1},
			},
			wantErr: true,
		},
		{
			name: "callback token required",
			cfg: &Bootstrap{
				Security: &Security{
					TokenEncryptKey: "12345678901234567890123456789012",
				},
				Logs: &Logs{DataDir: "/tmp/chronoflow", MaxLogBytes: 1, RetentionDays: 1},
			},
			wantErr: true,
		},
		{
			name: "log data dir required",
			cfg: &Bootstrap{
				Security: &Security{
					TokenEncryptKey: "12345678901234567890123456789012",
					CallbackToken:   "callback-token",
				},
				Logs: &Logs{MaxLogBytes: 1, RetentionDays: 1},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateChronoFlow(tt.cfg)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}

