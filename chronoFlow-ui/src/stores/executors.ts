import { defineStore } from 'pinia'
import { message } from 'ant-design-vue'
import * as executorApi from '@/api/executors'
import type { ExecutorForm, ExecutorInfo } from '@/types/executor'

interface ExecutorsState {
  items: ExecutorInfo[]
  total: number
  loading: boolean
  submitting: boolean
}

export const useExecutorsStore = defineStore('executors', {
  state: (): ExecutorsState => ({
    items: [],
    total: 0,
    loading: false,
    submitting: false,
  }),
  actions: {
    async fetchList() {
      this.loading = true
      try {
        const data = await executorApi.listExecutors()
        this.items = data.items
        this.total = data.total
      } finally {
        this.loading = false
      }
    },
    async save(form: ExecutorForm) {
      this.submitting = true
      try {
        if (form.id) {
          await executorApi.updateExecutor(form)
          message.success('执行器已更新')
        } else {
          await executorApi.createExecutor(form)
          message.success('执行器已创建')
        }
        await this.fetchList()
      } finally {
        this.submitting = false
      }
    },
    async remove(id: string) {
      this.submitting = true
      try {
        await executorApi.deleteExecutor(id)
        message.success('执行器已删除')
        await this.fetchList()
      } finally {
        this.submitting = false
      }
    },
  },
})
