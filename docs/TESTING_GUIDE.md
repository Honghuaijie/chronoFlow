# ChronoFlow V1 测试用例

本文档用于完整验收 ChronoFlow V1 是否好用。推荐按顺序执行：先基础构建，再本地 Docker 联调，最后逐条跑业务用例。

## 0. 本次实测结果

测试时间：2026-06-27

测试环境：

- MySQL：复用本机已有 Docker 容器 `your-mysql-container`，端口 `3306`，状态 healthy。
- Admin：`chronoflow-admin` Docker 容器，HTTP 端口 `10003`。
- Exec：`chronoflow-exec` Docker 容器，HTTP 端口 `10004`。
- UI：Docker/Nginx 或本地 Vite 服务 `http://127.0.0.1:5173`。
- 自动化测试数据后缀：`104240`。

总体结论：

- 主要业务链路通过：登录、执行器管理、任务管理、Glue、手动运行、定时调度、失败日志、callback 重试、日志文件存储、日志截断、Cron 可视化、下次运行时间展示均可用。
- 失败项：0 个。
- `TC-ERROR-003` 已按方案 A 调整为通过：Admin 重启期间如果执行器最终 callback 成功，则以真实执行结果为准。
- `TC-KILL-002` 已补充基于 Linux `/proc` 的进程组检查，不再依赖容器内的 `ps` 命令。
- `TC-UX-003` 未做完整多视口检查；当前系统主要提供电脑网页服务，该项暂不作为阻塞项。
- `TC-LOG-003` 自动脚本首次误判失败，原因是 `logSizeBytes` 返回为字符串；复核后确认 `logTruncated=true` 且大小为 `5242880`，按通过处理。

| 用例 | 结果 | 备注 |
| --- | --- | --- |
| 2.1 已有 MySQL 容器 | PASS | `your-mysql-container` 正常运行并映射 `3306`。 |
| 2.2 创建数据库 | PASS | Admin 已成功连接并完成全链路数据写入。 |
| 2.3 启动后端容器 | PASS | `chronoflow-admin`、`chronoflow-exec` 均正常运行。 |
| 2.4 启动前端 | PASS | Vite 服务可访问，已完成浏览器登录和页面操作。 |
| TC-BUILD-001 | PASS | Admin `go test ./internal/...` 和 `go build` 通过。 |
| TC-BUILD-002 | PASS | Exec `go test ./internal/...` 和 `go build` 通过；Exec 未连接数据库。 |
| TC-BUILD-003 | PASS | UI `npm run build` 通过，仅有 chunk size 提示。 |
| TC-API-001 | PASS | Admin 登录返回 token。 |
| TC-API-002 | PASS | Exec health 返回 online。 |
| TC-API-003 | PASS | 错误 token 被拒绝。 |
| TC-UI-001 | PASS | 浏览器登录成功，进入任务页，菜单可见。 |
| TC-UI-002 | PASS | 错误密码不能进入后台。 |
| TC-EXECUTOR-001 | PASS | 新增执行器成功。 |
| TC-EXECUTOR-002 | PASS | token 错误时执行器检测为 offline。 |
| TC-EXECUTOR-003 | PASS | 编辑执行器成功。 |
| TC-CRON-001 | PASS | 每 5 分钟表达式和最近 5 次运行时间展示正常。 |
| TC-CRON-002 | PASS | 每小时第 05 分钟表达式 `0 5 */1 * * *` 展示正常。 |
| TC-CRON-003 | PASS | 每天固定时间 tab 可用，预览正常。 |
| TC-CRON-004 | PARTIAL | 每周 tab 切换无异常；未逐项验证指定周几和时间组合。 |
| TC-CRON-005 | PARTIAL | 每月 tab 切换无异常；未逐项验证指定日期和时间组合。 |
| TC-CRON-006 | PASS | 手动输入 Cron 表达式可预览最近 5 次运行时间。 |
| TC-CRON-007 | PASS | 后端拒绝非法 Cron，返回 `cron_expr 必须是 6 位 Cron 表达式`。 |
| TC-JOB-001 | PASS | 新增任务成功。 |
| TC-JOB-002 | PASS | 编辑任务成功。 |
| TC-JOB-003 | PASS | 删除临时任务成功。 |
| TC-GLUE-001 | PASS | Glue 保存和读取成功。 |
| TC-GLUE-002 | PASS | Glue Shell 成功调用挂载目录中的 Python 脚本。 |
| TC-RUN-001 | PASS | 手动运行成功，日志状态 success。 |
| TC-RUN-002 | PASS | 脚本失败时记录 failed、exit code 和错误输出。 |
| TC-RUN-003 | PASS | 同一任务运行中再次运行被拒绝，未生成新的运行日志。 |
| TC-SCHEDULE-001 | PASS | 启动调度后按 Cron 生成运行日志。 |
| TC-SCHEDULE-002 | PASS | 停止调度后不再新增 Cron 运行日志。 |
| TC-KILL-001 | PASS | 长任务可终止，日志状态 killed。 |
| TC-KILL-002 | PASS | 状态变为 killed；已补充基于 Linux `/proc` 的进程组消失检查，不依赖 `ps`。 |
| TC-LOG-001 | PASS | 日志列表和筛选可用。 |
| TC-LOG-002 | PASS | 日志详情包含 Glue 快照和文件日志内容。 |
| TC-LOG-003 | PASS | 日志超过 5MB 后截断，`logTruncated=true`，`logSizeBytes=5242880`。 |
| TC-ERROR-001 | PASS | 执行器离线后运行任务失败。 |
| TC-ERROR-002 | PASS | Admin 暂停期间 callback 失败，恢复后执行器重试成功。 |
| TC-ERROR-003 | PASS | 按方案 A：Admin 重启后如果 Exec 最终 callback 成功，则日志以真实结果为准；本次实际为 success。 |
| TC-UX-001 | PASS | 任务列表展示下次运行时间，说明列不再过窄。 |
| TC-UX-002 | PASS | Cron 弹窗可用，浏览器控制台无 error。 |
| TC-UX-003 | NOT RUN | 未执行完整窄屏、多视口响应式检查。 |

