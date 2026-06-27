# ChronoFlow 定时任务平台 PRD V1

## 1. 项目背景

目前很多定时脚本分散运行在不同服务器上，常见方式是通过 `crontab` 或手动登录服务器执行脚本。

这种方式存在以下问题：

1. 脚本分散在不同服务器，不方便统一管理。
2. 定时规则不直观，查看和修改成本较高。
3. 脚本执行成功或失败后，不方便统一查看日志。
4. 多台服务器上的脚本缺少统一调度中心。
5. 任务执行异常时，排查成本较高。
6. 后续如果脚本数量增加，单纯依赖 `crontab` 管理会越来越混乱。

因此需要开发一个轻量级定时任务平台，用于统一管理执行器、定时任务、Glue Shell 脚本和执行日志。

---

## 2. 项目定位

项目总名：**ChronoFlow**

项目介绍：

> ChronoFlow 是一个使用 Go 编写的轻量级分布式定时任务平台，支持通过 Web 调度中心管理执行器、Cron 任务、Glue Shell 脚本和执行日志。

ChronoFlow V1 面向以下场景：

1. 内网单团队自用。
2. 任务规模为几十个以内。
3. 单调度器部署。
4. 多执行器部署。
5. 轻量优先，但需要支持长任务、任务终止和执行结果可靠回调。

ChronoFlow V1 不面向以下场景：

1. 多调度器高可用。
2. 多租户平台。
3. 公网暴露执行器。
4. Kubernetes 多副本调度器并发。
5. DAG 工作流或复杂任务编排。

---

## 3. 子项目命名

| 项目 | 说明 |
| --- | --- |
| `chronoflow-admin` | 调度器后端 |
| `chronoflow-ui` | 调度中心前端 |
| `chronoflow-exec` | 执行器后端 |

---

## 4. V1 项目目标

ChronoFlow V1 的目标是实现一个类似 XXL-JOB 的轻量级定时任务平台。

第一版主要支持：

1. 通过 Web 页面手动添加执行器。
2. 调度器定时检测执行器在线状态。
3. 按执行器维度管理定时任务。
4. 使用 6 位 Cron 表达式创建定时任务。
5. 支持 Glue Shell 脚本在线编辑。
6. 支持 Shell 脚本调用服务器上的 Python 脚本。
7. 支持手动运行任务。
8. 支持启动、停止任务。
9. 支持终止正在运行的任务。
10. 支持查看任务执行日志。
11. 支持任务执行超时控制。
12. 支持调度器重启后自动恢复已启动任务。
13. 支持异步执行和执行器结果回调。
14. 支持执行器回调失败后的本地落盘和重试。
15. 调度器调用执行器时使用 token 鉴权。
16. 执行器回调调度器时使用 callback token 鉴权。
17. 执行器 token 加密存储。

---

## 5. 核心架构

ChronoFlow 由三部分组成：

```text
ChronoFlow
├── chronoflow-admin
│   └── 调度器后端
│
├── chronoflow-ui
│   └── 调度中心前端
│
└── chronoflow-exec
    └── 执行器后端
```

整体调用关系：

```text
用户
 ↓
chronoflow-ui 调度中心前端
 ↓
chronoflow-admin 调度器后端
 ↓
chronoflow-exec 执行器后端
 ↓
Linux 服务器或执行器容器内执行 Shell 脚本
```

异步执行关系：

```text
chronoflow-admin -> chronoflow-exec：下发任务
chronoflow-exec -> chronoflow-admin：回调执行结果
```

---

## 6. 角色说明

### 6.1 调度器

调度器是平台的核心控制中心。

调度器负责：

1. 提供 Web 后端 API。
2. 管理执行器。
3. 管理定时任务。
4. 管理 Glue Shell 脚本。
5. 管理执行日志元数据。
6. 保存执行日志正文文件。
7. 根据 Cron 表达式触发任务。
8. 将任务下发给指定执行器。
9. 接收执行器 callback 返回的执行结果。
10. 调度器重启后恢复已启动任务。
11. 清理过期执行日志和日志文件。
12. 统一连接 MySQL 并完成所有数据库读写。

### 6.2 调度中心前端

调度中心前端用于用户操作平台。

前端负责：

1. 登录页面。
2. 执行器管理页面。
3. 任务管理页面。
4. Glue 脚本编辑页面。
5. 执行日志页面。
6. 系统设置页面。

### 6.3 执行器

执行器部署在需要运行脚本的 Linux 服务器上。

执行器负责：

1. 提供健康检查接口。
2. 提供执行 Shell 脚本接口。
3. 提供终止任务接口。
4. 校验调度器请求中的 `executor_token`。
5. 接收调度器下发的 Glue Shell 脚本。
6. 在本机或容器内临时生成 `.sh` 文件。
7. 使用 `os/exec` 执行 Shell 脚本。
8. 为每次任务创建独立 Linux 进程组。
9. 收集 `stdout`、`stderr`、`exit_code`。
10. 执行完成后先将 callback 结果落盘。
11. 回调调度器更新执行结果。
12. callback 失败时定期重试。

执行器不负责：

1. 连接 MySQL。
2. 直接读写调度器数据库。
3. 直接读取任务、Glue 或执行日志表。

所有数据库访问统一由 `chronoflow-admin` 完成，`chronoflow-exec` 只通过 HTTP 接口与 `chronoflow-admin` 通信。

---

## 7. 部署方式

### 7.1 单机部署

调度器、前端、执行器、MySQL 可以部署在同一台服务器上。

```text
同一台 Linux 服务器
├── chronoflow-admin
├── chronoflow-ui
├── chronoflow-exec
└── MySQL
```

### 7.2 多服务器部署

调度器和执行器可以部署在不同服务器。

```text
调度中心服务器
├── chronoflow-admin
├── chronoflow-ui
└── MySQL

执行服务器 A
└── chronoflow-exec

执行服务器 B
└── chronoflow-exec

执行服务器 C
└── chronoflow-exec
```

### 7.3 执行器系统要求

ChronoFlow V1 的执行器仅支持 Linux。

规则：

1. Glue Shell 默认使用 `/bin/bash` 执行。
2. 任务超时和用户终止基于 Linux 进程组信号机制实现。
3. V1 不支持 Windows 执行器。
4. `chronoflow-admin` 和 `chronoflow-ui` 可以在开发环境中运行，但 `chronoflow-exec` 的进程组终止能力需要在 Linux 环境验证。

### 7.4 Docker 部署约定

`chronoflow-exec` 支持 Docker 部署。

Docker 模式下：

