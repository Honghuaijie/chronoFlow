import { defineStore } from 'pinia'
import { message } from 'ant-design-vue'
import * as glueApi from '@/api/glues'
import type { GlueInfo } from '@/types/glue'

interface GluesState {
  current: GlueInfo | null
  content: string
  loading: boolean
  submitting: boolean
}

export const useGluesStore = defineStore('glues', {
  state: (): GluesState => ({
    current: null,
    content: '',
    loading: false,
    submitting: false,
  }),
  actions: {
    async fetchByJob(jobId: string) {
      this.loading = true
      try {
        this.current = await glueApi.getGlue(jobId)
        this.content = this.current?.content || ''
      } finally {
        this.loading = false
      }
    },
    async save(jobId: string, content: string) {
      this.submitting = true
      try {
        this.current = await glueApi.saveGlue(jobId, content)
        this.content = this.current.content
        message.success('Glue 脚本已保存')
      } finally {
        this.submitting = false
      }
    },
  },
})