待确认项：

- `TC-UX-003`：当前产品主要面向电脑网页服务，完整手机/平板多视口检查暂不作为阻塞项。
- `TC-CRON-004`、`TC-CRON-005`：每周、每月 tab 已验证切换无异常，后续可补更细的指定日期/时间组合测试。

## 1. 测试范围

本轮测试覆盖：

- Admin 后端：登录、执行器管理、任务管理、Glue、手动运行、定时运行、日志、终止任务。
- Exec 后端：健康检查、脚本执行、Python 脚本调用、日志回调、进程组终止。
- UI 前端：任务列表、执行器列表、Cron 可视化配置、下次运行时间、日志详情、使用说明。
- 本地 Docker：Admin 和 Exec 容器联调，Admin 连接已有 MySQL Docker 容器。

## 2. 前置条件

### 2.1 已有 MySQL 容器

确认你的 MySQL 容器正在运行，并映射宿主机 `3306`：

```bash
docker ps | grep your-mysql-container
```

期望看到类似：

```text
0.0.0.0:3306->3306/tcp   your-mysql-container
```

### 2.2 创建数据库

```bash
docker exec -it your-mysql-container mysql -uroot -p
```

进入 MySQL 后执行：

```sql
CREATE DATABASE IF NOT EXISTS chronoflow DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

复制部署配置，并确认 `deploy/.env` 中数据库账号密码正确：

```bash
cd /path/to/chronoFlow/deploy
cp .env.example .env
```

```env
DB_HOST=host.docker.internal
DB_PORT=3306
DB_USER=root
DB_PASSWORD=root
DB_NAME=chronoflow
```

### 2.3 启动容器

```bash
cd /path/to/chronoFlow/deploy
docker compose up -d --build
```

确认容器状态：

```bash
docker compose ps
```

期望：

```text
chronoflow-admin   Up
chronoflow-exec    Up
chronoflow-ui      Up
```

### 2.4 打开前端

打开：

```text
http://127.0.0.1:5173
```

默认账号：

```text
admin / admin123
```

说明：这是本地测试默认账号。公开部署或生产环境请在 `deploy/.env` 中修改 `CHRONOFLOW_ADMIN_PASSWORD`。

## 3. 基础构建测试

### TC-BUILD-001 Admin 单元测试和构建

步骤：

```bash
cd /path/to/chronoFlow/chronoFlow-admin
go test ./internal/... -count=1
go build -o /tmp/chronoflow-admin-build ./cmd/chronoFlow-admin
```

预期：

- 测试全部通过。
- 构建成功。

### TC-BUILD-002 Exec 单元测试和构建

步骤：

```bash
cd /path/to/chronoFlow/chronoFlow-exec
go test ./internal/... -count=1
go build -o /tmp/chronoflow-exec-build ./cmd/chronoFlow-exec
```

预期：

- 测试全部通过。
- 构建成功。
- Exec 不需要连接数据库。

### TC-BUILD-003 UI 构建

步骤：

```bash
cd /path/to/chronoFlow/chronoFlow-ui
npm run build
```

预期：

- TypeScript 检查通过。
- Vite 构建成功。

## 4. API 冒烟测试

### TC-API-001 Admin 登录

步骤：

```bash
curl -sS -X POST http://127.0.0.1:10003/v1/public/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}'
```

预期：

- 返回 `code=0`。
- 返回 `data.token`。

### TC-API-002 Exec 健康检查

步骤：

```bash
curl -i http://127.0.0.1:10004/health \
  -H 'X-Executor-Token: your-executor-token'
