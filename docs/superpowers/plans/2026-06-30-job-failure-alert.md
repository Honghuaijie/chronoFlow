# Job Failure Alert Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add V1 Feishu webhook failure alerts for ChronoFlow jobs, including encrypted global webhook settings, per-job alert switches, async alert sending, alert status in job logs, frontend settings UI, and documentation.

**Architecture:** Admin remains the only service that connects to MySQL and sends alerts. Exec does not know about alert configuration. Job runs snapshot the per-job alert switch into `job_logs`, and Admin asynchronously sends Feishu cards after a log first reaches `failed` or `timeout`.

**Tech Stack:** Go 1.22, Kratos, gRPC/HTTP proto generation, GORM/MySQL, Vue 3, Ant Design Vue, Docker Compose.

---

## File Map

Backend API and generated files:

- Modify `chronoFlow-admin/api/job/v1/job.proto`: add `failure_alert_enabled`.
- Modify `chronoFlow-admin/api/joblog/v1/job_log.proto`: add alert fields to `JobLogInfo`.
- Create `chronoFlow-admin/api/system/v1/system_settings.proto`: system settings API for Feishu alert webhook.
- Regenerate `chronoFlow-admin/api/all-pb-go/v1/*` and `chronoFlow-admin/openapi.yaml` with `make api`.

Backend data layer:

- Modify `chronoFlow-admin/internal/data/model.go`: add `failure_alert_enabled`, alert fields, and `SystemSetting`.
- Modify `chronoFlow-admin/internal/data/data.go`: migrate `SystemSetting` and bind repo interfaces.
- Modify `chronoFlow-admin/internal/data/job.go`: map `failure_alert_enabled`.
- Modify `chronoFlow-admin/internal/data/job_log.go`: map alert fields and add alert-specific update helpers.
- Create `chronoFlow-admin/internal/data/system_setting.go`: encrypted setting persistence.

Backend biz layer:

- Modify `chronoFlow-admin/internal/biz/job.go`: carry `FailureAlertEnabled`.
- Modify `chronoFlow-admin/internal/biz/job_run.go`: snapshot alert switch and mark direct dispatch failures for alert.
- Modify `chronoFlow-admin/internal/biz/callback.go`: trigger async alert after final status update.
- Modify `chronoFlow-admin/internal/biz/maintenance.go`: mark stale pending alerts failed on startup and trigger recovery failure alerts.
- Create `chronoFlow-admin/internal/biz/system_setting.go`: settings usecase.
- Create `chronoFlow-admin/internal/biz/alert.go`: alert sender orchestration.
- Create `chronoFlow-admin/internal/biz/feishu_alert.go`: Feishu card payload builder and HTTP sender.

Backend service and DI:

- Create `chronoFlow-admin/internal/service/system_settings.go`.
- Modify `chronoFlow-admin/internal/service/job.go`: request/response mapping.
- Modify `chronoFlow-admin/internal/service/job_log.go`: response mapping.
- Modify `chronoFlow-admin/internal/biz/biz.go`, `chronoFlow-admin/internal/data/data.go`, `chronoFlow-admin/internal/service/service.go`, and `chronoFlow-admin/cmd/chronoFlow-admin/wire.go`: providers and bindings.
- Regenerate `chronoFlow-admin/cmd/chronoFlow-admin/wire_gen.go` with `make wire`.

Frontend:

- Modify `chronoFlow-ui/src/types/job.ts` and `chronoFlow-ui/src/types/jobLog.ts`.
- Modify `chronoFlow-ui/src/api/jobs.ts`, `chronoFlow-ui/src/stores/jobs.ts`, `chronoFlow-ui/src/views/jobs/JobListView.vue`.
- Modify `chronoFlow-ui/src/api/jobLogs.ts`, `chronoFlow-ui/src/stores/jobLogs.ts`, `chronoFlow-ui/src/views/logs/JobLogDetailView.vue`.
- Create `chronoFlow-ui/src/types/systemSettings.ts`.
- Create `chronoFlow-ui/src/api/systemSettings.ts`.
- Modify `chronoFlow-ui/src/views/settings/SettingsView.vue`.
- Optionally add tests only where existing frontend test infrastructure exists; otherwise verify with `npm run build`.

Docs:

- Modify `README.md`, `README.en.md`, and `deploy/README.md`.
- Modify `docs/TESTING_GUIDE.md` with alert manual test cases.

---

## Task 1: Backend schema and proto fields

**Files:**
- Modify: `chronoFlow-admin/api/job/v1/job.proto`
- Modify: `chronoFlow-admin/api/joblog/v1/job_log.proto`
- Create: `chronoFlow-admin/api/system/v1/system_settings.proto`
- Modify: generated files under `chronoFlow-admin/api/all-pb-go/v1/`
- Modify: `chronoFlow-admin/openapi.yaml`
- Test: `chronoFlow-admin/internal/service/job_test.go`
- Test: `chronoFlow-admin/internal/service/job_log_test.go`

- [ ] **Step 1: Add failing service mapping expectations**

In `chronoFlow-admin/internal/service/job_test.go`, extend existing create/update/list assertions so they expect `FailureAlertEnabled` to round-trip. Use this pattern in the relevant tests:

```go
req := &v1.CreateJobRequest{
	ExecutorId:          1,
	Name:                "alert-job",
	CronExpr:            "0 */5 * * * *",
	TimeoutSeconds:      60,
	Description:         "with alert",
	FailureAlertEnabled: true,
}

reply, err := svc.CreateJob(context.Background(), req)
require.NoError(t, err)
require.True(t, reply.GetData().GetJob().GetFailureAlertEnabled())
```

In `chronoFlow-admin/internal/service/job_log_test.go`, extend `toJobLogInfo` tests to expect:

```go
AlertEnabledSnapshot: true,
AlertStatus:          "sent",
AlertError:           "",
AlertSentAt:          "2026-06-30 10:00:00",
```

- [ ] **Step 2: Run tests and verify failure**

Run:

```bash
cd chronoFlow-admin
go test ./internal/service -run 'Test.*Job' -count=1
```

Expected: FAIL because proto generated structs do not yet have alert fields.

