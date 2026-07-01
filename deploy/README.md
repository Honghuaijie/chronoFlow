# ChronoFlow 部署指南

本文档说明 ChronoFlow 的 Docker 部署方式、配置项、MySQL 初始化和常见网络场景。

## 文件说明

```text
deploy/
├── docker-compose.mysql.yml   # 只启动 MySQL，有状态基础设施
├── docker-compose.yml         # 启动 admin / exec / ui，源码构建镜像
├── docker-compose.image.yml   # 启动 admin / exec / ui，使用作者发布镜像
├── .env.example               # 配置模板，复制为 .env 后使用
├── mysql/init/001-init.sql    # MySQL 初始化 SQL
└── scripts/report.py          # 默认挂载到执行器容器的示例脚本
```

## 快速部署

### 1. 准备配置

```bash
cp .env.example .env
```

生产环境请至少修改 `.env` 中的管理员密码、JWT Secret、Callback Token 和执行器 Token。

### 2. 启动 MySQL

如果你需要使用 compose 内置 MySQL，先启动：

```bash
docker compose -f docker-compose.mysql.yml up -d
```

如果你已有外部 MySQL，可以跳过本步骤，直接修改 `.env` 里的数据库连接信息。

### 3. 启动应用

源码构建部署：

```bash
docker compose up -d --build
```

作者镜像部署：

```bash
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

## 服务说明

| 服务 | 所在 compose | 说明 | 默认访问 |
| --- | --- | --- | --- |
| `mysql` | `docker-compose.mysql.yml` | MySQL 8.0，保存元数据 | `127.0.0.1:3306` |
| `admin` | `docker-compose.yml` / `docker-compose.image.yml` | 调度器后端，连接 MySQL | `127.0.0.1:10003` |
| `exec` | `docker-compose.yml` / `docker-compose.image.yml` | 执行器后端，不连接数据库 | `127.0.0.1:10004` |
| `ui` | `docker-compose.yml` / `docker-compose.image.yml` | Nginx 托管的前端页面 | `127.0.0.1:5173` |

这些 compose 共用显式 Docker 网络：

```text
chronoflow
```

因此应用容器可以通过 `DB_HOST=mysql` 访问内置 MySQL。

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

`docker-compose.mysql.yml` 会通过官方 MySQL 镜像环境变量创建数据库和用户：

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

然后跳过 `docker-compose.mysql.yml`，只启动应用：

```bash
docker compose up -d --build
```

镜像部署模式：

```bash
docker compose -f docker-compose.image.yml up -d
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

## 飞书失败告警

ChronoFlow V1 支持飞书自定义机器人 Webhook 失败告警。配置步骤：

1. 在飞书群中进入群设置。
2. 添加自定义机器人。
3. 复制机器人 Webhook。
4. 登录 ChronoFlow，进入“系统设置”。
5. 粘贴 Webhook 并保存。
6. 点击“测试发送”，确认飞书群能收到测试卡片。
7. 创建或编辑任务时开启“失败告警”。

说明：

- V1 不支持飞书签名 Secret。
- 如需飞书安全策略，建议使用关键词校验，或暂不启用签名校验。
- 任务失败判断依赖进程退出码，不解析日志正文。
- Glue Shell 调用 Python 时推荐使用 `set -euo pipefail`，确保 Python 报错会让任务返回非 0 退出码。
- 页面不会回显 Webhook 明文，请自行保存好原始 Webhook。

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

默认使用 Docker volume：

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
docker compose -f docker-compose.mysql.yml down -v
docker compose down -v
```

## 常用命令

启动 MySQL：

```bash
docker compose -f docker-compose.mysql.yml up -d
```

启动源码构建应用：

```bash
docker compose up -d --build
```

启动镜像部署应用：

```bash
docker compose -f docker-compose.image.yml up -d
```

查看日志：

```bash
docker compose logs -f admin exec ui
docker compose -f docker-compose.mysql.yml logs -f mysql
```

重启 Admin：

```bash
docker compose restart admin
```

停止应用：

```bash
docker compose down
```

停止 MySQL：

```bash
docker compose -f docker-compose.mysql.yml down
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