```

预期：

- HTTP 200。
- 返回 `status=online`。
- 返回 `executorName=local-docker-exec`。

### TC-API-003 Exec token 错误

步骤：

```bash
curl -i http://127.0.0.1:10004/health \
  -H 'X-Executor-Token: wrong-token'
```

预期：

- 返回认证失败。
- 不应返回 `status=online`。

## 5. UI 登录和导航测试

### TC-UI-001 登录成功

步骤：

1. 打开前端页面。
2. 输入本地测试默认账号 `admin / admin123`，或你在 `deploy/.env` 中配置的管理员账号密码。
3. 点击登录。

预期：

- 登录成功。
- 跳转到任务页面。
- 左侧菜单可切换：任务、执行器、执行日志、设置、使用说明。

### TC-UI-002 登录失败

步骤：

1. 输入错误密码。
2. 点击登录。

预期：

- 页面提示登录失败。
- 不进入后台页面。

## 6. 执行器管理测试

### TC-EXECUTOR-001 新增执行器

步骤：

1. 进入“执行器”页面。
2. 点击“新增执行器”。
3. 填写：

```text
名称：local-docker-exec
地址：http://chronoflow-exec:10004
Token：your-executor-token
说明：本地 Docker 执行器
```

4. 保存。

预期：

- 保存成功。
- 列表出现 `local-docker-exec`。
- 等待健康检查后状态变为 `online`。

注意：

- 地址必须填 `http://chronoflow-exec:10004`，因为 Admin 容器通过 Docker 网络访问 Exec。
- 从宿主机 curl 才使用 `http://127.0.0.1:10004`。

### TC-EXECUTOR-002 执行器 token 错误

步骤：

1. 新增或编辑一个执行器。
2. 地址填 `http://chronoflow-exec:10004`。
3. Token 填 `wrong-token`。
4. 保存并等待约 30 秒。

预期：

- 健康检查失败。
- 执行器状态变为 `offline`。

### TC-EXECUTOR-003 编辑执行器

步骤：

1. 编辑 `local-docker-exec`。
2. 修改说明。
3. 保存。

预期：

- 保存成功。
- 列表说明更新。
- Token 未改错时，健康检查仍可恢复为 `online`。

## 7. Cron 可视化配置测试

### TC-CRON-001 每 5 分钟执行

步骤：

1. 新增任务或编辑任务。
2. 打开 Cron 配置。
3. 选择“分钟”。
4. 间隔分钟填 `5`。

预期：

- 表达式为 `0 */5 * * * *`。
- 说明为“每 5 分钟”。
- 最近 5 次运行时间每次间隔 5 分钟。

### TC-CRON-002 每小时第 05 分钟执行

步骤：

1. 打开 Cron 配置。
2. 选择“小时”。
3. 填写：

```text
间隔小时：1
分：5
秒：0
```

预期：

- 表达式为 `0 5 */1 * * *`。
- 说明为“每 1 小时的第 05 分钟”。
- 最近 5 次运行时间类似 `10:05、11:05、12:05`。

### TC-CRON-003 每天固定时间执行

步骤：

1. 打开 Cron 配置。
2. 选择“日”。
3. 填写：

```text
时：2
分：30
秒：0
```

预期：

- 表达式为 `0 30 2 * * *`。
- 说明为“每天 02:30:00”。
- 最近 5 次运行时间每天一次。

### TC-CRON-004 每周固定时间执行

步骤：

1. 打开 Cron 配置。
2. 选择“周”。
3. 选择 `周一`。
4. 填写 `09:00:00`。

预期：

- 表达式为 `0 0 9 * * 1`。
- 最近 5 次运行时间都在周一 09:00:00。

### TC-CRON-005 每月固定日期执行

步骤：

1. 打开 Cron 配置。
2. 选择“月”。
3. 日期填 `1`。
4. 填写 `08:00:00`。

预期：

- 表达式为 `0 0 8 1 * *`。
- 最近 5 次运行时间都在每月 1 日 08:00:00。

### TC-CRON-006 手动输入表达式

步骤：

1. 打开 Cron 配置。
2. 选择“手动”。
3. 输入：

