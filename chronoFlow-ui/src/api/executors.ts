import { getData, postData } from './request'
import type { PageResult } from '@/types/api'
import type { ExecutorForm, ExecutorInfo } from '@/types/executor'
import { normalizeId, toApiId } from '@/utils/id'

interface ExecutorPayload extends Omit<ExecutorInfo, 'id'> {
  id: string | number
}

interface ExecutorData {
  executor: ExecutorPayload
}

interface ExecutorListData {
  items: ExecutorPayload[]
  total: number
}

function mapExecutor(payload: ExecutorPayload): ExecutorInfo {
  return {
    ...payload,
    id: normalizeId(payload.id),
  }
}

export async function listExecutors(): Promise<PageResult<ExecutorInfo>> {
  const data = await getData<ExecutorListData>('/v1/admin/executors/list')
  return {
    items: data.items.map(mapExecutor),
    total: data.total,
  }
}

export async function createExecutor(form: ExecutorForm): Promise<ExecutorInfo> {
  const data = await postData<ExecutorData, ExecutorForm>('/v1/admin/executors/create', form)
  return mapExecutor(data.executor)
}

export async function updateExecutor(form: ExecutorForm): Promise<ExecutorInfo> {
  if (!form.id) {
    throw new Error('缺少执行器 ID')
  }
  const payload = {
    ...form,
    id: toApiId(form.id),
  }
  const data = await postData<ExecutorData, typeof payload>('/v1/admin/executors/update', payload)
  return mapExecutor(data.executor)
}

export async function deleteExecutor(id: string): Promise<void> {
  await postData<{ id: number }, { id: number }>('/v1/admin/executors/delete', { id: toApiId(id) })
}