1. Glue Shell 在 `chronoflow-exec` 容器内部执行。
2. 业务 Python 脚本、虚拟环境、配置文件、数据目录必须通过镜像内置或 volume mount 提供给容器。
3. V1 推荐将宿主机业务脚本目录挂载到执行器容器中。
4. Glue Shell 中填写的是容器内路径，不是宿主机路径。
5. 执行器只保证终止容器内由本次 Glue 启动的进程组。
6. V1 不支持通过容器内执行器直接管理宿主机任意进程。

推荐挂载示例：

```bash
docker run \
  -v /data/scripts:/data/scripts \
  -v /data/chronoflow-exec:/app/data \
  chronoflow-exec
```

Glue 示例：

```bash
#!/bin/bash

cd /data/scripts/data-process
source venv/bin/activate

python3 sync_card_status.py
```

### 7.5 日志目录挂载

`chronoflow-admin` 会将执行日志正文保存到本地文件系统。

Docker 部署时必须挂载数据目录，否则容器删除后日志文件会丢失。

推荐挂载示例：

```bash
docker run \
  -v /data/chronoflow-admin:/app/data \
  chronoflow-admin
```

---

## 8. 技术栈

| 类型 | 技术 |
| --- | --- |
| 后端 | Go + Kratos |
| 数据库 | MySQL |
| 前端 | Vue3 + Ant Design Vue + Pinia + TypeScript |
| 调度库 | robfig/cron |
| 脚本执行 | os/exec |
| 日志收集 | stdout + stderr 合并 |
| 日志正文存储 | chronoflow-admin 本地文件系统 |
| Token 加密 | AES-GCM |

说明：

1. `robfig/cron` 用于在调度器中根据 Cron 表达式触发任务。
2. `os/exec` 用于执行器执行 Shell 脚本。
3. `stdout + stderr` 合并为一份执行日志。
4. MySQL 只保存执行日志元数据，不保存完整日志正文。
5. 完整日志正文保存到 `chronoflow-admin` 本地文件系统。
6. 只有 `chronoflow-admin` 连接 MySQL，`chronoflow-exec` 不连接 MySQL。

---

## 9. 页面菜单

ChronoFlow V1 前端包含以下菜单：

```text
调度中心
├── 执行器管理
├── 任务管理
├── 执行日志
└── 系统设置
```

### 9.1 执行器管理

功能包括：

1. 新增执行器。
2. 编辑执行器。
3. 删除执行器。
4. 查看执行器在线状态。
5. 手动触发执行器健康检查。

### 9.2 任务管理

功能包括：

1. 选择执行器。
2. 查看当前执行器下的任务列表。
3. 新增任务。
4. 编辑任务基础信息。
5. 编辑 Glue Shell 脚本。
6. 手动运行任务。
7. 启动任务。
8. 停止任务。
9. 终止正在运行的任务。
10. 删除任务。

### 9.3 执行日志

功能包括：

1. 查看任务执行记录。
2. 查看日志详情。
3. 查看日志是否被截断。
4. 按任务筛选。
5. 按执行器筛选。
6. 按执行状态筛选。
7. 按触发方式筛选。
8. 按时间范围筛选。
9. 终止正在运行的日志。

### 9.4 系统设置

功能包括：

1. 查看管理员账号信息。
2. 修改密码。
3. 查看系统配置摘要。

V1 系统设置页面不要求支持在线修改调度器配置。

---

## 10. 登录与账号

### 10.1 V1 登录规则

第一版做简单登录。

规则：

1. 系统初始化一个管理员账号。
2. 用户需要登录后才能进入调度中心。
3. 第一版不做注册功能。
4. 第一版不做多用户管理页面。
5. 第一版不做复杂角色权限。
6. 账号信息保存到数据库 `users` 表中，为后续扩展做准备。

### 10.2 后续账号规划

后续版本支持管理员分配账号。

后续功能包括：

1. 管理员创建账号。
2. 管理员分配账号给其他用户。
3. 管理员启用或禁用账号。
4. 管理员重置用户密码。
5. 后续扩展角色权限。

### 10.3 users 表建议字段

| 字段 | 说明 |
| --- | --- |
| id | 用户 ID |
| username | 登录账号 |
| password_hash | 加密后的密码 |
| role | 角色，第一版只有 admin |
| status | 状态：enabled / disabled |
| created_at | 创建时间 |
| updated_at | 更新时间 |
| deleted_at | 删除时间，预留字段 |

---

## 11. 执行器管理

### 11.1 执行器添加方式

ChronoFlow V1 使用 Web 手动添加执行器。

第一版不支持执行器主动注册。

新增执行器流程：

```text
1. 用户登录调度中心
2. 进入执行器管理页面
3. 点击新增执行器
4. 填写执行器名称、地址、token、描述
5. 保存执行器
6. 调度器加密保存 token
7. 调度器后续定时检测执行器在线状态
```

### 11.2 执行器字段

| 字段 | 说明 |
| --- | --- |
| id | 执行器 ID |
| name | 执行器名称 |
| address | 执行器地址，例如 `http://192.168.1.10:9999` |
| token_ciphertext | 加密后的执行器 token |
| description | 描述 |
| status | 状态：online / offline |
| heartbeat_fail_count | 连续健康检查失败次数 |
| last_heartbeat_time | 最近一次检测成功时间 |
| created_at | 创建时间 |
| updated_at | 更新时间 |
| deleted_at | 删除时间 |

### 11.3 Token 展示与编辑规则

1. 执行器 token 不明文保存。
2. 执行器 token 入库前使用 AES-GCM 加密。
3. 调度器调用执行器时，从数据库读取密文并解密后放入 `X-Executor-Token`。
4. 前端列表和详情不展示完整 token，只显示脱敏值，例如 `****abcd`。
5. 编辑执行器时，token 字段默认为空。
6. 用户不填写 token 表示不修改 token。
7. 用户填写 token 表示更新 token，并重新加密入库。
8. 如果 token 加密密钥丢失，历史 token 无法解密，需要重新配置执行器 token。

### 11.4 在线检测规则

执行器需要提供健康检查接口。

接口：

```http
GET /health
X-Executor-Token: xxxxxx
```

在线检测规则：

1. 调度器每 10 秒请求一次执行器健康检查接口。
2. 请求成功，将执行器标记为 online。
3. 请求成功时，重置 `heartbeat_fail_count = 0`。
4. 请求失败时，`heartbeat_fail_count + 1`。
5. 连续 3 次健康检查失败后，将执行器标记为 offline。
6. 执行器 offline 后，后续任意一次健康检查成功，则标记为 online。
7. 删除后的执行器不再做在线检测。
8. 调度器调用执行器所有接口都必须携带 `X-Executor-Token`。

### 11.5 执行器失联处理

如果执行器从 online 变为 offline：

