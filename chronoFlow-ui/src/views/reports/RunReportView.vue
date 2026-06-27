<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  BarChartOutlined,
  ClockCircleOutlined,
  DatabaseOutlined,
  FileDoneOutlined,
  ReloadOutlined,
} from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import PageHeaderBar from '@/components/PageHeaderBar.vue'
import * as executorApi from '@/api/executors'
import * as jobApi from '@/api/jobs'
import * as jobLogApi from '@/api/jobLogs'
import type { JobLogInfo } from '@/types/jobLog'

interface ReportStats {
  jobTotal: number
  runTotal: number
  executorTotal: number
  recentSuccess: number
  recentFailed: number
}

interface TrendPoint {
  date: string
  label: string
  success: number
  failed: number
  running: number
}

const loading = ref(false)
const loadedAt = ref('')
const recentLogs = ref<JobLogInfo[]>([])
const hoveredTrendIndex = ref<number | null>(null)
const stats = ref<ReportStats>({
  jobTotal: 0,
  runTotal: 0,
  executorTotal: 0,
  recentSuccess: 0,
  recentFailed: 0,
})

const pieRadius = 70
const pieCircumference = 2 * Math.PI * pieRadius
const chartWidth = 760
const chartHeight = 300
const chartPadding = {
  top: 26,
  right: 20,
  bottom: 42,
  left: 48,
}

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
const trendPoints = computed<TrendPoint[]>(() => {
  const days = Array.from({ length: 7 }, (_, index) => {
    const date = dayjs().subtract(6 - index, 'day')
    return {
      date: date.format('YYYY-MM-DD'),
      label: date.format('MM-DD'),
      success: 0,
      failed: 0,
      running: 0,
    }
  })
  const byDate = new Map(days.map((item) => [item.date, item]))
  recentLogs.value.forEach((log) => {
    const date = dayjs(log.startTime || log.createdAt).format('YYYY-MM-DD')
    const point = byDate.get(date)
    if (!point) {
      return
    }
    if (log.status === 'success') {
      point.success += 1
    } else if (log.status === 'failed') {
      point.failed += 1
    } else if (log.status === 'running' || log.status === 'killing') {
      point.running += 1
    }
  })
  return days
})
const trendMax = computed(() => {
  const maxValue = Math.max(...trendPoints.value.flatMap((item) => [item.success, item.failed, item.running]), 0)
  return Math.max(maxValue, 1)
})
const yTicks = computed(() => {
  const max = trendMax.value
  return [max, Math.round(max * 0.75), Math.round(max * 0.5), Math.round(max * 0.25), 0]
})
const plotWidth = computed(() => chartWidth - chartPadding.left - chartPadding.right)
const plotHeight = computed(() => chartHeight - chartPadding.top - chartPadding.bottom)
const trendTotal = computed(() => trendPoints.value.reduce((sum, item) => sum + item.success + item.failed + item.running, 0))
const hoveredTrendPoint = computed(() => {
  if (hoveredTrendIndex.value === null) {
    return null
  }
  return trendPoints.value[hoveredTrendIndex.value] || null
})

function xForIndex(index: number): number {
  if (trendPoints.value.length <= 1) {
    return chartPadding.left
  }
  return chartPadding.left + (plotWidth.value / (trendPoints.value.length - 1)) * index
}

function yForValue(value: number): number {
  return chartPadding.top + plotHeight.value - (value / trendMax.value) * plotHeight.value
}

function linePath(key: 'success' | 'failed' | 'running'): string {
  return trendPoints.value
    .map((item, index) => `${index === 0 ? 'M' : 'L'} ${xForIndex(index)} ${yForValue(item[key])}`)
    .join(' ')
}

function areaPath(key: 'success' | 'failed' | 'running'): string {
  const points = trendPoints.value.map((item, index) => `${index === 0 ? 'M' : 'L'} ${xForIndex(index)} ${yForValue(item[key])}`).join(' ')
  const lastX = xForIndex(trendPoints.value.length - 1)
  const firstX = xForIndex(0)
  const bottomY = chartPadding.top + plotHeight.value
  return `${points} L ${lastX} ${bottomY} L ${firstX} ${bottomY} Z`
}