- [ ] **Step 3: Update job proto**

Add fields to `JobInfo`, `CreateJobRequest`, and `UpdateJobRequest` in `chronoFlow-admin/api/job/v1/job.proto`:

```proto
message JobInfo {
  int64 id = 1;
  int64 executor_id = 2;
  string name = 3;
  string cron_expr = 4;
  int32 timeout_seconds = 5;
  string schedule_status = 6;
  string description = 7;
  string created_at = 8;
  string updated_at = 9;
  bool failure_alert_enabled = 10;
}

message CreateJobRequest {
  int64 executor_id = 1;
  string name = 2;
  string cron_expr = 3;
  int32 timeout_seconds = 4;
  string description = 5;
  bool failure_alert_enabled = 6;
}

message UpdateJobRequest {
  int64 id = 1;
  int64 executor_id = 2;
  string name = 3;
  string cron_expr = 4;
  int32 timeout_seconds = 5;
  string description = 6;
  bool failure_alert_enabled = 7;
}
```

- [ ] **Step 4: Update job log proto**

Add fields to `JobLogInfo` in `chronoFlow-admin/api/joblog/v1/job_log.proto`:

```proto
message JobLogInfo {
  int64 id = 1;
  int64 job_id = 2;
  string job_name = 3;
  int64 executor_id = 4;
  string executor_name = 5;
  string executor_address = 6;
  string cron_expr = 7;
  int32 timeout_seconds = 8;
  string trigger_type = 9;
  string status = 10;
  string start_time = 11;
  string end_time = 12;
  int64 duration_ms = 13;
  int32 exit_code = 14;
  string log_path = 15;
  int64 log_size_bytes = 16;
  bool log_truncated = 17;
  string error_message = 18;
  string created_at = 19;
  string updated_at = 20;
  bool alert_enabled_snapshot = 21;
  string alert_status = 22;
  string alert_error = 23;
  string alert_sent_at = 24;
}
```

- [ ] **Step 5: Add system settings proto**

Create `chronoFlow-admin/api/system/v1/system_settings.proto`:

```proto
syntax = "proto3";

package system.v1;

import "google/api/annotations.proto";

option go_package = "chronoFlow-admin/api/all-pb-go/v1;v1";

service SystemSettings {
  rpc GetAlertSettings (GetAlertSettingsRequest) returns (GetAlertSettingsReply) {
    option (google.api.http) = {
      get: "/v1/admin/system/settings/alert"
    };
  }

  rpc SaveFeishuWebhook (SaveFeishuWebhookRequest) returns (SaveFeishuWebhookReply) {
    option (google.api.http) = {
      put: "/v1/admin/system/settings/alert/feishu"
      body: "*"
    };
  }

  rpc TestFeishuWebhook (TestFeishuWebhookRequest) returns (TestFeishuWebhookReply) {
    option (google.api.http) = {
      post: "/v1/admin/system/settings/alert/feishu/test"
      body: "*"
    };
  }

  rpc ClearFeishuWebhook (ClearFeishuWebhookRequest) returns (ClearFeishuWebhookReply) {
    option (google.api.http) = {
      delete: "/v1/admin/system/settings/alert/feishu"
    };
  }
}

message AlertSettingsInfo {
  bool feishu_webhook_configured = 1;
  string feishu_webhook_updated_at = 2;
}

message GetAlertSettingsRequest {}

message SaveFeishuWebhookRequest {
  string webhook = 1;
}

message TestFeishuWebhookRequest {}

message ClearFeishuWebhookRequest {}

message GetAlertSettingsReply {
  int32 code = 1;
  string message = 2;
  message Data {
    AlertSettingsInfo settings = 1;
  }
  Data data = 3;
}

message SaveFeishuWebhookReply {
  int32 code = 1;
  string message = 2;
  message Data {
    AlertSettingsInfo settings = 1;
  }
  Data data = 3;
}

message TestFeishuWebhookReply {
  int32 code = 1;
  string message = 2;
  message Data {
    string status = 1;
  }
  Data data = 3;
}

message ClearFeishuWebhookReply {
  int32 code = 1;
  string message = 2;
  message Data {
    AlertSettingsInfo settings = 1;
  }
  Data data = 3;
}
```

- [ ] **Step 6: Regenerate API files**

Run:

```bash
cd chronoFlow-admin
make api
```

Expected: generated `system_settings*.pb.go` files appear under `api/all-pb-go/v1/`, and existing job/job_log generated files include the new fields.

- [ ] **Step 7: Commit**

```bash
git add chronoFlow-admin/api chronoFlow-admin/openapi.yaml chronoFlow-admin/internal/service/job_test.go chronoFlow-admin/internal/service/job_log_test.go
git commit -m "feat: add alert api fields"
```

---

## Task 2: Persist job alert switch and log alert fields

**Files:**
- Modify: `chronoFlow-admin/internal/data/model.go`
- Modify: `chronoFlow-admin/internal/data/data.go`
- Modify: `chronoFlow-admin/internal/data/job.go`
- Modify: `chronoFlow-admin/internal/data/job_log.go`
- Modify: `chronoFlow-admin/internal/biz/job.go`
- Modify: `chronoFlow-admin/internal/biz/job_run.go`
- Modify: `chronoFlow-admin/internal/biz/job_log.go`
- Modify: `chronoFlow-admin/internal/service/job.go`
- Modify: `chronoFlow-admin/internal/service/job_log.go`
- Test: `chronoFlow-admin/internal/data/model_test.go`
- Test: `chronoFlow-admin/internal/biz/job_test.go`
- Test: `chronoFlow-admin/internal/biz/job_run_test.go`

- [ ] **Step 1: Add failing tests for persistence mapping**

In `chronoFlow-admin/internal/data/model_test.go`, add assertions to the existing model migration/mapping tests:

```go
require.True(t, db.Migrator().HasColumn(&Job{}, "failure_alert_enabled"))
require.True(t, db.Migrator().HasColumn(&JobLog{}, "alert_enabled_snapshot"))
require.True(t, db.Migrator().HasColumn(&JobLog{}, "alert_status"))
require.True(t, db.Migrator().HasColumn(&JobLog{}, "alert_error"))
require.True(t, db.Migrator().HasColumn(&JobLog{}, "alert_sent_at"))
```

