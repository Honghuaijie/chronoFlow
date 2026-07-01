<h1 align="center">ChronoFlow</h1>

<p align="center">
  A lightweight internal scheduled job platform for Shell / Python scripts and small-team operations.
</p>

<p align="center">
  <a href="README.md">简体中文</a>
  ·
  <a href="deploy/README.md">Deployment Guide</a>
  ·
  <a href="docs/TESTING_GUIDE.md">Testing Guide</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white" alt="Go 1.22">
  <img src="https://img.shields.io/badge/Vue-3-42b883?logo=vuedotjs&logoColor=white" alt="Vue 3">
  <img src="https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white" alt="Docker Compose">
  <img src="https://img.shields.io/badge/MySQL-8.0-4479A1?logo=mysql&logoColor=white" alt="MySQL 8.0">
</p>

<p align="center">
  <img src="docs/images/chronoflow-jobs.png" alt="ChronoFlow jobs" width="920">
</p>

## What is ChronoFlow?

ChronoFlow is a lightweight scheduled job platform for internal single-team use. It includes a scheduler backend, an executor backend, and a web console for Cron schedules, manual runs, Glue Shell scripts, async callbacks, kill operations, execution logs, and runtime reports.

It is designed for teams that currently manage many Shell / Python scripts with crontab but want:

- A visible job list with upcoming run times
- Manual run, pause, resume, and kill operations
- Centralized stdout / stderr logs and failure reasons
- A simple web console for non-ops teammates
- A lighter alternative to large distributed scheduling systems

## Features

| Capability | Description |
| --- | --- |
| Executor management | Create, edit, delete, and monitor executor heartbeat status. |
| Job management | Create, edit, start scheduling, stop scheduling, and manually run jobs. |
| Visual Cron picker | Configure minute, hour, day, week, and month schedules, or write a manual expression. |
| Glue Shell | Store one Shell script per job; scripts can call mounted Python or other local files. |
| Async callback | Admin dispatches a run request and returns immediately; Exec callbacks Admin after completion. |
| Per-job mutual exclusion | The same job cannot run concurrently, while different jobs can run in parallel. |
| Kill operation | Admin asks Exec to kill the process group, useful when Shell starts Python subprocesses. |
| File-based logs | MySQL stores metadata only; full log content is stored as files. |
| Runtime report | View job count, run count, executor count, success rate, and recent 7-day trends. |
| Feishu failure alerts | Configure a Feishu webhook in System Settings and send card alerts when jobs fail or time out. |

## Screenshots

| Login | Executors |
| --- | --- |
| <img src="docs/images/chronoflow-login.png" alt="Login" width="420"> | <img src="docs/images/chronoflow-executors.png" alt="Executors" width="620"> |

| Cron Picker | Glue Shell |
| --- | --- |
| <img src="docs/images/chronoflow-cron-picker.png" alt="Cron picker" width="520"> | <img src="docs/images/chronoflow-glue-shell.png" alt="Glue Shell" width="520"> |

| Log Detail | Runtime Report |
| --- | --- |
| <img src="docs/images/chronoflow-log-detail.png" alt="Log detail" width="620"> | <img src="docs/images/chronoflow-report.png" alt="Runtime report" width="620"> |

## Quick Start

ChronoFlow supports two Docker deployment modes:

- **Prebuilt-image deployment**: for servers where you do not want to pull the full source code.
- **Source-build deployment**: for developers who want to modify code and build images locally.

### Option 1: Prebuilt-Image Deployment

The server only needs the files under `deploy`; it does not need the full source code.

```bash
cd deploy
cp .env.example .env
```