```text
0 10,20,30 * * * *
```

预期：

- 最近 5 次运行时间只出现在每小时第 10、20、30 分钟。
- 点击 OK 后，任务表单中的 Cron 表达式同步为手动输入值。

### TC-CRON-007 无效表达式

步骤：

1. 打开 Cron 配置。
2. 选择“手动”。
3. 输入：

```text
bad cron
```

预期：

- 预览显示“无法计算”或“Cron 表达式需为 6 段”。
- 保存任务时后端返回 Cron 表达式不合法。

## 8. 任务管理测试

### TC-JOB-001 新增任务

步骤：

1. 进入“任务”页面。
2. 点击“新增任务”。
3. 填写：

```text
任务名称：test-basic
执行器：local-docker-exec
Cron：0 */5 * * * *
超时时间：3600
说明：测试任务
```

4. 保存。

预期：

- 保存成功。
- 任务列表出现 `test-basic`。
- 调度状态为 `已停止`。
- 执行状态为 `空闲`。
- 说明列不会被挤成竖排。

### TC-JOB-002 编辑任务

步骤：

1. 编辑 `test-basic`。
2. 修改说明为 `测试任务-编辑后`。
3. 保存。

预期：

- 保存成功。
- 任务列表说明更新。
- 如果当前有运行实例，新配置只对下次运行生效。

### TC-JOB-003 删除任务

步骤：

1. 新建一个临时任务 `test-delete`。
2. 点击删除。
3. 确认删除。

预期：

- 删除成功。
- 列表不再显示该任务。

## 9. Glue Shell 测试

### TC-GLUE-001 保存 Glue

步骤：

1. 在任务 `test-basic` 上点击 `Glue`。
2. 输入：

```bash
echo chronoflow-glue-start
pwd
echo chronoflow-glue-done
```

3. 保存。

预期：

- 保存成功。
- 关闭后再次打开 Glue，内容仍存在。

### TC-GLUE-002 调用挂载 Python 脚本

步骤：

1. 在 Glue 中输入：

```bash
echo chronoflow-python-start
python3 /scripts/report.py
echo chronoflow-python-done
```

2. 保存。
3. 手动运行任务。
4. 查看日志详情。

预期：

- 状态为 `success`。
- 日志正文包含：

```text
chronoflow-python-start
chronoflow docker script ok
chronoflow-python-done
```

## 10. 手动运行测试

### TC-RUN-001 手动运行成功

步骤：

1. 确认任务 `test-basic` 已保存 Glue。
2. 点击“运行”。
3. 进入“执行日志”页面。
4. 打开最新日志详情。

预期：

- 运行按钮下发成功。
- 最新日志状态最终变为 `success`。
- `exitCode=0`。
- 日志详情显示 Glue 快照和脚本输出。

### TC-RUN-002 脚本失败

步骤：

1. Glue 改为：

```bash
echo chronoflow-fail-start
exit 2
```

2. 保存并运行。
3. 查看日志详情。

预期：

- 状态为 `failed`。
- `exitCode=2`。
- 日志正文包含 `chronoflow-fail-start`。

### TC-RUN-003 同任务互斥

步骤：

1. Glue 改为：

```bash
echo chronoflow-lock-start
sleep 60
echo chronoflow-lock-done
```

2. 保存并点击运行。
3. 在任务仍运行中时，再次观察任务列表运行按钮。

预期：

- 同一个任务运行中时，“运行”按钮不可用。
- 后端不应生成第二条运行日志。
- 不同任务仍可并行运行。

## 11. 定时调度测试

### TC-SCHEDULE-001 启动调度

步骤：

1. Glue 改为：

```bash
echo chronoflow-schedule-start
date
echo chronoflow-schedule-done
```

2. Cron 设置为 `0 */1 * * * *`。
3. 点击“启动”。

预期：

- 调度状态变为 `运行中`。
- 任务列表显示“下次运行”。
- 到达下次运行时间后，自动生成一条触发类型为 `cron` 的日志。
- 日志状态为 `success`。

### TC-SCHEDULE-002 停止调度

步骤：

1. 对运行中的调度点击“停止”。
2. 等待超过一个 Cron 周期。

预期：

- 调度状态变为 `已停止`。
- 不再自动生成新的定时日志。
- 如果停止前已有运行实例，当前运行不受影响。

## 12. 终止任务测试

### TC-KILL-001 终止长任务

步骤：

1. Glue 改为：

```bash
echo chronoflow-kill-before
sleep 120
echo chronoflow-kill-after
```

2. 保存并运行。
3. 日志或任务列表显示运行中后，点击“终止”。
4. 查看日志详情。

