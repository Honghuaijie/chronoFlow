<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  BarChartOutlined,
  ClockCircleOutlined,
  DatabaseOutlined,
  FileDoneOutlined,
  ReloadOutlined,
} from '@ant-design/icons-vue'
import PageHeaderBar from '@/components/PageHeaderBar.vue'
import * as executorApi from '@/api/executors'
import * as jobApi from '@/api/jobs'
import * as jobLogApi from '@/api/jobLogs'

interface ReportStats {
  jobTotal: number
  runTotal: number
  executorTotal: number
  recentSuccess: number
  recentFailed: number
}

const loading = ref(false)
const loadedAt = ref('')
const stats = ref<ReportStats>({
  jobTotal: 0,
  runTotal: 0,
  executorTotal: 0,
  recentSuccess: 0,
  recentFailed: 0,
})

const pieRadius = 70
const pieCircumference = 2 * Math.PI * pieRadius

const cards = computed(() => [
  {
    title: '任务数量',
    value: stats.value.jobTotal,
    description: '当前已创建的定时任务总数',
    icon: ClockCircleOutlined,
  },
  {
    title: '调度次数',
    value: stats.value.runTotal,
    description: '所有手动和 Cron 触发产生的执行日志数',
    icon: FileDoneOutlined,
  },
  {
    title: '执行器数量',
    value: stats.value.executorTotal,
    description: '当前登记的执行器节点总数',
    icon: DatabaseOutlined,
  },
])

const recentFinishedTotal = computed(() => stats.value.recentSuccess + stats.value.recentFailed)
const successPercent = computed(() => {
  if (!recentFinishedTotal.value) {
    return 0
  }
  return Math.round((stats.value.recentSuccess / recentFinishedTotal.value) * 100)
})
const failedPercent = computed(() => (recentFinishedTotal.value ? 100 - successPercent.value : 0))
const successStroke = computed(() => (recentFinishedTotal.value ? pieCircumference * (stats.value.recentSuccess / recentFinishedTotal.value) : 0))
const failedStroke = computed(() => (recentFinishedTotal.value ? pieCircumference - successStroke.value : 0))

onMounted(() => {
  void refresh()
})

