# 任务失败告警实施计划

> **给 AI/开发者的要求：** 实施本计划时必须使用 `superpowers:subagent-driven-development`（推荐）或 `superpowers:executing-plans`，并按任务逐项推进。步骤使用 checkbox（`- [ ]`）格式，方便执行过程中跟踪进度。

**目标：** 为 ChronoFlow 增加 V1 飞书 Webhook 失败告警能力，包括全局加密 Webhook 配置、任务级告警开关、异步发送飞书卡片、日志详情展示告警状态、前端系统设置页面和文档更新。

**架构：** Admin 仍然是唯一连接 MySQL 和发送告警的服务，Exec 不感知告警配置。任务开始运行时把任务的失败告警开关快照到 `job_logs`，当日志首次进入 `failed` 或 `timeout` 最终状态后，由 Admin 异步发送飞书卡片。

**技术栈：** Go 1.22、Kratos、gRPC/HTTP proto 生成、GORM/MySQL、Vue 3、Ant Design Vue、Docker Compose。

---

## 文件地图

### 后端 API 和生成文件

- 修改 `chronoFlow-admin/api/job/v1/job.proto`：新增 `failure_alert_enabled`。
- 修改 `chronoFlow-admin/api/joblog/v1/job_log.proto`：给 `JobLogInfo` 增加告警字段。
- 新增 `chronoFlow-admin/api/system/v1/system_settings.proto`：系统设置 API，用于飞书告警 Webhook。
- 运行 `make api` 重新生成 `chronoFlow-admin/api/all-pb-go/v1/*` 和 `chronoFlow-admin/openapi.yaml`。

### 后端数据层

- 修改 `chronoFlow-admin/internal/data/model.go`：新增 `failure_alert_enabled`、告警字段和 `SystemSetting` 模型。
- 修改 `chronoFlow-admin/internal/data/data.go`：迁移 `SystemSetting`，绑定仓储接口。
- 修改 `chronoFlow-admin/internal/data/job.go`：映射 `failure_alert_enabled`。
- 修改 `chronoFlow-admin/internal/data/job_log.go`：映射告警字段，增加告警状态更新方法。
- 新增 `chronoFlow-admin/internal/data/system_setting.go`：系统设置持久化。

### 后端业务层

- 修改 `chronoFlow-admin/internal/biz/job.go`：增加 `FailureAlertEnabled`。
- 修改 `chronoFlow-admin/internal/biz/job_run.go`：运行时快照告警开关，并处理直接调度失败告警。
- 修改 `chronoFlow-admin/internal/biz/callback.go`：callback 更新最终状态后触发异步告警。
- 修改 `chronoFlow-admin/internal/biz/maintenance.go`：启动时处理历史 pending 告警，恢复失败时触发告警。
- 新增 `chronoFlow-admin/internal/biz/system_setting.go`：系统设置 usecase。
- 新增 `chronoFlow-admin/internal/biz/alert.go`：告警编排 usecase。
- 新增 `chronoFlow-admin/internal/biz/feishu_alert.go`：飞书卡片构建和 HTTP 发送器。

### 后端服务与依赖注入

- 新增 `chronoFlow-admin/internal/service/system_settings.go`。
- 修改 `chronoFlow-admin/internal/service/job.go`：任务请求/响应映射。
- 修改 `chronoFlow-admin/internal/service/job_log.go`：日志响应映射。
- 修改 `chronoFlow-admin/internal/biz/biz.go`、`chronoFlow-admin/internal/data/data.go`、`chronoFlow-admin/internal/service/service.go`、`chronoFlow-admin/cmd/chronoFlow-admin/wire.go`：Provider 和接口绑定。
- 运行 `make wire` 重新生成 `chronoFlow-admin/cmd/chronoFlow-admin/wire_gen.go`。

### 前端

- 修改 `chronoFlow-ui/src/types/job.ts`、`chronoFlow-ui/src/types/jobLog.ts`。
- 修改 `chronoFlow-ui/src/api/jobs.ts`、`chronoFlow-ui/src/stores/jobs.ts`、`chronoFlow-ui/src/views/jobs/JobListView.vue`。
- 修改 `chronoFlow-ui/src/api/jobLogs.ts`、`chronoFlow-ui/src/stores/jobLogs.ts`、`chronoFlow-ui/src/views/logs/JobLogDetailView.vue`。
- 新增 `chronoFlow-ui/src/types/systemSettings.ts`。
- 新增 `chronoFlow-ui/src/api/systemSettings.ts`。
- 修改 `chronoFlow-ui/src/views/settings/SettingsView.vue`。

### 文档

