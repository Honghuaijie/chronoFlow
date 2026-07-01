# ChronoFlow 任务失败告警设计

## 背景

ChronoFlow 第一版已经完成任务调度、执行器、Glue Shell、执行日志和运行报表等核心能力。后续需要新增任务失败告警能力，让用户在任务执行失败或超时时收到飞书群通知。

本设计面向 V1 轻量实现，继续保持 ChronoFlow 的定位：

- 内网单团队使用
- 单调度器
- 几十个以内任务
- 轻量部署和维护
- 不引入复杂多通道告警中心

## 目标

1. 支持全局配置一个飞书自定义机器人 Webhook。
2. 支持任务级别开启或关闭失败告警。
3. 任务最终状态为 `failed` 或 `timeout` 时发送飞书卡片告警。
4. 在任务日志详情中展示本次告警发送结果。
5. Webhook 加密存储，不在前端回显明文。
6. 告警发送不阻塞执行器 callback。

## 非目标

V1 不做以下能力：

- 邮件告警
- 多个告警通道
- 每个任务选择不同告警通道
- 飞书签名 Secret
- 告警降噪或冷却
- 告警记录独立列表页
- 告警可靠队列或持久化补发
- 飞书消息中的日志详情跳转链接
- 根据日志正文内容判断任务是否失败

## 失败判断规则

ChronoFlow 不解析日志内容判断失败。任务是否失败只以执行器返回的最终状态为准。

执行器根据进程退出码和运行控制结果生成最终状态：

| 条件 | 最终状态 |
| --- | --- |
| `exit_code == 0` | `success` |
| `exit_code != 0` | `failed` |
| 执行超时 | `timeout` |
| 用户主动终止 | `killed` |

如果 Glue Shell 中调用 Python 脚本，推荐写法：

```bash
#!/bin/bash
set -euo pipefail

python3 /scripts/report.py
```

如果脚本主动吞掉异常，例如：

```bash
python3 /scripts/report.py || true
```

则 Shell 最终可能返回 `0`，ChronoFlow 会认为任务执行成功。这属于用户脚本自身语义，不由告警功能额外识别。

## 配置与前端交互

### 系统设置

前端新增一级菜单：

```text
系统设置
```

系统设置页用于配置全局飞书 Webhook。

页面展示：

- 飞书 Webhook 配置状态：已配置 / 未配置
- 上次更新时间
- Webhook 输入框
- 保存按钮
- 测试发送按钮
- 清空配置按钮

交互规则：

1. Webhook 输入框使用密码框样式。
2. 输入框支持眼睛按钮临时查看当前输入内容。
3. Webhook 保存后不回显明文。
4. 页面提示用户自行保存 Webhook 原文。
5. 保存时只校验非空和 URL 格式，不主动请求飞书。
6. 测试发送必须先保存 Webhook，再测试当前已保存配置。
7. 清空配置前弹出确认。
8. 清空 Webhook 后，状态变为未配置，更新时间更新为清空时间。

系统设置页提供简短说明：

- 飞书群 -> 群设置 -> 机器人 -> 添加自定义机器人 -> 复制 Webhook。
- V1 只支持普通飞书 Webhook。
- V1 不支持飞书签名校验 Secret。
- 如果需要机器人安全策略，建议使用关键词校验，或先不要启用签名校验。

### 任务表单

任务创建和编辑表单新增字段：

```text
失败告警
```

默认关闭。

字段说明：

```text
任务执行失败或超时时发送飞书告警；需先在系统设置中配置飞书 Webhook。
```

任务运行期间编辑告警开关不影响当前运行，只影响下次运行。

### 任务列表

任务列表新增“失败告警”字段，位置放在“说明”前。

展示规则：

- 未开启：显示“关闭”。
- 已开启：显示“开启”。
- 已开启但全局 Webhook 未配置：仍显示“开启”，并提示“系统设置未配置飞书 Webhook，失败时不会发送”。

### 日志详情

日志详情展示本次告警发送结果。