1. 调度器查询该执行器下 `status = running` 或 `status = killing` 的执行日志。
2. 将这些执行日志标记为 failed。
3. `end_time` 设置为当前时间。
4. `duration_ms` 根据 `start_time` 到当前时间计算。
5. `error_message` 写：`执行器重启或失联，执行结果未知`。
6. 如果执行器后续恢复并 callback 这些 `log_id`，调度器因为日志已不是中间状态，应忽略重复 callback。

### 11.6 执行器删除规则

执行器删除规则：

1. 如果执行器下面还有未删除任务，不允许删除执行器。
2. 需要先删除该执行器下面的所有任务。
3. 执行器删除使用软删除。
4. 删除后执行器列表不再显示。
5. 删除后不再做在线检测。
6. 历史执行日志保留，直到日志保留策略清理。

---

## 12. 任务管理

### 12.1 任务创建流程

ChronoFlow V1 的任务创建流程如下：

```text
1. 进入调度中心 Web 页面
2. 先选择一个已经注册好的执行器
3. 选择执行器后，任务列表显示该执行器下的定时任务
4. 点击新增任务
5. 填写任务名称
6. 填写 Cron 表达式
7. 填写超时时间，可不填
8. 保存任务
9. 选中刚创建的任务
10. 点击 Glue
11. 编辑该任务的 Shell 脚本
12. 保存 Glue 脚本
13. 可以先手动运行一次
14. 查看执行日志
15. 如果没有问题，点击启动按钮
16. 调度器根据 Cron 表达式定时触发
```

### 12.2 任务字段

| 字段 | 说明 |
| --- | --- |
| id | 任务 ID |
| executor_id | 所属执行器 ID |
| name | 任务名称 |
| cron_expr | Cron 表达式 |
| timeout_seconds | 超时时间，默认 600 秒 |
| schedule_status | 调度状态：stopped / running |
| description | 描述 |
| created_at | 创建时间 |
| updated_at | 更新时间 |
| deleted_at | 删除时间 |

### 12.3 Cron 表达式规则

第一版直接支持 Cron 表达式。

Cron 表达式统一使用 6 位格式：

```text
秒 分 时 日 月 周
```

示例：

```text
*/10 * * * * *     每 10 秒执行一次
0 */5 * * * *      每 5 分钟执行一次
0 0 1 * * *        每天 01:00 执行
0 0 0 * * 1        每周一 00:00 执行
```

规则：

1. 创建任务时必须填写 Cron 表达式。
2. 创建任务时需要校验 Cron 表达式是否合法。
3. Cron 表达式不合法，不允许保存。
4. 编辑任务时如果修改 Cron 表达式，也必须重新校验。
5. Cron 调度默认使用 `Asia/Shanghai` 时区。
6. Cron 调度时区可通过 `chronoflow-admin` 配置修改。
7. V1 不支持每个任务单独设置时区。
8. 调度器停机期间错过的 Cron 触发不补偿执行。

### 12.4 任务状态

任务有两个不同概念：调度状态和最近执行状态。

#### 调度状态

字段：`schedule_status`

| 状态 | 含义 |
| --- | --- |
| stopped | 停止，不参与 Cron 调度 |
| running | 启动，参与 Cron 调度 |

#### 最近执行状态

最近执行状态来自最新一条执行日志。

| 状态 | 含义 |
| --- | --- |
| none | 从未执行 |
| running | 最近一次正在执行 |
| killing | 最近一次正在终止 |
| success | 最近一次执行成功 |
| failed | 最近一次执行失败 |
| timeout | 最近一次执行超时 |
| skipped | 最近一次调度跳过 |
| killed | 最近一次被用户终止 |

### 12.5 任务按钮规则

任务按钮规则：

1. 新建任务后默认停止状态。
2. 没有 Glue 脚本，不允许手动运行。
3. 没有 Glue 脚本，不允许启动。
4. 停止状态下，可以手动运行。
5. 启动状态下，也可以手动运行。
6. 点击启动后，任务才参与 Cron 调度。
7. 点击停止后，任务不参与 Cron 调度，但仍然可以手动运行。
8. 任务有运行中实例时，手动运行按钮置灰。
9. 任务有运行中实例时，可以显示终止按钮。
10. 终止按钮只对 `running` 状态的执行日志可用。

### 12.6 任务编辑规则

任务正在执行时允许编辑。

规则：

1. 任务执行开始时，调度器生成本次执行快照。
2. 执行快照包括任务名称、执行器、Cron、超时时间、Glue 内容。
3. 当前正在运行的实例不受后续编辑影响。
4. 修改后的任务配置只影响下一次手动运行或 Cron 调度。
5. 如果任务处于启动状态，修改 `cron_expr` 后，调度器立即重新注册 Cron。
6. 旧 Cron 规则停止生效，新 Cron 规则从保存成功后开始生效。

### 12.7 任务删除规则

任务删除规则：

1. 删除任务使用软删除。
2. 删除后任务列表不再显示。
3. 删除后不再参与调度。
4. 删除后历史执行日志保留，直到日志保留策略清理。
5. 日志详情里仍然能看到当时的任务名称、执行器、Cron、Glue 内容快照。

---

## 13. Glue Shell 脚本

### 13.1 Glue 定义

Glue 是任务对应的 Shell 脚本内容。

第一版 Glue 只支持 Shell 脚本。

Glue 内容是一整段 Shell 脚本，而不是单行命令。

典型场景是 Shell 脚本调用业务 Python 脚本。

示例：

```bash
#!/bin/bash

cd /data/scripts/data-process
source venv/bin/activate

python3 sync_card_status.py
```

### 13.2 Glue 存储

Glue 脚本保存到调度器数据库中。

任务执行时：

1. 调度器读取任务对应的 Glue 脚本。
2. 调度器将 Glue 内容写入本次执行日志快照。
3. 调度器把 Glue 脚本内容发送给执行器。
4. 执行器在本机或容器内生成临时 `.sh` 文件。
5. 执行器使用 `/bin/bash` 执行该 `.sh` 文件。
6. 执行完成后删除临时文件。
7. 执行器收集 stdout、stderr、exit_code。
8. 执行器通过 callback 将执行结果返回调度器。

### 13.3 Glue 字段

| 字段 | 说明 |
| --- | --- |
| id | Glue ID |
| job_id | 任务 ID |
| content | Shell 脚本内容 |
| created_at | 创建时间 |
| updated_at | 更新时间 |

### 13.4 Glue 运行限制

1. ChronoFlow 不负责同步 Python 脚本文件。
2. ChronoFlow 不负责创建 Python 虚拟环境。
3. 用户需要自己保证业务脚本目录和虚拟环境存在。
4. Docker 模式下，用户需要自己保证宿主机目录已挂载到执行器容器。
5. Glue Shell 中填写的是执行器运行环境内的路径。
6. V1 不做脚本沙箱，用户需要自行保证脚本安全。
7. V1 推荐执行器使用普通系统用户运行，不建议使用 root。