- 修改 `README.md`、`README.en.md`、`deploy/README.md`。
- 修改 `docs/TESTING_GUIDE.md`，增加失败告警测试用例。

---

## 任务 1：后端 Schema 和 Proto 字段

**文件：**

- 修改：`chronoFlow-admin/api/job/v1/job.proto`
- 修改：`chronoFlow-admin/api/joblog/v1/job_log.proto`
- 新增：`chronoFlow-admin/api/system/v1/system_settings.proto`
- 修改：`chronoFlow-admin/api/all-pb-go/v1/` 下生成文件
- 修改：`chronoFlow-admin/openapi.yaml`
- 测试：`chronoFlow-admin/internal/service/job_test.go`
- 测试：`chronoFlow-admin/internal/service/job_log_test.go`

- [ ] **步骤 1：先写失败的 service 映射测试**

在 `chronoFlow-admin/internal/service/job_test.go` 中扩展创建、更新、列表相关断言，验证 `FailureAlertEnabled` 可以完整往返。

示例断言：

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

在 `chronoFlow-admin/internal/service/job_log_test.go` 中扩展 `toJobLogInfo` 测试，验证：

```go
AlertEnabledSnapshot: true,
AlertStatus:          "sent",
AlertError:           "",
AlertSentAt:          "2026-06-30 10:00:00",
```

- [ ] **步骤 2：运行测试确认失败**

```bash
cd chronoFlow-admin
go test ./internal/service -run 'Test.*Job' -count=1
```

预期：失败，因为 proto 生成结构体还没有告警字段。

- [ ] **步骤 3：修改 job proto**

在 `chronoFlow-admin/api/job/v1/job.proto` 中给 `JobInfo`、`CreateJobRequest`、`UpdateJobRequest` 增加字段：

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

- [ ] **步骤 4：修改 job_log proto**

在 `chronoFlow-admin/api/joblog/v1/job_log.proto` 的 `JobLogInfo` 中增加字段：

```proto
bool alert_enabled_snapshot = 21;
string alert_status = 22;
string alert_error = 23;
string alert_sent_at = 24;
```

- [ ] **步骤 5：新增系统设置 proto**

新增 `chronoFlow-admin/api/system/v1/system_settings.proto`：

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

- [ ] **步骤 6：重新生成 API 文件**

```bash
cd chronoFlow-admin
make api
```

预期：`api/all-pb-go/v1/` 下出现 `system_settings*.pb.go`，并且 job/job_log 生成文件包含新增字段。

- [ ] **步骤 7：提交**

```bash
git add chronoFlow-admin/api chronoFlow-admin/openapi.yaml chronoFlow-admin/internal/service/job_test.go chronoFlow-admin/internal/service/job_log_test.go
git commit -m "feat: add alert api fields"
```

---

## 任务 2：持久化任务告警开关和日志告警字段

**文件：**

- 修改：`chronoFlow-admin/internal/data/model.go`
- 修改：`chronoFlow-admin/internal/data/data.go`
- 修改：`chronoFlow-admin/internal/data/job.go`
- 修改：`chronoFlow-admin/internal/data/job_log.go`
- 修改：`chronoFlow-admin/internal/biz/job.go`
- 修改：`chronoFlow-admin/internal/biz/job_run.go`
- 修改：`chronoFlow-admin/internal/biz/job_log.go`
- 修改：`chronoFlow-admin/internal/service/job.go`
- 修改：`chronoFlow-admin/internal/service/job_log.go`
- 测试：`chronoFlow-admin/internal/data/model_test.go`
- 测试：`chronoFlow-admin/internal/biz/job_test.go`
- 测试：`chronoFlow-admin/internal/biz/job_run_test.go`

- [ ] **步骤 1：先写失败测试**

在 `chronoFlow-admin/internal/data/model_test.go` 中增加迁移字段断言：

```go
require.True(t, db.Migrator().HasColumn(&Job{}, "failure_alert_enabled"))
require.True(t, db.Migrator().HasColumn(&JobLog{}, "alert_enabled_snapshot"))
require.True(t, db.Migrator().HasColumn(&JobLog{}, "alert_status"))
require.True(t, db.Migrator().HasColumn(&JobLog{}, "alert_error"))
require.True(t, db.Migrator().HasColumn(&JobLog{}, "alert_sent_at"))
```

在 `chronoFlow-admin/internal/biz/job_test.go` 中增加 create/update 测试：传入 `FailureAlertEnabled: true`，验证返回的 `Job` 仍然为 true。

在 `chronoFlow-admin/internal/biz/job_run_test.go` 中断言创建运行日志时写入快照：

```go
require.True(t, created.AlertEnabledSnapshot)
require.Equal(t, biz.AlertStatusNone, created.AlertStatus)
```

