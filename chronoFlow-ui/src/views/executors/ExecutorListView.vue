<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Modal } from 'ant-design-vue'
import PageHeaderBar from '@/components/PageHeaderBar.vue'
import StatusTag from '@/components/StatusTag.vue'
import { useExecutorsStore } from '@/stores/executors'
import type { ExecutorForm, ExecutorInfo } from '@/types/executor'
import { formatDateTime } from '@/utils/datetime'

const store = useExecutorsStore()
const modalOpen = ref(false)
const editingId = ref('')

const form = reactive<ExecutorForm>({
  name: '',
  address: '',
  token: '',
  description: '',
})

const modalTitle = computed(() => (editingId.value ? '编辑执行器' : '新增执行器'))

onMounted(() => {
  void store.fetchList()
})

function resetForm() {
  editingId.value = ''
  form.id = undefined
  form.name = ''
  form.address = ''
  form.token = ''
  form.description = ''
}

function openCreate() {
  resetForm()
  modalOpen.value = true
}

function openEdit(row: ExecutorInfo) {
  editingId.value = row.id
  form.id = row.id
  form.name = row.name
  form.address = row.address
  form.token = ''
  form.description = row.description
  modalOpen.value = true
}

async function submit() {
  await store.save({ ...form })
  modalOpen.value = false
  resetForm()
}

function confirmDelete(row: ExecutorInfo) {
  Modal.confirm({
    title: '删除执行器',
    content: `确认删除「${row.name}」？关联任务请先自行处理。`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: () => store.remove(row.id),
  })
}
</script>

<template>
  <div class="page-body">
    <PageHeaderBar title="执行器" description="管理可接收调度请求的执行器节点。">
      <a-button type="primary" @click="openCreate">新增执行器</a-button>
    </PageHeaderBar>

    <div class="table-shell">
      <a-table
        row-key="id"
        :data-source="store.items"
        :loading="store.loading"
        :pagination="false"
        size="middle"
        :scroll="{ x: 980 }"
      >
        <a-table-column title="名称" data-index="name" :width="180" fixed="left" />
        <a-table-column title="地址" data-index="address" :width="260">
          <template #default="{ text }">
            <span class="mono">{{ text }}</span>
          </template>
        </a-table-column>
        <a-table-column title="状态" data-index="status" :width="110">
          <template #default="{ text }">
            <StatusTag :status="text" />
          </template>
        </a-table-column>
        <a-table-column title="心跳失败" data-index="heartbeatFailCount" :width="100" />
        <a-table-column title="最近心跳" data-index="lastHeartbeatTime" :width="180">
          <template #default="{ text }">{{ formatDateTime(text) }}</template>
        </a-table-column>
        <a-table-column title="说明" data-index="description" />
        <a-table-column title="操作" :width="150" fixed="right">
          <template #default="{ record }">
            <a-space>
              <a-button type="link" size="small" @click="openEdit(record as ExecutorInfo)">编辑</a-button>
              <a-button danger type="link" size="small" @click="confirmDelete(record as ExecutorInfo)">删除</a-button>
            </a-space>
          </template>
        </a-table-column>
      </a-table>
    </div>

    <a-modal v-model:open="modalOpen" :title="modalTitle" :confirm-loading="store.submitting" @ok="submit">
      <a-form layout="vertical" :model="form">
        <a-form-item label="名称" required>
          <a-input v-model:value="form.name" placeholder="如：exec-local" />
        </a-form-item>
        <a-form-item label="地址" required>
          <a-input v-model:value="form.address" placeholder="http://127.0.0.1:9000" />
        </a-form-item>
        <a-form-item :label="editingId ? 'Token（留空表示不修改）' : 'Token'" :required="!editingId">
          <a-input-password v-model:value="form.token" placeholder="执行器访问 token" />
        </a-form-item>
        <a-form-item label="说明">
          <a-textarea v-model:value="form.description" :rows="3" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>