可能状态：

- 未启用
- 未发送：任务未开启失败告警
- 未发送：系统未配置飞书 Webhook
- 发送中
- 已发送
- 发送失败：具体错误

## 触发规则

告警只在任务日志最终状态变为以下状态时触发：

```text
failed
timeout
```

以下状态不触发：

```text
success
killed
skipped
```

同一个 `log_id` 最多发送一次告警。重复 callback 不重复发送。

任务执行结果和告警发送结果完全分离。告警发送失败不会改变任务日志的最终状态。

## 运行快照

任务开始运行时，需要把任务当时的失败告警开关写入本次 `job_logs`。

原因：

- 当前运行不受后续任务编辑影响。
- 行为与 Glue Shell 快照一致。
- 用户在任务运行期间关闭告警，只影响下次运行。

快照字段建议：

```text
alert_enabled_snapshot
```

## 告警发送流程

告警由 Admin 发送，Exec 不感知飞书配置。

典型流程：

1. Exec 执行任务。
2. Exec callback Admin。
3. Admin 更新 `job_logs` 最终状态。
4. 如果最终状态为 `failed` 或 `timeout`，并且本次运行快照启用了失败告警，则把 `alert_status` 标记为 `pending`。
5. Admin 启动后台 goroutine 异步发送飞书卡片。
6. callback 接口立即返回成功给 Exec，不等待飞书请求。
7. 发送成功后更新 `alert_status=sent` 和 `alert_sent_at`。
8. 发送失败时最多重试 3 次，每次间隔 2 秒。
9. 最终仍失败则更新 `alert_status=failed` 和 `alert_error`。

如果任务开启告警但 Webhook 未配置：

```text
alert_status = skipped
alert_error = 系统未配置飞书 Webhook
```

如果 Admin 重启恢复时把遗留 `running` / `killing` 日志标记为 `failed`，也触发失败告警。告警文案中应体现“执行结果未知”。

如果 Admin 在 `alert_status=pending` 时重启，V1 不补发。Admin 启动时把历史 pending 标记为：

```text
alert_status = failed
alert_error = Admin 重启，告警发送结果未知
```

## 飞书卡片内容

飞书消息使用卡片，不使用纯文本。

标题按状态区分：

| 状态 | 标题 |
| --- | --- |
| `failed` | `ChronoFlow 任务执行失败` |
| `timeout` | `ChronoFlow 任务执行超时` |

卡片字段：

- 任务名称
- 执行器名称
- 日志 ID
- 状态：`failed` / `timeout`
- 开始时间
- 结束时间
- 耗时
- Exit Code
- 错误信息

卡片不包含：

- 日志正文
- 查看详情按钮
- 前端跳转链接

错误信息最多显示 500 字符，超出截断并追加 `...`。完整日志仍在 ChronoFlow 日志详情中查看。

时间使用系统调度时区，默认 `Asia/Shanghai`。

测试发送使用同样的卡片样式，但标题为：

```text
ChronoFlow 飞书告警测试
```

测试卡片需明确标注：

```text
这是一条测试消息，不代表任务失败。
```

## 数据结构

### jobs

新增字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `failure_alert_enabled` | boolean | 是否开启失败告警，默认 false |

### job_logs

新增字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `alert_enabled_snapshot` | boolean | 本次运行是否启用失败告警 |
| `alert_status` | varchar(32) | `none` / `pending` / `sent` / `failed` / `skipped` |
| `alert_error` | text | 失败或跳过原因 |
| `alert_sent_at` | datetime nullable | 告警发送成功时间 |

`alert_status` 语义：

| 状态 | 说明 |
| --- | --- |
| `none` | 未启用告警，或最终状态不需要告警 |
| `pending` | 等待发送或正在发送 |
| `sent` | 已发送 |
| `failed` | 发送失败 |
| `skipped` | 跳过发送，例如 Webhook 未配置 |

### system_settings

新增通用系统设置表，用于保存全局配置。

