<script setup lang="ts">
import {
  ApiOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CodeOutlined,
  DatabaseOutlined,
  FileSearchOutlined,
  InfoCircleOutlined,
  ToolOutlined,
} from '@ant-design/icons-vue'
import PageHeaderBar from '@/components/PageHeaderBar.vue'

const quickStartSteps = [
  '启动 admin、exec 和 UI。',
  '在执行器页面新增执行器，地址填写 http://127.0.0.1:10004，Token 填写执行器配置里的 token。',
  '在任务页面新增任务，选择执行器，填写 6 位 Cron 表达式和超时时间。',
  '打开任务的 Glue 编辑器，保存 Shell 脚本。',
  '点击运行，在执行日志页面查看结果和日志正文。',
]

const operationRules = [
  '同一个任务运行中时不能再次手动运行，不同任务可以并行。',
  '停止调度只影响后续 Cron 触发，不会终止当前运行实例。',
  '终止任务会请求执行器 kill 进程组，状态会从 running 进入 killing，再进入 killed 或 failed。',
  '编辑任务配置不会影响当前运行实例，新配置下次执行生效。',
]

const troubleshooting = [
  {
    title: '执行器离线',
    detail: '确认执行器进程、地址、Token 和网络连通性。Admin 默认每 10 秒检查一次，连续 3 次失败后标记 offline。',
  },
  {
    title: '任务一直运行中',
    detail: '检查执行器是否能 callback Admin。Admin 启动恢复会在等待窗口后把遗留 active 日志置为 failed。',
  },
  {
    title: '看不到日志正文',
    detail: '日志正文保存在 Admin 的 logs.data_dir 文件目录中，MySQL 只保存元数据和 log_path。',
  },
  {
    title: 'Shell 找不到 Python 脚本',
    detail: '如果执行器跑在 Docker 里，需要把宿主机脚本目录通过 volume 挂载到容器内，并在 Glue 中使用容器内路径。',
  },
]
</script>

<template>
  <div class="page-body">
    <PageHeaderBar title="使用说明" description="ChronoFlow V1 的日常操作、状态规则和联调要点。" />

    <a-row :gutter="[16, 16]">
      <a-col :xs="24" :xl="14">
        <section class="help-section">
          <div class="section-title">
            <CheckCircleOutlined />
            <h2>快速开始</h2>
          </div>
          <a-steps direction="vertical" size="small" :current="-1">
            <a-step v-for="item in quickStartSteps" :key="item" :title="item" />
          </a-steps>
        </section>
      </a-col>

      <a-col :xs="24" :xl="10">
        <section class="help-section">
          <div class="section-title">
            <InfoCircleOutlined />
            <h2>默认信息</h2>
          </div>
          <a-descriptions bordered size="small" :column="1">
            <a-descriptions-item label="Admin HTTP">10003</a-descriptions-item>
            <a-descriptions-item label="Exec HTTP">10004</a-descriptions-item>
            <a-descriptions-item label="默认账号">admin / admin123</a-descriptions-item>
            <a-descriptions-item label="默认执行器 Token">change-me</a-descriptions-item>
            <a-descriptions-item label="默认时区">Asia/Shanghai</a-descriptions-item>
          </a-descriptions>
        </section>
      </a-col>
    </a-row>

    <a-row :gutter="[16, 16]">
      <a-col :xs="24" :lg="12">
        <section class="help-section">
          <div class="section-title">
            <ClockCircleOutlined />
            <h2>任务运行规则</h2>
          </div>
          <a-list size="small" :data-source="operationRules">
            <template #renderItem="{ item }">
              <a-list-item>{{ item }}</a-list-item>
            </template>
          </a-list>
        </section>
      </a-col>

      <a-col :xs="24" :lg="12">
        <section class="help-section">
          <div class="section-title">
            <FileSearchOutlined />
            <h2>日志规则</h2>
          </div>
          <a-list size="small">
            <a-list-item>MySQL 只保存日志元数据，完整日志正文保存在 Admin 文件目录。</a-list-item>
            <a-list-item>单次日志默认最多保存 5MB，超出后执行器会做截断。</a-list-item>
            <a-list-item>日志默认保留 30 天，pending callback 默认保留 7 天。</a-list-item>
            <a-list-item>日志详情包含执行快照、Glue 快照、退出码、耗时和错误信息。</a-list-item>
          </a-list>
        </section>
      </a-col>
    </a-row>

    <section class="help-section">
      <div class="section-title">
        <CodeOutlined />
        <h2>Glue Shell 示例</h2>
      </div>
      <pre class="code-box mono">echo chronoflow-demo-start
pwd
python3 /scripts/report.py
echo chronoflow-demo-done</pre>
    </section>

    <a-row :gutter="[16, 16]">
      <a-col :xs="24" :lg="12">
        <section class="help-section">
          <div class="section-title">
            <DatabaseOutlined />
            <h2>服务边界</h2>
          </div>
          <a-list size="small">
            <a-list-item>Admin 连接 MySQL，保存任务、执行器和日志元数据。</a-list-item>
            <a-list-item>Exec 不连接数据库，只执行脚本、采集日志、保存 pending callback。</a-list-item>
            <a-list-item>UI 只调用 Admin API，不直接访问 Exec。</a-list-item>
          </a-list>
        </section>
      </a-col>

      <a-col :xs="24" :lg="12">
        <section class="help-section">
          <div class="section-title">
            <ApiOutlined />
            <h2>鉴权</h2>
          </div>
          <a-list size="small">
            <a-list-item>UI 登录后使用 JWT 调用 /v1/admin/*。</a-list-item>
            <a-list-item>Admin 调用 Exec 时携带 X-Executor-Token。</a-list-item>
            <a-list-item>Exec 回调 Admin 时携带 X-Callback-Token。</a-list-item>
          </a-list>
        </section>
      </a-col>
    </a-row>

    <section class="help-section">
      <div class="section-title">
        <ToolOutlined />
        <h2>常见问题</h2>
      </div>
      <a-collapse>
        <a-collapse-panel v-for="item in troubleshooting" :key="item.title" :header="item.title">
          {{ item.detail }}
        </a-collapse-panel>
      </a-collapse>
    </section>
  </div>
</template>

<style scoped>
.help-section {
  padding: 16px;
  background: #fff;
  border: 1px solid #d9e2f2;
  border-radius: 6px;
}

.section-title {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 12px;
  color: #1e40af;
}

.section-title h2 {
  margin: 0;
  color: #172033;
  font-size: 16px;
  font-weight: 650;
  line-height: 1.4;
}

.code-box {
  min-height: 128px;
  margin: 0;
}
</style>