预期：

- 状态先进入 `killing`。
- 最终变为 `killed`。
- 错误信息为 `任务被终止`。
- 日志正文包含 `chronoflow-kill-before`。
- 日志正文不包含 `chronoflow-kill-after`。

### TC-KILL-002 终止含子进程任务

步骤：

1. Glue 改为：

```bash
echo chronoflow-process-group-before
sh -c "sleep 120" &
wait
echo chronoflow-process-group-after
```

2. 运行后点击终止。

预期：

- 任务最终变为 `killed`。
- 子进程应被一并终止。
- 不应残留持续运行的 `sleep 120`。

可用命令辅助检查：

```bash
docker exec chronoflow-exec ps -ef | grep sleep
```

## 13. 日志测试

### TC-LOG-001 日志列表

步骤：

1. 进入“执行日志”页面。
2. 使用任务、状态、触发类型筛选。

预期：

- 日志列表正常展示。
- 筛选条件生效。
- 运行中和终止中任务能显示操作按钮。

### TC-LOG-002 日志详情

步骤：

1. 打开一条成功日志。
2. 查看元数据、Glue 快照、日志正文。

预期：

- 元数据完整。
- Glue 快照与运行时保存的脚本一致。
- 日志正文从文件读取，不依赖 MySQL 保存完整正文。

### TC-LOG-003 日志截断

步骤：

1. Glue 改为大量输出：

```bash
for i in $(seq 1 200000); do
  echo "line-$i xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
done
```

2. 运行并查看日志详情。

预期：

- 任务能结束。
- 日志大小不超过配置限制。
- 日志详情显示截断标记。

## 14. 异常场景测试

### TC-ERROR-001 执行器离线后运行任务

步骤：

1. 停止 Exec 容器：

```bash
docker stop chronoflow-exec
```

2. 等待约 30 秒。
3. 手动运行任务。

预期：

- 执行器健康状态变为 `offline`。
- 运行失败，并有明确错误提示。

恢复：

```bash
cd /path/to/chronoFlow/deploy
docker compose up -d exec
```

### TC-ERROR-002 callback 失败后重试

步骤：

1. 运行一个短任务。
2. 在任务运行期间临时停止 Admin 容器。
3. 等待 Exec 生成 pending callback。
4. 恢复 Admin。

预期：

- Exec 将待回调结果落盘。
- Admin 恢复后，Exec 后台重试 callback。
- 日志最终更新为正确状态。

辅助查看：

```bash
docker volume inspect chronoflow_chronoflow-exec-data
```

### TC-ERROR-003 Admin 重启恢复 running 日志

步骤：

1. 运行一个长任务。
2. 重启 Admin：

```bash
docker restart chronoflow-admin
```

3. 等待启动恢复逻辑完成。

预期：

- 如果执行结果未知，运行中日志被标记为 `failed`。
- 错误信息为 `执行器重启或失联，执行结果未知`。

## 15. 页面体验验收

### TC-UX-001 任务列表可读性

步骤：

1. 创建说明较长的任务。
2. 查看任务列表。

预期：

- 说明列不会挤成竖排。
- 说明最多展示两行。
- 鼠标悬浮可查看完整说明。
- 操作按钮不与其他列重叠。

### TC-UX-002 Cron 弹窗可读性

步骤：

1. 打开 Cron 配置。
2. 切换分钟、小时、日、周、月、手动。
3. 调整各输入值。

预期：

- 表达式实时变化。
- 说明实时变化。
- 最近 5 次运行时间实时变化。
- 弹窗内容不溢出、不遮挡按钮。

### TC-UX-003 响应式检查

步骤：

1. 浏览器切换到较窄宽度。
2. 查看任务列表、弹窗、日志详情。

预期：

- 表格可横向滚动。
- 弹窗内容仍可操作。
- 文本不重叠。

## 16. 测试完成标准

全部通过时应满足：

- Admin、Exec、UI 构建通过。
- 本地 Docker 后端可启动。
- UI 可登录。
- 执行器健康检查正常。
- 手动运行成功。
- 定时调度成功。
- 终止任务成功。
- Glue 调 Python 脚本成功。
- 日志列表和详情正确。
- Cron 可视化配置可用，并能展示最近 5 次运行时间。
- 任务列表下次运行时间、说明列展示正常。

## 17. 提交前检查

```bash
git status --short
```

不要提交：

- `*.log`
- `data/`
- `node_modules/`
- `dist/`
- `/tmp/chronoflow-*`

可以提交：

- 源码。
- README / TESTING_GUIDE。
- 配置模板。
- Docker 本地调试配置。
