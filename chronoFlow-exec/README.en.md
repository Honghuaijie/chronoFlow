# ChronoFlow Exec

Executor backend for ChronoFlow. It receives dispatch requests from Admin, runs Glue Shell on Linux, captures logs, kills process groups, and asynchronously callbacks execution results to Admin.

## Responsibilities

- Provide `/health`, `/run`, and `/kill`.
- Authenticate Admin requests with `X-Executor-Token`.
- Enforce per-job mutual exclusion.
- Start scripts with Shell and try to terminate the whole process group.
- Capture stdout/stderr, with a default 5 MB limit.
- Write pending callback files after execution and try to callback Admin.
- Retry failed callbacks in the background. Default pending retention is 7 days.

Out of scope:

- Connecting to MySQL or any Admin database.
- Storing job definitions, executor definitions, or log metadata.
- Cron scheduling.

## Default Ports

| Protocol | Address |
| --- | --- |
| HTTP | `0.0.0.0:10004` |
| gRPC | `0.0.0.0:11004` |

## Local Startup

```bash
go run ./cmd/chronoFlow-exec -conf ./configs
```

Default token:

```text
change-me
```

Health check:

```bash
curl -i http://127.0.0.1:10004/health \
  -H 'X-Executor-Token: change-me'
```

## Key Configuration

- `executor.name`: executor name.
- `executor.token`: token used by Admin when calling Exec.
- `executor.data_dir`: local data directory for pending callbacks.
- `executor.shell_path`: Shell path. Default is `/bin/bash`.
- `executor.temp_dir`: temporary script directory.
- `executor.kill_grace_seconds`: wait time between SIGTERM and SIGKILL.
- `executor.max_log_bytes`: max log content per run. Default is 5 MB.
- `callback.retry_interval_seconds`: pending callback retry interval.
- `callback.pending_retention_days`: pending callback retention days. Default is 7.

`${ENV:default}` placeholders can be overridden with environment variables.

## Docker and Script Mounts

Exec can run inside Docker. Mount host script directories into the container and call them from Glue Shell.

```bash
docker run --rm \
  -p 10004:10004 \
  -v /opt/chronoflow/scripts:/scripts \
  -v /opt/chronoflow/exec-data:/app/data \
  chronoFlow-exec:latest
```

Glue example:

```bash
python3 /scripts/report.py
```

## Verification

```bash
go test ./internal/... -count=1
go build -o /tmp/chronoflow-exec-build ./cmd/chronoFlow-exec
```

## Development Notes

- Exec does not need a database. Template database initialization code must not remain as business code.
- Template user APIs are examples only and are not ChronoFlow business code.
- V1 only requires real process-group kill semantics on Linux.
- `/run` must execute asynchronously and must not depend on the HTTP request context staying alive.
- Callback result should be written to disk before sending, so short Admin outages do not lose results.
