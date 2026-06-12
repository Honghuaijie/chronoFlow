import type { Id } from './api'

export type ExecutorStatus = 'online' | 'offline' | string

export interface ExecutorInfo {
  id: Id
  name: string
  address: string
  description: string
  status: ExecutorStatus
  heartbeatFailCount: number
  lastHeartbeatTime: string
  createdAt: string
  updatedAt: string
}

export interface ExecutorForm {
  id?: Id
  name: string
  address: string
  token: string
  description: string
}
