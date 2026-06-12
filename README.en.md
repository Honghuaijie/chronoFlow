# ChronoFlow

ChronoFlow is a lightweight internal scheduled job platform for a single team. V1 targets dozens of jobs, one scheduler, Linux executors, asynchronous callbacks, file-based log content, and a basic web console.

## Project Layout

```text
chronoFlow/
├── chronoFlow-admin/   # Scheduler backend. Owns MySQL, jobs, executors, log metadata, and scheduling.
├── chronoFlow-exec/    # Executor backend. No database. Runs Shell, kills process groups, and callbacks results.
├── chronoFlow-ui/      # Admin console. Vue 3 + Ant Design Vue.
├── prd-v1.md           # V1 product requirements.
├── task_plan.md        # Project plan.
├── progress.md         # Development and integration progress.
└── findings.md         # Notes, findings, and decisions.
```

## V1 Features

- Executor management: create, edit, delete, and heartbeat status.
- Job management: create, edit, delete, start scheduling, stop scheduling, and manual run.
- Glue Shell: save a Shell script for each job. The script may call Python scripts mounted on the host or inside a container.
- Asynchronous execution: Admin dispatches a run request and returns immediately; Exec callbacks Admin after completion.
- Per-job mutual exclusion: the same job cannot run concurrently; different jobs may run in parallel.
- Kill running jobs: Admin asks Exec to kill the running process group. Log status moves from `running` to `killing`, then to `killed` or `failed`.
- Log storage: MySQL stores metadata only. Full log content is stored as files on the Admin side.
- Log console: list, filter, detail, Glue snapshot, and log content viewer.
- Lightweight auth: a config-based admin account and JWT for `/v1/admin/*`.

## Architecture

```text
UI -> Admin -> Exec
       ^        |
       |        v
       +-- callback
```

- `chronoFlow-admin` is the only service that connects to MySQL.
- `chronoFlow-exec` does not connect to MySQL and must not read or write Admin database tables.
- Admin calls Exec with an executor-specific `X-Executor-Token`.
- Exec callbacks Admin with the global `X-Callback-Token`.
- If callback fails, Exec writes the pending callback to disk and keeps retrying in the background. Default retention is 7 days.

## Default Ports

| Service | HTTP | gRPC |
| --- | --- | --- |
| Admin | `10003` | `11003` |
| Exec | `10004` | `11004` |
| UI | `5173`; Vite chooses another port if occupied | - |

## Prerequisites

- Go 1.22 or later.
- Node.js 18 or later.
- MySQL 8.x or a compatible database.
- Linux for production executor process-group semantics. Basic HTTP integration can be tested locally on macOS.

Default Admin MySQL config:

```yaml
host: 127.0.0.1
port: 3306
username: root
password: root
database: chronoflow
```

Create the database first:

```sql
CREATE DATABASE IF NOT EXISTS chronoflow DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## Local Startup

### 1. Start Admin

```bash
cd chronoFlow-admin
go run ./cmd/chronoFlow-admin -conf ./configs
```

Default admin account:

```text
username: admin
password: admin123
```

### 2. Start Exec

```bash
cd chronoFlow-exec
go run ./cmd/chronoFlow-exec -conf ./configs
```

Default executor token:

```text
change-me
```

### 3. Start UI

```bash
cd chronoFlow-ui
npm install
npm run dev
```

The UI proxies API requests to:

```text
http://127.0.0.1:10003
```

Override it when needed:

```bash
VITE_API_PROXY_TARGET=http://127.0.0.1:10003 npm run dev
```

## Minimal Integration Flow

1. Open the UI: `http://127.0.0.1:5173/`. If the port is occupied, use the URL printed by Vite.
2. Log in with `admin / admin123`.
3. Create an executor:
   - Name: `local-exec`
   - Address: `http://127.0.0.1:10004`
   - Token: `change-me`
4. Create a job and select that executor.
5. Open Glue editor for the job and save:

```bash
echo chronoflow-demo-start
pwd
echo chronoflow-demo-done
```

6. Click Run.
7. Open the job log detail page and verify that status is `success` and log content is visible.

## Verification Commands

Admin:

```bash
cd chronoFlow-admin
go test ./internal/... -count=1
go build -o /tmp/chronoflow-admin-build ./cmd/chronoFlow-admin
```

Exec:

```bash
cd chronoFlow-exec
go test ./internal/... -count=1
go build -o /tmp/chronoflow-exec-build ./cmd/chronoFlow-exec
```

UI:

```bash
cd chronoFlow-ui
npm run build
```

## Key Configuration

Admin:

- `security.admin_username` / `security.admin_password`: built-in admin account.
- `security.jwt_secret`: JWT secret.
- `security.token_encrypt_key`: executor token encryption key. Must be 32 bytes.
- `security.callback_token`: global token for Exec-to-Admin callbacks.
- `logs.data_dir`: directory for full log content files.
- `logs.retention_days`: retention days for log metadata and files. Default is 30.
- `scheduler.timezone`: Cron timezone. Default is `Asia/Shanghai`.

Exec:

- `executor.token`: token used by Admin when calling Exec.
- `executor.data_dir`: directory for pending callback files.
- `executor.shell_path`: Shell path. Default is `/bin/bash`.
- `executor.max_log_bytes`: max log content per run. Default is 5 MB.
- `callback.retry_interval_seconds`: pending callback retry interval.
- `callback.pending_retention_days`: pending callback retention days. Default is 7.

Environment variables override placeholders in config files.

## Docker and Script Mounts

V1 supports running Exec inside Docker. Put business scripts on the host and mount them into the executor container through a Docker volume.

Example:

```bash
docker run --rm \
  -p 10004:10004 \
  -v /opt/chronoflow/scripts:/scripts \
  -v /opt/chronoflow/exec-data:/app/data \
  chronoFlow-exec:latest
```

Glue Shell can call mounted Python scripts:

```bash
python3 /scripts/report.py
```

## Notes

- Real process-group kill semantics are Linux-only.
- Do not store full log content in MySQL. MySQL stores metadata only.
- Exec does not need a database and should not connect to Admin's database.
- Template user APIs are examples only and are not ChronoFlow business APIs.
