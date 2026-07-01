import type { Id } from './api'

export type JobScheduleStatus = 'running' | 'stopped' | string

export interface JobInfo {
  id: Id
  executorId: Id
  name: string
  cronExpr: string
  timeoutSeconds: number
  scheduleStatus: JobScheduleStatus
  description: string
  failureAlertEnabled: boolean
  createdAt: string
  updatedAt: string
}

export interface JobForm {
  id?: Id
  executorId: Id
  name: string
  cronExpr: string
  timeoutSeconds: number
  description: string
  failureAlertEnabled: boolean
}

export interface RunJobResult {
  logId: Id
  status: string
}