async function refresh() {
  loading.value = true
  try {
    const [jobs, logs, executors] = await Promise.all([
      jobApi.listJobs(),
      jobLogApi.listJobLogs({ page: 1, pageSize: 1 }),
      executorApi.listExecutors(),
    ])
    const recentLogs = await jobLogApi.listJobLogs({ page: 1, pageSize: 100 })
    const recentSuccess = recentLogs.items.filter((item) => item.status === 'success').length
    const recentFailed = recentLogs.items.filter((item) => item.status === 'failed').length
    stats.value = {
      jobTotal: jobs.total,
      runTotal: logs.total,
      executorTotal: executors.total,
      recentSuccess,
      recentFailed,
    }
    loadedAt.value = new Date().toLocaleString()
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page-body">
    <PageHeaderBar title="运行报表" description="查看调度平台当前的任务、执行和执行器概览。">
      <a-button type="primary" :loading="loading" @click="refresh">
        <template #icon>
          <ReloadOutlined />
        </template>
        刷新
      </a-button>
    </PageHeaderBar>

    <a-spin :spinning="loading">
      <div class="report-grid">
        <a-card v-for="card in cards" :key="card.title" class="report-card" :bordered="false">
          <div class="report-card-content">
            <div class="report-icon">
              <component :is="card.icon" />
            </div>
            <div class="report-main">
              <div class="report-title">{{ card.title }}</div>
              <div class="report-value">{{ card.value }}</div>
              <div class="report-description">{{ card.description }}</div>
            </div>
          </div>
        </a-card>
      </div>

      <div class="chart-panel">
        <div class="chart-section">
          <div class="chart-header">
            <div>
              <h2>最近执行结果</h2>
              <p>统计最近 100 条执行日志中的成功与失败次数。</p>
            </div>
            <div class="chart-total">
              <span>完成记录</span>
              <strong>{{ recentFinishedTotal }}</strong>
            </div>
          </div>
          <div class="pie-layout">
            <div class="pie-wrap">
              <svg class="pie-chart" viewBox="0 0 180 180" role="img" aria-label="最近执行成功失败比例图">
                <circle class="pie-bg" cx="90" cy="90" :r="pieRadius" />
                <template v-if="recentFinishedTotal">
                  <circle
                    class="pie-segment pie-success"
                    cx="90"
                    cy="90"
                    :r="pieRadius"
                    :stroke-dasharray="`${successStroke} ${pieCircumference - successStroke}`"
                  />
                  <circle
                    v-if="stats.recentFailed > 0"
                    class="pie-segment pie-failed"
                    cx="90"
                    cy="90"
                    :r="pieRadius"
                    :stroke-dasharray="`${failedStroke} ${pieCircumference - failedStroke}`"
                    :stroke-dashoffset="-successStroke"
                  />
                </template>
                <text class="pie-percent" x="90" y="84" text-anchor="middle">{{ successPercent }}%</text>
                <text class="pie-label" x="90" y="108" text-anchor="middle">成功率</text>
              </svg>
            </div>
            <div class="pie-legend">
              <div class="legend-row">
                <span class="legend-dot success-dot"></span>
                <span class="legend-label">成功</span>
                <strong>{{ stats.recentSuccess }}</strong>
                <span class="legend-percent">{{ successPercent }}%</span>
              </div>
              <div class="legend-row">
                <span class="legend-dot failed-dot"></span>
                <span class="legend-label">失败</span>
                <strong>{{ stats.recentFailed }}</strong>
                <span class="legend-percent">{{ failedPercent }}%</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="report-panel">
        <div class="report-panel-header">
          <BarChartOutlined />
          <span>统计口径</span>
        </div>
        <a-descriptions :column="1" bordered size="middle">
          <a-descriptions-item label="任务数量">统计当前任务列表中的全部任务。</a-descriptions-item>
          <a-descriptions-item label="调度次数">统计执行日志总数，包含手动运行和 Cron 定时触发。</a-descriptions-item>
          <a-descriptions-item label="执行器数量">统计当前执行器列表中的全部执行器。</a-descriptions-item>
          <a-descriptions-item label="成功比例图">统计最近 100 条执行日志中状态为 success 和 failed 的记录。</a-descriptions-item>
          <a-descriptions-item label="最近刷新">{{ loadedAt || '-' }}</a-descriptions-item>
        </a-descriptions>
      </div>
    </a-spin>
  </div>
</template>

<style scoped>
.report-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.report-card {
  border: 1px solid #d9e2f2;
  border-radius: 6px;
}

.report-card-content {
  display: flex;
  gap: 14px;
  align-items: flex-start;
}

.report-icon {
  display: grid;
  flex: 0 0 auto;
  width: 40px;
  height: 40px;
  color: #1e40af;
  font-size: 20px;
  background: #eaf1ff;
  border-radius: 6px;
  place-items: center;
}

.report-main {
  min-width: 0;
}

.report-title {
  color: #475569;
  font-weight: 600;
}

.report-value {
  margin-top: 6px;
  color: #172033;
  font-size: 32px;
  font-weight: 700;
  line-height: 1.15;
}

.report-description {
  margin-top: 8px;
  color: #64748b;
  line-height: 1.5;
}

.report-panel {
  padding: 16px;
  background: #fff;
  border: 1px solid #d9e2f2;
  border-radius: 6px;
}

.chart-panel {
  padding: 16px;
  background: #fff;
  border: 1px solid #d9e2f2;
  border-radius: 6px;
}

.chart-section {
  min-height: 300px;
}

.chart-header {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 18px;
}

.chart-header h2 {
  margin: 0;
  color: #172033;
  font-size: 18px;
  line-height: 1.4;
}

.chart-header p {
  margin: 6px 0 0;
  color: #64748b;
}

.chart-total {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  color: #64748b;
}

.chart-total strong {
  margin-top: 2px;
  color: #172033;
  font-size: 26px;
  line-height: 1.2;
}

.pie-layout {
  display: grid;
  grid-template-columns: minmax(220px, 360px) minmax(180px, 1fr);
  gap: 32px;
  align-items: center;
}

.pie-wrap {
  display: grid;
  place-items: center;
}

.pie-chart {
  width: min(300px, 100%);
  height: auto;
}

.pie-bg {
  fill: none;
  stroke: #e5ecf6;
  stroke-width: 26;
}

.pie-segment {
  fill: none;
  stroke-width: 26;
  transform: rotate(-90deg);
  transform-origin: 90px 90px;
  transition:
    stroke-dasharray 0.2s ease,
    stroke-dashoffset 0.2s ease;
}

.pie-success {
  stroke: #0ea66a;
}

.pie-failed {
  stroke: #d93f3f;
}

.pie-percent {
  fill: #172033;
  font-size: 28px;
  font-weight: 700;
}

.pie-label {
  fill: #64748b;
  font-size: 13px;
}

.pie-legend {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.legend-row {
  display: grid;
  grid-template-columns: 12px minmax(48px, 1fr) auto auto;
  gap: 10px;
  align-items: center;
  min-height: 40px;
  padding: 10px 12px;
  background: #f8fafc;
  border: 1px solid #e5ecf6;
  border-radius: 6px;
}

.legend-dot {
  width: 10px;
  height: 10px;
  border-radius: 999px;
}

.success-dot {
  background: #0ea66a;
}

.failed-dot {
  background: #d93f3f;
}

.legend-label,
.legend-percent {
  color: #64748b;
}

.report-panel-header {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 14px;
  color: #172033;
  font-weight: 700;
}

@media (max-width: 960px) {
  .report-grid {
    grid-template-columns: 1fr;
  }

  .pie-layout {
    grid-template-columns: 1fr;
  }

  .chart-header {
    flex-direction: column;
  }

  .chart-total {
    align-items: flex-start;
  }
}
</style>