Published images are available on [GitHub Packages](https://github.com/Honghuaijie?tab=packages).

Fixed image versions are recommended because they are easier to roll back and troubleshoot:

```env
CHRONOFLOW_ADMIN_IMAGE=ghcr.io/honghuaijie/chronoflow-admin:v0.1.3
CHRONOFLOW_EXEC_IMAGE=ghcr.io/honghuaijie/chronoflow-exec:v0.1.3
CHRONOFLOW_UI_IMAGE=ghcr.io/honghuaijie/chronoflow-ui:v0.1.3
```

If you want to use the bundled MySQL service:

```bash
docker compose -f docker-compose.mysql.yml up -d
```

Start the application:

```bash
docker compose -f docker-compose.image.yml up -d
```

### Option 2: Source-Build Deployment

```bash
git clone https://github.com/Honghuaijie/chronoFlow.git chronoflow
cd chronoflow/deploy
cp .env.example .env
```

If you want to use the bundled MySQL service:

```bash
docker compose -f docker-compose.mysql.yml up -d
```

Start the application:

```bash
docker compose up -d --build
```

Open:

```text
http://127.0.0.1:5173
```

Default account:

```text
admin / admin123
```

For production, change the default admin password, JWT secret, callback token, executor token, and database password in `.env`.

For detailed deployment, ports, MySQL, external database, script mounts, and troubleshooting, see [deploy/README.md](deploy/README.md).

## First Use

### 1. Create an executor

If Admin and Exec are started by the same compose stack, create an executor in the UI with:

```text
Name: exec-default
Address: http://chronoflow-exec:10004
Token: the EXECUTOR_TOKEN value from .env
```

Do not use `http://127.0.0.1:10004`, because from the Admin container, `127.0.0.1` means the Admin container itself, not the Exec container.

### 2. Create a test job

Create a Glue Shell job to verify the full run path:

```bash
#!/bin/bash
set -e

echo "hello chronoflow"
echo "run time: $(date '+%Y-%m-%d %H:%M:%S')"
echo "hostname: $(hostname)"
python3 --version
echo "done"
```

After a manual run, the execution log should be `success` and include the script output.

### 3. Configure failure alerts

Open **System Settings**, paste the Feishu custom bot webhook, and save it. Then enable **Failure Alert** when creating or editing a job. ChronoFlow sends a Feishu card when the final job status is `failed` or `timeout`.

If Feishu keyword verification is enabled, configure the bot keyword as `ChronoFlow`. V1 does not support Feishu signature secrets. Failure detection depends on the process exit code, not log text parsing. When Glue Shell calls Python scripts, `set -euo pipefail` is recommended so Python errors produce a non-zero task exit code.

## Architecture

```text
UI -> Admin -> Exec
       ^        |
       |        v
       +-- callback
```

- `chronoFlow-admin` is the only service that connects to MySQL.
- `chronoFlow-exec` does not connect to MySQL.
- Admin calls Exec with the executor's `X-Executor-Token`.
- Exec callbacks Admin with the global `X-Callback-Token`.
- If callback fails, Exec stores the pending callback locally and retries in the background. The default retention is 7 days.

## Project Layout

```text
chronoFlow/
├── chronoFlow-admin/        # Scheduler backend. Connects to MySQL.
├── chronoFlow-exec/         # Executor backend. No database connection.
├── chronoFlow-ui/           # Web console.
├── deploy/                  # Docker Compose, env template, MySQL init, and scripts.
└── docs/                    # PRD, testing guide, development plan, and notes.
```

## Development

Run modules separately:

```bash
cd chronoFlow-admin
go run ./cmd/chronoFlow-admin -conf ./configs
```

```bash
cd chronoFlow-exec
go run ./cmd/chronoFlow-exec -conf ./configs
```

```bash
cd chronoFlow-ui
npm install
VITE_API_PROXY_TARGET=http://127.0.0.1:10003 npm run dev
```

## Verification

```bash
cd chronoFlow-admin
go test ./internal/... -count=1
```

```bash
cd chronoFlow-exec
go test ./internal/... -count=1
```

```bash
cd chronoFlow-ui
npm run build
```

For the full testing guide, see [docs/TESTING_GUIDE.md](docs/TESTING_GUIDE.md).

## Suitable Scenarios

ChronoFlow is currently best suited for:

- Internal networks
- Single-team use
- Dozens of jobs or fewer
- A single scheduler
- Shell / Python script scheduling
- Lightweight Docker deployment
- Web-based operation and log inspection

It is not positioned as a large-scale distributed scheduler, and it intentionally avoids complex multi-tenant permission systems and massive scheduling workloads in the first version.

## Production Notes

- Change the default admin password, JWT secret, callback token, executor token, and database password.
- `CHRONOFLOW_TOKEN_ENCRYPT_KEY` must be 32 bytes. If changed, existing encrypted executor tokens cannot be decrypted with the new key.
- Real process-group kill semantics require Linux.
- MySQL stores log metadata only; full log content is stored as files.
- Exec does not need database configuration and should not connect to Admin's database.
- Do not commit `deploy/.env`, runtime logs, database passwords, or GitHub tokens.
