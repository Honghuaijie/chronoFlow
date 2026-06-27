# ChronoFlow Admin Testing Guide

本文件只记录 Admin 模块验证。完整三端联调见 `../docs/TESTING_GUIDE.md`。

## 基础验证

```bash
go test ./internal/... -count=1
go build -o /tmp/chronoflow-admin-build ./cmd/chronoFlow-admin
```

## 本地启动验证

```bash
go run ./cmd/chronoFlow-admin -conf ./configs
```

登录接口：

```bash
curl -sS -X POST http://127.0.0.1:10003/v1/public/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}'
```

## 重点场景

- 配置启动校验。
- 执行器 token 加密存储。
- 执行器健康检查。
- 任务创建、编辑、启动、停止。
- 手动运行同任务互斥。
- Glue 保存和执行快照。
- 执行器 callback 更新日志状态。
- 日志正文写入文件，MySQL 只保存元数据。
- kill 状态从 `running` 到 `killing`，最终到 `killed` 或 `failed`。
- 启动恢复把遗留 active 日志标记为 failed。

## 注意事项

- Admin 连接 MySQL。
- Admin 不直接执行 Shell。
- 模板 user 示例接口不属于 ChronoFlow 业务。
