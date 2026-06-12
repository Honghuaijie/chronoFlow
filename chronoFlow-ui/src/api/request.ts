import axios, { AxiosError } from 'axios'
import { message } from 'ant-design-vue'
import type { ApiEnvelope } from '@/types/api'

const TOKEN_KEY = 'chronoflow_token'

export class ApiError extends Error {
  code: number

  constructor(code: number, messageText: string) {
    super(messageText)
    this.name = 'ApiError'
    this.code = code
  }
}

export function getStoredToken(): string {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function setStoredToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearStoredToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 15000,
})

request.interceptors.request.use((config) => {
  const token = getStoredToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    const envelope = response.data as ApiEnvelope<unknown>
    if (typeof envelope?.code === 'number' && envelope.code !== 0) {
      throw new ApiError(envelope.code, envelope.message || '请求失败')
    }
    return response
  },
  (error: AxiosError<ApiEnvelope<unknown>>) => {
    const status = error.response?.status
    const msg = error.response?.data?.message || error.message || '网络请求失败'
    if (status === 401) {
      clearStoredToken()
      message.warning('登录已失效，请重新登录')
      window.location.href = '/login'
    }
    throw new ApiError(status || -1, msg)
  },
)

export async function getData<T>(url: string, params?: object): Promise<T> {
  const response = await request.get<ApiEnvelope<T>>(url, { params })
  return response.data.data
}

export async function postData<T, P extends object>(url: string, payload: P): Promise<T> {
  const response = await request.post<ApiEnvelope<T>>(url, payload)
  return response.data.data
}