---

## 14. 异步任务执行模型

ChronoFlow V1 使用异步执行模型。

调度器调用执行器 `/run` 时，执行器只接收任务并立即返回 accepted，不等待脚本执行完成。

脚本执行完成后，执行器通过 callback 接口将执行结果回调给调度器。

### 14.1 整体流程

```text
1. 调度器触发任务
2. 调度器创建 running 执行日志
3. 调度器读取任务和 Glue 快照
4. 调度器调用执行器 /run
5. 执行器校验 X-Executor-Token
6. 执行器创建本地运行实例
7. 执行器立即返回 accepted
8. 执行器后台执行 Shell 脚本
9. 执行器收集执行结果和日志
10. 执行器将 callback 结果写入本地 pending 文件
11. 执行器回调调度器 callback 接口
12. 调度器校验 X-Callback-Token
13. 调度器写入日志正文文件
14. 调度器更新 job_logs 元数据
15. callback 成功后，执行器删除 pending 文件
```

### 14.2 手动执行流程

```text
1. 用户在任务列表点击手动运行
2. 调度器检查任务是否存在
3. 调度器检查任务是否已删除
4. 调度器检查任务是否有 Glue 脚本
5. 调度器检查绑定执行器是否在线
6. 调度器检查该任务当前是否已有实例正在执行
7. 如果已有实例正在执行，返回错误：任务正在执行中
8. 如果没有运行中实例，创建 running 执行日志
9. 调度器读取任务和 Glue 快照
10. 调度器调用执行器 /run
11. 执行器 accepted 后，手动执行请求返回成功
12. 用户进入日志详情查看执行状态
13. 执行器完成后 callback 更新日志
```

手动运行冲突规则：

1. 如果任务当前没有运行实例，允许手动运行。
2. 如果任务当前已有运行实例，前端手动运行按钮置灰。
3. 后端仍需做并发校验，避免绕过前端直接调用接口。
4. 后端发现任务正在执行时，返回错误：任务正在执行中。
5. 手动运行冲突不生成执行日志。

### 14.3 定时执行流程

```text
1. robfig/cron 根据 Cron 表达式触发任务
2. 调度器检查任务是否仍然存在
3. 调度器检查任务是否未删除
4. 调度器检查任务是否处于启动状态
5. 调度器检查任务是否有 Glue 脚本
6. 调度器检查绑定执行器是否在线
7. 调度器检查该任务当前是否已有实例正在执行
8. 如果已有实例正在执行，创建 skipped 执行日志
9. 如果没有运行中实例，创建 running 执行日志
10. 调度器下发任务到执行器 /run
11. 执行器 accepted 后后台执行
12. 执行器完成后 callback 更新日志
```

Cron 触发冲突规则：

1. 同一个任务同一时间只允许一个实例执行。
2. 如果上一次执行还未结束，下一次 Cron 触发到来，则跳过本次调度。
3. Cron 跳过时需要生成 `skipped` 执行日志。
4. `skipped` 日志不调用执行器。

### 14.4 /run accepted 语义

执行器 `/run` 接口返回 accepted 只代表执行器已经接收任务，不代表脚本执行成功。

脚本最终结果以执行器 callback 为准。

---

## 15. 执行日志

### 15.1 日志存储原则

ChronoFlow V1 不把完整日志正文保存到 MySQL。

存储规则：

1. MySQL 只保存执行日志元数据。
2. 完整日志正文保存到 `chronoflow-admin` 本地文件系统。
3. 日志详情由 `chronoflow-admin` 查询 MySQL 元数据，并读取本地日志文件后返回。
4. 删除任务不会立即删除历史日志。
5. 历史日志按全局日志保留策略清理。

### 15.2 日志目录结构

推荐日志目录：

```text
/app/data/logs/
└── 2026/
    └── 06/
        └── 10/
            └── job-1001/
                └── log-2001.log
```

MySQL 保存相对路径：

```text
logs/2026/06/10/job-1001/log-2001.log
```

### 15.3 日志内容

每执行一次任务，都生成一条执行日志元数据。

日志元数据需要保存：

| 字段 | 说明 |
| --- | --- |
| id | 日志 ID |
| job_id | 任务 ID |
| job_name | 执行时的任务名称快照 |
| executor_id | 执行器 ID |
| executor_name | 执行时的执行器名称快照 |
| executor_address | 执行时的执行器地址快照 |
| cron_expr | 执行时的 Cron 表达式快照 |
| timeout_seconds | 执行时的超时时间快照 |
| glue_snapshot | 执行时的 Shell 脚本快照 |
| trigger_type | 触发方式：manual / cron |
| status | 执行状态 |
| start_time | 开始时间 |
| end_time | 结束时间 |
| duration_ms | 执行耗时 |
| exit_code | Shell 退出码 |
| log_path | 日志文件相对路径 |
| log_size_bytes | 日志文件大小 |
| log_truncated | 日志是否被截断 |
| error_message | 错误信息 |
| created_at | 创建时间 |
| updated_at | 更新时间 |

### 15.4 日志正文

日志正文内容为 stdout + stderr 合并后的文本。

规则：

1. 执行器收集 Shell 脚本执行过程中的 stdout 和 stderr。
2. stdout 和 stderr 合并为一份日志正文。
3. 单次任务日志最大保存 5MB。
4. 如果日志超过 5MB，执行器截断日志内容。
5. 截断后，回调结果中 `log_truncated = true`。
6. 调度器写入截断后的日志文件。
7. 日志详情页需要提示“日志已截断”。

推荐截断策略：

```text
保留前 2.5MB + 后 2.5MB，中间插入截断提示。
```

说明：

1. Python 报错堆栈通常在日志尾部。
2. 保留前后内容比只保留前 5MB 更便于排查。

### 15.5 触发方式

字段：`trigger_type`

| 值 | 说明 |
| --- | --- |
| manual | 手动触发 |
| cron | 定时触发 |

### 15.6 执行状态

字段：`status`

| 状态 | 含义 |
| --- | --- |
| running | 正在执行 |
| killing | 用户已发起终止，等待执行器确认 |
| success | 执行成功 |
| failed | 执行失败或执行结果未知 |
| timeout | 执行超时 |
| skipped | 本次 Cron 调度被跳过 |
| killed | 用户主动终止成功 |

状态说明：

1. `success` 表示脚本退出码为 0。
2. `failed` 表示脚本退出码非 0、调用执行器失败、执行器离线、执行器失联、调度器重启后结果未知等。
3. `timeout` 表示超过任务超时时间后被系统终止。
4. `skipped` 只用于 Cron 触发冲突。
5. `killed` 表示用户主动终止成功。

### 15.7 日志筛选

执行日志页面支持筛选：

