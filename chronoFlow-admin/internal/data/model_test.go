package data

import "testing"

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
