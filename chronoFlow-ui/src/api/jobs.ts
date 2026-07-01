import { getData, postData } from './request'
import type { PageResult } from '@/types/api'
import type { JobForm, JobInfo, RunJobResult } from '@/types/job'
import { normalizeId, toApiId } from '@/utils/id'

interface JobPayload extends Omit<JobInfo, 'id' | 'executorId' | 'failureAlertEnabled'> {
  id: string | number
  executorId?: string | number
  executor_id?: string | number
  failureAlertEnabled?: boolean
  failure_alert_enabled?: boolean
}

interface JobData {
  job: JobPayload
}

interface JobListData {
  items: JobPayload[]
  total: number
}

interface RunJobData {
  logId?: string | number
  log_id?: string | number
  status: string
}

interface JobListParams {
  executorId?: string
}

function mapJob(payload: JobPayload): JobInfo {
  return {
    ...payload,
    id: normalizeId(payload.id),
    executorId: normalizeId(payload.executorId ?? payload.executor_id),
    failureAlertEnabled: Boolean(payload.failureAlertEnabled ?? payload.failure_alert_enabled),
  }
}

function toApiJobForm(form: JobForm): Omit<JobForm, 'id' | 'executorId'> & { id?: number; executorId: number } {
  return {
    id: form.id ? toApiId(form.id) : undefined,
    executorId: toApiId(form.executorId),
    name: form.name,
    cronExpr: form.cronExpr,
    timeoutSeconds: form.timeoutSeconds,
    description: form.description,
    failureAlertEnabled: form.failureAlertEnabled,
  }
}

function mapRunJob(data: RunJobData): RunJobResult {
  return {
    logId: normalizeId(data.logId ?? data.log_id),
    status: data.status,
  }
}

export async function listJobs(params: JobListParams = {}): Promise<PageResult<JobInfo>> {
  const data = await getData<JobListData>('/v1/admin/jobs/list', {
    executorId: params.executorId ? toApiId(params.executorId) : undefined,
  })
  return {
    items: data.items.map(mapJob),
    total: data.total,
  }
}

export async function createJob(form: JobForm): Promise<JobInfo> {
  const data = await postData<JobData, ReturnType<typeof toApiJobForm>>('/v1/admin/jobs/create', toApiJobForm(form))
  return mapJob(data.job)
}

export async function updateJob(form: JobForm): Promise<JobInfo> {
  if (!form.id) {
    throw new Error('缺少任务 ID')
  }
  const data = await postData<JobData, ReturnType<typeof toApiJobForm>>('/v1/admin/jobs/update', toApiJobForm(form))
  return mapJob(data.job)
}

export async function deleteJob(id: string): Promise<void> {
  await postData<{ id: number }, { id: number }>('/v1/admin/jobs/delete', { id: toApiId(id) })
}

export async function startJob(id: string): Promise<JobInfo> {
  const data = await postData<JobData, { id: number }>('/v1/admin/jobs/start', { id: toApiId(id) })
  return mapJob(data.job)
}

export async function stopJob(id: string): Promise<JobInfo> {
  const data = await postData<JobData, { id: number }>('/v1/admin/jobs/stop', { id: toApiId(id) })
  return mapJob(data.job)
}

export async function runJob(id: string): Promise<RunJobResult> {
  const data = await postData<RunJobData, { id: number }>('/v1/admin/jobs/run', { id: toApiId(id) })
  return mapRunJob(data)
}

export async function killJob(id: string): Promise<RunJobResult> {
  const data = await postData<RunJobData, { id: number }>('/v1/admin/jobs/kill', { id: toApiId(id) })
  return mapRunJob(data)
}
