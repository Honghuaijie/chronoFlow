export interface AlertSettings {
  feishuWebhookConfigured: boolean
  feishuWebhookUpdatedAt: string
}

export interface SaveFeishuWebhookPayload {
  webhook: string
}
