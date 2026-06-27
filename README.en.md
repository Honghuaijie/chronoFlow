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

Source build deployment:

```bash
git clone <your-repo-url> chronoflow
cd chronoflow/deploy
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

For prebuilt images, edit image variables in `deploy/.env`, then run:

```bash
cd deploy
docker compose -f docker-compose.image.yml up -d
```

For detailed deployment, ports, MySQL, external database, script mounts, and first-job setup, see [deploy/README.md](deploy/README.md).

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

## Verification

```bash
cd chronoFlow-admin && go test ./internal/... -count=1
cd chronoFlow-exec && go test ./internal/... -count=1
cd chronoFlow-ui && npm run build
```

For the full testing guide, see [docs/TESTING_GUIDE.md](docs/TESTING_GUIDE.md).

## Production Notes

- Change the default admin password, JWT secret, callback token, and executor token.
- `CHRONOFLOW_TOKEN_ENCRYPT_KEY` must be 32 bytes.
- Real process-group kill semantics require Linux.
- MySQL stores log metadata only; full log content is stored as files.
- Exec does not need database configuration.
