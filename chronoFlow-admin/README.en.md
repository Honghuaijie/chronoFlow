# ChronoFlow Admin

Scheduler backend for ChronoFlow. It owns job management, executor management, Cron scheduling, job log metadata, log file storage, executor health checks, startup recovery, and internal callback APIs.

## Responsibilities

- Connect to MySQL and own ChronoFlow business tables.
- Store encrypted executor tokens.
- Call executor `/run`, `/kill`, and `/health`.
- Receive asynchronous callbacks from executors.
- Store full log content as local files. MySQL stores metadata only.
- Provide `/v1/public/*` and `/v1/admin/*` HTTP APIs for the UI.

Out of scope:

- Running Shell scripts directly.
- Managing executor pending callback files.
- Managing executor processes.

## Default Ports

| Protocol | Address |
| --- | --- |
| HTTP | `0.0.0.0:10003` |
| gRPC | `0.0.0.0:11003` |

## Prerequisites

- Go 1.22 or later.
- MySQL 8.x or compatible.

Default database config is in `configs/config.yaml`:

```yaml
database: chronoflow
username: root
password: root
```

Create the database before starting:

```sql
CREATE DATABASE IF NOT EXISTS chronoflow DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## Local Startup

```bash
go run ./cmd/chronoFlow-admin -conf ./configs
```

Or use the template Makefile:

```bash
make run env=local
```

Default admin account:

```text
admin / admin123
```

## Key Configuration

- `server.public_base_url`: public URL used by executors for callbacks.
- `security.jwt_secret`: JWT signing secret.
- `security.token_encrypt_key`: executor token encryption key. Must be 32 bytes.
- `security.callback_token`: global token for executor callbacks.
- `security.admin_username` / `security.admin_password`: built-in admin account.
- `logs.data_dir`: directory for log content files.
- `logs.max_log_bytes`: max log content per run. Default is 5 MB.
- `logs.retention_days`: log retention days. Default is 30.
- `scheduler.timezone`: Cron timezone. Default is `Asia/Shanghai`.
- `executor.health_check_interval_seconds`: executor health check interval.
- `recovery.startup_running_grace_seconds`: startup recovery grace window.
- `recovery.killing_timeout_seconds`: timeout for `killing` logs.

`${ENV:default}` placeholders can be overridden with environment variables.

## Main APIs

Public:

- `POST /v1/public/auth/login`

Admin:

- `GET /v1/admin/auth/current`
- `/v1/admin/executors/*`
- `/v1/admin/jobs/*`
- `/v1/admin/glues/*`
- `/v1/admin/jobLogs/*`

Internal:

- Executor callback endpoint, authenticated by `X-Callback-Token`.

## Verification

```bash
go test ./internal/... -count=1
go build -o /tmp/chronoflow-admin-build ./cmd/chronoFlow-admin
```

## Development Notes

- Template user APIs are examples only and are not ChronoFlow business code.
- Do not store full log content in MySQL.
- Executor tokens must be encrypted before storage.
- If the same job is already running, manual run should return “任务正在执行中” and must not create a new log.
- Job configuration can be edited while a job is running. The running instance is unaffected; the new config applies next time.
