import { deleteData, getData, postData, putData } from './request'
import type { AlertSettings, SaveFeishuWebhookPayload } from '@/types/systemSettings'

interface AlertSettingsPayload {
  feishuWebhookConfigured?: boolean
  feishu_webhook_configured?: boolean
  feishuWebhookUpdatedAt?: string
  feishu_webhook_updated_at?: string
}

interface AlertSettingsData {
  settings?: AlertSettingsPayload
}

function mapAlertSettings(raw?: AlertSettingsPayload): AlertSettings {
  return {
    feishuWebhookConfigured: Boolean(raw?.feishuWebhookConfigured ?? raw?.feishu_webhook_configured),
    feishuWebhookUpdatedAt: raw?.feishuWebhookUpdatedAt ?? raw?.feishu_webhook_updated_at ?? '',
  }
}

export async function getAlertSettings(): Promise<AlertSettings> {
  const data = await getData<AlertSettingsData>('/v1/admin/system/settings/alert')
  return mapAlertSettings(data.settings)
}

export async function saveFeishuWebhook(payload: SaveFeishuWebhookPayload): Promise<AlertSettings> {
  const data = await putData<AlertSettingsData, { webhook: string }>('/v1/admin/system/settings/alert/feishu', {
    webhook: payload.webhook,
  })
  return mapAlertSettings(data.settings)
}

export async function testFeishuWebhook(): Promise<void> {
  await postData<{ status: string }, Record<string, never>>('/v1/admin/system/settings/alert/feishu/test', {})
}

export async function clearFeishuWebhook(): Promise<AlertSettings> {
  const data = await deleteData<AlertSettingsData>('/v1/admin/system/settings/alert/feishu')
  return mapAlertSettings(data.settings)
}
