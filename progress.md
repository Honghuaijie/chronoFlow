# ChronoFlow 进度

## 总览

| 模块 | 状态 | 说明 |
| --- | --- | --- |
| 调度器后端 `chronoFlow-admin` | 基础完成，待联调 | 核心接口、执行闭环、Cron、callback、worker、轻量鉴权已实现 |
| 执行器后端 `chronoFlow-exec` | 基础完成，待联调 | `/health`、`/run`、`/kill`、进程组执行、pending callback、重试已实现 |
| 调度器前端 `chronoFlow-ui` | 基础完成，已联调 | Vue3 管理台、登录、执行器、任务、Glue、日志页面已实现并完成三端联调 |

## 调度器后端

### 技术方案

1. 已创建 `chronoFlow-admin/技术方案.md`。
2. 文档覆盖 admin 的模块边界、接口、数据模型、调度、callback、日志文件和测试方案。
3. 明确模板 `user` 模块只作为写法参考，不作为 ChronoFlow 业务代码。

### 已完成

1. 新增 token 加密和日志文件存储基础能力：`internal/security`、`internal/logstore`。
2. 新增状态常量和 callback 状态判断：`internal/biz/status.go`。
3. 新增 ChronoFlow 核心 GORM model，并将 AutoMigrate 扩展到 `executors`、`jobs`、`job_glues`、`job_logs`。
4. 新增执行器 HTTP client：封装 `/health`、`/run`、`/kill`，携带 `X-Executor-Token`。
5. 扩展 `internal/conf/conf.proto` 和 `configs/config.yaml`：加入 `scheduler`、`executor`、`security`、`logs`、`recovery`、`server.public_base_url`。
6. 新增 `ValidateChronoFlow` 启动校验。
7. 调整 admin 启动流程：先加载配置、校验核心配置，再按配置时区设置 `time.Local`。
8. 新增执行器管理接口契约 `api/executor/v1/executor.proto`，路径使用 `/v1/admin/executors/*`。
9. 新增执行器管理基础链路：`ExecutorUsecase`、`ExecutorRepo`、`ExecutorService`。
10. 执行器创建时 token 加密，默认 `offline`；更新时空 token 不覆盖旧密文。
11. 新增任务、Glue、执行日志接口契约：`api/job/v1/job.proto`、`api/glue/v1/glue.proto`、`api/joblog/v1/job_log.proto`。
12. 新增 `JobUsecase`、`GlueUsecase`、`JobLogUsecase`。
13. 任务创建默认 `stopped`，Cron 做 6 段校验，启动前要求 Glue 存在。
14. Glue 支持按 `job_id` upsert。
15. 日志详情从文件存储读取正文，文件不存在返回固定提示。
16. 新增 `JobRepo`、`GlueRepo`、`JobLogRepo`，覆盖任务 CRUD、Glue upsert、执行日志列表筛选和详情查询。
17. 新增 `JobService`、`GlueService`、`JobLogService`，并注册到 HTTP/gRPC server。
18. 新增任务执行闭环：`JobRunUsecase` 支持手动/cron 触发、同任务互斥、创建 running 执行日志、解密执行器 token、调用执行器 `/run`。
19. 支持 `KillJob`：将日志置为 `killing` 后调用执行器 `/kill`。
20. 新增内部 callback 接口 `api/internal/v1/job_run_callback.proto`。
21. 新增 `CallbackUsecase`：校验最终状态、保护已终态日志不被覆盖、写入日志正文文件、更新日志元数据和最终状态。
22. 扩展 `JobLogRepo`：支持创建日志、查询同任务 active 日志、更新日志、执行器离线批量失败、启动恢复批量失败、killing 超时失败和过期日志元数据清理。
23. 新增后台 worker：作为 Kratos server 启动，负责启动恢复、执行器健康检查、killing 超时扫描和日志保留清理。
24. 执行器连续健康检查失败达到阈值后标记 offline，并将该执行器 active 日志标记 failed。
25. 新增 `scheduler.Manager`：基于 `robfig/cron/v3` 的 6 位 Cron 单调度器。
26. 任务 start 注册 Cron，stop 移除 Cron，Cron 触发时按异步执行链路创建执行日志并下发执行器。
27. 新增轻量后台鉴权：`api/auth/v1/auth.proto`，支持 `/v1/public/auth/login` 和 `/v1/admin/auth/current`。
28. V1 使用配置里的内置管理员账号，登录后返回 JWT。
29. 新增 HTTP admin 鉴权中间件：保护 `/v1/admin/*`，校验 `Authorization: Bearer <token>`。
30. 公共登录、健康检查、内部 callback 和模板示例接口不受 admin 鉴权影响。
31. 通过 wire 接入所有新增 service、repo、client、worker、scheduler。

