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

## 快速部署

项目支持两种 Docker 部署方式。

### 方式一：源码构建部署

适合想本地构建镜像、二次开发或内网无法拉取作者镜像的用户。

```bash
git clone <your-repo-url> chronoflow
cd chronoflow
cd deploy
cp .env.example .env
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

### 方式二：作者镜像部署

适合只想快速部署使用的用户。进入 `deploy/`，复制 `.env.example` 后，把镜像地址改成作者发布的镜像：

```env
CHRONOFLOW_ADMIN_IMAGE=ghcr.io/your-name/chronoflow-admin:latest
CHRONOFLOW_EXEC_IMAGE=ghcr.io/your-name/chronoflow-exec:latest
CHRONOFLOW_UI_IMAGE=ghcr.io/your-name/chronoflow-ui:latest
```

启动：

```bash
docker compose -f docker-compose.image.yml up -d
```

## 端口配置

所有端口都在 `.env` 中配置：

```env
CHRONOFLOW_UI_PORT=5173
CHRONOFLOW_ADMIN_HTTP_PORT=10003
CHRONOFLOW_ADMIN_GRPC_PORT=11003
CHRONOFLOW_EXEC_HTTP_PORT=10004
CHRONOFLOW_EXEC_GRPC_PORT=11004
MYSQL_HOST_PORT=3306
```

如果本机 `5173` 或 `3306` 被占用，直接改 `.env` 后重新启动即可。

## MySQL 配置

`deploy/docker-compose.yml` 默认会启动一个 MySQL 8.0 容器，并通过环境变量自动创建数据库和用户：

```env
DB_HOST=mysql
DB_PORT=3306
DB_NAME=chronoflow
DB_USER=chronoflow
DB_PASSWORD=chronoflow123
MYSQL_ROOT_PASSWORD=root123456
```

Admin 启动时会自动迁移表结构。`deploy/mysql/init/001-init.sql` 提供默认数据库初始化 SQL。

### 使用外部 MySQL

如果你已有 MySQL，可以把 `.env` 改成外部地址：

```env
DB_HOST=host.docker.internal
DB_PORT=3306
DB_NAME=chronoflow
DB_USER=root
DB_PASSWORD=root
```

Linux 服务器也可以直接写数据库 IP：

```env
DB_HOST=192.168.1.20
```

使用外部 MySQL 时，请先创建数据库：

```sql
CREATE DATABASE IF NOT EXISTS chronoflow DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

如果不需要 compose 内置 MySQL，可以只启动：

```bash
docker compose up -d --build --no-deps admin exec ui
```

## 创建第一个执行器

Docker compose 部署后，Admin 在容器网络内访问 Exec，所以执行器地址填写：

```text
名称：default-exec
地址：http://chronoflow-exec:10004
Token：default-exec-token
```

Token 来自 `.env`：

```env
EXECUTOR_TOKEN=default-exec-token
```

## 创建第一个任务

1. 在任务页面新增任务，选择 `default-exec`。
2. 打开 Glue 编辑器，保存：

```bash
echo chronoflow-demo-start
python3 /scripts/report.py
echo chronoflow-demo-done
```

3. 点击“运行”。
4. 在执行日志中查看状态和日志正文。

默认 compose 会把宿主机目录挂载到执行器容器：

```text
deploy/scripts -> /scripts
```

## 项目结构

```text
chronoFlow/
├── chronoFlow-admin/        # 调度器后端，连接 MySQL
├── chronoFlow-exec/         # 执行器后端，不连接数据库
├── chronoFlow-ui/           # 调度中心前端
├── deploy/
│   ├── docker-compose.yml       # 源码构建部署
│   ├── docker-compose.image.yml # 作者镜像部署
│   ├── docker-compose.local.yml # 本地开发调试部署
│   ├── .env.example             # 部署配置模板
│   ├── mysql/init/              # MySQL 初始化 SQL
│   └── scripts/                 # 默认挂载到执行器的脚本目录
└── docs/                        # PRD、测试指南、开发计划和过程记录
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

你当前本地已有 MySQL 容器的调试方式仍可使用：

```bash
cd deploy
docker compose -f docker-compose.local.yml up -d --build --remove-orphans
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

## 生产注意事项

- 修改默认管理员密码、JWT Secret、Callback Token 和执行器 Token。
- `CHRONOFLOW_TOKEN_ENCRYPT_KEY` 必须是 32 字节。
- 执行器真实进程组终止语义只支持 Linux。
- 不要把完整日志正文写入 MySQL；MySQL 只保存元数据。
- 执行器不需要数据库配置，也不应该连接 Admin 的数据库。
