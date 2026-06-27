# ChronoFlow Exec Testing Guide

本文件只记录 Exec 模块验证。完整三端联调见 `../docs/TESTING_GUIDE.md`。

## 基础验证

```bash
go test ./internal/... -count=1
go build -o /tmp/chronoflow-exec-build ./cmd/chronoFlow-exec
```

## 本地启动验证

```bash
go run ./cmd/chronoFlow-exec -conf ./configs
```

健康检查：

```bash
curl -i http://127.0.0.1:10004/health \
  -H 'X-Executor-Token: change-me'
```

## 重点场景

- 执行器不连接数据库。
- token 鉴权。
- `/run` 异步执行。
- HTTP request context 取消后任务仍能继续执行。
- 同任务互斥。
- stdout/stderr 日志采集。
- 日志超过 `max_log_bytes` 后截断。
- `/kill` 终止运行进程组。
- 执行完成先写 pending callback 文件，再尝试回调。
- callback 成功后删除 pending 文件。
- callback 失败后后台重试。

## 注意事项

- V1 只支持 Linux 服务器上的真实进程组 kill 语义。
- Docker 部署时需要把脚本目录和 `data_dir` 挂载出来。
- 模板 user 示例接口不属于 ChronoFlow 业务。