### 已验证

```bash
cd /Users/hhj/dev/codexDemo/chronoFlow/chronoFlow-admin
go test ./internal/... -count=1
go build -o /tmp/chronoflow-admin-build ./cmd/chronoFlow-admin
```

### 待办

1. 后续可将配置内置管理员升级为数据库管理员账号体系。
2. 补充更完整的集成测试。
3. 前端联调时根据页面使用情况继续修字段、状态、路径或错误码细节。

## 执行器后端

### 技术方案

1. 已创建 `chronoFlow-exec/技术方案.md`。
2. 文档重点说明执行器不连接数据库、不读写 admin 数据库。
3. 明确模板中的数据库连接和 `user` 示例只作为写法参考，不能作为业务代码。

### 已完成

1. 重写配置契约：执行器不再注入 `Data/GORM/MySQL`。
2. 配置改为 `executor`、`callback`、`logging`。
3. 新增执行器协议 `api/executor/v1/executor.proto`。
4. 实现 `GET /health`、`POST /run`、`POST /kill`。
5. 新增 `process.LogBuffer`，支持 head+tail 日志截断。
6. 新增 `process.Manager`，支持同任务互斥、Linux 进程组执行、超时、kill。
7. 新增 `store.PendingStore`，支持 pending callback JSON 文件保存、加载、删除。
8. 新增 `callback.Client`，支持携带 `X-Callback-Token` 回调 admin。
9. 新增 `ExecutorService`：实现 health、run、kill。
10. run 异步执行 Shell，完成后写 pending callback 文件并立即尝试回调。
11. callback 成功后删除 pending 文件。
12. 新增 callback retry worker：执行器重启后定时重试 pending 文件。
13. 新增执行器 HTTP token 鉴权中间件：保护 `/health`、`/run`、`/kill`。
14. `X-Executor-Token` 使用常量时间比较。
15. 删除执行器模板中的 GORM data 实现文件。
16. 保留 `internal/data/README.md` 作为说明目录。

### 已验证

```bash
cd /Users/hhj/dev/codexDemo/chronoFlow/chronoFlow-exec
go test ./internal/... -count=1
go build -o /tmp/chronoflow-exec-build ./cmd/chronoFlow-exec
```

### 待办

1. 验证 Docker volume 挂载脚本目录和数据目录。
2. 根据联调结果补充端到端测试。

## 调度器前端

### 技术方案

1. 已创建 `chronoFlow-ui/技术方案.md`。
2. 已使用 `ui-ux-pro-max` 增强 UI 方案。
3. 方案补充了设计系统、布局密度、交互反馈、可访问性和 UI 交付检查清单。

### 已完成

