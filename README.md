# ChronoFlow

ChronoFlow 是一个面向内网单团队使用的轻量定时任务平台。V1 目标是支持几十个以内任务、单调度器、Linux 执行器、异步回调、日志文件存储和基础 Web 管理台。

## 项目结构

```text
chronoFlow/
├── chronoFlow-admin/   # 调度器后端，连接 MySQL，负责任务、执行器、日志元数据和调度
├── chronoFlow-exec/    # 执行器后端，不连接数据库，负责运行 Shell、kill 进程组和回调结果
├── chronoFlow-ui/      # 调度中心前端，Vue3 + Ant Design Vue
├── prd-v1.md           # V1 产品需求
├── task_plan.md        # 项目计划
├── progress.md         # 开发与联调进度
└── findings.md         # 过程发现和决策记录
```

## V1 功能

- 执行器管理：新增、编辑、删除、心跳状态展示。
- 任务管理：新增、编辑、删除、启动调度、停止调度、手动运行。
- Glue Shell：按任务保存 Shell 脚本，脚本可以调用宿主机或容器内挂载的 Python 脚本。
- 异步执行：调度器下发执行请求后立即返回，执行器完成后回调调度器。
- 同任务互斥：同一个任务不能并发运行，不同任务可以并行。
- 终止任务：调度器请求执行器 kill 运行进程组，日志状态从 `running` 进入 `killing`，最终变为 `killed` 或 `failed`。
- 日志存储：MySQL 只保存日志元数据，完整日志正文保存到调度器本地文件目录。
- 日志查看：前端支持日志列表、筛选、详情、Glue 快照和日志正文查看。
- 轻量鉴权：配置内置管理员账号，登录后使用 JWT 访问 `/v1/admin/*`。

## 架构约定

```text
UI -> Admin -> Exec
       ^        |
       |        v
       +-- callback
```

- `chronoFlow-admin` 是唯一连接 MySQL 的服务。
- `chronoFlow-exec` 不连接 MySQL，不读写调度器数据库。
- 调度器调用执行器时使用每个执行器自己的 `X-Executor-Token`。
- 执行器回调调度器时使用全局 `X-Callback-Token`。
- 执行器回调失败时会把待回调结果落盘，并在后台持续重试，默认保留 7 天。

## 默认端口

| 服务 | HTTP | gRPC |
| --- | --- | --- |
| Admin | `10003` | `11003` |
| Exec | `10004` | `11004` |
| UI | `5173`，被占用时 Vite 自动换端口 | - |

## 前置依赖

- Go 1.22 或更高版本。
- Node.js 18 或更高版本。
- MySQL 8.x 或兼容版本。
- Linux 执行环境用于真实执行器部署；本地开发可在 macOS 上联调基础 HTTP 流程。

Admin 默认 MySQL 配置：

```yaml
host: 127.0.0.1
port: 3306
username: root
password: root
database: chronoflow
```

请先创建数据库：

```sql
CREATE DATABASE IF NOT EXISTS chronoflow DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## 本地启动

### 1. 启动 Admin

```bash
cd chronoFlow-admin
go run ./cmd/chronoFlow-admin -conf ./configs
```

默认管理员：

```text
username: admin
password: admin123
```

### 2. 启动 Exec

```bash
cd chronoFlow-exec
go run ./cmd/chronoFlow-exec -conf ./configs
```

默认执行器 token：

```text
change-me
```

### 3. 启动 UI

```bash
cd chronoFlow-ui
npm install
npm run dev
```

前端默认通过 Vite proxy 访问：

```text
http://127.0.0.1:10003
```

如需覆盖：

```bash
VITE_API_PROXY_TARGET=http://127.0.0.1:10003 npm run dev
```

## 最小联调流程

1. 打开 UI：`http://127.0.0.1:5173/`，如果端口被占用，以 Vite 输出为准。
2. 使用 `admin / admin123` 登录。
3. 新增执行器：
   - 名称：`local-exec`
   - 地址：`http://127.0.0.1:10004`
   - Token：`change-me`
4. 新增任务，选择该执行器。
5. 打开任务的 Glue 编辑器，保存 Shell：

```bash
echo chronoflow-demo-start
pwd
echo chronoflow-demo-done
```

6. 点击任务“运行”。
7. 打开执行日志详情，确认状态为 `success` 且日志正文可见。

## 常用验证命令

Admin：

```bash
cd chronoFlow-admin
go test ./internal/... -count=1
go build -o /tmp/chronoflow-admin-build ./cmd/chronoFlow-admin
```

Exec：

```bash
cd chronoFlow-exec
go test ./internal/... -count=1
go build -o /tmp/chronoflow-exec-build ./cmd/chronoFlow-exec
```

UI：

```bash
cd chronoFlow-ui
npm run build
```

## 关键配置

Admin：

- `security.admin_username` / `security.admin_password`：内置管理员账号。
- `security.jwt_secret`：JWT 密钥。
- `security.token_encrypt_key`：执行器 token 加密密钥，需 32 字节。
- `security.callback_token`：执行器回调调度器的全局 token。
- `logs.data_dir`：调度器保存日志正文的目录。
- `logs.retention_days`：日志元数据和日志文件保留天数，默认 30 天。
- `scheduler.timezone`：Cron 时区，默认 `Asia/Shanghai`。

Exec：

- `executor.token`：调度器访问执行器时使用的 token。
- `executor.data_dir`：执行器 pending callback 存储目录。
- `executor.shell_path`：Shell 路径，默认 `/bin/bash`。
- `executor.max_log_bytes`：单次执行最大日志正文，默认 5MB。
- `callback.retry_interval_seconds`：pending callback 重试间隔。
- `callback.pending_retention_days`：pending callback 保留天数，默认 7 天。

配置文件中的环境变量占位符支持环境变量覆盖。

## Docker 和脚本目录

V1 支持执行器跑在 Docker 容器中。推荐把业务脚本放在宿主机目录，并通过 Docker volume 挂载到执行器容器内。

本地 Docker 调试可以直接使用：

```bash
docker compose -f docker-compose.local.yml up -d --build --remove-orphans
```

本地 compose 不启动新的 MySQL 容器，Admin 默认连接宿主机 `3306` 上已有的 MySQL，例如已映射端口的 `boke-mysql`。详细步骤见 `DEPLOYMENT.md`。

示例：

```bash
docker run --rm \
  -p 10004:10004 \
  -v /opt/chronoflow/scripts:/scripts \
  -v /opt/chronoflow/exec-data:/app/data \
  chronoFlow-exec:latest
```

Glue Shell 中可以调用挂载目录里的 Python 脚本：

```bash
python3 /scripts/report.py
```

## 注意事项

- 执行器只支持 Linux 服务器上的真实进程组 kill 语义。
- 不要把完整日志正文写入 MySQL；MySQL 只保存元数据。
- 执行器不需要数据库配置，也不应该连接 Admin 的数据库。
- 模板自带的 user 示例接口只作为写法参考，不属于 ChronoFlow 业务。
