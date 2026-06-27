# ChronoFlow

ChronoFlow 是一个面向内网单团队使用的轻量定时任务平台。它包含调度器后端、执行器后端和 Web 调度中心，支持 Cron 定时、手动运行、Glue Shell、异步回调、任务终止、执行日志和运行报表。

## 功能概览

- 执行器管理：新增、编辑、删除、心跳状态展示。
- 任务管理：新增、编辑、删除、启动调度、停止调度、手动运行。
- Cron 可视化配置：支持常用分钟、小时、日、周、月配置和手动表达式。
- Glue Shell：按任务保存 Shell 脚本，脚本可调用挂载进执行器容器的 Python 或其他脚本。
- 异步执行：Admin 下发执行请求后立即返回，Exec 完成后回调 Admin。
- 同任务互斥：同一个任务不能并发运行，不同任务可以并行。
- 终止任务：Admin 请求 Exec kill 进程组，日志状态进入 `killing`，最终变为 `killed` 或 `failed`。
- 日志存储：MySQL 只保存日志元数据，完整日志正文保存为文件。
- 运行报表：展示任务数量、近 7 天调度次数、执行器数量、成功率和日期趋势。

## 快速开始

源码构建部署：

```bash
git clone <your-repo-url> chronoflow
cd chronoflow/deploy
cp .env.example .env
docker compose -f docker-compose.mysql.yml up -d
docker compose up -d --build
```

打开：

```text
http://127.0.0.1:5173
```

默认账号：

```text
admin / admin123
```

如果使用作者提前发布的镜像，修改 `deploy/.env` 中的镜像地址后运行：

```bash
cd deploy
docker compose -f docker-compose.image.yml up -d
```

详细部署、端口、MySQL、外部数据库、脚本挂载和首个任务创建说明见 [deploy/README.md](deploy/README.md)。

## 项目结构

```text
chronoFlow/
├── chronoFlow-admin/        # 调度器后端，连接 MySQL
├── chronoFlow-exec/         # 执行器后端，不连接数据库
├── chronoFlow-ui/           # 调度中心前端
├── deploy/                  # Docker Compose、env 模板、MySQL 初始化和脚本挂载
└── docs/                    # PRD、测试指南、开发计划和过程记录
```

## 架构约定

```text
UI -> Admin -> Exec
       ^        |
       |        v
       +-- callback
```

- `chronoFlow-admin` 是唯一连接 MySQL 的服务。
- `chronoFlow-exec` 不连接 MySQL。
- Admin 调用 Exec 使用每个执行器自己的 `X-Executor-Token`。
- Exec 回调 Admin 使用全局 `X-Callback-Token`。
- Exec 回调失败时会把待回调结果落盘，并在后台持续重试，默认保留 7 天。

## 开发模式

如果你要改代码，可以分别启动三个模块：

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

## 验证命令

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

完整测试指南见 [docs/TESTING_GUIDE.md](docs/TESTING_GUIDE.md)。

## 生产注意事项

- 修改默认管理员密码、JWT Secret、Callback Token 和执行器 Token。
- `CHRONOFLOW_TOKEN_ENCRYPT_KEY` 必须是 32 字节。
- 执行器真实进程组终止语义只支持 Linux。
- 不要把完整日志正文写入 MySQL；MySQL 只保存元数据。
- 执行器不需要数据库配置，也不应该连接 Admin 的数据库。