1. 按任务筛选。
2. 按执行器筛选。
3. 按触发方式筛选。
4. 按执行状态筛选。
5. 按时间范围筛选。

### 15.8 日志详情

日志详情页面显示：

1. 执行状态。
2. 触发方式。
3. 开始时间。
4. 结束时间。
5. 执行耗时。
6. 退出码。
7. 错误信息。
8. 任务名称快照。
9. 执行器快照。
10. Cron 快照。
11. Glue 快照。
12. 日志正文。
13. 日志是否被截断。

如果日志文件不存在，页面显示：

```text
日志文件不存在或已被清理。
```

### 15.9 日志清理

V1 默认保留最近 30 天执行日志。

规则：

1. 调度器启动一个定时清理任务，每天执行一次。
2. 默认清理时间为每天 03:00。
3. 清理对象为 `created_at` 早于保留天数的 `job_logs`。
4. 清理 `job_logs` 时，同时删除对应日志文件。
5. 如果日志文件已经不存在，清理任务忽略文件删除错误。
6. 日志保留天数可通过配置文件或环境变量修改。
7. 删除任务不会立即删除历史日志，历史日志仍按全局日志保留策略清理。

---

## 16. 执行器 callback 可靠性

### 16.1 pending 文件

执行器执行完成后，先将执行结果写入本地 pending callback 文件，再回调调度器。

推荐目录：

```text
/app/data/callbacks/pending/{log_id}.json
```

pending 文件内容包括：

```json
{
  "log_id": 2001,
  "job_id": 1001,
  "status": "success",
  "exit_code": 0,
  "log_content": "hello\n",
  "log_truncated": false,
  "start_time": "2026-06-10 10:00:00",
  "end_time": "2026-06-10 10:00:01",
  "duration_ms": 1000,
  "error_message": ""
}
```

说明：

1. pending 文件是执行器本地临时结果。
2. pending 文件不是最终日志存储。
3. 最终日志正文由调度器保存到 admin 本地日志目录。
4. callback 成功后，执行器删除 pending 文件。

### 16.2 callback 重试

规则：

1. callback 成功后，执行器删除 pending 文件。
2. callback 失败后，执行器保留 pending 文件。
3. 执行器后台任务定期扫描 pending 文件并重试 callback。
4. 默认重试间隔为 30 秒。
5. 执行器重启后，也会扫描 pending 文件继续回调。
6. pending 结果默认保留 7 天。
7. 超过 7 天仍未回调成功的 pending 文件，移动到 expired 目录或删除，并记录执行器本地日志。

### 16.3 callback 幂等

调度器 callback 接口必须幂等。

规则：

1. 同一个 `log_id` 可能收到多次 callback。
2. 只有 `running` 或 `killing` 状态允许被 callback 更新为最终状态。
3. 最终状态包括 `success`、`failed`、`timeout`、`killed`。
4. 如果日志已经是最终状态，重复 callback 直接忽略并返回成功。
5. 如果日志不存在，返回错误。

---

## 17. 超时控制

### 17.1 超时时间规则

任务超时时间规则：

1. 每个任务都有一个超时时间。
2. 默认超时时间为 10 分钟。
3. 默认值为 600 秒。
4. 创建任务时用户可以手动填写超时时间。
5. 如果用户不填，使用默认 600 秒。
6. 执行脚本超过超时时间后，执行器强制终止脚本进程组。
7. 本次执行日志状态标记为 `timeout`。
8. 错误信息记录：任务执行超时。

### 17.2 超时处理流程

```text
1. 执行器开始执行脚本
2. 执行器启动超时计时
3. 如果脚本在超时时间内结束，正常生成执行结果
4. 如果超过超时时间仍未结束，执行器终止脚本进程组
5. 执行器收集已产生的 stdout / stderr
6. 执行器生成 timeout 结果
7. 执行器写入 pending 文件
8. 执行器 callback 调度器
9. 调度器将执行日志更新为 timeout
```

---

## 18. 任务终止

### 18.1 终止能力

ChronoFlow V1 支持用户终止正在运行的任务。

入口：

1. 任务列表。
2. 执行日志详情。

只有 `status = running` 的执行日志可以终止。

### 18.2 终止流程

```text
1. 用户点击终止
2. 调度器校验日志 status = running
3. 调度器将日志状态更新为 killing
4. 调度器调用执行器 POST /kill
5. 执行器根据 log_id 找到本地运行进程组
6. 执行器终止进程组
7. 执行器收集已产生的 stdout / stderr
8. 执行器生成 killed 结果
9. 执行器写入 pending 文件
10. 执行器 callback 调度器
11. 调度器将日志状态更新为 killed
```

### 18.3 进程组终止规则

执行器启动脚本时，为每次任务创建独立 Linux 进程组。

任务超时或用户终止时：

1. 优先对整个进程组发送 `SIGTERM`。
2. 等待 grace period，默认 5 秒。
3. 如果进程组仍未退出，再发送 `SIGKILL`。
4. 收集已产生的 stdout / stderr。
5. 根据场景回调 `timeout` 或 `killed`。

限制：

1. V1 尽力终止进程组。
2. V1 不保证清理脚本主动脱离进程组、`nohup`、`setsid`、daemon 化产生的进程。
3. V1 推荐 Glue 脚本不要自行 daemonize 长驻进程。
4. 如果 Python 脚本需要清理资源，建议自行处理 `SIGTERM`。

### 18.4 killing 超时规则

用户发起终止后，日志状态从 `running` 变为 `killing`。

规则：

1. `killing` 状态最多等待 60 秒。
2. 60 秒内收到执行器 callback，则按 callback 更新为 `killed`、`failed` 或其他最终状态。
3. 超过 60 秒仍未收到 callback，调度器将日志标记为 failed。
4. 错误信息写：`终止任务超时，执行结果未知`。

---

## 19. 并发控制

### 19.1 同一任务并发策略

第一版采用同任务互斥策略。

规则：

```text
同一个任务同一时间只允许一个实例执行。
```

不同任务可以在同一个执行器上并行执行。

### 19.2 调度器侧控制

调度器负责主并发判断。

规则：

1. 手动运行前，检查该任务是否存在 `running` 或 `killing` 日志。
2. Cron 触发前，检查该任务是否存在 `running` 或 `killing` 日志。
3. 如果手动运行发现任务正在执行，返回错误，不生成日志。
4. 如果 Cron 触发发现任务正在执行，生成 `skipped` 日志。

### 19.3 执行器侧兜底

执行器也需要维护本地运行中的 `job_id` 集合。

规则：

1. 执行器收到 `/run` 后，检查同一 `job_id` 是否正在本地执行。
2. 如果同一 `job_id` 正在本地执行，拒绝本次 `/run` 请求。
3. 执行器侧控制是兜底，主要并发控制仍由调度器负责。

### 19.4 跳过日志

