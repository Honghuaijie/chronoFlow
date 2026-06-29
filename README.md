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
- 运行报表：展示任务数量、调度次数、执行器数量、成功率和近 7 天日期趋势。

## 快速开始

ChronoFlow 支持两种 Docker 部署方式：

- 源码构建部署：适合开发者本地修改代码后自行构建镜像。
- 作者镜像部署：适合服务器空间较小、不想拉源码的场景。

### 源码构建部署

```bash
git clone https://github.com/Honghuaijie/chronoFlow.git chronoflow
cd chronoflow/deploy
cp .env.example .env
```

如果需要使用项目内置 MySQL：

```bash
docker compose -f docker-compose.mysql.yml up -d
```

如果你已有 MySQL，请跳过上一步，并修改 `.env` 中的 `DB_HOST`、`DB_PORT`、`DB_NAME`、`DB_USER`、`DB_PASSWORD`。

启动应用：

```bash
docker compose up -d --build
```

### 作者镜像部署

服务器只需要复制 `deploy` 目录中的部署文件，不需要拉完整源码。推荐使用固定版本镜像：

```env
CHRONOFLOW_ADMIN_IMAGE=ghcr.io/honghuaijie/chronoflow-admin:v0.1.2
CHRONOFLOW_EXEC_IMAGE=ghcr.io/honghuaijie/chronoflow-exec:v0.1.2
CHRONOFLOW_UI_IMAGE=ghcr.io/honghuaijie/chronoflow-ui:latest
```

启动应用：

```bash
cd deploy
docker compose -f docker-compose.image.yml up -d
```

打开：

```text
http://127.0.0.1:5173
```

默认账号：

```text
admin / admin123
```

生产环境请在 `.env` 中修改默认管理员密码、JWT Secret、Callback Token、执行器 Token 和数据库密码。

详细部署、端口、MySQL、外部数据库、脚本挂载和首个任务创建说明见 [deploy/README.md](deploy/README.md)。

## 首个执行器怎么填

如果 Admin 和 Exec 都由同一个 compose 启动，在 UI 新增执行器时填写：

```text
名称：exec-default
地址：http://chronoflow-exec:10004
Token：填写 .env 中的 EXECUTOR_TOKEN
```

不要填写 `http://127.0.0.1:10004`，因为对 Admin 容器来说，`127.0.0.1` 是 Admin 容器自己，不是 Exec 容器。

## 首个测试任务

可以创建一个 Glue Shell 任务验证完整链路：

```bash
#!/bin/bash
set -e

echo "hello chronoflow"
echo "run time: $(date '+%Y-%m-%d %H:%M:%S')"
echo "hostname: $(hostname)"
python3 --version
echo "done"
```

手动运行后，在执行日志中应看到状态为 `success`，并能看到脚本输出。

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

- 修改默认管理员密码、JWT Secret、Callback Token、执行器 Token 和数据库密码。
- `CHRONOFLOW_TOKEN_ENCRYPT_KEY` 必须是 32 字节；修改后，已保存的执行器 token 密文无法用新密钥解密。
- 执行器真实进程组终止语义只支持 Linux。
- 不要把完整日志正文写入 MySQL；MySQL 只保存元数据。
- 执行器不需要数据库配置，也不应该连接 Admin 的数据库。
- 不要提交 `deploy/.env`、运行日志、数据库密码或 GitHub Token。
