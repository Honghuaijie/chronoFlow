import { defineStore } from 'pinia'
import { message } from 'ant-design-vue'
import * as jobApi from '@/api/jobs'
import type { JobForm, JobInfo } from '@/types/job'

interface JobsState {
  items: JobInfo[]
  total: number
  loading: boolean
  submitting: boolean
  actionLoadingIds: string[]
  filters: {
    executorId?: string
  }
}

export const useJobsStore = defineStore('jobs', {
  state: (): JobsState => ({
    items: [],
    total: 0,
    loading: false,
    submitting: false,
    actionLoadingIds: [],
    filters: {},
  }),
  getters: {
    runningJobIds: (state) => new Set(state.items.filter((job) => job.scheduleStatus === 'running').map((job) => job.id)),
  },
  actions: {
    async fetchList() {
      this.loading = true
      try {
        const data = await jobApi.listJobs(this.filters)
        this.items = data.items
        this.total = data.total
      } finally {
        this.loading = false
      }
    },
    async save(form: JobForm) {
      this.submitting = true
      try {
        if (form.id) {
          await jobApi.updateJob(form)
          message.success('任务已更新，新配置下次执行生效')
        } else {
          await jobApi.createJob(form)
          message.success('任务已创建')
        }
        await this.fetchList()
      } finally {
        this.submitting = false
      }
    },
    async remove(id: string) {
      this.submitting = true
      try {
        await jobApi.deleteJob(id)
        message.success('任务已删除')
        await this.fetchList()
      } finally {
        this.submitting = false
      }
    },
    async start(id: string) {
      await this.withActionLoading(id, async () => {
        await jobApi.startJob(id)
        message.success('任务已启动')
        await this.fetchList()
      })
    },
    async stop(id: string) {
      await this.withActionLoading(id, async () => {
        await jobApi.stopJob(id)
        message.success('任务已停止')
        await this.fetchList()
      })
    },
    async run(id: string) {
      await this.withActionLoading(id, async () => {
        const result = await jobApi.runJob(id)
        message.success(`任务已下发，日志 #${result.logId}`)
        await this.fetchList()
      })
    },
    async kill(id: string) {
      await this.withActionLoading(id, async () => {
        const result = await jobApi.killJob(id)
        message.warning(`终止请求已下发，日志 #${result.logId}`)
        await this.fetchList()
      })
    },
    async withActionLoading(id: string, action: () => Promise<void>) {
      this.actionLoadingIds.push(id)
      try {
        await action()
      } finally {
        this.actionLoadingIds = this.actionLoadingIds.filter((item) => item !== id)
      }
    },
  },
})
