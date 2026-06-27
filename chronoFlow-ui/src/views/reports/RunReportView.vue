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
}

const loading = ref(false)
const loadedAt = ref('')
const stats = ref<ReportStats>({
  jobTotal: 0,
  runTotal: 0,
  executorTotal: 0,
})

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
    stats.value = {
      jobTotal: jobs.total,
      runTotal: logs.total,
      executorTotal: executors.total,
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

      <div class="report-panel">
        <div class="report-panel-header">
          <BarChartOutlined />
          <span>统计口径</span>
        </div>
        <a-descriptions :column="1" bordered size="middle">
          <a-descriptions-item label="任务数量">统计当前任务列表中的全部任务。</a-descriptions-item>
          <a-descriptions-item label="调度次数">统计执行日志总数，包含手动运行和 Cron 定时触发。</a-descriptions-item>
          <a-descriptions-item label="执行器数量">统计当前执行器列表中的全部执行器。</a-descriptions-item>
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
}
</style>
