<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { Modal } from 'ant-design-vue'
import CronExpressionPicker from '@/components/CronExpressionPicker.vue'
import PageHeaderBar from '@/components/PageHeaderBar.vue'
import PollingIndicator from '@/components/PollingIndicator.vue'
import StatusTag from '@/components/StatusTag.vue'
import { useExecutorsStore } from '@/stores/executors'
import { useGluesStore } from '@/stores/glues'
import { useJobLogsStore } from '@/stores/jobLogs'
import { useJobsStore } from '@/stores/jobs'
import type { ExecutorInfo } from '@/types/executor'
import type { JobForm, JobInfo } from '@/types/job'
import { formatDateTime } from '@/utils/datetime'
import { isActiveLogStatus } from '@/utils/status'

const jobsStore = useJobsStore()
const executorsStore = useExecutorsStore()
const gluesStore = useGluesStore()
const logsStore = useJobLogsStore()

const jobModalOpen = ref(false)
const glueDrawerOpen = ref(false)
const editingId = ref('')
const glueJob = ref<JobInfo | null>(null)
const pollingTimer = ref<number | null>(null)

const form = reactive<JobForm>({
  executorId: '',
  name: '',
  cronExpr: '0 */5 * * * *',
  timeoutSeconds: 3600,
  description: '',
})

const executorOptions = computed(() =>
  executorsStore.items.map((item: ExecutorInfo) => ({
    label: `${item.name} (${item.status})`,
    value: item.id,
  })),
)

const activeJobIds = computed(() => {
  return new Set(logsStore.activeItems.filter((log) => isActiveLogStatus(log.status)).map((log) => log.jobId))
})

const modalTitle = computed(() => (editingId.value ? '编辑任务' : '新增任务'))

onMounted(async () => {
  await Promise.all([executorsStore.fetchList(), jobsStore.fetchList(), refreshActiveLogs()])
  pollingTimer.value = window.setInterval(() => {
    void refreshActiveLogs()
  }, 5000)
})

onBeforeUnmount(() => {
  if (pollingTimer.value) {
    window.clearInterval(pollingTimer.value)
  }
})

async function refreshActiveLogs() {
  await logsStore.fetchActiveList()
}

function resetForm() {
  editingId.value = ''
  form.id = undefined
  form.executorId = executorsStore.items[0]?.id || ''
  form.name = ''
  form.cronExpr = '0 */5 * * * *'
  form.timeoutSeconds = 3600
  form.description = ''
}

function openCreate() {
  resetForm()
  jobModalOpen.value = true
}

function openEdit(row: JobInfo) {
  editingId.value = row.id
  form.id = row.id
  form.executorId = row.executorId
  form.name = row.name
  form.cronExpr = row.cronExpr
  form.timeoutSeconds = row.timeoutSeconds
  form.description = row.description
  jobModalOpen.value = true
}

async function submitJob() {
  await jobsStore.save({ ...form })
  jobModalOpen.value = false
  resetForm()
}

async function openGlue(row: JobInfo) {
  glueJob.value = row
  glueDrawerOpen.value = true
  await gluesStore.fetchByJob(row.id)
}

async function saveGlue() {
  if (!glueJob.value) {
    return
  }
  await gluesStore.save(glueJob.value.id, gluesStore.content)
}

function confirmDelete(row: JobInfo) {
  Modal.confirm({
    title: '删除任务',
    content: `确认删除「${row.name}」？执行日志会按后端保留策略清理。`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: () => jobsStore.remove(row.id),
  })
}

function confirmStop(row: JobInfo) {
  Modal.confirm({
    title: '停止调度',
    content: `确认停止「${row.name}」的定时调度？当前运行中的实例不受影响。`,
    okText: '停止',
    cancelText: '取消',
    onOk: () => jobsStore.stop(row.id),
  })
}

function confirmKill(row: JobInfo) {
  Modal.confirm({
    title: '终止运行中任务',
    content: `确认请求执行器终止「${row.name}」当前运行实例？`,
    okText: '终止',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      await jobsStore.kill(row.id)
      await refreshActiveLogs()
    },
  })
}

async function runNow(row: JobInfo) {
  await jobsStore.run(row.id)
  await refreshActiveLogs()
}
</script>