- [ ] **步骤 2：运行测试确认失败**

```bash
cd chronoFlow-admin
go test ./internal/data ./internal/biz -run 'Test.*(Job|Model)' -count=1
```

预期：失败，因为字段和常量还不存在。

- [ ] **步骤 3：增加 biz 常量和字段**

在 `chronoFlow-admin/internal/biz/status.go` 中增加：

```go
const (
	AlertStatusNone    = "none"
	AlertStatusPending = "pending"
	AlertStatusSent    = "sent"
	AlertStatusFailed  = "failed"
	AlertStatusSkipped = "skipped"
)
```

在 `chronoFlow-admin/internal/biz/job.go` 中给 `Job`、`CreateJobInput`、`UpdateJobInput` 增加：

```go
FailureAlertEnabled bool
```

并在 `CreateJob`、`UpdateJob`、`normalizeJobInput` 中传递该字段。

在 `chronoFlow-admin/internal/biz/job_log.go` 中给 `JobLog` 增加：

```go
AlertEnabledSnapshot bool
AlertStatus          string
AlertError           string
AlertSentAt          *time.Time
```

- [ ] **步骤 4：增加数据模型字段**

在 `chronoFlow-admin/internal/data/model.go` 的 `Job` 中增加：

```go
FailureAlertEnabled bool `json:"failureAlertEnabled" gorm:"column:failure_alert_enabled;not null;default:false"`
```

在 `JobLog` 中增加：

```go
AlertEnabledSnapshot bool       `json:"alertEnabledSnapshot" gorm:"column:alert_enabled_snapshot;not null;default:false"`
AlertStatus          string     `json:"alertStatus" gorm:"column:alert_status;size:32;not null;default:'none';index"`
AlertError           string     `json:"alertError" gorm:"column:alert_error;type:text"`
AlertSentAt          *time.Time `json:"alertSentAt" gorm:"column:alert_sent_at"`
```

- [ ] **步骤 5：更新映射和迁移**

`chronoFlow-admin/internal/data/data.go` 中 `AutoMigrate` 仍然包含 `&Job{}`、`&JobLog{}`，无需新增模型。

在 `chronoFlow-admin/internal/data/job.go` 中双向映射 `FailureAlertEnabled`，并在 `Update` 中保留更新。

在 `chronoFlow-admin/internal/data/job_log.go` 中更新：

- `Update`
- `toJobLogModel`
- `toBizJobLog`

创建日志时如果 `AlertStatus == ""`，设置为 `biz.AlertStatusNone`。

- [ ] **步骤 6：创建运行日志时写入告警快照**

在 `chronoFlow-admin/internal/biz/job_run.go` 的 `RunJob` 中创建日志时增加：

```go
AlertEnabledSnapshot: job.FailureAlertEnabled,
AlertStatus:          AlertStatusNone,
```

- [ ] **步骤 7：更新 service 映射**

在 `chronoFlow-admin/internal/service/job.go` 中映射 `FailureAlertEnabled`：

- `validateCreateJobRequest`
- `validateUpdateJobRequest`
- `toJobInfo`

在 `chronoFlow-admin/internal/service/job_log.go` 的 `toJobLogInfo` 中映射告警字段。

- [ ] **步骤 8：运行测试**

```bash
cd chronoFlow-admin
go test ./internal/data ./internal/biz ./internal/service -count=1
```

预期：通过。

- [ ] **步骤 9：提交**

```bash
git add chronoFlow-admin/internal/data chronoFlow-admin/internal/biz chronoFlow-admin/internal/service
git commit -m "feat: persist job failure alert fields"
```

---

## 任务 3：系统设置存储和加密飞书 Webhook API

**文件：**

- 修改：`chronoFlow-admin/internal/data/model.go`
- 修改：`chronoFlow-admin/internal/data/data.go`
- 新增：`chronoFlow-admin/internal/data/system_setting.go`
- 新增：`chronoFlow-admin/internal/biz/system_setting.go`
- 新增：`chronoFlow-admin/internal/service/system_settings.go`
- 修改：`chronoFlow-admin/internal/biz/biz.go`
- 修改：`chronoFlow-admin/internal/data/data.go`
- 修改：`chronoFlow-admin/internal/service/service.go`
- 修改：`chronoFlow-admin/cmd/chronoFlow-admin/wire.go`
- 修改：`chronoFlow-admin/cmd/chronoFlow-admin/wire_gen.go`
- 测试：`chronoFlow-admin/internal/data/system_setting_test.go`
- 测试：`chronoFlow-admin/internal/biz/system_setting_test.go`
- 测试：`chronoFlow-admin/internal/service/system_settings_test.go`