任务被 Cron 跳过时，需要生成执行日志。

日志内容：

| 字段 | 示例 |
| --- | --- |
| trigger_type | cron |
| status | skipped |
| error_message | 上一次任务仍在执行，本次调度跳过 |
| start_time | 当前触发时间 |
| end_time | 当前触发时间 |

---

## 20. 执行器离线处理

任务触发时，如果绑定执行器离线：

1. 不执行任务。
2. 不下发任务到执行器。
3. 生成一条执行日志。
4. 日志状态为 `failed`。
5. 错误信息写：执行器离线，任务无法执行。

第一版不支持：

1. 失败重试。
2. 自动切换执行器。
3. 故障转移。
4. 补偿执行。

---

## 21. 调度器重启恢复机制

### 21.1 Cron 恢复

调度器启动时需要恢复已启动任务。

规则：

```text
1. 调度器启动后，初始化 Cron 调度引擎
2. 从数据库查询所有未删除的任务
3. 只加载调度状态为 running 的任务
4. 读取任务的 Cron 表达式
5. 将任务重新注册到调度器内存中
6. 后续按 Cron 表达式继续触发
7. 停止状态任务不加载
8. 已删除任务不加载
```

### 21.2 running 日志恢复

调度器重启时，数据库中可能存在 `status = running` 或 `status = killing` 的执行日志。

规则：

1. 调度器启动后，先恢复 Cron 调度任务。
2. 调度器查询数据库中 `status = running` 或 `status = killing` 的执行日志。
3. 对这些执行日志启动 120 秒恢复宽限期。
4. 宽限期内，如果执行器 callback 返回结果，则按 callback 更新为最终状态。
5. 宽限期结束后，仍为 `running` 或 `killing` 的日志标记为 failed。
6. `end_time` 设置为当前时间。
7. `duration_ms` 根据 `start_time` 到当前时间计算。
8. `error_message` 写：`调度器重启，执行结果未知`。

### 21.3 Cron 校验

为了避免调度器启动失败：

1. 创建任务时必须校验 Cron 表达式。
2. 编辑任务时必须校验 Cron 表达式。
3. Cron 表达式非法，不允许保存。
4. 理论上调度器启动恢复时不应该加载到非法 Cron。

---

## 22. 安全认证

### 22.1 调度器调用执行器认证

第一版执行器使用 token 认证。

规则：

1. 每个执行器启动时配置 `executor_token`。
2. 调度器在 Web 页面添加执行器时，需要填写该执行器的 token。
3. 调度器加密保存执行器 token。
4. 调度器调用执行器接口时，在 Header 中携带 token。
5. 执行器收到请求后校验 token。
6. token 正确，允许执行。
7. token 错误或缺失，直接拒绝请求。

请求 Header：

```http
X-Executor-Token: xxxxxx
```

需要认证的执行器接口：

```text
GET /health
POST /run
POST /kill
```

### 22.2 执行器回调调度器认证

执行器回调调度器使用全局 `callback_token`。

规则：

1. `chronoflow-admin` 配置全局 `callback_token`。
2. 调度器调用执行器 `/run` 时，把 `callback_url` 和 `callback_token` 下发给执行器。
3. 执行器 callback 调度器时，在 Header 中携带 `X-Callback-Token`。
4. 调度器校验 token 正确后，才允许更新执行日志。
5. token 错误或缺失，调度器拒绝 callback。
6. V1 为了轻量，所有执行器共用一个 callback token。
7. 后续版本可扩展为每执行器独立 callback token 或签名机制。

请求 Header：

```http
X-Callback-Token: xxxxxx
```

### 22.3 Token 加密密钥

执行器 token 加密密钥支持配置文件和环境变量。

规则：

1. 加密算法使用 AES-GCM。
2. 加密密钥需要是 32 字节。
3. 支持在配置文件中设置。
4. 支持通过环境变量设置。
5. 环境变量优先覆盖配置文件。

环境变量示例：

```bash
CHRONOFLOW_TOKEN_ENCRYPT_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
CHRONOFLOW_CALLBACK_TOKEN=xxxxxxxx
```

### 22.4 安全边界

V1 安全边界：

1. 执行器端口建议只在内网开放。
2. 不建议将执行器直接暴露到公网。
3. 执行器建议使用普通系统用户运行，不建议使用 root。
4. 临时脚本文件权限建议为 `0600` 或 `0700`。
5. V1 不做 Shell 脚本沙箱。
6. 用户需要自行保证 Glue 脚本安全。

---

## 23. 执行器接口

### 23.1 健康检查接口

接口：

```http
GET /health
```

Header：

```http
X-Executor-Token: xxxxxx
```

返回示例：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "status": "online",
    "executor_name": "executor-01"
  }
}
```

### 23.2 执行脚本接口

接口：

```http
POST /run
```

Header：

```http
X-Executor-Token: xxxxxx
```

请求参数：

```json
{
  "job_id": 1001,
  "log_id": 2001,
  "script": "#!/bin/bash\necho hello",
  "timeout_seconds": 600,
  "callback_url": "http://10.0.0.10:8080/internal/job-runs/callback",
  "callback_token": "xxxxxx"
}
```

返回参数：

```json
{
  "code": 0,
  "msg": "accepted",
  "data": {
    "log_id": 2001,
    "status": "accepted"
  }
}
```

说明：

1. `accepted` 只表示执行器已接收任务。
2. 脚本最终执行结果通过 callback 返回。

### 23.3 终止任务接口

接口：

```http
POST /kill
```

Header：

```http
X-Executor-Token: xxxxxx
```

请求参数：

```json
{
  "job_id": 1001,
  "log_id": 2001
}
```

返回参数：

```json
{
  "code": 0,
  "msg": "killing",
  "data": {
    "log_id": 2001,
    "status": "killing"
  }
}
```

说明：

1. `/kill` 返回成功只表示执行器已接收终止请求。
2. 最终是否终止成功以 callback 为准。

---

## 24. 调度器内部接口

### 24.1 执行结果 callback 接口

接口：

```http
POST /internal/job-runs/callback
```

Header：

```http
X-Callback-Token: xxxxxx
```

请求参数：

```json
{
  "log_id": 2001,
  "job_id": 1001,
  "status": "success",
  "exit_code": 0,
  "log_content": "hello\n",
  "log_truncated": false,
  "start_time": "2026-06-10 10:00:00",
  "end_time": "2026-06-10 10:00:01",
  "duration_ms": 1000,
  "error_message": ""
}
```

返回参数：

```json
{
  "code": 0,
  "msg": "ok"
}
```

处理规则：

1. 校验 `X-Callback-Token`。
2. 查询 `job_logs`。
3. 如果日志状态为 `running` 或 `killing`，允许更新。
4. 将 `log_content` 写入 admin 本地日志文件。
5. 更新日志元数据，包括状态、时间、退出码、日志路径、日志大小、是否截断。
6. 如果日志已是最终状态，忽略重复 callback 并返回成功。

---

## 25. 后端接口范围

### 25.1 认证模块

功能：

1. 登录。
2. 退出登录。
3. 获取当前管理员信息。
4. 修改密码。

### 25.2 执行器模块

功能：

1. 新增执行器。
2. 编辑执行器。
3. 删除执行器。
4. 查询执行器列表。
5. 检测执行器在线状态。

### 25.3 任务模块

功能：

1. 根据执行器查询任务列表。
2. 新增任务。
3. 编辑任务基础信息。
4. 删除任务。
5. 启动任务。
6. 停止任务。
7. 编辑 Glue 脚本。
8. 查询 Glue 脚本。
9. 手动运行任务。
10. 终止正在运行的任务。

### 25.4 执行日志模块

功能：

1. 查询执行日志列表。
2. 查询日志详情。
3. 按任务筛选。
4. 按执行器筛选。
5. 按执行状态筛选。
6. 按触发方式筛选。
7. 按时间范围筛选。

### 25.5 执行器内部接口

功能：

1. 健康检查。
2. 执行 Shell 脚本。
3. 终止任务。

### 25.6 调度器内部接口

功能：

1. 接收执行器执行结果 callback。

---

## 26. 配置项

### 26.1 chronoflow-admin 配置

建议配置：

```yaml
server:
  http_addr: "0.0.0.0:8080"
  public_base_url: "http://10.0.0.10:8080"