建议字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | bigint | 主键 |
| `setting_key` | varchar(128) | 设置 key，唯一 |
| `value_encrypted` | text nullable | 加密后的值 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |
| `deleted` | boolean | 软删除标记，沿用现有模型风格 |

飞书 Webhook key：

```text
alert.feishu.webhook
```

Webhook 使用现有 `CHRONOFLOW_TOKEN_ENCRYPT_KEY` 加密。

清空 Webhook 时不删除记录，只把 `value_encrypted` 置空，并更新 `updated_at`。

## API 设计

底层表使用通用 `system_settings`，但 API 使用明确的系统设置接口，不做过度抽象。

建议接口：

```text
GET    /v1/system/settings/alert
PUT    /v1/system/settings/alert/feishu
POST   /v1/system/settings/alert/feishu/test
DELETE /v1/system/settings/alert/feishu
```

### GET /v1/system/settings/alert

返回：

```json
{
  "feishu_webhook_configured": true,
  "feishu_webhook_updated_at": "2026-06-30 10:00:00"
}
```

不返回 Webhook 明文。

### PUT /v1/system/settings/alert/feishu

请求：

```json
{
  "webhook": "https://open.feishu.cn/open-apis/bot/v2/hook/..."
}
```

后端只校验非空和 URL 格式，不强制校验飞书域名。

### POST /v1/system/settings/alert/feishu/test

使用已保存 Webhook 发送测试卡片。

如果 Webhook 未配置，返回明确错误。

### DELETE /v1/system/settings/alert/feishu

清空已保存 Webhook。

## 错误处理

1. Webhook 未配置时，不发送告警，日志记录 `alert_status=skipped`。
2. 飞书返回非 2xx 或业务错误时，记录到 `alert_error`。
3. 网络错误或超时最多重试 3 次，每次间隔 2 秒。
4. 告警错误信息不影响任务最终状态。
5. Admin 启动时把历史 `alert_status=pending` 标记为 failed，避免状态长期停留在 pending。

## 测试范围

### 后端

- 保存 Webhook 时加密存储。
- 查询系统设置时不回显 Webhook。
- 清空 Webhook 后状态为未配置。
- 测试发送在 Webhook 未配置时返回错误。
- 任务创建默认关闭失败告警。
- 任务运行时写入 `alert_enabled_snapshot`。
- `failed` 和 `timeout` 触发告警。
- `success`、`killed`、`skipped` 不触发告警。
- 同一个 `log_id` 不重复发送告警。
- 告警失败重试 3 次后记录 `alert_status=failed`。
- Webhook 未配置时记录 `alert_status=skipped`。
- Admin 启动时把历史 pending 标记为 failed。

### 前端

- 系统设置菜单可访问。
- Webhook 保存后只显示已配置，不显示明文。
- Webhook 输入框支持显示/隐藏当前输入。
- 清空配置前有确认。
- 测试发送按钮在未配置时给出明确提示。
- 任务创建/编辑有失败告警开关，默认关闭。
- 任务列表在说明前展示失败告警开/关。
- Webhook 未配置时，已开启告警的任务显示提示。
- 日志详情展示告警状态和错误信息。

### 联调

- 配置飞书 Webhook 后，失败任务能收到飞书卡片。
- 超时任务能收到飞书卡片。
- 成功任务不发送飞书卡片。
- Webhook 清空后，任务失败不发送卡片，日志显示未配置原因。

## 文档更新

README 只做简短说明：

- 支持任务失败飞书告警。
- 在系统设置中配置飞书 Webhook。

`deploy/README.md` 增加详细说明：

- 如何创建飞书自定义机器人。
- 如何填写 Webhook。
- V1 不支持签名 Secret。
- 测试发送方式。
- 任务失败判断依赖 exit code，推荐 Glue Shell 使用 `set -euo pipefail`。

## 分支与开发流程

该功能在独立分支开发，建议分支名：

```text
feat/alarm
```

开发完成并通过测试后，再合并回 `master`。
