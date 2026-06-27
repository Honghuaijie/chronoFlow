# ChronoFlow 技术方案文档调研记录

## 已确认需求

1. 文档使用中文文件名：`技术方案.md`。
2. 文档定位为给 AI/开发者照着实现的详细开发方案。
3. 三个子项目都需要独立技术方案。
4. `chronoFlow-admin` 和 `chronoFlow-exec` 已经使用用户自己的后端模板。
5. 后端模板中的示例接口、示例模块、示例代码只用于学习写法，不是业务需求。
6. `chronoFlow-exec` 不需要数据库连接；模板默认数据库能力需要移除或禁用。

## PRD 关键约束

1. V1 面向内网单团队、几十个任务以内、单调度器。
2. 采用异步执行：admin 下发 `/run`，exec accepted，exec 完成后 callback admin。
3. MySQL 只存日志元数据，日志正文存在 admin 本地文件系统。
4. 执行器不连接 MySQL。
5. 执行器 callback 失败时写 pending 文件并重试，默认保留 7 天。
6. 执行状态包含 `running / killing / success / failed / timeout / skipped / killed`。
7. 执行器仅支持 Linux，支持 Docker 部署。

## 待补充

1. admin 模板目录和编码规则。
2. exec 模板目录和编码规则。
3. ui 项目技术栈和目录结构。

## 目录结构发现

1. `chronoFlow-admin` 使用 Kratos 后端模板，已有 `api/`、`cmd/`、`configs/`、`internal/service`、`internal/biz`、`internal/data`、`internal/server`、`internal/errors` 等目录。
2. `chronoFlow-exec` 使用同一套后端模板，也包含 `internal/data` 和 `user` 示例代码，但根据 PRD 执行器不能连接 MySQL，后续方案必须明确移除或禁用数据库链路。
3. 两个后端模板都包含 `api/user/v1/user.proto` 和 `internal/*/user.go` 示例，方案中需要强调这些只作为编码风格参考。
4. `chronoFlow-ui` 当前只发现 `FRONTEND_AI_CODING_RULES.md`，需要继续读取前端规则后再规划目录。

## 模板规则发现

1. 后端模板要求普通接口 proto-first，接口路径使用方法级风格，不使用 REST 风格。
2. 后端模板分层为 `service -> biz -> data`，`service` 做参数校验和响应组装，`biz` 做业务规则和流程编排，`data` 做数据库访问。
3. 后端模板要求新增依赖同步维护 `ProviderSet` 和 `cmd/<service>/wire.go`。
4. 后端模板的 `user` 模块是示例，应该作为分层、错误处理、依赖注入风格参考，不是 ChronoFlow 业务模块。
5. 前端规则要求固定调用链为 `page/view -> store -> api`，页面不能直接请求接口，API/Store/页面都必须有 TypeScript 类型。
6. 前端技术栈为 `Vue3 + Ant Design Vue + Pinia + TypeScript`。

## 配置与依赖发现

1. admin 当前 `conf.proto` 包含 `Server/Data/Logging`，其中 `Data` 有 MySQL 和 Redis 配置。
2. admin 当前 `internal/data/data.go` 会创建 GORM MySQL 连接，并默认 AutoMigrate 示例 `User`。
3. exec 当前 `conf.proto` 和 `configs/config.yaml` 也包含模板默认 MySQL/Redis 配置。
4. exec 当前 `internal/data/data.go` 也会创建 GORM MySQL 连接并 AutoMigrate 示例 `User`，这与 PRD 冲突，执行器技术方案必须要求移除或禁用。
5. Makefile 支持 `make api`、`make config`、`make wire`、`make test`，技术方案应要求修改 proto/config/wire 后运行生成命令。

## 后端实现细节发现

1. `internal/server/http.go` 当前注册 `UserService` 和手写 `/health`、上传示例路由；ChronoFlow 业务接口需要替换成自己的 service 注册。
2. `internal/service/service.go` 当前 `ProviderSet` 只有 `NewUserService`，新增业务 service 后必须加入。
3. `internal/biz/biz.go` 当前 `ProviderSet` 只有 `NewUserUsecase`，新增 usecase 后必须加入。
4. `internal/errors/codes.go` 已有通用错误码，ChronoFlow 需要追加任务、执行器、Glue、日志、callback、kill 等业务错误。
5. `TESTING_GUIDE.md` 推荐 service/biz/server 分层测试，biz 使用 fake repo/fake tx，不连接真实数据库。