<template>
  <div class="page-body">
    <PageHeaderBar title="任务" description="管理 Cron 调度、Glue Shell 脚本和运行控制。">
      <a-space>
        <PollingIndicator :active="true" />
        <a-button @click="jobsStore.fetchList">刷新</a-button>
        <a-button type="primary" @click="openCreate">新增任务</a-button>
      </a-space>
    </PageHeaderBar>

    <div class="table-shell">
      <a-table
        row-key="id"
        :data-source="jobsStore.items"
        :loading="jobsStore.loading"
        :pagination="false"
        size="middle"
        :scroll="{ x: 1220 }"
      >
        <a-table-column title="任务" data-index="name" :width="190" fixed="left" />
        <a-table-column title="执行器" data-index="executorId" :width="170">
          <template #default="{ text }">
            {{ executorsStore.items.find((item) => item.id === text)?.name || text }}
          </template>
        </a-table-column>
        <a-table-column title="Cron" data-index="cronExpr" :width="170">
          <template #default="{ text }">
            <span class="mono">{{ text }}</span>
          </template>
        </a-table-column>
        <a-table-column title="超时" data-index="timeoutSeconds" :width="90">
          <template #default="{ text }">{{ text }}s</template>
        </a-table-column>
        <a-table-column title="调度状态" data-index="scheduleStatus" :width="110">
          <template #default="{ text }">
            <StatusTag :status="text" />
          </template>
        </a-table-column>
        <a-table-column title="执行状态" :width="110">
          <template #default="{ record }">
            <StatusTag v-if="activeJobIds.has((record as JobInfo).id)" status="running" />
            <span v-else class="muted-text">空闲</span>
          </template>
        </a-table-column>
        <a-table-column title="更新时间" data-index="updatedAt" :width="180">
          <template #default="{ text }">{{ formatDateTime(text) }}</template>
        </a-table-column>
        <a-table-column title="说明" data-index="description" />
        <a-table-column title="操作" :width="360" fixed="right">
          <template #default="{ record }">
            <a-space wrap>
              <a-button type="link" size="small" @click="openEdit(record as JobInfo)">编辑</a-button>
              <a-button type="link" size="small" @click="openGlue(record as JobInfo)">Glue</a-button>
              <a-button
                v-if="(record as JobInfo).scheduleStatus === 'stopped'"
                type="link"
                size="small"
                :loading="jobsStore.actionLoadingIds.includes((record as JobInfo).id)"
                @click="jobsStore.start((record as JobInfo).id)"
              >
                启动
              </a-button>
              <a-button v-else type="link" size="small" @click="confirmStop(record as JobInfo)">停止</a-button>
              <a-tooltip :title="activeJobIds.has((record as JobInfo).id) ? '任务正在执行中，不能重复手动运行' : ''">
                <a-button
                  type="link"
                  size="small"
                  :disabled="activeJobIds.has((record as JobInfo).id)"
                  :loading="jobsStore.actionLoadingIds.includes((record as JobInfo).id)"
                  @click="runNow(record as JobInfo)"
                >
                  运行
                </a-button>
              </a-tooltip>
              <a-button
                v-if="activeJobIds.has((record as JobInfo).id)"
                danger
                type="link"
                size="small"
                @click="confirmKill(record as JobInfo)"
              >
                终止
              </a-button>
              <a-button danger type="link" size="small" @click="confirmDelete(record as JobInfo)">删除</a-button>
            </a-space>
          </template>
        </a-table-column>
      </a-table>
    </div>

    <a-modal v-model:open="jobModalOpen" :title="modalTitle" :confirm-loading="jobsStore.submitting" @ok="submitJob">
      <a-form layout="vertical" :model="form">
        <a-form-item label="任务名称" required>
          <a-input v-model:value="form.name" placeholder="如：daily-report" />
        </a-form-item>
        <a-form-item label="执行器" required>
          <a-select v-model:value="form.executorId" :options="executorOptions" placeholder="选择执行器" />
        </a-form-item>
        <a-form-item label="Cron 表达式" required>
          <CronExpressionPicker v-model:value="form.cronExpr" />
        </a-form-item>
        <a-form-item label="超时时间（秒）" required>
          <a-input-number v-model:value="form.timeoutSeconds" :min="1" :max="604800" style="width: 100%" />
        </a-form-item>
        <a-form-item label="说明">
          <a-textarea v-model:value="form.description" :rows="3" />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-drawer
      v-model:open="glueDrawerOpen"
      :title="glueJob ? `Glue Shell - ${glueJob.name}` : 'Glue Shell'"
      width="min(760px, 100vw)"
    >
      <a-spin :spinning="gluesStore.loading">
        <a-textarea v-model:value="gluesStore.content" class="glue-editor mono" />
      </a-spin>
      <template #extra>
        <a-button type="primary" :loading="gluesStore.submitting" @click="saveGlue">保存</a-button>
      </template>
    </a-drawer>
  </div>
</template>

<style scoped>
.glue-editor {
  min-height: calc(100vh - 148px);
  resize: vertical;
}
</style>