In `chronoFlow-admin/internal/biz/job_test.go`, add a create/update test that sets `FailureAlertEnabled: true` and verifies the returned `Job` keeps it true.

In `chronoFlow-admin/internal/biz/job_run_test.go`, update the fake created log assertion:

```go
require.True(t, created.AlertEnabledSnapshot)
require.Equal(t, biz.AlertStatusNone, created.AlertStatus)
```

- [ ] **Step 2: Run tests and verify failure**

Run:

```bash
cd chronoFlow-admin
go test ./internal/data ./internal/biz -run 'Test.*(Job|Model)' -count=1
```

Expected: FAIL because fields and constants do not exist yet.

- [ ] **Step 3: Add biz constants and fields**

In `chronoFlow-admin/internal/biz/status.go`, add:

```go
const (
	AlertStatusNone    = "none"
	AlertStatusPending = "pending"
	AlertStatusSent    = "sent"
	AlertStatusFailed  = "failed"
	AlertStatusSkipped = "skipped"
)
```

In `chronoFlow-admin/internal/biz/job.go`, add `FailureAlertEnabled bool` to:

```go
type Job struct { ... }
type CreateJobInput struct { ... }
type UpdateJobInput struct { ... }
```

Pass it through `CreateJob`, `UpdateJob`, and `normalizeJobInput`.

In `chronoFlow-admin/internal/biz/job_log.go`, add fields to `JobLog`:

```go
AlertEnabledSnapshot bool
AlertStatus          string
AlertError           string
AlertSentAt          *time.Time
```

- [ ] **Step 4: Add data model fields**

In `chronoFlow-admin/internal/data/model.go`, update `Job`:

```go
FailureAlertEnabled bool `json:"failureAlertEnabled" gorm:"column:failure_alert_enabled;not null;default:false"`
```

Update `JobLog`:

```go
AlertEnabledSnapshot bool       `json:"alertEnabledSnapshot" gorm:"column:alert_enabled_snapshot;not null;default:false"`
AlertStatus          string     `json:"alertStatus" gorm:"column:alert_status;size:32;not null;default:'none';index"`
AlertError           string     `json:"alertError" gorm:"column:alert_error;type:text"`
AlertSentAt          *time.Time `json:"alertSentAt" gorm:"column:alert_sent_at"`
```

- [ ] **Step 5: Update migration and mappings**

In `chronoFlow-admin/internal/data/data.go`, `AutoMigrate` still uses `&Job{}` and `&JobLog{}`; no new model is required for this task.

In `chronoFlow-admin/internal/data/job.go`, map `FailureAlertEnabled` both ways and preserve it in `Update`.

In `chronoFlow-admin/internal/data/job_log.go`, map the new alert fields in:

- `Update`
- `toJobLogModel`
- `toBizJobLog`

When creating a log in `toJobLogModel`, if `AlertStatus == ""`, set `AlertStatus = biz.AlertStatusNone`.

- [ ] **Step 6: Snapshot alert switch when creating a job log**

In `chronoFlow-admin/internal/biz/job_run.go`, update `RunJob` log creation:

```go
created, err := uc.logRepo.CreateRunningIfNoActive(ctx, &JobLog{
	JobID:                job.ID,
	JobName:              job.Name,
	ExecutorID:           executor.ID,
	ExecutorName:         executor.Name,
	ExecutorAddress:      executor.Address,
	CronExpr:             job.CronExpr,
	TimeoutSeconds:       job.TimeoutSeconds,
	GlueSnapshot:         glue.Content,
	TriggerType:          triggerType,
	Status:               JobLogStatusRunning,
	StartTime:            time.Now(),
	AlertEnabledSnapshot: job.FailureAlertEnabled,
	AlertStatus:          AlertStatusNone,
})
```

- [ ] **Step 7: Update service mappings**

In `chronoFlow-admin/internal/service/job.go`, map `FailureAlertEnabled` in:

- `validateCreateJobRequest`
- `validateUpdateJobRequest`
- `toJobInfo`

In `chronoFlow-admin/internal/service/job_log.go`, map alert fields in `toJobLogInfo`.

- [ ] **Step 8: Run tests**

Run:

```bash
cd chronoFlow-admin
go test ./internal/data ./internal/biz ./internal/service -count=1
```

Expected: PASS.

- [ ] **Step 9: Commit**

```bash
git add chronoFlow-admin/internal/data chronoFlow-admin/internal/biz chronoFlow-admin/internal/service
git commit -m "feat: persist job failure alert fields"
```

---

## Task 3: System setting storage and encrypted Feishu webhook API

**Files:**
- Modify: `chronoFlow-admin/internal/data/model.go`
- Modify: `chronoFlow-admin/internal/data/data.go`
- Create: `chronoFlow-admin/internal/data/system_setting.go`
- Create: `chronoFlow-admin/internal/biz/system_setting.go`
- Create: `chronoFlow-admin/internal/service/system_settings.go`
- Modify: `chronoFlow-admin/internal/biz/biz.go`
- Modify: `chronoFlow-admin/internal/data/data.go`
- Modify: `chronoFlow-admin/internal/service/service.go`
- Modify: `chronoFlow-admin/cmd/chronoFlow-admin/wire.go`
- Modify: generated `wire_gen.go`
- Test: `chronoFlow-admin/internal/data/system_setting_test.go`
- Test: `chronoFlow-admin/internal/biz/system_setting_test.go`
- Test: `chronoFlow-admin/internal/service/system_settings_test.go`

- [ ] **Step 1: Write failing tests**

Create `chronoFlow-admin/internal/biz/system_setting_test.go` with tests:

