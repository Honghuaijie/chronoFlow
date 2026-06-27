# ChronoFlow 部署指南

本文档说明 ChronoFlow 的 Docker 部署方式、配置项、MySQL 初始化和常见网络场景。

## 部署方式

ChronoFlow 支持两种部署方式。

### 1. 源码构建部署

适合开发者、本地改代码或内网构建镜像：

```bash
cp .env.example .env
docker compose up -d --build
```

使用的文件：

```text
docker-compose.yml
.env
```

### 2. 作者镜像部署

适合普通用户直接使用已发布镜像：

```bash
cp .env.example .env
```

修改 `.env`：

```env
CHRONOFLOW_ADMIN_IMAGE=ghcr.io/your-name/chronoflow-admin:latest
CHRONOFLOW_EXEC_IMAGE=ghcr.io/your-name/chronoflow-exec:latest
CHRONOFLOW_UI_IMAGE=ghcr.io/your-name/chronoflow-ui:latest
```

启动：

```bash
docker compose -f docker-compose.image.yml up -d
```

使用的文件：

```text
docker-compose.image.yml
.env
```

## 服务说明

| 服务 | 说明 | 默认访问 |
| --- | --- | --- |
| `mysql` | MySQL 8.0，保存元数据 | `127.0.0.1:3306` |
| `admin` | 调度器后端，连接 MySQL | `127.0.0.1:10003` |
| `exec` | 执行器后端，不连接数据库 | `127.0.0.1:10004` |
| `ui` | Nginx 托管的前端页面 | `127.0.0.1:5173` |

## 端口配置

修改 `.env`：

```env
CHRONOFLOW_UI_PORT=5173
CHRONOFLOW_ADMIN_HTTP_PORT=10003
CHRONOFLOW_ADMIN_GRPC_PORT=11003
CHRONOFLOW_EXEC_HTTP_PORT=10004
CHRONOFLOW_EXEC_GRPC_PORT=11004
MYSQL_HOST_PORT=3306
```

如果你把 Admin HTTP 改为 `18003`，同时建议改：

```env
CHRONOFLOW_ADMIN_HTTP_PORT=18003
PUBLIC_BASE_URL=http://chronoflow-admin:18003
CHRONOFLOW_ADMIN_UPSTREAM=http://chronoflow-admin:18003
```

如果你把 Exec HTTP 改为 `18004`，在 UI 新增执行器时地址也要填写：

```text
http://chronoflow-exec:18004
```

## MySQL 初始化

默认 compose 会启动 MySQL，并通过官方 MySQL 镜像环境变量创建数据库和用户：

```env
DB_NAME=chronoflow
DB_USER=chronoflow
DB_PASSWORD=chronoflow123
MYSQL_ROOT_PASSWORD=root123456
```

初始化 SQL 目录：

```text
mysql/init
```

Admin 启动时会自动创建或迁移业务表，所以默认只需要数据库存在。

## 使用外部 MySQL

如果 MySQL 在 Docker 网络外，比如宿主机 MySQL 或已有 MySQL 容器，修改 `.env`：

```env
DB_HOST=host.docker.internal
DB_PORT=3306
DB_NAME=chronoflow
DB_USER=root
DB_PASSWORD=root
```

Linux 服务器可以直接使用 IP：

```env
DB_HOST=192.168.1.20
```

先执行：

```sql
CREATE DATABASE IF NOT EXISTS chronoflow DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

然后只启动应用服务：

```bash
docker compose up -d --build --no-deps admin exec ui
```

镜像部署模式：

```bash
docker compose -f docker-compose.image.yml up -d --no-deps admin exec ui
```

## 执行器地址怎么填

如果 Admin 和 Exec 都在 compose 里，Admin 调用 Exec 应使用容器网络地址：

```text
http://chronoflow-exec:10004
```

如果 Exec 部署在另一台服务器，填写那台服务器对 Admin 可访问的地址：

```text
http://192.168.1.30:10004
```

Token 填 `.env` 中的：

```env
EXECUTOR_TOKEN=default-exec-token
```

## 脚本挂载

默认 compose 挂载：

```text
scripts:/scripts:ro
```

Glue Shell 可以直接调用：

```bash
python3 /scripts/report.py
```

生产环境可以把自己的脚本目录挂载进去：

```yaml
volumes:
  - /opt/chronoflow/scripts:/scripts:ro
```

## 数据目录

源码构建和镜像部署都使用 Docker volume：

```text
chronoflow-mysql-data
chronoflow-admin-data
chronoflow-admin-logs
chronoflow-exec-data
chronoflow-exec-logs
```

查看：

```bash
docker volume ls | grep chronoflow
```

删除所有数据需谨慎：

```bash
docker compose down -v
```

## 常用命令

启动：

```bash
docker compose up -d --build
```

查看日志：

```bash
docker compose logs -f admin exec ui
```

重启 Admin：

```bash
docker compose restart admin
```

停止：

```bash
docker compose down
```

## 健康检查

登录：

```bash
curl -sS -X POST http://127.0.0.1:10003/v1/public/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}'
```

Exec：

```bash
curl -i http://127.0.0.1:10004/health \
  -H 'X-Executor-Token: default-exec-token'
```

## 安全配置

生产环境至少修改：

```env
CHRONOFLOW_ADMIN_PASSWORD=change-me
JWT_SECRET=change-me
CHRONOFLOW_CALLBACK_TOKEN=change-me
EXECUTOR_TOKEN=change-me
```

`CHRONOFLOW_TOKEN_ENCRYPT_KEY` 必须是 32 字节，例如：

```env
CHRONOFLOW_TOKEN_ENCRYPT_KEY=12345678901234567890123456789012
```
