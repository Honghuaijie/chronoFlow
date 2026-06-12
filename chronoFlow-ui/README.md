# ChronoFlow UI

调度中心前端，面向内网单团队使用。技术栈为 Vue3、TypeScript、Ant Design Vue、Pinia、Vue Router 和 Axios。

## 功能

- 登录。
- 执行器列表、新增、编辑、删除。
- 任务列表、新增、编辑、删除、启动、停止、手动运行、终止。
- Glue Shell 编辑。
- 执行日志列表、筛选、详情、Glue 快照和日志正文查看。
- 运行中日志自动轮询。

## 本地启动

```bash
npm install
npm run dev
```

默认 API 代理到：

```text
http://127.0.0.1:10003
```

如需覆盖：

```bash
VITE_API_PROXY_TARGET=http://127.0.0.1:10003 npm run dev
```

## 构建

```bash
npm run build
```

## 目录约定

```text
src/
├── api/          # 请求封装，只处理 HTTP 和响应解包
├── stores/       # Pinia 状态、loading、分页和请求编排
├── views/        # 页面容器
├── components/   # 可复用组件
├── types/        # TypeScript 类型
├── utils/        # 工具函数
├── router/       # 路由
└── layouts/      # 后台布局
```

固定调用链：

```text
view -> store -> api
```

页面不要直接请求接口，不要直接处理原始 HTTP Response。

## 默认账号

账号由 Admin 后端配置决定。默认：

```text
admin / admin123
```

## 开发注意事项

- UI 是运维控制台，不是营销站点。
- 优先使用表格、筛选、状态标签和清晰的危险操作确认。
- 同任务运行中时，任务页手动运行按钮需要置灰。
- 终止任务应从 `running` 进入 `killing`，最终展示 `killed` 或 `failed`。
- 后端 protobuf JSON 的 int64 可能以字符串返回，前端 ID 统一按字符串处理。
