import { getData } from './request'
import type { PageResult } from '@/types/api'
import type { AlertStatus, JobLogDetail, JobLogFilters, JobLogInfo } from '@/types/jobLog'
import { normalizeId, toApiId } from '@/utils/id'

interface JobLogPayload
  extends Omit<JobLogInfo, 'id' | 'jobId' | 'executorId' | 'durationMs' | 'logSizeBytes' | 'alertEnabledSnapshot' | 'alertStatus' | 'alertError' | 'alertSentAt'> {
  id: string | number
  jobId?: string | number
  job_id?: string | number
  executorId?: string | number
  executor_id?: string | number
  durationMs: string | number
  logSizeBytes: string | number
  alertEnabledSnapshot?: boolean
  alert_enabled_snapshot?: boolean
  alertStatus?: string
  alert_status?: string
  alertError?: string
  alert_error?: string
  alertSentAt?: string
  alert_sent_at?: string
}

interface JobLogListData {
  items: JobLogPayload[]
  total: number
}

interface JobLogDetailData {
  log: JobLogPayload
  glueSnapshot?: string
  glue_snapshot?: string
  logContent?: string
  log_content?: string
}

export interface ListJobLogsParams extends JobLogFilters {
  page: number
  pageSize: number
}

function mapJobLog(payload: JobLogPayload): JobLogInfo {
  return {
    ...payload,
    id: normalizeId(payload.id),
    jobId: normalizeId(payload.jobId ?? payload.job_id),
    executorId: normalizeId(payload.executorId ?? payload.executor_id),
    durationMs: Number(payload.durationMs || 0),
    logSizeBytes: Number(payload.logSizeBytes || 0),
    alertEnabledSnapshot: Boolean(payload.alertEnabledSnapshot ?? payload.alert_enabled_snapshot),
    alertStatus: mapAlertStatus(payload.alertStatus ?? payload.alert_status),
    alertError: payload.alertError ?? payload.alert_error ?? '',
    alertSentAt: payload.alertSentAt ?? payload.alert_sent_at ?? '',
  }
}

function mapAlertStatus(status?: string): AlertStatus {
  if (status === 'pending' || status === 'sent' || status === 'failed' || status === 'skipped' || status === 'none' || status === '') {
    return status
  }
  return 'none'
}

export async function listJobLogs(params: ListJobLogsParams): Promise<PageResult<JobLogInfo>> {
  const data = await getData<JobLogListData>('/v1/admin/jobLogs/list', {
    page: params.page,
    pageSize: params.pageSize,
    jobId: params.jobId ? toApiId(params.jobId) : undefined,
    executorId: params.executorId ? toApiId(params.executorId) : undefined,
    status: params.status || undefined,
    triggerType: params.triggerType || undefined,
  })
  return {
    items: data.items.map(mapJobLog),
    total: data.total,
  }
}

export async function getJobLogDetail(id: string): Promise<JobLogDetail> {
  const data = await getData<JobLogDetailData>(`/v1/admin/jobLogs/detail/${toApiId(id)}`)
  return {
    log: mapJobLog(data.log),
    glueSnapshot: data.glueSnapshot ?? data.glue_snapshot ?? '',
    logContent: data.logContent ?? data.log_content ?? '',
  }
}