function hoverX(index: number): number {
  if (trendPoints.value.length <= 1) {
    return chartPadding.left
  }
  const step = plotWidth.value / (trendPoints.value.length - 1)
  if (index === 0) {
    return chartPadding.left
  }
  return xForIndex(index) - step / 2
}

function hoverWidth(index: number): number {
  if (trendPoints.value.length <= 1) {
    return plotWidth.value
  }
  const step = plotWidth.value / (trendPoints.value.length - 1)
  if (index === 0 || index === trendPoints.value.length - 1) {
    return step / 2
  }
  return step
}

function tooltipX(index: number): number {
  const width = 156
  return Math.min(Math.max(xForIndex(index) - width / 2, chartPadding.left), chartWidth - chartPadding.right - width)
}

onMounted(() => {
  void refresh()
})

async function refresh() {
  loading.value = true
  try {
    const [jobs, logs, executors, latestLogs] = await Promise.all([
      jobApi.listJobs(),
      jobLogApi.listJobLogs({ page: 1, pageSize: 1 }),
      executorApi.listExecutors(),
      jobLogApi.listJobLogs({ page: 1, pageSize: 100 }),
    ])
    recentLogs.value = latestLogs.items
    const recentSuccess = recentLogs.value.filter((item) => item.status === 'success').length
    const recentFailed = recentLogs.value.filter((item) => item.status === 'failed').length
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
        <div class="trend-section">
          <div class="chart-header">
            <div>
              <h2>日期分布图</h2>
              <p>最近 7 天执行结果趋势，按日志开始时间统计。</p>
            </div>
            <div class="chart-total">
              <span>样本记录</span>
              <strong>{{ trendTotal }}</strong>
            </div>
          </div>
          <div class="trend-legend">
            <span><i class="legend-line success-line"></i>成功</span>
            <span><i class="legend-line failed-line"></i>失败</span>
            <span><i class="legend-line running-line"></i>进行中</span>
          </div>
          <div class="trend-wrap">
            <svg class="trend-chart" :viewBox="`0 0 ${chartWidth} ${chartHeight}`" role="img" aria-label="最近 7 天执行结果趋势图">
              <g class="grid-lines">
                <template v-for="(tick, tickIndex) in yTicks" :key="`${tick}-${tickIndex}`">
                  <line
                    :x1="chartPadding.left"
                    :x2="chartWidth - chartPadding.right"
                    :y1="yForValue(tick)"
                    :y2="yForValue(tick)"
                  />
                  <text :x="chartPadding.left - 12" :y="yForValue(tick) + 4" text-anchor="end">{{ tick }}</text>
                </template>
              </g>
              <path class="trend-area success-area" :d="areaPath('success')" />
              <path class="trend-path success-path" :d="linePath('success')" />
              <path class="trend-path failed-path" :d="linePath('failed')" />
              <path class="trend-path running-path" :d="linePath('running')" />
              <g v-if="hoveredTrendPoint && hoveredTrendIndex !== null" class="trend-tooltip-layer">
                <line
                  class="hover-guide"
                  :x1="xForIndex(hoveredTrendIndex)"
                  :x2="xForIndex(hoveredTrendIndex)"
                  :y1="chartPadding.top"
                  :y2="chartPadding.top + plotHeight"
                />
                <g :transform="`translate(${tooltipX(hoveredTrendIndex)}, 16)`">
                  <rect class="trend-tooltip-box" width="156" height="96" rx="6" />
                  <text class="trend-tooltip-title" x="12" y="22">{{ hoveredTrendPoint.date }}</text>
                  <text class="trend-tooltip-success" x="12" y="46">成功：{{ hoveredTrendPoint.success }}</text>
                  <text class="trend-tooltip-failed" x="12" y="66">失败：{{ hoveredTrendPoint.failed }}</text>
                  <text class="trend-tooltip-running" x="12" y="86">进行中：{{ hoveredTrendPoint.running }}</text>
                </g>
              </g>
              <g v-for="(point, index) in trendPoints" :key="point.date">
                <circle class="trend-dot success-dot-stroke" :cx="xForIndex(index)" :cy="yForValue(point.success)" r="4" />
                <circle class="trend-dot failed-dot-stroke" :cx="xForIndex(index)" :cy="yForValue(point.failed)" r="4" />
                <circle class="trend-dot running-dot-stroke" :cx="xForIndex(index)" :cy="yForValue(point.running)" r="4" />
                <text class="axis-label" :x="xForIndex(index)" :y="chartHeight - 12" text-anchor="middle">{{ point.label }}</text>
              </g>
              <g class="hover-zones">
                <rect
                  v-for="(_point, index) in trendPoints"
                  :key="`hover-${index}`"
                  class="hover-zone"
                  :x="hoverX(index)"
                  :y="chartPadding.top"
                  :width="hoverWidth(index)"
                  :height="plotHeight"
                  @mouseenter="hoveredTrendIndex = index"
                  @mousemove="hoveredTrendIndex = index"
                  @mouseleave="hoveredTrendIndex = null"
                />
              </g>
            </svg>
          </div>
        </div>

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
          <a-descriptions-item label="日期分布图">统计最近 100 条执行日志中最近 7 天的 success、failed、running、killing 状态。</a-descriptions-item>
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
  display: grid;
  grid-template-columns: minmax(520px, 1.7fr) minmax(320px, 0.9fr);
  gap: 24px;
  padding: 16px;
  background: #fff;
  border: 1px solid #d9e2f2;
  border-radius: 6px;
}

.trend-section,
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

.trend-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 18px;
  align-items: center;
  justify-content: center;
  margin: -6px 0 10px;
  color: #334155;
}