- [ ] **步骤 1：写失败测试**

新增 `chronoFlow-admin/internal/biz/system_setting_test.go`，覆盖保存、查询、清空 Webhook，以及无效 URL。

核心断言：

```go
settings, err := uc.SaveFeishuWebhook(ctx, "https://open.feishu.cn/open-apis/bot/v2/hook/abc")
require.NoError(t, err)
require.True(t, settings.FeishuWebhookConfigured)
require.NotContains(t, repo.savedCiphertext, "open.feishu.cn")
```

无效 URL：

```go
_, err := uc.SaveFeishuWebhook(context.Background(), "not-url")
require.Error(t, err)
```

- [ ] **步骤 2：运行测试确认失败**

```bash
cd chronoFlow-admin
go test ./internal/biz -run TestSystemSettingUsecase -count=1
```

预期：失败，因为 usecase 还不存在。

- [ ] **步骤 3：增加 system_settings 模型和仓储**

在 `chronoFlow-admin/internal/data/model.go` 中增加：

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

在 `chronoFlow-admin/internal/data/data.go` 的 `AutoMigrate` 中增加 `&SystemSetting{}`。

新增 `chronoFlow-admin/internal/data/system_setting.go`，实现：

```go
type SystemSettingRepo struct {
	data *Data
	log  *log.Helper
}

func NewSystemSettingRepo(data *Data, logger log.Logger) *SystemSettingRepo
func (r *SystemSettingRepo) GetByKey(ctx context.Context, key string) (*biz.SystemSetting, error)
func (r *SystemSettingRepo) Upsert(ctx context.Context, setting *biz.SystemSetting) (*biz.SystemSetting, error)
```

`Upsert` 使用 `clause.OnConflict`，根据 `setting_key` 冲突更新 `value_encrypted`。

- [ ] **步骤 4：增加 biz usecase**

新增 `chronoFlow-admin/internal/biz/system_setting.go`，包含：

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
```

实现：

```go
func (uc *SystemSettingUsecase) GetAlertSettings(ctx context.Context) (*AlertSettings, error)
func (uc *SystemSettingUsecase) SaveFeishuWebhook(ctx context.Context, webhook string) (*AlertSettings, error)
func (uc *SystemSettingUsecase) ClearFeishuWebhook(ctx context.Context) (*AlertSettings, error)
func (uc *SystemSettingUsecase) GetFeishuWebhook(ctx context.Context) (string, bool, error)
```

URL 校验使用 `url.ParseRequestURI`，只允许 `http` 或 `https`。

- [ ] **步骤 5：增加 service**

新增 `chronoFlow-admin/internal/service/system_settings.go`，实现：

```go
type SystemSettingsService struct {
	v1.UnimplementedSystemSettingsServer
	uc      *biz.SystemSettingUsecase
	alertUC *biz.AlertUsecase
}
```

实现接口：

- `GetAlertSettings`
- `SaveFeishuWebhook`
- `TestFeishuWebhook`
- `ClearFeishuWebhook`

如果此任务先于告警发送器实现，`TestFeishuWebhook` 可以暂时通过窄接口或 nil 检查处理，任务 4 再接入真实发送。

- [ ] **步骤 6：更新依赖注入和路由注册**

更新：

- `internal/data/data.go`
- `internal/biz/biz.go`
- `internal/service/service.go`
- `internal/server/http.go`
- `internal/server/grpc.go`

注册生成的：

- `RegisterSystemSettingsHTTPServer`
- `RegisterSystemSettingsServer`

运行：

```bash
cd chronoFlow-admin
make wire
```

- [ ] **步骤 7：运行测试**

```bash
cd chronoFlow-admin
go test ./internal/data ./internal/biz ./internal/service -run 'Test.*SystemSetting|Test.*SystemSettings' -count=1
```

预期：通过。

- [ ] **步骤 8：提交**

```bash
git add chronoFlow-admin/api chronoFlow-admin/internal chronoFlow-admin/cmd/chronoFlow-admin
git commit -m "feat: add encrypted alert settings api"
```

---

## 任务 4：飞书卡片发送器和告警 Usecase

**文件：**

- 新增：`chronoFlow-admin/internal/biz/feishu_alert.go`
- 新增：`chronoFlow-admin/internal/biz/alert.go`
- 修改：`chronoFlow-admin/internal/biz/biz.go`
- 修改：`chronoFlow-admin/internal/service/system_settings.go`
- 测试：`chronoFlow-admin/internal/biz/feishu_alert_test.go`
- 测试：`chronoFlow-admin/internal/biz/alert_test.go`

- [ ] **步骤 1：写飞书卡片 payload 失败测试**

新增 `chronoFlow-admin/internal/biz/feishu_alert_test.go`，验证：

- 标题包含 `ChronoFlow 任务执行失败`
- 包含任务名称
- 包含日志 ID
- 错误信息超过 500 字符会截断并追加 `...`

- [ ] **步骤 2：写发送重试失败测试**

在 `chronoFlow-admin/internal/biz/alert_test.go` 中使用 `httptest.NewServer`，前两次返回 500，第三次成功，断言：

- 实际请求 3 次
- 最终 `alert_status=sent`

- [ ] **步骤 3：运行测试确认失败**

```bash
cd chronoFlow-admin
go test ./internal/biz -run 'Test.*Alert|Test.*Feishu' -count=1
```

预期：失败，因为发送器和告警 usecase 不存在。

- [ ] **步骤 4：实现飞书卡片构建和发送器**

新增 `chronoFlow-admin/internal/biz/feishu_alert.go`，包含：

```go
type FeishuAlertSender struct {
	client *http.Client
}