```go
func TestSystemSettingUsecase_SaveGetAndClearFeishuWebhook(t *testing.T) {
	ctx := context.Background()
	repo := newFakeSystemSettingRepo()
	cipher := security.NewTokenCipherForTest("12345678901234567890123456789012")
	uc := biz.NewSystemSettingUsecase(repo, cipher, log.NewStdLogger(io.Discard))

	settings, err := uc.SaveFeishuWebhook(ctx, "https://open.feishu.cn/open-apis/bot/v2/hook/abc")
	require.NoError(t, err)
	require.True(t, settings.FeishuWebhookConfigured)
	require.NotEmpty(t, settings.FeishuWebhookUpdatedAt)
	require.NotContains(t, repo.savedCiphertext, "open.feishu.cn")

	settings, err = uc.GetAlertSettings(ctx)
	require.NoError(t, err)
	require.True(t, settings.FeishuWebhookConfigured)

	settings, err = uc.ClearFeishuWebhook(ctx)
	require.NoError(t, err)
	require.False(t, settings.FeishuWebhookConfigured)
}
```

Create a second test:

```go
func TestSystemSettingUsecase_SaveFeishuWebhookRejectsInvalidURL(t *testing.T) {
	uc := biz.NewSystemSettingUsecase(newFakeSystemSettingRepo(), fakeCipher{}, log.NewStdLogger(io.Discard))
	_, err := uc.SaveFeishuWebhook(context.Background(), "not-url")
	require.Error(t, err)
}
```

Use small local fakes in the test file rather than a database.

- [ ] **Step 2: Run tests and verify failure**

Run:

```bash
cd chronoFlow-admin
go test ./internal/biz -run TestSystemSettingUsecase -count=1
```

Expected: FAIL because usecase does not exist.

- [ ] **Step 3: Add system setting model and repo**

In `chronoFlow-admin/internal/data/model.go`, add:

```go
type SystemSetting struct {
	ID uint64 `json:"id" gorm:"primaryKey;autoIncrement"`
	ModelOpt
	SettingKey     string `json:"settingKey" gorm:"column:setting_key;size:128;not null;uniqueIndex"`
	ValueEncrypted string `json:"valueEncrypted" gorm:"column:value_encrypted;type:text"`
}

func (SystemSetting) TableName() string {
	return "system_settings"
}
```

In `chronoFlow-admin/internal/data/data.go`, add `&SystemSetting{}` to `AutoMigrate`.

Create `chronoFlow-admin/internal/data/system_setting.go` with:

```go
type SystemSettingRepo struct {
	data *Data
	log  *log.Helper
}

func NewSystemSettingRepo(data *Data, logger log.Logger) *SystemSettingRepo
func (r *SystemSettingRepo) GetByKey(ctx context.Context, key string) (*biz.SystemSetting, error)
func (r *SystemSettingRepo) Upsert(ctx context.Context, setting *biz.SystemSetting) (*biz.SystemSetting, error)
func toBizSystemSetting(model *SystemSetting) *biz.SystemSetting
func toSystemSettingModel(setting *biz.SystemSetting) *SystemSetting
```

`Upsert` should use `clause.OnConflict` on `setting_key` and update `value_encrypted`.

- [ ] **Step 4: Add biz usecase**

Create `chronoFlow-admin/internal/biz/system_setting.go` with:

```go
const FeishuWebhookSettingKey = "alert.feishu.webhook"

type SystemSetting struct {
	ID             int64
	SettingKey     string
	ValueEncrypted string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type AlertSettings struct {
	FeishuWebhookConfigured bool
	FeishuWebhookUpdatedAt  time.Time
}

type SystemSettingRepo interface {
	GetByKey(context.Context, string) (*SystemSetting, error)
	Upsert(context.Context, *SystemSetting) (*SystemSetting, error)
}

type SystemSettingUsecase struct { ... }

func (uc *SystemSettingUsecase) GetAlertSettings(ctx context.Context) (*AlertSettings, error)
func (uc *SystemSettingUsecase) SaveFeishuWebhook(ctx context.Context, webhook string) (*AlertSettings, error)
func (uc *SystemSettingUsecase) ClearFeishuWebhook(ctx context.Context) (*AlertSettings, error)
func (uc *SystemSettingUsecase) GetFeishuWebhook(ctx context.Context) (string, bool, error)
```

Validate URL with `url.ParseRequestURI` and require `http` or `https`.

- [ ] **Step 5: Add service**

Create `chronoFlow-admin/internal/service/system_settings.go` implementing:

```go
type SystemSettingsService struct {
	v1.UnimplementedSystemSettingsServer
	uc *biz.SystemSettingUsecase
	alertUC *biz.AlertUsecase
}

func NewSystemSettingsService(uc *biz.SystemSettingUsecase, alertUC *biz.AlertUsecase) *SystemSettingsService
func (s *SystemSettingsService) GetAlertSettings(...)
func (s *SystemSettingsService) SaveFeishuWebhook(...)
func (s *SystemSettingsService) TestFeishuWebhook(...)
func (s *SystemSettingsService) ClearFeishuWebhook(...)
```

For this task, `TestFeishuWebhook` can temporarily return not implemented until `AlertUsecase` exists in Task 4. If implementing before Task 4, create a narrow interface:

```go
type FeishuTester interface {
	SendTestFeishuAlert(context.Context) error
}
```

and pass nil safely until Task 4.

- [ ] **Step 6: Wire providers**

Update provider sets:

- Add `NewSystemSettingRepo` and binding in `internal/data/data.go`.
- Add `NewSystemSettingUsecase` in `internal/biz/biz.go`.
- Add `NewSystemSettingsService` in `internal/service/service.go`.
- Add registration in `internal/server/http.go` and `internal/server/grpc.go` using generated `RegisterSystemSettingsHTTPServer` and `RegisterSystemSettingsServer`.

Run:

```bash
cd chronoFlow-admin
make wire
```

- [ ] **Step 7: Run tests**

Run:

```bash
cd chronoFlow-admin
go test ./internal/data ./internal/biz ./internal/service -run 'Test.*SystemSetting|Test.*SystemSettings' -count=1
```

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add chronoFlow-admin/api chronoFlow-admin/internal chronoFlow-admin/cmd/chronoFlow-admin
git commit -m "feat: add encrypted alert settings api"
```

---

## Task 4: Feishu card sender and alert usecase

**Files:**
- Create: `chronoFlow-admin/internal/biz/feishu_alert.go`
- Create: `chronoFlow-admin/internal/biz/alert.go`
- Modify: `chronoFlow-admin/internal/biz/biz.go`
- Modify: `chronoFlow-admin/internal/service/system_settings.go`
- Test: `chronoFlow-admin/internal/biz/feishu_alert_test.go`
- Test: `chronoFlow-admin/internal/biz/alert_test.go`

- [ ] **Step 1: Write failing Feishu payload tests**

Create `chronoFlow-admin/internal/biz/feishu_alert_test.go`:

```go
func TestBuildFeishuAlertCardTruncatesErrorAndUsesTitle(t *testing.T) {
	longErr := strings.Repeat("x", 600)
	card := buildFeishuAlertCard(FeishuAlertCardInput{
		Title:        "ChronoFlow 任务执行失败",
		JobName:      "daily-report",
		ExecutorName: "exec-1",
		LogID:        42,
		Status:       JobLogStatusFailed,
		StartTime:    time.Date(2026, 6, 30, 10, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		EndTime:      time.Date(2026, 6, 30, 10, 0, 5, 0, time.FixedZone("CST", 8*3600)),
		DurationMS:   5000,
		ExitCode:     ptrInt32(1),
		ErrorMessage: longErr,
	})

	body, err := json.Marshal(card)
	require.NoError(t, err)
	require.Contains(t, string(body), "ChronoFlow 任务执行失败")
	require.Contains(t, string(body), "daily-report")
	require.Contains(t, string(body), "日志 ID")
	require.NotContains(t, string(body), strings.Repeat("x", 520))
	require.Contains(t, string(body), "...")
}
```

- [ ] **Step 2: Write failing sender retry tests**

In `chronoFlow-admin/internal/biz/alert_test.go`, create a fake HTTP server that fails twice then succeeds:

```go
func TestAlertUsecaseSendRetriesAndMarksSent(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) < 3 {
			http.Error(w, "temporary", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{"StatusCode":0,"msg":"success"}`))
	}))
	defer server.Close()

	settings := fakeSettings{webhook: server.URL}
	repo := newFakeAlertLogRepo(logWithStatus(JobLogStatusFailed, true))
	uc := biz.NewAlertUsecase(repo, settings, biz.NewFeishuAlertSender(server.Client()), log.NewStdLogger(io.Discard))

	uc.DispatchJobLogAlert(context.Background(), 1)
	require.Eventually(t, func() bool {
		return repo.status == biz.AlertStatusSent
	}, time.Second, 10*time.Millisecond)
	require.Equal(t, int32(3), calls.Load())
}
```

- [ ] **Step 3: Run tests and verify failure**

Run:

```bash
cd chronoFlow-admin
go test ./internal/biz -run 'Test.*Alert|Test.*Feishu' -count=1
```

Expected: FAIL because alert sender does not exist.

- [ ] **Step 4: Implement Feishu card builder and sender**

Create `chronoFlow-admin/internal/biz/feishu_alert.go` with:

```go
type FeishuAlertSender struct {
	client *http.Client
}

func NewFeishuAlertSender() *FeishuAlertSender
func (s *FeishuAlertSender) SendCard(ctx context.Context, webhook string, payload any) error
```

Use JSON body:

```json
{
  "msg_type": "interactive",
  "card": { ... }
}
```

Accept HTTP 2xx only. Also parse common Feishu response fields `StatusCode`, `code`, `msg`, and treat non-zero as error.

Add card builder helpers:

```go
type FeishuAlertCardInput struct { ... }
func buildFailureAlertTitle(status string) string
func buildFeishuAlertCard(input FeishuAlertCardInput) map[string]any
func truncateAlertError(message string) string
```

- [ ] **Step 5: Implement AlertUsecase**

Create `chronoFlow-admin/internal/biz/alert.go`:

```go
type AlertJobLogRepo interface {
	GetByID(context.Context, int64) (*JobLog, error)
	MarkAlertPending(context.Context, int64) error
	MarkAlertSent(context.Context, int64, time.Time) error
	MarkAlertFailed(context.Context, int64, string) error
	MarkAlertSkipped(context.Context, int64, string) error
	MarkPendingAlertsFailed(context.Context, string) error
}

type AlertSettingsProvider interface {
	GetFeishuWebhook(context.Context) (string, bool, error)
}

type AlertUsecase struct { ... }

func (uc *AlertUsecase) DispatchJobLogAlert(ctx context.Context, logID int64)
func (uc *AlertUsecase) SendTestFeishuAlert(ctx context.Context) error
func (uc *AlertUsecase) MarkPendingAlertsFailedOnStartup(ctx context.Context) error
```

`DispatchJobLogAlert` should:

1. Load job log.
2. Return if log is nil.
3. Return if `AlertStatus` is already `pending`, `sent`, `failed`, or `skipped`.
4. Return if `AlertEnabledSnapshot` is false; set `skipped` with reason if desired by design.
5. Return if status is not `failed` or `timeout`; set `none`.
6. Check webhook; if not configured, mark skipped.
7. Mark pending.
8. Start goroutine that retries 3 times with 2 second interval.

For tests, keep retry interval injectable:

```go
type AlertConfig struct {
	MaxAttempts int
	RetryDelay  time.Duration
}
```

Production default: `MaxAttempts=3`, `RetryDelay=2*time.Second`.

- [ ] **Step 6: Implement repo alert update methods**

In `chronoFlow-admin/internal/data/job_log.go`, add methods required by `AlertJobLogRepo`.

Use conditional updates to prevent duplicate send:

```go
Where("id = ? AND alert_status IN ?", logID, []string{biz.AlertStatusNone, ""})
```

- [ ] **Step 7: Connect test send API**

Update `chronoFlow-admin/internal/service/system_settings.go` `TestFeishuWebhook` to call `alertUC.SendTestFeishuAlert(ctx)`.

- [ ] **Step 8: Wire provider**

Add `NewFeishuAlertSender`, `NewAlertUsecase`, and repo binding to `biz.ProviderSet` / `data.ProviderSet`.

Run:

```bash
cd chronoFlow-admin
make wire
```

- [ ] **Step 9: Run tests**

Run:

```bash
cd chronoFlow-admin
go test ./internal/biz ./internal/data ./internal/service -run 'Test.*Alert|Test.*Feishu|Test.*SystemSetting' -count=1
```

Expected: PASS.

- [ ] **Step 10: Commit**

```bash
git add chronoFlow-admin/internal chronoFlow-admin/cmd/chronoFlow-admin
git commit -m "feat: add feishu alert sender"
```

---

## Task 5: Trigger alerts from callbacks, direct run failures, and recovery

**Files:**
- Modify: `chronoFlow-admin/internal/biz/callback.go`
- Modify: `chronoFlow-admin/internal/biz/job_run.go`
- Modify: `chronoFlow-admin/internal/biz/maintenance.go`
- Modify: `chronoFlow-admin/internal/worker/server.go`
- Test: `chronoFlow-admin/internal/biz/callback_test.go`
- Test: `chronoFlow-admin/internal/biz/job_run_test.go`
- Test: `chronoFlow-admin/internal/biz/maintenance_test.go`

- [ ] **Step 1: Write failing callback trigger test**

In `chronoFlow-admin/internal/biz/callback_test.go`, add:

```go
func TestCallbackUsecaseDispatchesAlertWhenFinalFailed(t *testing.T) {
	alerts := &fakeAlertDispatcher{}
	uc := NewCallbackUsecase(logRepo, store, CallbackConfig{MaxLogBytes: 1024}, log.NewStdLogger(io.Discard))
	uc.alerts = alerts

	_, err := uc.ApplyCallback(context.Background(), &CallbackInput{
		LogID: 1, JobID: 10, Status: JobLogStatusFailed, ExitCode: 1,
		ErrorMessage: "boom",
	})
	require.NoError(t, err)
	require.Equal(t, []int64{1}, alerts.dispatched)
}
```

Use the existing fake repo style and inject alert dispatcher through constructor rather than direct field if possible.

- [ ] **Step 2: Write non-trigger test**

Add:

```go
func TestCallbackUsecaseDoesNotDispatchAlertWhenSuccess(t *testing.T) {
	alerts := &fakeAlertDispatcher{}
	// callback status success
	require.Empty(t, alerts.dispatched)
}
```

- [ ] **Step 3: Run tests and verify failure**

Run:

```bash
cd chronoFlow-admin
go test ./internal/biz -run 'TestCallbackUsecase.*Alert|TestJobRunUsecase.*Alert|TestMaintenance.*Alert' -count=1
```

Expected: FAIL because callback does not dispatch alerts.

- [ ] **Step 4: Add alert dispatcher interface**

In `chronoFlow-admin/internal/biz/alert.go`, add:

```go
type AlertDispatcher interface {
	DispatchJobLogAlert(context.Context, int64)
	MarkPendingAlertsFailedOnStartup(context.Context) error
}
```

`AlertUsecase` implements it.

- [ ] **Step 5: Inject alert dispatcher into callback**

Modify `CallbackUsecase` constructor:

```go
func NewCallbackUsecase(logRepo JobRunLogRepo, store LogWriter, alerts AlertDispatcher, config CallbackConfig, logger log.Logger) *CallbackUsecase
```

After successful update in `ApplyCallback`, call:

```go
if ShouldTriggerFailureAlert(updated.Status) && uc.alerts != nil {
	uc.alerts.DispatchJobLogAlert(context.Background(), updated.ID)
}
```

Use `context.Background()` for goroutine-safe async work; do not let request cancellation stop alert sending after callback response.

- [ ] **Step 6: Trigger alerts for direct dispatch failures**

In `JobRunUsecase.markLogFailed`, after updating log to failed, dispatch alert if `uc.alerts != nil`.

Update `NewJobRunUsecase` to accept `AlertDispatcher`.

- [ ] **Step 7: Trigger alerts for recovery failures**

Current repo methods `MarkAllActiveLogsFailed` and related bulk update do not return log IDs. Replace or supplement them with methods that return affected IDs:

```go
MarkAllActiveLogsFailedReturningIDs(ctx, message string) ([]int64, error)
MarkKillingTimeoutLogsFailedReturningIDs(ctx, timeoutSeconds int32, message string) ([]int64, error)
```

After marking failed in maintenance/recovery, dispatch alert for each affected log ID.

- [ ] **Step 8: Mark pending alerts failed on startup**

In `chronoFlow-admin/internal/worker/server.go`, call:

```go
if alerts != nil {
	_ = alerts.MarkPendingAlertsFailedOnStartup(ctx)
}
```

Place this near existing startup recovery/maintenance initialization so stale pending alerts are cleaned once Admin starts.

- [ ] **Step 9: Run focused tests**

Run:

```bash
cd chronoFlow-admin
go test ./internal/biz ./internal/worker -run 'Test.*Alert|Test.*Recovery|Test.*Maintenance' -count=1
```

Expected: PASS.

- [ ] **Step 10: Run broader backend tests**

Run:

```bash
cd chronoFlow-admin
go test ./internal/... -count=1
```

Expected: PASS.

- [ ] **Step 11: Commit**

```bash
git add chronoFlow-admin/internal chronoFlow-admin/cmd/chronoFlow-admin
git commit -m "feat: trigger alerts from failed jobs"
```

---

## Task 6: Frontend API types and job alert switch

**Files:**
- Modify: `chronoFlow-ui/src/types/job.ts`
- Modify: `chronoFlow-ui/src/api/jobs.ts`
- Modify: `chronoFlow-ui/src/stores/jobs.ts`
- Modify: `chronoFlow-ui/src/views/jobs/JobListView.vue`
- Test: `chronoFlow-ui`

- [ ] **Step 1: Add frontend types**

In `chronoFlow-ui/src/types/job.ts`, add:

```ts
failureAlertEnabled: boolean
```

to `Job`, `CreateJobPayload`, and `UpdateJobPayload`.

- [ ] **Step 2: Update API mapping**

In `chronoFlow-ui/src/api/jobs.ts`, map backend snake_case:

```ts
failureAlertEnabled: item.failure_alert_enabled ?? false
```

When sending create/update:

```ts
failure_alert_enabled: payload.failureAlertEnabled,
```

- [ ] **Step 3: Update store defaults**

In `chronoFlow-ui/src/stores/jobs.ts`, ensure new job defaults include:

```ts
failureAlertEnabled: false,
```

- [ ] **Step 4: Update job form**

In `chronoFlow-ui/src/views/jobs/JobListView.vue`, add form field:

```vue
<a-form-item label="失败告警" name="failureAlertEnabled">
  <a-switch v-model:checked="formState.failureAlertEnabled" />
  <div class="form-help">任务执行失败或超时时发送飞书告警；需先在系统设置中配置飞书 Webhook。</div>
</a-form-item>
```

Ensure edit form initializes from selected job.

- [ ] **Step 5: Update job table**

In `JobListView.vue`, add a “失败告警” column before “说明”.

Render:

```vue
<a-tooltip v-if="record.failureAlertEnabled" title="开启后，任务失败或超时时发送飞书告警。若系统设置未配置 Webhook，失败时不会发送。">
  <a-tag color="blue">开启</a-tag>
</a-tooltip>
<a-tag v-else>关闭</a-tag>
```

If settings API is available in this task, fetch `feishuWebhookConfigured` and use a warning tooltip when false. If not, leave generic tooltip and refine in Task 7.

- [ ] **Step 6: Build frontend**

Run:

```bash
cd chronoFlow-ui
npm run build
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add chronoFlow-ui/src/types/job.ts chronoFlow-ui/src/api/jobs.ts chronoFlow-ui/src/stores/jobs.ts chronoFlow-ui/src/views/jobs/JobListView.vue
git commit -m "feat: add job alert switch ui"
```

---

## Task 7: Frontend system settings page

**Files:**
- Create: `chronoFlow-ui/src/types/systemSettings.ts`
- Create: `chronoFlow-ui/src/api/systemSettings.ts`
- Modify: `chronoFlow-ui/src/views/settings/SettingsView.vue`
- Modify: `chronoFlow-ui/src/views/jobs/JobListView.vue`
- Test: `chronoFlow-ui`

- [ ] **Step 1: Add system settings types**

Create `chronoFlow-ui/src/types/systemSettings.ts`:

```ts
export interface AlertSettings {
  feishuWebhookConfigured: boolean
  feishuWebhookUpdatedAt: string
}

export interface SaveFeishuWebhookPayload {
  webhook: string
}
```

- [ ] **Step 2: Add system settings API**

Create `chronoFlow-ui/src/api/systemSettings.ts`:

```ts
import { http } from './request'
import type { AlertSettings, SaveFeishuWebhookPayload } from '@/types/systemSettings'

function mapAlertSettings(raw: any): AlertSettings {
  return {
    feishuWebhookConfigured: Boolean(raw?.feishu_webhook_configured),
    feishuWebhookUpdatedAt: raw?.feishu_webhook_updated_at || '',
  }
}

export async function getAlertSettings(): Promise<AlertSettings> {
  const res = await http.get('/v1/admin/system/settings/alert')
  return mapAlertSettings(res.data?.settings)
}

export async function saveFeishuWebhook(payload: SaveFeishuWebhookPayload): Promise<AlertSettings> {
  const res = await http.put('/v1/admin/system/settings/alert/feishu', {
    webhook: payload.webhook,
  })
  return mapAlertSettings(res.data?.settings)
}

export async function testFeishuWebhook(): Promise<void> {
  await http.post('/v1/admin/system/settings/alert/feishu/test', {})
}

export async function clearFeishuWebhook(): Promise<AlertSettings> {
  const res = await http.delete('/v1/admin/system/settings/alert/feishu')
  return mapAlertSettings(res.data?.settings)
}
```

- [ ] **Step 3: Replace settings placeholder page**

Rewrite `chronoFlow-ui/src/views/settings/SettingsView.vue` to include:

- Page title “系统设置”
- Feishu alert settings section
- Status tag: 已配置 / 未配置
- Updated time display
- Password input with eye icon support from Ant Design Vue
- Save button
- Test send button disabled when not configured
- Clear button with confirm
- Explanation alert for Feishu robot creation and V1 no Secret support

Use existing visual style from other views: `PageHeaderBar`, `a-card`, `a-alert`, `a-space`.

- [ ] **Step 4: Use settings status in job list tooltip**

In `JobListView.vue`, call `getAlertSettings()` on mount and use the result for tooltip:

```ts
const alertSettings = ref<AlertSettings | null>(null)
const feishuWebhookConfigured = computed(() => alertSettings.value?.feishuWebhookConfigured ?? false)
```

Tooltip for enabled alert and missing webhook:

```text
系统设置未配置飞书 Webhook，失败时不会发送。
```

- [ ] **Step 5: Build frontend**

Run:

```bash
cd chronoFlow-ui
npm run build
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add chronoFlow-ui/src/types/systemSettings.ts chronoFlow-ui/src/api/systemSettings.ts chronoFlow-ui/src/views/settings/SettingsView.vue chronoFlow-ui/src/views/jobs/JobListView.vue
git commit -m "feat: add alert settings ui"
```

---

## Task 8: Frontend job log alert status display

**Files:**
- Modify: `chronoFlow-ui/src/types/jobLog.ts`
- Modify: `chronoFlow-ui/src/api/jobLogs.ts`
- Modify: `chronoFlow-ui/src/stores/jobLogs.ts`
- Modify: `chronoFlow-ui/src/views/logs/JobLogDetailView.vue`
- Test: `chronoFlow-ui`

- [ ] **Step 1: Add job log alert fields**

In `chronoFlow-ui/src/types/jobLog.ts`, add:

```ts
alertEnabledSnapshot: boolean
alertStatus: 'none' | 'pending' | 'sent' | 'failed' | 'skipped' | ''
alertError: string
alertSentAt: string
```

- [ ] **Step 2: Map API fields**

In `chronoFlow-ui/src/api/jobLogs.ts`, map:

```ts
alertEnabledSnapshot: Boolean(item.alert_enabled_snapshot),
alertStatus: item.alert_status || 'none',
alertError: item.alert_error || '',
alertSentAt: item.alert_sent_at || '',
```

- [ ] **Step 3: Add display helpers**

In `JobLogDetailView.vue`, add helper:

```ts
function alertStatusText(log: JobLog): string {
  if (!log.alertEnabledSnapshot) return '未启用'
  if (log.alertStatus === 'sent') return '已发送'
  if (log.alertStatus === 'pending') return '发送中'
  if (log.alertStatus === 'failed') return `发送失败${log.alertError ? `：${log.alertError}` : ''}`
  if (log.alertStatus === 'skipped') return `未发送${log.alertError ? `：${log.alertError}` : ''}`
  return '未发送'
}
```

- [ ] **Step 4: Render alert status in detail**

Add a row near execution status/error info:

```vue
<a-descriptions-item label="失败告警">
  <a-tag :color="alertTagColor(detail.log)">{{ alertStatusText(detail.log) }}</a-tag>
</a-descriptions-item>
```

Use `green` for sent, `orange` for pending/skipped, `red` for failed, default for none.

- [ ] **Step 5: Build frontend**

Run:

```bash
cd chronoFlow-ui
npm run build
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add chronoFlow-ui/src/types/jobLog.ts chronoFlow-ui/src/api/jobLogs.ts chronoFlow-ui/src/stores/jobLogs.ts chronoFlow-ui/src/views/logs/JobLogDetailView.vue
git commit -m "feat: show alert status in logs"
```

---

## Task 9: Documentation and manual testing guide

**Files:**
- Modify: `README.md`
- Modify: `README.en.md`
- Modify: `deploy/README.md`
- Modify: `docs/TESTING_GUIDE.md`

- [ ] **Step 1: Update root README files**

In `README.md`, add failure alert to feature table:

```markdown
| 飞书失败告警 | 在系统设置中配置飞书 Webhook，任务失败或超时时发送卡片告警。 |
```

In `README.en.md`, add:

```markdown
| Feishu failure alerts | Configure a Feishu webhook in System Settings and send card alerts when jobs fail or time out. |
```

- [ ] **Step 2: Update deploy README**

In `deploy/README.md`, add section:

```markdown
## 飞书失败告警

1. 在飞书群中进入群设置。
2. 添加自定义机器人。
3. 复制机器人 Webhook。
4. 登录 ChronoFlow，进入“系统设置”。
5. 粘贴 Webhook 并保存。
6. 点击“测试发送”确认群里能收到卡片。

V1 不支持飞书签名 Secret。如果启用机器人安全策略，建议使用关键词校验，或先不启用签名校验。

任务失败判断依赖进程退出码，不解析日志正文。Glue Shell 调用 Python 脚本时推荐：

```bash
#!/bin/bash
set -euo pipefail

python3 /scripts/report.py
```
```

- [ ] **Step 3: Update testing guide**

Add test cases to `docs/TESTING_GUIDE.md`:

```markdown
### TC-ALERT-001 保存飞书 Webhook
- 进入系统设置。
- 填写飞书 Webhook 并保存。
- 预期：页面显示已配置，不回显 Webhook 明文。

### TC-ALERT-002 测试发送
- 点击测试发送。
- 预期：飞书群收到 ChronoFlow 测试卡片。

### TC-ALERT-003 failed 任务发送告警
- 创建 Glue Shell：`set -euo pipefail; exit 1`
- 开启失败告警并运行。
- 预期：任务日志 failed，飞书群收到失败卡片，日志详情显示告警已发送。

### TC-ALERT-004 timeout 任务发送告警
- 创建超时任务。
- 开启失败告警并运行。
- 预期：任务日志 timeout，飞书群收到超时卡片。

### TC-ALERT-005 Webhook 未配置
- 清空 Webhook。
- 运行开启失败告警的失败任务。
- 预期：不发送飞书，日志详情显示未发送：系统未配置飞书 Webhook。
```

- [ ] **Step 4: Commit**

```bash
git add README.md README.en.md deploy/README.md docs/TESTING_GUIDE.md
git commit -m "docs: document failure alerts"
```

---

## Task 10: Full verification and release readiness

**Files:**
- No required source edits unless verification finds defects.

- [ ] **Step 1: Run backend tests**

Run:

```bash
cd chronoFlow-admin
go test ./internal/... -count=1
```

Expected: PASS.

- [ ] **Step 2: Run frontend build**

Run:

```bash
cd chronoFlow-ui
npm run build
```

Expected: PASS.

- [ ] **Step 3: Generate API and wire one final time**

Run:

```bash
cd chronoFlow-admin
make api
make wire
go test ./internal/... -count=1
```

Expected: generated files unchanged after final run; tests PASS.

- [ ] **Step 4: Check git diff**

Run:

```bash
git status --short
git diff --stat
```

Expected: no unexpected untracked files. Any generated changes should already belong to prior commits or be committed here.

- [ ] **Step 5: Manual local smoke test**

Start local stack using the existing deployment docs, then verify:

1. Login works.
2. System settings page loads.
3. Webhook save shows configured.
4. Test send reaches Feishu.
5. Failed task sends card.
6. Log detail shows sent.
7. Clear webhook.
8. Failed task does not send card and log detail shows skipped reason.

- [ ] **Step 6: Final commit if verification caused changes**

If verification required fixes:

```bash
git add <changed-files>
git commit -m "fix: stabilize failure alert flow"
```

If no changes were required, do not create an empty commit.

---

## Self-Review

- Spec coverage: The plan covers global encrypted Feishu Webhook, system settings UI, per-job switch, job list column, log detail alert status, failed/timeout triggers, async sending, retry behavior, pending cleanup, no Secret, no URL link, no log parsing, and documentation.
- Placeholder scan: This plan contains no unresolved placeholders. Steps include concrete files, code snippets, commands, and expected outcomes.
- Type consistency: The selected names are consistent across tasks: `failure_alert_enabled`, `FailureAlertEnabled`, `alert_enabled_snapshot`, `alert_status`, `alert_error`, `alert_sent_at`, `SystemSettingUsecase`, `AlertUsecase`, and `FeishuAlertSender`.