.trend-legend span {
  display: inline-flex;
  gap: 7px;
  align-items: center;
}

.legend-line {
  display: inline-block;
  width: 24px;
  height: 3px;
  border-radius: 999px;
}

.success-line {
  background: #0ea66a;
}

.failed-line {
  background: #d93f3f;
}

.running-line {
  background: #f59e0b;
}

.trend-wrap {
  width: 100%;
  overflow: hidden;
}

.trend-chart {
  width: 100%;
  min-height: 260px;
}

.grid-lines line {
  stroke: #d8e1ee;
  stroke-width: 1;
}

.grid-lines text,
.axis-label {
  fill: #475569;
  font-size: 13px;
}

.trend-area {
  opacity: 0.32;
}

.success-area {
  fill: #0ea66a;
}

.trend-path {
  fill: none;
  stroke-width: 3;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.success-path {
  stroke: #0ea66a;
}

.failed-path {
  stroke: #d93f3f;
}

.running-path {
  stroke: #f59e0b;
}

.trend-dot {
  fill: #fff;
  stroke-width: 2;
}

.success-dot-stroke {
  stroke: #0ea66a;
}

.failed-dot-stroke {
  stroke: #d93f3f;
}

.running-dot-stroke {
  stroke: #f59e0b;
}

.hover-guide {
  stroke: #94a3b8;
  stroke-dasharray: 4 4;
  stroke-width: 1.5;
}

.trend-tooltip-box {
  fill: #fff;
  stroke: #d9e2f2;
  stroke-width: 1;
  filter: drop-shadow(0 8px 18px rgb(15 23 42 / 14%));
}

.trend-tooltip-title {
  fill: #172033;
  font-size: 13px;
  font-weight: 700;
}

.trend-tooltip-success,
.trend-tooltip-failed,
.trend-tooltip-running {
  font-size: 13px;
}

.trend-tooltip-success {
  fill: #0ea66a;
}

.trend-tooltip-failed {
  fill: #d93f3f;
}

.trend-tooltip-running {
  fill: #b87500;
}

.hover-zone {
  fill: transparent;
  cursor: crosshair;
}

.pie-layout {
  display: flex;
  flex-direction: column;
  gap: 18px;
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
    align-items: stretch;
  }

  .chart-header {
    flex-direction: column;
  }

  .chart-total {
    align-items: flex-start;
  }
}

@media (max-width: 1120px) {
  .chart-panel {
    grid-template-columns: 1fr;
  }
}
</style>
