# ChronoFlow 本地 Docker 调试

本文档用于本机 Docker 调试，不是生产部署手册。

## 1. 准备本机 MySQL

本地 Docker 调试默认不启动新的 MySQL 容器。Admin 容器会连接宿主机 `3306` 端口上的 MySQL：

```text
host.docker.internal:3306
```

这适用于两种情况：

- 你的电脑原生安装了 MySQL，并监听 `3306`。
- 你已经有现成的 MySQL Docker 容器，例如 `boke-mysql`，并且它映射了 `0.0.0.0:3306->3306/tcp`。

请先在这个 MySQL 里创建数据库：

```sql
CREATE DATABASE IF NOT EXISTS chronoflow DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

默认连接信息在 `deploy/local/admin-conf/config.yaml`：

```yaml
host: host.docker.internal
port: "3306"
username: root
password: root
database: chronoflow
```

如果你的现有 MySQL 容器端口、账号或密码不同，改这个文件即可。

## 2. 启动容器

在项目根目录执行：

```bash
docker compose -f docker-compose.local.yml up -d --build --remove-orphans
```

服务端口：

| 服务 | Host 访问地址 | 容器内访问地址 |
| --- | --- | --- |
| Admin | `http://127.0.0.1:10003` | `http://chronoflow-admin:10003` |
| Exec | `http://127.0.0.1:10004` | `http://chronoflow-exec:10004` |

默认账号：

```text
admin / admin123
```

本地 Docker 调试用执行器 Token：

```text
local-exec-token
```

本地调试配置文件放在：

```text
deploy/local/admin-conf
deploy/local/exec-conf
```

compose 会把它们挂载到容器 `/data/conf`，避免本地调试时依赖环境变量占位符解析。

## 3. 启动 UI

UI 继续在本机用 Vite 跑，方便热更新：

```bash
cd chronoFlow-ui
npm install
VITE_API_PROXY_TARGET=http://127.0.0.1:10003 npm run dev
```

打开 Vite 输出的地址，一般是：

```text
http://127.0.0.1:5173/
```

## 4. 在 UI 新增执行器

执行器页面新增：

```text
名称：local-docker-exec
地址：http://chronoflow-exec:10004
Token：local-exec-token
```

注意：这里地址要填容器网络地址 `http://chronoflow-exec:10004`，因为 Admin 在容器里调用 Exec。

如果你只是从宿主机 curl Exec，才使用：

```text
http://127.0.0.1:10004
```

## 5. 创建任务并测试脚本挂载

在任务页面创建任务，选择 `local-docker-exec`。

Glue Shell 示例：

```bash
echo chronoflow-docker-start
python3 /scripts/report.py
echo chronoflow-docker-done
```

本地 Exec 镜像已内置 `python3`，可以直接调用挂载进容器的 Python 脚本。

`/scripts/report.py` 来自宿主机目录：

```text
deploy/local/scripts/report.py
```

该目录通过 compose 挂载到 Exec 容器：

```yaml
./deploy/local/scripts:/scripts:ro
```

## 6. 查看日志

Admin 日志正文目录：

```text
deploy/local/admin-data
```

Exec pending callback 数据目录：

```text
deploy/local/exec-data
```

服务日志目录：

```text
deploy/local/admin-logs
deploy/local/exec-logs
```

## 7. 常用命令

后台启动：

```bash
docker compose -f docker-compose.local.yml up -d --build --remove-orphans
```

查看日志：

```bash
docker compose -f docker-compose.local.yml logs -f admin exec
```

重启某个服务：

```bash
docker compose -f docker-compose.local.yml restart exec
```

停止服务：

```bash
docker compose -f docker-compose.local.yml down
```

## 8. 快速健康检查

Admin 登录：

```bash
curl -sS -X POST http://127.0.0.1:10003/v1/public/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}'
```

Exec health：

```bash
curl -i http://127.0.0.1:10004/health \
  -H 'X-Executor-Token: local-exec-token'
```

## 9. 注意事项

- 本地 compose 不启动新的 MySQL 容器，Admin 连接宿主机 `3306` 上已有的 MySQL，例如已映射端口的 `boke-mysql`。
- UI 从宿主机访问 Admin 使用 `127.0.0.1:10003`。
- Admin 调用 Exec 使用 `chronoflow-exec:10004`。
- Exec callback Admin 使用 `PUBLIC_BASE_URL=http://chronoflow-admin:10003`。
- 生产环境不要使用默认密码、默认 JWT secret、默认 callback token。
