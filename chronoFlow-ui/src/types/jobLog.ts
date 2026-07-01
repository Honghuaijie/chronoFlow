import type { Id } from './api'

export type JobLogStatus = 'running' | 'killing' | 'success' | 'failed' | 'timeout' | 'killed' | 'skipped' | string
export type TriggerType = 'manual' | 'cron' | string
export type AlertStatus = 'none' | 'pending' | 'sent' | 'failed' | 'skipped' | ''

export interface JobLogInfo {
  id: Id
  jobId: Id
  jobName: string
  executorId: Id
  executorName: string
  executorAddress: string
  cronExpr: string
  timeoutSeconds: number
  triggerType: TriggerType
  status: JobLogStatus
  startTime: string
  endTime: string
  durationMs: number
  exitCode: number
  logPath: string
  logSizeBytes: number
  logTruncated: boolean
  errorMessage: string
  alertEnabledSnapshot: boolean
  alertStatus: AlertStatus
  alertError: string
  alertSentAt: string
  createdAt: string
  updatedAt: string
}

export interface JobLogFilters {
  jobId?: Id
  executorId?: Id
  status?: string
  triggerType?: string
}

export interface JobLogDetail {
  log: JobLogInfo
  glueSnapshot: string
  logContent: string
}
