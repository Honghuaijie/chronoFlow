# ChronoFlow Exec

执行器后端，负责接收调度器下发的执行请求，在 Linux 服务器上运行 Glue Shell，采集日志，终止进程组，并把执行结果异步回调给调度器。

## 职责边界

- 提供 `/health`、`/run`、`/kill`。
- 使用 `X-Executor-Token` 校验调度器请求。
- 按任务维度限制并发：同一个任务不能同时运行。
- 使用 Shell 启动脚本，并尽量终止整个进程组。
- 采集 stdout/stderr，最多保留 5MB。
- 执行完成后写 pending callback 文件，并尝试回调调度器。
- callback 失败时后台重试，默认保留 7 天。

不负责：

- 连接 MySQL 或任何调度器数据库。
- 保存任务定义、执行器定义、日志元数据。
- Cron 调度。

## 默认端口

| 协议 | 地址 |
| --- | --- |
| HTTP | `0.0.0.0:10004` |
| gRPC | `0.0.0.0:11004` |

## 本地启动

```bash
go run ./cmd/chronoFlow-exec -conf ./configs
```

默认 token：

```text
change-me
```

健康检查：

```bash
curl -i http://127.0.0.1:10004/health \
  -H 'X-Executor-Token: change-me'
```

## 关键配置

- `executor.name`：执行器名称。
- `executor.token`：调度器访问执行器的 token。
- `executor.data_dir`：执行器本地数据目录，用于 pending callback。
- `executor.shell_path`：Shell 路径，默认 `/bin/bash`。
- `executor.temp_dir`：临时脚本目录。
- `executor.kill_grace_seconds`：先 SIGTERM 后 SIGKILL 的等待时间。
- `executor.max_log_bytes`：单次执行最大日志内容，默认 5MB。
- `callback.retry_interval_seconds`：pending callback 重试间隔。
- `callback.pending_retention_days`：pending callback 保留天数，默认 7 天。

配置文件中的 `${ENV:default}` 支持环境变量覆盖。

## Docker 与脚本挂载

执行器可以跑在 Docker 容器里。推荐把宿主机脚本目录挂载进容器，然后在 Glue Shell 中调用。

```bash
docker run --rm \
  -p 10004:10004 \
  -v /opt/chronoflow/scripts:/scripts \
  -v /opt/chronoflow/exec-data:/app/data \
  chronoFlow-exec:latest
```

Glue 示例：

```bash
python3 /scripts/report.py
```

## 验证命令

```bash
go test ./internal/... -count=1
go build -o /tmp/chronoflow-exec-build ./cmd/chronoFlow-exec
```

## 开发注意事项

- 执行器不需要数据库，模板里的数据库初始化代码不能作为业务代码保留。
- 模板中的 user 示例接口只作为写法参考，不属于 ChronoFlow 业务代码。
- V1 只要求 Linux 服务器上的真实进程组 kill 语义。
- `/run` 必须异步执行，不能依赖 HTTP request context 存活。
- callback 结果应先落盘再发送，避免调度器短暂故障导致结果丢失。
