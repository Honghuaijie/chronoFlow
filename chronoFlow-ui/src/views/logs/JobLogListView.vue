<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import PageHeaderBar from '@/components/PageHeaderBar.vue'
import PollingIndicator from '@/components/PollingIndicator.vue'
import StatusTag from '@/components/StatusTag.vue'
import { useExecutorsStore } from '@/stores/executors'
import { useJobLogsStore } from '@/stores/jobLogs'
import { useJobsStore } from '@/stores/jobs'
import type { JobLogFilters, JobLogInfo } from '@/types/jobLog'
import { formatBytes, formatDateTime, formatDuration } from '@/utils/datetime'
import { isActiveLogStatus } from '@/utils/status'

const router = useRouter()
const logsStore = useJobLogsStore()
const jobsStore = useJobsStore()
const executorsStore = useExecutorsStore()
const pollingTimer = ref<number | null>(null)

const filters = reactive<JobLogFilters>({
  jobId: undefined,
  executorId: undefined,
  status: undefined,
  triggerType: undefined,
})

const pollingActive = computed(() => logsStore.items.some((log) => isActiveLogStatus(log.status)))
const jobOptions = computed(() => jobsStore.items.map((item) => ({ label: item.name, value: item.id })))
const executorOptions = computed(() => executorsStore.items.map((item) => ({ label: item.name, value: item.id })))

onMounted(async () => {
  await Promise.all([jobsStore.fetchList(), executorsStore.fetchList(), refresh()])
  pollingTimer.value = window.setInterval(() => {
    if (pollingActive.value) {
      void refresh()
    }
  }, 5000)
})

onBeforeUnmount(() => {
  if (pollingTimer.value) {
    window.clearInterval(pollingTimer.value)
  }
})

async function refresh() {
  await logsStore.fetchList()
}

async function applyFilters() {
  logsStore.setFilters({ ...filters })
  await refresh()
}

async function resetFilters() {
  filters.jobId = undefined
  filters.executorId = undefined
  filters.status = undefined
  filters.triggerType = undefined
  await applyFilters()
}

async function onPageChange(page: number, pageSize: number) {
  logsStore.setPage(page, pageSize)
  await refresh()
}
</script>

<template>
  <div class="page-body">
    <PageHeaderBar title="执行日志" description="查看任务执行状态、错误信息和文件日志正文。">
      <a-space>
        <PollingIndicator :active="pollingActive" />
        <a-button @click="refresh">刷新</a-button>
      </a-space>
    </PageHeaderBar>

    <div class="toolbar">
      <div class="toolbar-left">
        <a-select
          v-model:value="filters.jobId"
          allow-clear
          show-search
          placeholder="任务"
          style="width: 180px"
          :options="jobOptions"
        />
        <a-select
          v-model:value="filters.executorId"
          allow-clear
          show-search
          placeholder="执行器"
          style="width: 180px"
          :options="executorOptions"
        />
        <a-select v-model:value="filters.status" allow-clear placeholder="状态" style="width: 140px">
          <a-select-option value="running">运行中</a-select-option>
          <a-select-option value="killing">终止中</a-select-option>
          <a-select-option value="success">成功</a-select-option>
          <a-select-option value="failed">失败</a-select-option>
          <a-select-option value="killed">已终止</a-select-option>
          <a-select-option value="skipped">已跳过</a-select-option>
        </a-select>
        <a-select v-model:value="filters.triggerType" allow-clear placeholder="触发方式" style="width: 140px">
          <a-select-option value="manual">手动</a-select-option>
          <a-select-option value="cron">定时</a-select-option>
        </a-select>
      </div>
      <div class="toolbar-right">
        <a-button type="primary" @click="applyFilters">筛选</a-button>
        <a-button @click="resetFilters">重置</a-button>
      </div>
    </div>

    <div class="table-shell">
      <a-table
        row-key="id"
        :data-source="logsStore.items"
        :loading="logsStore.loading"
        size="middle"
        :scroll="{ x: 1280 }"
        :pagination="{
          current: logsStore.pagination.page,
          pageSize: logsStore.pagination.pageSize,
          total: logsStore.pagination.total,
          showSizeChanger: true,
          showTotal: (total: number) => `共 ${total} 条`,
          onChange: onPageChange,
        }"
      >
        <a-table-column title="ID" data-index="id" :width="86" fixed="left" />
        <a-table-column title="任务" data-index="jobName" :width="180" />
        <a-table-column title="执行器" data-index="executorName" :width="160" />
        <a-table-column title="触发" data-index="triggerType" :width="90">
          <template #default="{ text }"><StatusTag :status="text" /></template>
        </a-table-column>
        <a-table-column title="状态" data-index="status" :width="100">
          <template #default="{ text }"><StatusTag :status="text" /></template>
        </a-table-column>
        <a-table-column title="开始时间" data-index="startTime" :width="180">
          <template #default="{ text }">{{ formatDateTime(text) }}</template>
        </a-table-column>
        <a-table-column title="耗时" data-index="durationMs" :width="100">
          <template #default="{ text }">{{ formatDuration(text) }}</template>
        </a-table-column>
        <a-table-column title="日志大小" data-index="logSizeBytes" :width="110">
          <template #default="{ text }">{{ formatBytes(text) }}</template>
        </a-table-column>
        <a-table-column title="错误" data-index="errorMessage" />
        <a-table-column title="操作" :width="140" fixed="right">
          <template #default="{ record }">
            <a-space>
              <a-button type="link" size="small" @click="router.push(`/logs/${(record as JobLogInfo).id}`)">详情</a-button>
              <a-button
                v-if="isActiveLogStatus((record as JobLogInfo).status)"
                danger
                type="link"
                size="small"
                :loading="logsStore.actionLoading"
                @click="logsStore.killJob((record as JobLogInfo).jobId)"
              >
                终止
              </a-button>
            </a-space>
          </template>
        </a-table-column>
      </a-table>
    </div>
  </div>
</template>