1. 技术方案已完成。
2. 已按 `Vue3 + TypeScript + Ant Design Vue + Pinia + Vue Router + Axios` 创建前端工程。
3. 已接入 Vite，默认代理 admin 后端 `http://127.0.0.1:10003`，可通过 `VITE_API_PROXY_TARGET` 覆盖。
4. 已实现统一请求封装：自动携带 JWT、统一解包 `{code,message,data}`、401 自动清理登录态。
5. 已实现登录页：调用 `/v1/public/auth/login`，登录后进入调度中心。
6. 已实现基础后台布局：左侧导航、顶部用户区、执行器/任务/日志/设置路由。
7. 已实现执行器管理页面：列表、新增、编辑、删除、在线/离线状态展示。
8. 已实现任务管理页面：列表、新增、编辑、删除、启动、停止、手动运行、终止。
9. 已实现 Glue Shell 编辑抽屉：按任务读取和保存 Glue 内容。
10. 已实现日志列表页面：分页、任务/执行器/状态/触发方式筛选、运行中日志轮询、详情跳转、终止入口。
11. 已实现日志详情页面：元数据、状态、Glue 快照、日志正文、运行中轮询和终止入口。
12. 已按 V1 规则在任务页基于 active 日志置灰同任务手动运行按钮。
13. 已补充基础组件：`PageHeaderBar`、`StatusTag`、`PollingIndicator`、`LogViewer`。
14. 已按 `ui-ux-pro-max` 建议采用浅色、表格优先、数据密集的运维控制台风格。

### 待办

1. 后续可按需增加代码分包，降低 Ant Design Vue 首包体积 warning。
2. 后续可补充更完整的前端自动化测试。

### 已验证

```bash
cd /Users/hhj/dev/codexDemo/chronoFlow/chronoFlow-ui
npm run build
```

浏览器首屏验证：

1. 已启动 Vite dev server，当前地址 `http://127.0.0.1:5174/`。
2. 已打开 `/login`，登录页正常渲染。
3. 控制台无业务错误，仅有 Vite 连接 debug 日志。
4. 已完成 admin、exec、ui 三端真实联调。

## 文档与决策记录

1. 用户确认需要为 admin、ui、exec 三个子项目编写中文技术方案文档。
2. 用户确认文档应偏向给 AI/开发者照着实现的详细开发方案。
3. 用户强调后端模板自带示例接口只学习写法，不当作业务代码。
4. 用户强调执行器不需要数据库，方案中需要特别说明模板数据库连接要移除或禁用。
5. 已创建 `task_plan.md`、`findings.md` 和本进度文件。
6. 已扫描三个子项目目录：admin 和 exec 均为 Kratos 后端模板，ui 当前为前端项目。
7. 已读取后端和前端 AI 编码规则，确认后端 proto-first 和前端 page/store/api 分层要求。
8. 已读取 admin/exec 配置和 data 初始化，确认 exec 模板数据库连接需要作为强制改造点。
9. 已完成三份技术方案校验：文件存在；占位符扫描无命中；UI 标题编号已修正；关键负向约束只出现在禁止项或契约提醒中。
10. 已创建根目录中文版 README：`README.md`。
11. 已创建根目录英文版 README：`README.en.md`。
12. 已创建根目录测试与联调指南：`TESTING_GUIDE.md`。
13. 已将 admin、exec 的 README 从模板说明改为 ChronoFlow 模块说明。
14. 已为 admin、exec、ui 分别补充英文 README。
15. 已将 admin、exec 的 TESTING_GUIDE 从模板测试说明改为 ChronoFlow 模块验证说明。
16. 已补充根目录 `.gitignore`，并更新 admin、exec `.gitignore` 忽略本地 `data/`。
17. 已确认 README/TESTING_GUIDE 中不再残留模板 `v1/users`、`UserUsecase` 等示例接口说明。

## 联调记录

### 2026-06-11 admin + exec

1. 已启动 `chronoFlow-admin` 和 `chronoFlow-exec` 做本地真实联调。
2. 已验证 `exec /health`：携带 `X-Executor-Token: change-me` 返回 `online`。
3. 已验证 admin 登录：`/v1/public/auth/login` 返回 JWT。
4. 已验证 admin 创建执行器、创建任务、保存 Glue。
5. 首次运行任务时发现 exec 异步执行使用了 HTTP request context，`/run` 返回后 context 被取消，导致任务立刻变成 `killed` 且日志为空。
6. 已修复 exec：`ExecutorService.Run` 下发后台任务时改用独立 `context.Background()`，并新增回归测试 `TestExecutorServiceRunSurvivesRequestContextCancel`。
7. 联调期间发现自定义 worker 作为 Kratos server 时生命周期需要保持阻塞；已调整 admin worker 和 exec callback worker 的 `Start` 行为。
8. 已验证普通任务完整链路：admin `/jobs/run` -> exec `/run` -> Shell 执行 -> exec callback admin -> admin 更新 `job_logs.status=success` -> 日志详情可读取日志正文。
9. 已验证终止链路：长任务运行后调用 admin `/jobs/kill`，状态从 `running/killing` 最终变为 `killed`，日志正文保留终止前输出。
10. 本次联调验证命令覆盖：