func NewFeishuAlertSender() *FeishuAlertSender
func (s *FeishuAlertSender) SendCard(ctx context.Context, webhook string, payload any) error
```

请求体使用：

```json
{
  "msg_type": "interactive",
  "card": {}
}
```

仅接受 HTTP 2xx。解析飞书常见响应字段 `StatusCode`、`code`、`msg`，非 0 视为失败。

同时实现：

```go
type FeishuAlertCardInput struct { ... }
func buildFailureAlertTitle(status string) string
func buildFeishuAlertCard(input FeishuAlertCardInput) map[string]any
func truncateAlertError(message string) string
```

- [ ] **步骤 5：实现 AlertUsecase**

新增 `chronoFlow-admin/internal/biz/alert.go`，包含：

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
```

实现：

```go
func (uc *AlertUsecase) DispatchJobLogAlert(ctx context.Context, logID int64)
func (uc *AlertUsecase) SendTestFeishuAlert(ctx context.Context) error
func (uc *AlertUsecase) MarkPendingAlertsFailedOnStartup(ctx context.Context) error
```

生产默认：

```go
MaxAttempts = 3
RetryDelay = 2 * time.Second
```

测试中允许注入更短的重试间隔。

- [ ] **步骤 6：实现 JobLogRepo 告警更新方法**

在 `chronoFlow-admin/internal/data/job_log.go` 中实现 `AlertJobLogRepo` 所需方法。

为避免重复发送，`MarkAlertPending` 使用条件更新：

```go
WHERE id = ? AND alert_status IN ('', 'none')
```

- [ ] **步骤 7：接入测试发送 API**

在 `chronoFlow-admin/internal/service/system_settings.go` 中，让 `TestFeishuWebhook` 调用：

```go
alertUC.SendTestFeishuAlert(ctx)
```

- [ ] **步骤 8：更新依赖注入**

把 `NewFeishuAlertSender`、`NewAlertUsecase` 以及相关 repo 绑定加入 ProviderSet，然后运行：

```bash
cd chronoFlow-admin
make wire
```

- [ ] **步骤 9：运行测试**

```bash
cd chronoFlow-admin
go test ./internal/biz ./internal/data ./internal/service -run 'Test.*Alert|Test.*Feishu|Test.*SystemSetting' -count=1
```

预期：通过。

- [ ] **步骤 10：提交**

```bash
git add chronoFlow-admin/internal chronoFlow-admin/cmd/chronoFlow-admin
git commit -m "feat: add feishu alert sender"
```

---

## 任务 5：从 callback、直接运行失败和恢复流程触发告警

**文件：**

- 修改：`chronoFlow-admin/internal/biz/callback.go`
- 修改：`chronoFlow-admin/internal/biz/job_run.go`
- 修改：`chronoFlow-admin/internal/biz/maintenance.go`
- 修改：`chronoFlow-admin/internal/worker/server.go`
- 测试：`chronoFlow-admin/internal/biz/callback_test.go`
- 测试：`chronoFlow-admin/internal/biz/job_run_test.go`
- 测试：`chronoFlow-admin/internal/biz/maintenance_test.go`

- [ ] **步骤 1：写 callback 触发告警失败测试**

在 `callback_test.go` 中添加测试：当 callback 最终状态为 `failed` 时，断言 fake alert dispatcher 收到对应 `log_id`。

- [ ] **步骤 2：写 success 不触发测试**

添加测试：当 callback 状态为 `success` 时，不触发告警。

- [ ] **步骤 3：运行测试确认失败**

