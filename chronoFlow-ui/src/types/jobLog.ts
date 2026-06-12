import type { Id } from './api'

export type JobLogStatus = 'running' | 'killing' | 'success' | 'failed' | 'killed' | 'skipped' | string
export type TriggerType = 'manual' | 'cron' | string

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
