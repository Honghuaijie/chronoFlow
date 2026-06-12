import { defineStore } from 'pinia'
import { message } from 'ant-design-vue'
import * as jobLogApi from '@/api/jobLogs'
import * as jobApi from '@/api/jobs'
import type { PaginationState } from '@/types/api'
import type { JobLogDetail, JobLogFilters, JobLogInfo } from '@/types/jobLog'

interface JobLogsState {
  items: JobLogInfo[]
  activeItems: JobLogInfo[]
  detail: JobLogDetail | null
  loading: boolean
  detailLoading: boolean
  actionLoading: boolean
  pagination: PaginationState
  filters: JobLogFilters
}

export const useJobLogsStore = defineStore('jobLogs', {
  state: (): JobLogsState => ({
    items: [],
    activeItems: [],
    detail: null,
    loading: false,
    detailLoading: false,
    actionLoading: false,
    pagination: {
      page: 1,
      pageSize: 20,
      total: 0,
    },
    filters: {},
  }),
  actions: {
    async fetchList() {
      this.loading = true
      try {
        const data = await jobLogApi.listJobLogs({
          ...this.filters,
          page: this.pagination.page,
          pageSize: this.pagination.pageSize,
        })
        this.items = data.items
        this.pagination.total = data.total
      } finally {
        this.loading = false
      }
    },
    async fetchActiveList() {
      const [running, killing] = await Promise.all([
        jobLogApi.listJobLogs({ page: 1, pageSize: 1000, status: 'running' }),
        jobLogApi.listJobLogs({ page: 1, pageSize: 1000, status: 'killing' }),
      ])
      this.activeItems = [...running.items, ...killing.items]
    },
    async fetchDetail(id: string) {
      this.detailLoading = true
      try {
        this.detail = await jobLogApi.getJobLogDetail(id)
      } finally {
        this.detailLoading = false
      }
    },
    async killJob(jobId: string, logId?: string) {
      this.actionLoading = true
      try {
        await jobApi.killJob(jobId)
        message.warning('终止请求已下发')
        if (logId) {
          await this.fetchDetail(logId)
        }
        await this.fetchList()
      } finally {
        this.actionLoading = false
      }
    },
    setPage(page: number, pageSize: number) {
      this.pagination.page = page
      this.pagination.pageSize = pageSize
    },
    setFilters(filters: JobLogFilters) {
      this.filters = filters
      this.pagination.page = 1
    },
  },
})