```bash
cd chronoFlow-admin
go test ./internal/biz -run 'TestCallbackUsecase.*Alert|TestJobRunUsecase.*Alert|TestMaintenance.*Alert' -count=1
```

预期：失败，因为 callback 尚未触发告警。

- [ ] **步骤 4：增加 AlertDispatcher 接口**

在 `chronoFlow-admin/internal/biz/alert.go` 中增加：

```go
type AlertDispatcher interface {
	DispatchJobLogAlert(context.Context, int64)
	MarkPendingAlertsFailedOnStartup(context.Context) error
}
```

`AlertUsecase` 实现该接口。

- [ ] **步骤 5：callback 注入 alert dispatcher**

修改 `NewCallbackUsecase` 构造函数，增加 `alerts AlertDispatcher`。

在 `ApplyCallback` 成功更新最终状态后：

```go
if ShouldTriggerFailureAlert(updated.Status) && uc.alerts != nil {
	uc.alerts.DispatchJobLogAlert(context.Background(), updated.ID)
}
```

使用 `context.Background()`，避免请求取消影响异步告警发送。

- [ ] **步骤 6：直接调度失败触发告警**

在 `JobRunUsecase.markLogFailed` 中，更新日志为 `failed` 后触发告警。

同时修改 `NewJobRunUsecase`，注入 `AlertDispatcher`。

- [ ] **步骤 7：恢复失败触发告警**

当前批量方法不返回受影响日志 ID，需要补充返回 ID 的方法，例如：

```go
MarkAllActiveLogsFailedReturningIDs(ctx, message string) ([]int64, error)
MarkKillingTimeoutLogsFailedReturningIDs(ctx, timeoutSeconds int32, message string) ([]int64, error)
```

恢复或 killing 超时标记失败后，对每个返回的 log ID 触发告警。

- [ ] **步骤 8：启动时清理 pending 告警**

在 `chronoFlow-admin/internal/worker/server.go` 启动流程中调用：

```go
_ = alerts.MarkPendingAlertsFailedOnStartup(ctx)
```

把历史 `alert_status=pending` 标记为 failed，错误信息为“Admin 重启，告警发送结果未知”。

- [ ] **步骤 9：运行聚焦测试**

```bash
cd chronoFlow-admin
go test ./internal/biz ./internal/worker -run 'Test.*Alert|Test.*Recovery|Test.*Maintenance' -count=1
```

预期：通过。

- [ ] **步骤 10：运行后端完整测试**

```bash
cd chronoFlow-admin
go test ./internal/... -count=1
```

预期：通过。

- [ ] **步骤 11：提交**

```bash
git add chronoFlow-admin/internal chronoFlow-admin/cmd/chronoFlow-admin
git commit -m "feat: trigger alerts from failed jobs"
```

---

## 任务 6：前端任务告警开关

**文件：**

- 修改：`chronoFlow-ui/src/types/job.ts`
- 修改：`chronoFlow-ui/src/api/jobs.ts`
- 修改：`chronoFlow-ui/src/stores/jobs.ts`
- 修改：`chronoFlow-ui/src/views/jobs/JobListView.vue`

- [ ] **步骤 1：增加前端类型**

在 `chronoFlow-ui/src/types/job.ts` 中给 `Job`、`CreateJobPayload`、`UpdateJobPayload` 增加：

```ts
failureAlertEnabled: boolean
```

- [ ] **步骤 2：更新 API 映射**

在 `chronoFlow-ui/src/api/jobs.ts` 中映射后端字段：

```ts
failureAlertEnabled: item.failure_alert_enabled ?? false
```

创建/更新时发送：

```ts
failure_alert_enabled: payload.failureAlertEnabled,
```

- [ ] **步骤 3：更新 store 默认值**

在 `chronoFlow-ui/src/stores/jobs.ts` 中给新建任务默认值增加：

```ts
failureAlertEnabled: false,
```

- [ ] **步骤 4：更新任务表单**

在 `chronoFlow-ui/src/views/jobs/JobListView.vue` 中增加表单项：

```vue
<a-form-item label="失败告警" name="failureAlertEnabled">
  <a-switch v-model:checked="formState.failureAlertEnabled" />
  <div class="form-help">任务执行失败或超时时发送飞书告警；需先在系统设置中配置飞书 Webhook。</div>
</a-form-item>
```

编辑任务时从当前任务初始化该值。

- [ ] **步骤 5：更新任务列表**

在任务列表“说明”前新增“失败告警”列。

开启时显示：

```vue
<a-tag color="blue">开启</a-tag>
```

关闭时显示：

```vue
<a-tag>关闭</a-tag>
```

如果此时已能查询系统设置，则在 Webhook 未配置时展示提示：

