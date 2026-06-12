# ChronoFlow 测试与联调指南

本文档记录 ChronoFlow V1 的推荐验证方式。提交前至少应完成“基础验证”，发布前应完成“三端联调”和“部署验证”。

## 基础验证

### Admin

```bash
cd chronoFlow-admin
go test ./internal/... -count=1
go build -o /tmp/chronoflow-admin-build ./cmd/chronoFlow-admin
```

验证重点：

- 配置校验。
- token 加密。
- 日志文件存储。
- job run / callback / kill 状态流转。
- worker 生命周期。

### Exec

```bash
cd chronoFlow-exec
go test ./internal/... -count=1
go build -o /tmp/chronoflow-exec-build ./cmd/chronoFlow-exec
```

验证重点：

- 执行器不连接数据库。
- `/run` 异步执行，不依赖 HTTP request context。
- 同任务互斥。
- 日志截断。
- kill 进程组。
- pending callback 落盘与重试。

### UI

```bash
cd chronoFlow-ui
npm install
npm run build
```

验证重点：

- TypeScript 类型检查。
- Vue 模板编译。
- API 字段映射。
- 路由和页面基础渲染。

## 本地三端联调

### 1. 启动 Admin

```bash
cd chronoFlow-admin
go run ./cmd/chronoFlow-admin -conf ./configs
```

确认：

```bash
curl -sS -X POST http://127.0.0.1:10003/v1/public/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}'
```

### 2. 启动 Exec

```bash
cd chronoFlow-exec
go run ./cmd/chronoFlow-exec -conf ./configs
```

确认：

```bash
curl -i http://127.0.0.1:10004/health \
  -H 'X-Executor-Token: change-me'
```

### 3. 启动 UI

```bash
cd chronoFlow-ui
npm run dev
```

打开 Vite 输出的地址，默认通常是：

```text
http://127.0.0.1:5173/
```

如果端口被占用，以 Vite 控制台输出为准。

## 联调用例

### 登录与列表

1. 使用 `admin / admin123` 登录。
2. 进入任务列表。
3. 确认前端能读取 `/v1/admin/jobs/list`。
4. 浏览器控制台不应有业务错误。

### 普通运行

1. 创建执行器：
   - 地址：`http://127.0.0.1:10004`
   - Token：`change-me`
2. 创建任务。
3. 保存 Glue：

```bash
echo chronoflow-e2e-start
pwd
echo chronoflow-e2e-done
```

4. 点击“运行”。
5. 打开日志详情。
6. 期望：
   - 状态为 `success`。
   - 日志正文包含 `chronoflow-e2e-done`。
   - Glue 快照可见。

### 终止运行中任务

1. 创建任务并保存 Glue：

```bash
echo chronoflow-kill-before
sleep 120
echo chronoflow-kill-after
```

2. 点击“运行”。
3. 打开日志详情，确认状态为 `running`。
4. 点击“终止”。
5. 期望：
   - 状态最终变为 `killed`。
   - 错误信息为 `任务被终止`。
   - 日志正文包含 `chronoflow-kill-before`。
   - 日志正文不包含 `chronoflow-kill-after`。

### 同任务互斥

1. 对一个正在运行的任务再次点击“运行”。
2. 期望：
   - 前端按钮置灰或后端返回“任务正在执行中”。
   - 不创建新的执行日志。

## 部署验证

执行器 Docker 部署需要单独验证：

1. 宿主机脚本目录通过 volume 挂载到执行器容器。
2. Glue Shell 可以调用挂载目录中的 Python 脚本。
3. 执行器 `data_dir` 挂载为持久目录。
4. callback 失败后 pending 文件能保留并在恢复后重试。
5. Linux 环境下 kill 能终止整个进程组。

示例：

```bash
docker run --rm \
  -p 10004:10004 \
  -v /opt/chronoflow/scripts:/scripts \
  -v /opt/chronoflow/exec-data:/app/data \
  chronoFlow-exec:latest
```

## 提交前检查

```bash
git status --short
```

确认不要提交：

- `*.log`
- `data/`
- `node_modules/`
- `dist/`
- `/tmp/chronoflow-*`

确认需要提交：

- 源码。
- README / TESTING_GUIDE。
- 配置模板。
- `package-lock.json`。
