import { getData, postData } from './request'
import type { GlueInfo } from '@/types/glue'
import { normalizeId, toApiId } from '@/utils/id'

interface GluePayload extends Omit<GlueInfo, 'id' | 'jobId'> {
  id: string | number
  jobId?: string | number
  job_id?: string | number
}

interface GlueData {
  glue: GluePayload
}

function mapGlue(payload: GluePayload): GlueInfo {
  return {
    ...payload,
    id: normalizeId(payload.id),
    jobId: normalizeId(payload.jobId ?? payload.job_id),
  }
}

export async function getGlue(jobId: string): Promise<GlueInfo | null> {
  const data = await getData<Partial<GlueData>>('/v1/admin/glues/get', { jobId: toApiId(jobId) })
  return data.glue ? mapGlue(data.glue) : null
}

export async function saveGlue(jobId: string, content: string): Promise<GlueInfo> {
  const data = await postData<GlueData, { jobId: number; content: string }>('/v1/admin/glues/save', {
    jobId: toApiId(jobId),
    content,
  })
  return mapGlue(data.glue)
}
