# ChronoFlow 技术方案文档任务计划

## 目标

为 ChronoFlow 三个子项目分别编写中文技术方案文档：

1. `chronoFlow-admin/技术方案.md`
2. `chronoFlow-exec/技术方案.md`
3. `chronoFlow-ui/技术方案.md`

文档定位为给 AI/开发者照着实现的详细开发方案，而不是泛泛架构介绍。

## 关键约束

1. 以 `prd-v1.md` 为需求来源。
2. 贴合现有项目目录和模板写法。
3. 后端模板自带示例接口只作为写法参考，不作为 ChronoFlow 业务代码。
4. `chronoFlow-exec` 不连接 MySQL，即使模板默认包含数据库连接，也需要在方案中明确移除或禁用。
5. 执行器只通过 HTTP 与 admin 通信，本地持久化只使用 pending callback 文件。

## 阶段

| 阶段 | 状态 |
| --- | --- |
| 梳理 PRD 和目录结构 | complete |
| 阅读 admin/exec 模板规则 | complete |
| 阅读 ui 项目结构 | complete |
| 编写 admin 技术方案 | complete |
| 编写 exec 技术方案 | complete |
| 编写 ui 技术方案 | complete |
| 校验文档一致性 | complete |

## 错误记录

暂无。
