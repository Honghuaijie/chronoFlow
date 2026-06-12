# ChronoFlow Admin

调度器后端，负责 ChronoFlow 的任务管理、执行器管理、Cron 调度、执行日志元数据、日志文件存储、执行器健康检查、启动恢复和内部 callback 接口。

## 职责边界

- 连接 MySQL，并拥有 ChronoFlow 的业务表。
- 保存执行器 token 的加密密文。
- 调用执行器 `/run`、`/kill`、`/health`。
- 接收执行器异步 callback。
- 保存完整日志正文到本地文件目录，MySQL 只保存日志元数据。
- 提供前端访问的 `/v1/public/*` 和 `/v1/admin/*` HTTP API。

不负责：

- 直接执行 Shell 脚本。
- 连接执行器本地 pending callback 文件。
- 执行器进程管理。

## 默认端口

| 协议 | 地址 |
| --- | --- |
| HTTP | `0.0.0.0:10003` |
| gRPC | `0.0.0.0:11003` |

## 前置依赖

- Go 1.22 或更高版本。
- MySQL 8.x 或兼容版本。

默认数据库配置在 `configs/config.yaml`：

```yaml
database: chronoflow
username: root
password: root
```

启动前创建数据库：

```sql
CREATE DATABASE IF NOT EXISTS chronoflow DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## 本地启动

```bash
go run ./cmd/chronoFlow-admin -conf ./configs
```

也可以使用模板 Makefile：

```bash
make run env=local
```

默认管理员账号：

```text
admin / admin123
```

## 关键配置

- `server.public_base_url`：执行器 callback 访问调度器的公开地址。
- `security.jwt_secret`：JWT 签名密钥。
- `security.token_encrypt_key`：执行器 token 加密密钥，必须是 32 字节。
- `security.callback_token`：执行器回调使用的全局 token。
- `security.admin_username` / `security.admin_password`：内置管理员账号。
- `logs.data_dir`：日志正文文件目录。
- `logs.max_log_bytes`：单次日志最大保存大小，默认 5MB。
- `logs.retention_days`：日志保留天数，默认 30 天。
- `scheduler.timezone`：Cron 时区，默认 `Asia/Shanghai`。
- `executor.health_check_interval_seconds`：执行器健康检查间隔。
- `recovery.startup_running_grace_seconds`：Admin 启动后等待恢复窗口。
- `recovery.killing_timeout_seconds`：`killing` 状态超时时间。

配置文件中的 `${ENV:default}` 支持环境变量覆盖。

## 主要接口

公共接口：

- `POST /v1/public/auth/login`

后台接口：

- `GET /v1/admin/auth/current`
- `/v1/admin/executors/*`
- `/v1/admin/jobs/*`
- `/v1/admin/glues/*`
- `/v1/admin/jobLogs/*`

内部接口：

- 执行器 callback 接口由执行器调用，使用 `X-Callback-Token` 鉴权。

## 验证命令

```bash
go test ./internal/... -count=1
go build -o /tmp/chronoflow-admin-build ./cmd/chronoFlow-admin
```

## 开发注意事项

- 模板中的 user 示例接口只作为写法参考，不属于 ChronoFlow 业务代码。
- 不要把完整日志正文写入 MySQL。
- 执行器 token 入库前必须加密。
- 同一个任务运行中时，手动运行应返回“任务正在执行中”，不创建新日志。
- 任务配置允许编辑，当前运行实例不受影响，新配置下次执行生效。
