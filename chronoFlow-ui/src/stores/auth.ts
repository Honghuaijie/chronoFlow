import { defineStore } from 'pinia'
import { message } from 'ant-design-vue'
import { clearStoredToken, getStoredToken, setStoredToken } from '@/api/request'
import * as authApi from '@/api/auth'
import type { CurrentUser, LoginParams } from '@/types/auth'

interface AuthState {
  token: string
  user: CurrentUser | null
  loading: boolean
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: getStoredToken(),
    user: null,
    loading: false,
  }),
  getters: {
    isLoggedIn: (state) => Boolean(state.token),
  },
  actions: {
    async login(params: LoginParams) {
      this.loading = true
      try {
        const data = await authApi.login(params)
        this.token = data.token
        this.user = {
          userId: 0,
          username: data.username,
          role: 'admin',
        }
        setStoredToken(data.token)
        message.success('登录成功')
      } finally {
        this.loading = false
      }
    },
    async fetchCurrent() {
      if (!this.token) {
        return
      }
      this.loading = true
      try {
        this.user = await authApi.currentUser()
      } finally {
        this.loading = false
      }
    },
    logout() {
      this.token = ''
      this.user = null
      clearStoredToken()
    },
  },
})