```text
系统设置未配置飞书 Webhook，失败时不会发送。
```

- [ ] **步骤 6：构建前端**

```bash
cd chronoFlow-ui
npm run build
```

预期：通过。

- [ ] **步骤 7：提交**

```bash
git add chronoFlow-ui/src/types/job.ts chronoFlow-ui/src/api/jobs.ts chronoFlow-ui/src/stores/jobs.ts chronoFlow-ui/src/views/jobs/JobListView.vue
git commit -m "feat: add job alert switch ui"
```

---

## 任务 7：前端系统设置页面

**文件：**

- 新增：`chronoFlow-ui/src/types/systemSettings.ts`
- 新增：`chronoFlow-ui/src/api/systemSettings.ts`
- 修改：`chronoFlow-ui/src/views/settings/SettingsView.vue`
- 修改：`chronoFlow-ui/src/views/jobs/JobListView.vue`

- [ ] **步骤 1：增加系统设置类型**

新增 `chronoFlow-ui/src/types/systemSettings.ts`：

```ts
export interface AlertSettings {
  feishuWebhookConfigured: boolean
  feishuWebhookUpdatedAt: string
}

export interface SaveFeishuWebhookPayload {
  webhook: string
}
```

- [ ] **步骤 2：增加系统设置 API**

新增 `chronoFlow-ui/src/api/systemSettings.ts`：

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

- [ ] **步骤 3：替换系统设置占位页**

重写 `chronoFlow-ui/src/views/settings/SettingsView.vue`，包含：

- 页面标题“系统设置”
- 飞书告警配置区域
- 状态标签：已配置 / 未配置
- 更新时间
- 密码输入框
- 保存按钮
- 测试发送按钮
- 清空配置按钮和确认弹窗
- 飞书机器人配置说明
- V1 不支持签名 Secret 的说明

复用现有风格：`PageHeaderBar`、`a-card`、`a-alert`、`a-space`。

- [ ] **步骤 4：任务列表使用系统设置状态**

在 `JobListView.vue` 中加载 `getAlertSettings()`，用于已开启告警但 Webhook 未配置时展示更准确提示。

- [ ] **步骤 5：构建前端**

```bash
cd chronoFlow-ui
npm run build
```

预期：通过。

- [ ] **步骤 6：提交**

```bash
git add chronoFlow-ui/src/types/systemSettings.ts chronoFlow-ui/src/api/systemSettings.ts chronoFlow-ui/src/views/settings/SettingsView.vue chronoFlow-ui/src/views/jobs/JobListView.vue
git commit -m "feat: add alert settings ui"
```

---

## 任务 8：前端日志详情展示告警状态

**文件：**

- 修改：`chronoFlow-ui/src/types/jobLog.ts`
- 修改：`chronoFlow-ui/src/api/jobLogs.ts`
- 修改：`chronoFlow-ui/src/stores/jobLogs.ts`
- 修改：`chronoFlow-ui/src/views/logs/JobLogDetailView.vue`

- [ ] **步骤 1：增加日志告警字段类型**

在 `chronoFlow-ui/src/types/jobLog.ts` 中增加：

```ts
alertEnabledSnapshot: boolean
alertStatus: 'none' | 'pending' | 'sent' | 'failed' | 'skipped' | ''
alertError: string
alertSentAt: string
```

- [ ] **步骤 2：映射 API 字段**

在 `chronoFlow-ui/src/api/jobLogs.ts` 中映射：

```ts
alertEnabledSnapshot: Boolean(item.alert_enabled_snapshot),
alertStatus: item.alert_status || 'none',
alertError: item.alert_error || '',
alertSentAt: item.alert_sent_at || '',
```

- [ ] **步骤 3：增加展示辅助函数**

在 `JobLogDetailView.vue` 中增加：

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

- [ ] **步骤 4：日志详情页面展示失败告警**

在执行状态/错误信息附近增加：

```vue
<a-descriptions-item label="失败告警">
  <a-tag :color="alertTagColor(detail.log)">{{ alertStatusText(detail.log) }}</a-tag>
</a-descriptions-item>
```

颜色建议：

- `sent`：green
- `pending` / `skipped`：orange
- `failed`：red
- `none`：default

- [ ] **步骤 5：构建前端**

```bash
cd chronoFlow-ui
npm run build
```

预期：通过。

- [ ] **步骤 6：提交**

```bash
git add chronoFlow-ui/src/types/jobLog.ts chronoFlow-ui/src/api/jobLogs.ts chronoFlow-ui/src/stores/jobLogs.ts chronoFlow-ui/src/views/logs/JobLogDetailView.vue
git commit -m "feat: show alert status in logs"
```

