<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PageHeaderBar from '@/components/PageHeaderBar.vue'
import LogViewer from '@/components/LogViewer.vue'
import PollingIndicator from '@/components/PollingIndicator.vue'
import StatusTag from '@/components/StatusTag.vue'
import { useJobLogsStore } from '@/stores/jobLogs'
import type { JobLogInfo } from '@/types/jobLog'
import { formatBytes, formatDateTime, formatDuration } from '@/utils/datetime'
import { isActiveLogStatus } from '@/utils/status'

const route = useRoute()
const router = useRouter()
const store = useJobLogsStore()
const pollingTimer = ref<number | null>(null)

const logId = computed(() => String(route.params.id || ''))
const log = computed(() => store.detail?.log)
const pollingActive = computed(() => (log.value ? isActiveLogStatus(log.value.status) : false))

onMounted(async () => {
  await store.fetchDetail(logId.value)
  pollingTimer.value = window.setInterval(() => {
    if (pollingActive.value) {
      void store.fetchDetail(logId.value)
    }
  }, 5000)
})

onBeforeUnmount(() => {
  if (pollingTimer.value) {
    window.clearInterval(pollingTimer.value)
  }
})

async function kill() {
  if (!log.value) {
    return
  }
  await store.killJob(log.value.jobId, log.value.id)
}

function alertStatusText(item: JobLogInfo): string {
  if (!item.alertEnabledSnapshot) {
    return '未启用'
  }
  if (item.alertStatus === 'sent') {
    return '已发送'
  }
  if (item.alertStatus === 'pending') {
    return '发送中'
  }
  if (item.alertStatus === 'failed') {
    return '发送失败'
  }
  if (item.alertStatus === 'skipped') {
    return '未发送'
  }
  return '未发送'
}

function alertTagColor(item: JobLogInfo): string {
  if (!item.alertEnabledSnapshot) {
    return 'default'
  }
  if (item.alertStatus === 'sent') {
    return 'green'
  }
  if (item.alertStatus === 'failed') {
    return 'red'
  }
  if (item.alertStatus === 'pending' || item.alertStatus === 'skipped') {
    return 'orange'
  }
  return 'default'
}
</script>

<template>
  <div class="page-body">
    <PageHeaderBar title="日志详情" description="查看执行元数据、Glue 快照和脚本输出。">
      <a-space>
        <PollingIndicator :active="pollingActive" />
        <a-button @click="router.push('/logs')">返回列表</a-button>
        <a-button v-if="log && isActiveLogStatus(log.status)" danger :loading="store.actionLoading" @click="kill">终止</a-button>
      </a-space>
    </PageHeaderBar>

    <a-skeleton v-if="store.detailLoading && !store.detail" active />
    <template v-else-if="log">
      <div class="detail-grid">
        <a-descriptions bordered size="small" :column="{ xs: 1, sm: 2, xl: 3 }">
          <a-descriptions-item label="日志 ID">{{ log.id }}</a-descriptions-item>
          <a-descriptions-item label="任务">{{ log.jobName }}</a-descriptions-item>
          <a-descriptions-item label="状态"><StatusTag :status="log.status" /></a-descriptions-item>
          <a-descriptions-item label="执行器">{{ log.executorName }}</a-descriptions-item>
          <a-descriptions-item label="触发"><StatusTag :status="log.triggerType" /></a-descriptions-item>
          <a-descriptions-item label="Cron"><span class="mono">{{ log.cronExpr }}</span></a-descriptions-item>
          <a-descriptions-item label="开始">{{ formatDateTime(log.startTime) }}</a-descriptions-item>
          <a-descriptions-item label="结束">{{ formatDateTime(log.endTime) }}</a-descriptions-item>
          <a-descriptions-item label="耗时">{{ formatDuration(log.durationMs) }}</a-descriptions-item>
          <a-descriptions-item label="退出码">{{ log.exitCode }}</a-descriptions-item>
          <a-descriptions-item label="日志大小">{{ formatBytes(log.logSizeBytes) }}</a-descriptions-item>
          <a-descriptions-item label="日志截断">{{ log.logTruncated ? '是' : '否' }}</a-descriptions-item>
          <a-descriptions-item label="失败告警">
            <a-space direction="vertical" size="small">
              <a-tag :color="alertTagColor(log)">{{ alertStatusText(log) }}</a-tag>
              <span v-if="log.alertSentAt" class="muted-text">{{ formatDateTime(log.alertSentAt) }}</span>
              <span v-if="log.alertError" class="alert-error">{{ log.alertError }}</span>
            </a-space>
          </a-descriptions-item>
          <a-descriptions-item label="错误信息" :span="3">{{ log.errorMessage || '-' }}</a-descriptions-item>
          <a-descriptions-item label="日志路径" :span="3"><span class="mono">{{ log.logPath || '-' }}</span></a-descriptions-item>
        </a-descriptions>
      </div>

      <a-tabs>
        <a-tab-pane key="log" tab="日志正文">
          <LogViewer :content="store.detail?.logContent || ''" :loading="store.detailLoading" />
        </a-tab-pane>
        <a-tab-pane key="glue" tab="Glue 快照">
          <pre class="code-box mono">{{ store.detail?.glueSnapshot || '暂无 Glue 快照' }}</pre>
        </a-tab-pane>
      </a-tabs>
    </template>
    <a-empty v-else description="日志不存在" />
  </div>
</template>

<style scoped>
.detail-grid {
  overflow: hidden;
  background: #fff;
  border: 1px solid #d9e2f2;
  border-radius: 6px;
}

.alert-error {
  display: inline-block;
  max-width: 420px;
  color: #dc2626;
  font-size: 12px;
  line-height: 1.5;
  word-break: break-word;
}
</style>
