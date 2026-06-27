# ChronoFlow

ChronoFlow is a lightweight scheduled job platform for an internal single-team environment. It includes a scheduler backend, an executor backend, and a web console for Cron jobs, manual runs, Glue Shell scripts, async callbacks, kill operations, logs, and reports.

## Features

- Executor management: create, edit, delete, and heartbeat status.
- Job management: create, edit, delete, start scheduling, stop scheduling, and manual run.
- Visual Cron picker for common schedules plus manual expressions.
- Glue Shell scripts per job. Scripts can call Python or other files mounted into the executor container.
- Async execution: Admin dispatches a run request; Exec callbacks Admin after completion.
- Per-job mutual exclusion: the same job cannot run concurrently.
- Kill running jobs by asking Exec to kill the process group.
- Log storage: MySQL stores metadata only; full log content is stored as files.
- Reports for job count, recent run count, executor count, success rate, and daily trend.

## Quick Start

ChronoFlow supports two Docker deployment modes.

### Source Build Deployment

Use this when you want to build images locally or modify the code.

```bash
git clone <your-repo-url> chronoflow
cd chronoflow
cp .env.example .env
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

### Prebuilt Image Deployment

Use this when the author has published images. Edit `.env`:

```env
CHRONOFLOW_ADMIN_IMAGE=ghcr.io/your-name/chronoflow-admin:latest
CHRONOFLOW_EXEC_IMAGE=ghcr.io/your-name/chronoflow-exec:latest
CHRONOFLOW_UI_IMAGE=ghcr.io/your-name/chronoflow-ui:latest
```

Start:

```bash
docker compose -f docker-compose.image.yml up -d
```

## Ports

All host ports are configured in `.env`:

```env
CHRONOFLOW_UI_PORT=5173
CHRONOFLOW_ADMIN_HTTP_PORT=10003
CHRONOFLOW_ADMIN_GRPC_PORT=11003
CHRONOFLOW_EXEC_HTTP_PORT=10004
CHRONOFLOW_EXEC_GRPC_PORT=11004
MYSQL_HOST_PORT=3306
```

## MySQL

By default, Compose starts a MySQL 8.0 container:

```env
DB_HOST=mysql
DB_PORT=3306
DB_NAME=chronoflow
DB_USER=chronoflow
DB_PASSWORD=chronoflow123
MYSQL_ROOT_PASSWORD=root123456
```

Admin auto-migrates tables on startup. `deploy/mysql/init/001-init.sql` provides the default database initialization SQL.

For an external MySQL, edit `.env`:

```env
DB_HOST=host.docker.internal
DB_PORT=3306
DB_NAME=chronoflow
DB_USER=root
DB_PASSWORD=root
```

Create the database first:

```sql
CREATE DATABASE IF NOT EXISTS chronoflow DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

Then start only app services:

```bash
docker compose up -d --build --no-deps admin exec ui
```

## First Executor

With Docker Compose, Admin reaches Exec through the Docker network:

```text
Name: default-exec
Address: http://chronoflow-exec:10004
Token: default-exec-token
```

The token comes from `.env`:

```env
EXECUTOR_TOKEN=default-exec-token
```

## First Job

Create a job, open the Glue editor, and save:

```bash
echo chronoflow-demo-start
python3 /scripts/report.py
echo chronoflow-demo-done
```

Run it manually and check the job log detail page.

The default Compose file mounts:

```text
deploy/scripts -> /scripts
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
- Admin calls Exec with `X-Executor-Token`.
- Exec callbacks Admin with `X-Callback-Token`.

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

The previous local debugging Compose file is still available:

```bash
docker compose -f docker-compose.local.yml up -d --build --remove-orphans
```

## Verification

```bash
cd chronoFlow-admin && go test ./internal/... -count=1
cd chronoFlow-exec && go test ./internal/... -count=1
cd chronoFlow-ui && npm run build
```

## Production Notes

- Change the default admin password, JWT secret, callback token, and executor token.
- `CHRONOFLOW_TOKEN_ENCRYPT_KEY` must be 32 bytes.
- Real process-group kill semantics require Linux.
- MySQL stores log metadata only; full log content is stored as files.
- Exec does not need database configuration.