```bash
cd /Users/hhj/dev/codexDemo/chronoFlow/chronoFlow-exec
go test ./internal/... -count=1
go build -o /tmp/chronoflow-exec-build ./cmd/chronoFlow-exec

cd /Users/hhj/dev/codexDemo/chronoFlow/chronoFlow-admin
go test ./internal/worker -count=1
go build -o /tmp/chronoflow-admin-build ./cmd/chronoFlow-admin
```

### 2026-06-12 admin + exec + ui

1. 已启动 `chronoFlow-admin`：HTTP `10003`，gRPC `11003`。
2. 已启动 `chronoFlow-exec`：HTTP `10004`，gRPC `11004`。
3. 已复用 Vite dev server：`http://127.0.0.1:5174/`。
4. 已验证 exec health：`GET /health` 携带 `X-Executor-Token: change-me` 返回 `online`。
5. 已验证 admin 登录：`POST /v1/public/auth/login` 返回 JWT。
6. 已通过前端登录页完成登录，自动跳转 `/jobs`。
7. 已验证任务列表页面能通过前端代理读取 admin API，并展示历史任务。
8. 已创建联调任务 `ui-e2e-091103`，保存 Glue 并手动运行。
9. 已验证执行链路：admin `/jobs/run` -> exec `/run` -> exec callback admin -> 日志状态 `success`。
10. 已通过前端 `/logs/6` 验证日志详情：状态成功、Glue 快照和日志正文均正常展示。
11. 已创建长任务 `ui-kill-091340` 并通过前端 `/logs/8` 的“终止”按钮触发 kill。
12. 已验证终止链路：页面从 `运行中` 自动刷新为 `已终止`，日志正文保留终止前输出，错误信息为 `任务被终止`。
13. 已检查浏览器控制台：无业务错误，仅有 Vite debug 连接日志。
14. 已检查联调期间关键 XHR/fetch：`auth/current`、`jobLogs/detail`、`jobs/kill`、`jobLogs/list` 均为 200。
15. 已修正前端日志 API 数值映射：`durationMs`、`logSizeBytes` 从 protobuf JSON 字符串显式转为 number。
16. 已再次验证前端构建通过：

```bash
cd /Users/hhj/dev/codexDemo/chronoFlow/chronoFlow-ui
npm run build
```

### 2026-06-12 文档与提交前收尾

1. 已补充项目级中英文 README。
2. 已补充项目级测试与联调指南。
3. 已补充 admin、exec、ui 模块 README，其中 admin/exec/ui 均有英文版。
4. 已更新 admin、exec 模块测试指南。
5. 已补充根目录 `.gitignore`，并确认本地日志、data、node_modules、dist、生成物会被忽略。
6. 已将曾被跟踪的 `chronoFlow-admin.log`、`chronoFlow-exec.log` 从 git 索引移除，本地文件保留并由 `.gitignore` 忽略。
7. 已执行文档残留扫描，未发现旧模板 user API 说明残留。
8. 已完成提交前验证：

```bash
cd /Users/hhj/dev/codexDemo/chronoFlow/chronoFlow-admin
go test ./internal/... -count=1
go build -o /tmp/chronoflow-admin-build ./cmd/chronoFlow-admin

cd /Users/hhj/dev/codexDemo/chronoFlow/chronoFlow-exec
go test ./internal/... -count=1
go build -o /tmp/chronoflow-exec-build ./cmd/chronoFlow-exec

cd /Users/hhj/dev/codexDemo/chronoFlow/chronoFlow-ui
npm run build
```

