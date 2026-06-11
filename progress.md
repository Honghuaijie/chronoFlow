# ChronoFlow 技术方案文档进度

## 2026-06-10

1. 用户确认需要为 admin、ui、exec 三个子项目编写中文技术方案文档。
2. 用户确认文档应偏向给 AI/开发者照着实现的详细开发方案。
3. 用户强调后端模板自带示例接口只学习写法，不当作业务代码。
4. 用户强调执行器不需要数据库，方案中需要特别说明模板数据库连接要移除或禁用。
5. 已创建本任务计划文件、调研记录文件和进度文件。
6. 已扫描三个子项目目录：admin 和 exec 均为 Kratos 后端模板，ui 当前只有前端编码规则文件。
7. 已读取后端和前端 AI 编码规则，确认后端 proto-first 和前端 page/store/api 分层要求。
8. 已读取 admin/exec 配置和 data 初始化，确认 exec 模板数据库连接需要在方案中作为强制改造点。
9. 已读取后端服务注册、错误码和测试指南，准备进入三份技术方案编写。
10. 已创建 `chronoFlow-admin/技术方案.md`，覆盖 admin 的模块、接口、数据模型、调度、callback、日志文件和测试方案。
11. 已创建 `chronoFlow-exec/技术方案.md`，重点说明执行器不连接数据库、移除模板 DB 链路、进程组执行、pending callback 和重试。
12. 已创建并用 `ui-ux-pro-max` 增强 `chronoFlow-ui/技术方案.md`，补充设计系统、布局密度、交互反馈、可访问性和 UI 交付检查清单。
13. 已完成三份技术方案校验：文件存在；占位符扫描无命中；UI 标题编号已修正；关键负向约束只出现在禁止项或契约提醒中。
14. 开始实现 `chronoFlow-admin`：按 TDD 新增 token 加密和日志文件存储测试，确认 RED 后实现 `internal/security` 与 `internal/logstore`，两个包测试已通过。
15. 新增 `internal/biz/status.go` 状态常量和 callback 状态判断，测试通过。
16. 新增 ChronoFlow 核心 GORM model，并将 AutoMigrate 扩展到 executors/jobs/job_glues/job_logs。期间发现 `internal/conf/conf.pb.go` 生成物导致 data 包初始化 panic，运行 `make config` 重新生成后模型测试通过。
17. 新增执行器 HTTP client：封装 `/health`、`/run`、`/kill`，携带 `X-Executor-Token`，单元测试通过。
18. 扩展 `internal/conf/conf.proto` 和 `configs/config.yaml`：加入 `scheduler`、`executor`、`security`、`logs`、`recovery`、`server.public_base_url`，并新增 `ValidateChronoFlow` 启动校验。
19. 调整 admin 启动流程：先加载配置、校验核心配置，再按配置时区设置 `time.Local`。
20. 新增执行器管理 proto 契约 `api/executor/v1/executor.proto`，生成 pb/http/grpc/openapi 代码，接口路径使用 `/v1/admin/executors/*`。
21. 新增执行器管理基础链路：`biz.ExecutorUsecase`、`data.ExecutorRepo`、`service.ExecutorService`，实现创建、更新、删除、详情、列表；创建时 token 加密，默认 `offline`，更新时空 token 不覆盖旧密文。
22. 将执行器 service 注册到 HTTP/gRPC server，并通过 wire 生成 `cmd/chronoFlow-admin/wire_gen.go`。
23. 当前验证通过：`go test ./internal/... -count=1`、`go build ./cmd/chronoFlow-admin`。
24. 新增任务、Glue、执行日志 proto：`api/job/v1/job.proto`、`api/glue/v1/glue.proto`、`api/joblog/v1/job_log.proto`，并生成 pb/http/grpc/openapi 代码。
25. 新增 `JobUsecase`、`GlueUsecase`、`JobLogUsecase`：任务创建默认 `stopped`，Cron 做 6 段校验，启动前要求 Glue 存在；Glue 支持按 `job_id` upsert；日志详情从文件存储读取正文，文件不存在返回固定提示。
26. 新增 `JobRepo`、`GlueRepo`、`JobLogRepo`：覆盖任务 CRUD、Glue upsert、执行日志列表筛选和详情查询，并用 sqlite 内存库测试。
27. 新增 `JobService`、`GlueService`、`JobLogService`，注册到 HTTP/gRPC server，并通过 wire 接入 `logstore.FileStore` 作为日志正文读取器。
28. 当前验证通过：`go test ./internal/... -count=1`、`go build -o /tmp/chronoflow-admin-build ./cmd/chronoFlow-admin`。
29. 新增任务执行闭环：`JobRunUsecase` 支持手动/cron 触发、同任务互斥、创建 running 执行日志、解密执行器 token、调用执行器 `/run`，并支持 `KillJob` 将日志置为 `killing` 后调用执行器 `/kill`。
30. 新增内部 callback 接口 `api/internal/v1/job_run_callback.proto` 和 `CallbackUsecase`：校验最终状态、保护已终态日志不被覆盖、写入日志正文文件、更新日志元数据和最终状态。
31. 扩展 `JobLogRepo`：支持创建日志、查询同任务 active 日志、更新日志、执行器离线批量失败、启动恢复批量失败、killing 超时失败和过期日志元数据清理。
32. 新增后台 worker：作为 Kratos server 启动，负责启动恢复、执行器健康检查、killing 超时扫描和日志保留清理；执行器连续健康检查失败达到阈值后标记 offline，并将该执行器 active 日志标记 failed。
33. 新增 `scheduler.Manager`：基于 `robfig/cron/v3` 的 6 位 Cron 单调度器，任务 start 注册 Cron，stop 移除 Cron，Cron 触发时按异步执行链路创建执行日志并下发执行器。
34. 当前验证通过：`go test ./internal/... -count=1`、`go build -o /tmp/chronoflow-admin-build ./cmd/chronoFlow-admin`。
35. 新增轻量后台鉴权：`api/auth/v1/auth.proto`，支持 `/v1/public/auth/login` 和 `/v1/admin/auth/current`；V1 使用配置里的内置管理员账号，登录后返回 JWT。
36. 新增 HTTP admin 鉴权中间件：保护 `/v1/admin/*`，校验 `Authorization: Bearer <token>`；公共登录、健康检查、内部 callback 和模板示例接口不受影响。
37. 当前最终验证通过：`go test ./internal/... -count=1`、`go build -o /tmp/chronoflow-admin-build ./cmd/chronoFlow-admin`。