---

## 任务 9：文档和测试指南

**文件：**

- 修改：`README.md`
- 修改：`README.en.md`
- 修改：`deploy/README.md`
- 修改：`docs/TESTING_GUIDE.md`

- [ ] **步骤 1：更新根 README**

`README.md` 功能表增加：

```markdown
| 飞书失败告警 | 在系统设置中配置飞书 Webhook，任务失败或超时时发送卡片告警。 |
```

`README.en.md` 增加：

```markdown
| Feishu failure alerts | Configure a Feishu webhook in System Settings and send card alerts when jobs fail or time out. |
```

- [ ] **步骤 2：更新部署文档**

在 `deploy/README.md` 中增加“飞书失败告警”章节，说明：

1. 飞书群进入群设置。
2. 添加自定义机器人。
3. 复制机器人 Webhook。
4. 登录 ChronoFlow，进入“系统设置”。
5. 粘贴 Webhook 并保存。
6. 点击“测试发送”确认群里能收到卡片。

同时说明：

- V1 不支持飞书签名 Secret。
- 如需安全策略，建议使用关键词校验或先不启用签名校验。
- 任务失败判断依赖进程退出码，不解析日志正文。
- Glue Shell 调用 Python 时推荐 `set -euo pipefail`。

- [ ] **步骤 3：更新测试指南**

在 `docs/TESTING_GUIDE.md` 中增加：

```markdown
### TC-ALERT-001 保存飞书 Webhook
进入系统设置，填写飞书 Webhook 并保存。
预期：页面显示已配置，不回显 Webhook 明文。

### TC-ALERT-002 测试发送
点击测试发送。
预期：飞书群收到 ChronoFlow 测试卡片。

### TC-ALERT-003 failed 任务发送告警
创建 Glue Shell：`set -euo pipefail; exit 1`
开启失败告警并运行。
预期：任务日志 failed，飞书群收到失败卡片，日志详情显示告警已发送。

### TC-ALERT-004 timeout 任务发送告警
创建超时任务。
开启失败告警并运行。
预期：任务日志 timeout，飞书群收到超时卡片。

### TC-ALERT-005 Webhook 未配置
清空 Webhook。
运行开启失败告警的失败任务。
预期：不发送飞书，日志详情显示未发送：系统未配置飞书 Webhook。
```

- [ ] **步骤 4：提交**

```bash
git add README.md README.en.md deploy/README.md docs/TESTING_GUIDE.md
git commit -m "docs: document failure alerts"
```

---

## 任务 10：完整验证和发布前检查

**文件：**

- 默认不需要改源文件，除非验证发现问题。

- [ ] **步骤 1：运行后端测试**

```bash
cd chronoFlow-admin
go test ./internal/... -count=1
```

预期：通过。

- [ ] **步骤 2：运行前端构建**

```bash
cd chronoFlow-ui
npm run build
```

预期：通过。

- [ ] **步骤 3：最终重新生成 API 和 wire**

```bash
cd chronoFlow-admin
make api
make wire
go test ./internal/... -count=1
```

预期：生成文件没有意外差异，测试通过。

- [ ] **步骤 4：检查 Git 状态**

```bash
git status --short
git diff --stat
```

预期：没有意外未跟踪文件。所有应提交的生成文件都已提交。

- [ ] **步骤 5：本地手动冒烟测试**

按现有部署文档启动本地环境，验证：

1. 登录正常。
2. 系统设置页面可打开。
3. 保存 Webhook 后显示已配置。
4. 测试发送能到飞书。
5. failed 任务能发送卡片。
6. 日志详情显示已发送。
7. 清空 Webhook。
8. failed 任务不再发送卡片，日志详情显示 skipped 原因。

- [ ] **步骤 6：如验证产生修复则提交**

如果验证过程中产生修复：

```bash
git add <changed-files>
git commit -m "fix: stabilize failure alert flow"
```

如果没有改动，不创建空提交。

---

## 自检

- 需求覆盖：本计划覆盖全局加密飞书 Webhook、系统设置页面、任务告警开关、任务列表展示、日志详情告警状态、failed/timeout 触发、异步发送、重试、pending 清理、不支持 Secret、不包含跳转链接、不解析日志正文、文档更新。
- 占位检查：本计划不包含未解决占位符。每个任务都包含明确文件、代码片段、命令和预期结果。
- 命名一致性：计划中统一使用 `failure_alert_enabled`、`FailureAlertEnabled`、`alert_enabled_snapshot`、`alert_status`、`alert_error`、`alert_sent_at`、`SystemSettingUsecase`、`AlertUsecase`、`FeishuAlertSender`。