database:
  dsn: "root:password@tcp(127.0.0.1:3306)/chronoflow"

scheduler:
  timezone: "Asia/Shanghai"

executor:
  health_check_interval_seconds: 10
  health_check_fail_threshold: 3

security:
  token_encrypt_key: "32-byte-secret-key"
  callback_token: "callback-token"

logs:
  data_dir: "/app/data"
  max_log_bytes: 5242880
  retention_days: 30
  cleanup_cron: "0 0 3 * * *"

recovery:
  startup_running_grace_seconds: 120
  killing_timeout_seconds: 60
```

环境变量优先覆盖配置文件。

建议环境变量：

```bash
CHRONOFLOW_TOKEN_ENCRYPT_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
CHRONOFLOW_CALLBACK_TOKEN=xxxxxxxx
```

### 26.2 chronoflow-exec 配置

建议配置：

```yaml
server:
  http_addr: "0.0.0.0:9999"

executor:
  name: "executor-01"
  token: "executor-token"
  data_dir: "/app/data"
  shell_path: "/bin/bash"
  kill_grace_seconds: 5

callback:
  retry_interval_seconds: 30
  pending_retention_days: 7
```

---

## 27. 数据表设计初稿

### 27.1 users

用户表。

| 字段 | 类型建议 | 说明 |
| --- | --- | --- |
| id | bigint | 主键 |
| username | varchar(64) | 用户名 |
| password_hash | varchar(255) | 密码 hash |
| role | varchar(32) | 角色 |
| status | varchar(32) | enabled / disabled |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |
| deleted_at | datetime | 删除时间 |

建议索引：

1. `uk_username(username)`。

### 27.2 executors

执行器表。

| 字段 | 类型建议 | 说明 |
| --- | --- | --- |
| id | bigint | 主键 |
| name | varchar(100) | 执行器名称 |
| address | varchar(255) | 执行器地址 |
| token_ciphertext | varchar(1000) | 加密后的执行器 token |
| description | varchar(500) | 描述 |
| status | varchar(32) | online / offline |
| heartbeat_fail_count | int | 连续健康检查失败次数 |
| last_heartbeat_time | datetime | 最近一次检测成功时间 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |
| deleted_at | datetime | 删除时间 |

建议索引：

1. `idx_status(status)`。
2. `idx_deleted_at(deleted_at)`。

### 27.3 jobs

任务表。

| 字段 | 类型建议 | 说明 |
| --- | --- | --- |
| id | bigint | 主键 |
| executor_id | bigint | 执行器 ID |
| name | varchar(100) | 任务名称 |
| cron_expr | varchar(100) | Cron 表达式 |
| timeout_seconds | int | 超时时间 |
| schedule_status | varchar(32) | stopped / running |
| description | varchar(500) | 描述 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |
| deleted_at | datetime | 删除时间 |

建议索引：

1. `idx_executor_id(executor_id)`。
2. `idx_schedule_status(schedule_status)`。
3. `idx_deleted_at(deleted_at)`。

### 27.4 job_glues

Glue 脚本表。

| 字段 | 类型建议 | 说明 |
| --- | --- | --- |
| id | bigint | 主键 |
| job_id | bigint | 任务 ID |
| content | text | Shell 脚本内容 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

建议索引：

1. `uk_job_id(job_id)`。

### 27.5 job_logs

执行日志元数据表。

| 字段 | 类型建议 | 说明 |
| --- | --- | --- |
| id | bigint | 主键 |
| job_id | bigint | 任务 ID |
| job_name | varchar(100) | 任务名称快照 |
| executor_id | bigint | 执行器 ID |
| executor_name | varchar(100) | 执行器名称快照 |
| executor_address | varchar(255) | 执行器地址快照 |
| cron_expr | varchar(100) | Cron 表达式快照 |
| timeout_seconds | int | 超时时间快照 |
| glue_snapshot | mediumtext | Glue 内容快照 |
| trigger_type | varchar(32) | manual / cron |
| status | varchar(32) | running / killing / success / failed / timeout / skipped / killed |
| start_time | datetime | 开始时间 |
| end_time | datetime | 结束时间 |
| duration_ms | bigint | 执行耗时 |
| exit_code | int | Shell 退出码 |
| log_path | varchar(500) | 日志文件相对路径 |
| log_size_bytes | bigint | 日志文件大小 |
| log_truncated | tinyint | 是否截断 |
| error_message | text | 错误信息 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

说明：

1. `job_logs` 不保存完整日志正文。
2. 完整日志正文保存到 `chronoflow-admin` 本地文件系统。

建议索引：

1. `idx_job_created(job_id, created_at)`。
2. `idx_executor_created(executor_id, created_at)`。
3. `idx_status_created(status, created_at)`。
4. `idx_trigger_created(trigger_type, created_at)`。
5. `idx_created_at(created_at)`。

---

## 28. V1 暂不支持功能

ChronoFlow V1 暂不支持以下功能：

1. 执行器主动注册。
2. 执行器分组。
3. 调度器自动选择执行器。
4. 失败重试。
5. 失败后切换执行器。
6. 实时日志推送。
7. WebSocket 滚动日志。
8. 多用户角色权限。
9. 任务依赖编排。
10. DAG 工作流。
11. 分片任务。
12. 广播任务。
13. Python / Go / Java 原生任务类型。
14. 文件上传脚本。
15. Git 脚本同步。
16. 多调度器高可用。
17. Kubernetes 多 Pod 调度器并发。
18. 执行器主动心跳注册。
19. 任务失败告警。
20. 邮件、飞书、企业微信通知。
21. Windows 执行器。
22. Shell 脚本沙箱。
23. Cron 错过补偿执行。
24. 每任务独立时区。

---

## 29. V1 验收标准

### 29.1 执行器管理验收

1. 可以新增执行器。
2. 可以编辑执行器。
3. 可以删除没有任务的执行器。
4. 执行器下面有未删除任务时，不允许删除。
5. 可以看到执行器在线或离线状态。
6. 删除后的执行器不再显示。
7. 删除后的执行器不再做在线检测。
8. 调度器每 10 秒健康检查一次执行器。
9. 连续 3 次健康检查失败后，执行器变为 offline。
10. 执行器恢复健康检查成功后，状态变为 online。

### 29.2 任务管理验收

1. 进入任务管理页面后，必须先选择执行器。
2. 选择执行器后，只显示该执行器下的任务。
3. 可以新增任务。
4. 新增任务时可以填写任务名称、Cron 表达式、超时时间。
5. Cron 表达式非法时，不允许保存。
6. 新建任务默认停止状态。
7. 没有 Glue 脚本时，不允许启动。
8. 没有 Glue 脚本时，不允许手动运行。
9. 有 Glue 脚本后，可以手动运行。
10. 点击启动后，任务参与 Cron 调度。
11. 点击停止后，任务不参与 Cron 调度。
12. 停止状态下仍然可以手动运行。
13. 启动状态下也可以手动运行。
14. 删除任务后，任务列表不再显示。
15. 删除任务后，不再参与调度。
16. 任务运行中允许编辑配置，当前运行不受影响。
17. 已启动任务修改 Cron 后，新 Cron 立即重新注册。

### 29.3 Glue 验收

1. 可以编辑一整段 Shell 脚本。
2. Glue 内容保存到数据库。
3. 再次打开任务 Glue 页面时，可以看到上次保存的脚本。
4. 手动执行和定时执行时，使用数据库中的 Glue 内容。
5. 执行日志详情可以看到执行开始时的 Glue 快照。
6. Glue Shell 可以调用执行器环境中的 Python 脚本。

### 29.4 异步执行验收

1. 手动执行任务时，调度器创建 running 日志。
2. 调度器调用执行器 `/run`。
3. 执行器 `/run` 返回 accepted。
4. 脚本执行完成前，日志状态保持 running。
5. 执行器执行完成后 callback 调度器。
6. 调度器收到 callback 后更新执行日志最终状态。
7. callback 重复发送不会重复更新最终状态日志。

### 29.5 执行日志验收

1. 每次手动执行都会生成执行日志。
2. 每次定时执行都会生成执行日志。
3. 执行成功时，日志状态为 success。
4. 执行失败时，日志状态为 failed。
5. 执行超时时，日志状态为 timeout。
6. 用户终止成功时，日志状态为 killed。
7. 用户发起终止后，日志状态先变为 killing。
8. 上一次未执行完成，本次 Cron 调度跳过时，日志状态为 skipped。
9. 日志详情可以看到 stdout + stderr 合并内容。
10. 日志详情可以看到 exit_code。
11. 日志详情可以看到任务名称、执行器、Cron、Glue 快照。
12. 删除任务后，历史日志仍然保留到清理周期。
13. MySQL 只保存日志元数据，不保存完整日志正文。
14. 日志正文保存到 admin 本地文件。
15. 日志超过 5MB 时会截断并提示。
16. 日志默认保留 30 天，清理元数据时同步删除日志文件。

### 29.6 并发控制验收

1. 不同任务可以在同一执行器上并行执行。
2. 同一任务不允许同时执行多个实例。
3. 手动运行遇到同任务正在执行时，后端返回“任务正在执行中”。
4. 手动运行冲突不生成执行日志。
5. Cron 触发遇到同任务正在执行时，生成 skipped 日志。

### 29.7 任务终止验收

1. running 日志可以点击终止。
2. 点击终止后，日志状态变为 killing。
3. 调度器调用执行器 `/kill`。
4. 执行器终止任务进程组。
5. 执行器 callback 后，日志状态变为 killed。
6. killing 超过 60 秒未收到 callback，日志状态变为 failed。
7. 超时和终止都应尽量终止整个 Linux 进程组。

### 29.8 调度恢复验收

1. 调度器重启后，会重新加载所有未删除且已启动的任务。
2. 停止状态任务不会被加载到调度器。
3. 已删除任务不会被加载到调度器。
4. 重启后，已启动任务可以继续按 Cron 表达式触发。
5. 调度器重启后，running 和 killing 日志有 120 秒恢复宽限期。
6. 宽限期内收到 callback，则正常更新日志。
7. 宽限期后仍未完成，则日志标记为 failed。

### 29.9 callback 可靠性验收

1. 执行器执行完成后，先写入 pending 文件。
2. callback 成功后，执行器删除 pending 文件。
3. callback 失败后，执行器保留 pending 文件。
4. 执行器每 30 秒重试 pending callback。
5. 执行器重启后会继续扫描 pending 文件并重试。
6. pending 文件默认保留 7 天。

### 29.10 安全认证验收

1. 调度器调用执行器 `/health` 时必须携带 token。
2. 调度器调用执行器 `/run` 时必须携带 token。
3. 调度器调用执行器 `/kill` 时必须携带 token。
4. token 正确时，执行器允许请求。
5. token 错误或缺失时，执行器拒绝请求。
6. 执行器 callback 调度器时必须携带 `X-Callback-Token`。
7. callback token 错误或缺失时，调度器拒绝 callback。
8. 未登录用户不能访问调度中心页面和后端管理接口。
9. 执行器 token 加密保存到数据库。
10. 前端不展示完整执行器 token。

---

## 30. 后续版本规划

后续版本可以扩展：

1. 管理员分配账号。
2. 多用户管理。
3. 角色权限管理。
4. 执行器主动注册。
5. 执行器分组。
6. 失败重试。
7. 失败转移执行器。
8. 实时日志。
9. WebSocket 日志推送。
10. 任务失败通知。
11. 飞书通知。
12. 邮件通知。
13. 任务依赖编排。
14. DAG 工作流。
15. 分片任务。
16. 广播任务。
17. Git 脚本同步。
18. 文件上传脚本。
19. 多调度器高可用。
20. Kubernetes 部署支持。
21. 对象存储保存日志正文。
22. 每执行器独立 callback token。
23. 调度器主动查询执行器运行结果。
24. 每任务独立时区。
