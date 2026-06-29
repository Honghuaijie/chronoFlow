# ChronoFlow

ChronoFlow is a lightweight scheduled job platform for internal single-team use. It includes a scheduler backend, an executor backend, and a web console for Cron schedules, manual runs, Glue Shell scripts, async callbacks, kill operations, execution logs, and runtime reports.

## Features

- Executor management: create, edit, delete, and heartbeat status.
- Job management: create, edit, delete, start scheduling, stop scheduling, and manual run.
- Visual Cron picker for common minute, hour, day, week, and month schedules, plus manual expressions.
- Glue Shell scripts per job. Scripts can call Python or other files mounted into the executor container.
- Async execution: Admin dispatches a run request and returns immediately; Exec callbacks Admin after completion.
- Per-job mutual exclusion: the same job cannot run concurrently, while different jobs can run in parallel.
- Kill running jobs: Admin asks Exec to kill the process group. Logs enter `killing` and eventually become `killed` or `failed`.
- Log storage: MySQL stores metadata only; full log content is stored as files.
- Reports: job count, run count, executor count, success rate, and recent 7-day trend.

## Quick Start

ChronoFlow supports two Docker deployment modes:

- Source-build deployment: for developers who want to modify code and build images locally.
- Prebuilt-image deployment: for servers where you do not want to pull the full source code.

### Source-Build Deployment

```bash
git clone https://github.com/Honghuaijie/chronoFlow.git chronoflow
cd chronoflow/deploy
cp .env.example .env
```

If you want to use the bundled MySQL service:

```bash
docker compose -f docker-compose.mysql.yml up -d
```

If you already have an external MySQL instance, skip the previous command and update `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, and `DB_PASSWORD` in `.env`.

Start the application:

```bash
docker compose up -d --build
```

### Prebuilt-Image Deployment

The server only needs the files under `deploy`; it does not need the full source code. Fixed image versions are recommended:

```env
CHRONOFLOW_ADMIN_IMAGE=ghcr.io/honghuaijie/chronoflow-admin:v0.1.2
CHRONOFLOW_EXEC_IMAGE=ghcr.io/honghuaijie/chronoflow-exec:v0.1.2
CHRONOFLOW_UI_IMAGE=ghcr.io/honghuaijie/chronoflow-ui:latest
```

Start the application:

```bash
cd deploy
docker compose -f docker-compose.image.yml up -d
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

For detailed deployment, ports, MySQL, external database, script mounts, and first-job setup, see [deploy/README.md](deploy/README.md).

## First Executor

If Admin and Exec are started by the same compose stack, create an executor in the UI with:

```text
Name: exec-default
Address: http://chronoflow-exec:10004
Token: the EXECUTOR_TOKEN value from .env
```

Do not use `http://127.0.0.1:10004`, because from the Admin container, `127.0.0.1` means the Admin container itself, not the Exec container.

## First Test Job

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

## Project Layout

```text
chronoFlow/
├── chronoFlow-admin/        # Scheduler backend. Connects to MySQL.
├── chronoFlow-exec/         # Executor backend. No database connection.
├── chronoFlow-ui/           # Web console.
├── deploy/                  # Docker Compose, env template, MySQL init, and scripts.
└── docs/                    # PRD, testing guide, development plan, and notes.
```

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

## Production Notes

- Change the default admin password, JWT secret, callback token, executor token, and database password.
- `CHRONOFLOW_TOKEN_ENCRYPT_KEY` must be 32 bytes. If changed, existing encrypted executor tokens cannot be decrypted with the new key.
- Real process-group kill semantics require Linux.
- MySQL stores log metadata only; full log content is stored as files.
- Exec does not need database configuration and should not connect to Admin's database.
- Do not commit `deploy/.env`, runtime logs, database passwords, or GitHub tokens.
