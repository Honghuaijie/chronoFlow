<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Modal, message } from 'ant-design-vue'
import { clearFeishuWebhook, getAlertSettings, saveFeishuWebhook, testFeishuWebhook } from '@/api/systemSettings'
import PageHeaderBar from '@/components/PageHeaderBar.vue'
import type { AlertSettings } from '@/types/systemSettings'

const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const clearing = ref(false)
const settings = ref<AlertSettings>({
  feishuWebhookConfigured: false,
  feishuWebhookUpdatedAt: '',
})
const form = reactive({
  webhook: '',
})

onMounted(() => {
  void loadSettings()
})

async function loadSettings() {
  loading.value = true
  try {
    settings.value = await getAlertSettings()
  } finally {
    loading.value = false
  }
}

async function saveWebhook() {
  saving.value = true
  try {
    settings.value = await saveFeishuWebhook({ webhook: form.webhook })
    form.webhook = ''
    message.success('飞书 Webhook 已保存')
  } finally {
    saving.value = false
  }
}

async function sendTest() {
  testing.value = true
  try {
    await testFeishuWebhook()
    message.success('测试告警已发送')
  } finally {
    testing.value = false
  }
}

function confirmClear() {
  Modal.confirm({
    title: '清空飞书 Webhook',
    content: '清空后，已开启失败告警的任务也不会发送飞书消息。',
    okText: '清空',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      clearing.value = true
      try {
        settings.value = await clearFeishuWebhook()
        form.webhook = ''
        message.success('飞书 Webhook 已清空')
      } finally {
        clearing.value = false
      }
    },
  })
}
</script>

<template>
  <div class="page-body">
    <PageHeaderBar title="系统设置" description="配置调度中心的全局运行选项和告警渠道。">
      <a-button @click="loadSettings">刷新</a-button>
    </PageHeaderBar>

    <a-spin :spinning="loading">
      <a-card title="飞书失败告警" class="settings-card">
        <a-alert
          type="info"
          show-icon
          message="Webhook 不会明文回显"
          description="保存后页面只显示已配置或未配置。请你自己妥善保存飞书机器人 Webhook；需要更换时直接粘贴新的 Webhook 覆盖保存。V1 不支持飞书签名 Secret。"
        />

        <a-alert
          class="settings-tip"
          type="warning"
          show-icon
          message="飞书关键词校验"
          description="如果你的飞书机器人开启了关键词校验，请在飞书机器人安全设置中把关键词配置为 ChronoFlow。测试发送或任务告警返回 Key Words Not Found 时，通常就是关键词没有配置或不匹配。"
        />

        <a-descriptions class="settings-status" bordered size="small" :column="2">
          <a-descriptions-item label="配置状态">
            <a-tag v-if="settings.feishuWebhookConfigured" color="green">已配置</a-tag>
            <a-tag v-else color="orange">未配置</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="更新时间">
            {{ settings.feishuWebhookUpdatedAt || '-' }}
          </a-descriptions-item>
        </a-descriptions>

        <a-form layout="vertical" :model="form" class="settings-form">
          <a-form-item label="飞书 Webhook">
            <a-input-password
              v-model:value="form.webhook"
              placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/..."
              autocomplete="off"
            />
            <div class="form-help">
              任务日志最终状态为 failed 或 timeout，且任务开启失败告警时，会向该 Webhook 发送卡片消息。
            </div>
          </a-form-item>

          <a-space wrap>
            <a-button type="primary" :loading="saving" :disabled="!form.webhook.trim()" @click="saveWebhook">保存</a-button>
            <a-button :loading="testing" :disabled="!settings.feishuWebhookConfigured" @click="sendTest">测试发送</a-button>
            <a-button danger :loading="clearing" :disabled="!settings.feishuWebhookConfigured" @click="confirmClear">
              清空配置
            </a-button>
          </a-space>
        </a-form>
      </a-card>

      <a-alert
        type="warning"
        show-icon
        message="失败判断说明"
        description="ChronoFlow 根据进程退出码判断任务是否失败，不解析日志正文。Glue Shell 调用 Python 脚本时，建议使用 set -euo pipefail，确保 Python 报错会让 Shell 返回非 0 退出码。"
      />
    </a-spin>
  </div>
</template>

<style scoped>
.settings-card {
  margin-bottom: 16px;
}

.settings-tip {
  margin-top: 12px;
}

.settings-status,
.settings-form {
  margin-top: 16px;
}

.form-help {
  margin-top: 6px;
  color: #64748b;
  font-size: 12px;
  line-height: 1.5;
}
</style>