9. 验证结果：admin 和 exec 测试/构建通过；ui 构建通过，仅保留 Ant Design Vue 首包体积 warning。

### 2026-06-12 review 问题修复

1. 修复同任务并发运行窗口：Admin 新增 `CreateRunningIfNoActive`，data 层在事务内锁定任务行后检查 active 日志并创建 running 日志。
2. 修复执行器下发失败后日志卡在 `running`：`RunJob` 创建日志后如果调用执行器失败，会把日志置为 `failed` 并记录错误信息。
3. 修复 kill 下发失败后日志卡在 `killing`：调用执行器 kill 失败时将日志置为 `failed` 并记录错误信息。
4. 修复 callback 可靠性校验：Admin callback 会校验 `job_id` 必须匹配日志所属任务。
5. 修复 gRPC callback 鉴权绕过：非 HTTP request context 的 callback 请求现在会返回 invalid token。
6. 修复 Exec pending callback 过期文件不清理：过期 pending 文件会被删除。
7. 修复前端任务页运行状态推断：改为单独查询 `running` 和 `killing` 日志，不再用普通日志列表第一页 50 条推断。
8. 已补充回归测试：
   - `TestJobRunUsecaseRunMarksLogFailedWhenDispatchFails`
   - `TestJobRunUsecaseKillMarksFailedWhenExecutorKillFails`
   - `TestCallbackUsecaseRejectsMismatchedJobID`
   - `TestCallbackServiceRejectsContextWithoutHTTPToken`
   - `TestJobLogRepoCreateRunningIfNoActiveRejectsActiveJob`
   - `TestWorkerDeletesExpiredPendingCallbacks`
9. 已完成验证：

```bash
cd /Users/hhj/dev/codexDemo/chronoFlow/chronoFlow-admin
go test ./internal/... -count=1
go build -o /tmp/chronoflow-admin-build ./cmd/chronoFlow-admin

cd /Users/hhj/dev/codexDemo/chronoFlow/chronoFlow-exec
go test ./internal/... -count=1
go build -o /tmp/chronoflow-exec-build ./cmd/chronoFlow-exec

cd /Users/hhj/dev/codexDemo/chronoFlow/chronoFlow-ui
npm run build
```

10. 已重启 admin/exec 并完成短任务冒烟：`fix-smoke-142926` 执行成功，日志 ID `9`，日志正文正常写入。

### 2026-06-12 第二轮轻量 review 和 UI 冒烟

1. 已确认 `第一轮review` 提交后工作区干净。
2. 已启动 ChronoFlow UI 到 `http://127.0.0.1:5174/`，避免占用已有 `5173` todo 前端。
3. 已打开 `/jobs` 验证任务列表正常渲染。
4. 已确认任务页 active 查询走独立接口：
   - `/v1/admin/jobLogs/list?page=1&pageSize=1000&status=running`
   - `/v1/admin/jobLogs/list?page=1&pageSize=1000&status=killing`
5. 已确认浏览器控制台无业务错误。
6. 已创建 `ui-active-smoke-144428` 长任务并手动运行，前端显示 `运行中`，运行按钮置灰，终止按钮可见。
7. 已通过前端点击终止，任务行回到 `空闲`，运行按钮恢复。
8. 已通过日志详情 API 确认日志 ID `10` 最终状态为 `killed`，错误信息为 `任务被终止`，日志正文只包含终止前输出。
9. 已重新执行最终验证：

```bash
cd /Users/hhj/dev/codexDemo/chronoFlow/chronoFlow-admin
go test ./internal/... -count=1
go build -o /tmp/chronoflow-admin-build ./cmd/chronoFlow-admin

cd /Users/hhj/dev/codexDemo/chronoFlow/chronoFlow-exec
go test ./internal/... -count=1
go build -o /tmp/chronoflow-exec-build ./cmd/chronoFlow-exec

cd /Users/hhj/dev/codexDemo/chronoFlow/chronoFlow-ui
npm run build
```

10. 验证结果：admin 和 exec 测试/构建通过；ui 构建通过，仅保留 Ant Design Vue 首包体积 warning。
