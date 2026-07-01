package data

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestChronoFlowModelsTableNames(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "executor", got: Executor{}.TableName(), want: "executors"},
		{name: "job", got: Job{}.TableName(), want: "jobs"},
		{name: "job_glue", got: JobGlue{}.TableName(), want: "job_glues"},
		{name: "job_log", got: JobLog{}.TableName(), want: "job_logs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("TableName() = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestChronoFlowAlertColumnsMigrate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&Job{}, &JobLog{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	assertColumn := func(model any, name string) {
		t.Helper()
		if !db.Migrator().HasColumn(model, name) {
			t.Fatalf("expected column %s", name)
		}
	}
	assertColumn(&Job{}, "failure_alert_enabled")
	assertColumn(&JobLog{}, "alert_enabled_snapshot")
	assertColumn(&JobLog{}, "alert_status")
	assertColumn(&JobLog{}, "alert_error")
	assertColumn(&JobLog{}, "alert_sent_at")
}
