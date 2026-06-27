<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { EditOutlined } from '@ant-design/icons-vue'
import { describeCron, formatNextRunTimes } from '@/utils/cron'

const props = defineProps<{
  value: string
}>()

const emit = defineEmits<{
  'update:value': [value: string]
}>()

const open = ref(false)
const activeTab = ref('minute')
const customExpr = ref(props.value || '0 */5 * * * *')
const config = reactive({
  minuteInterval: 5,
  hourInterval: 1,
  hour: 0,
  minute: 0,
  second: 0,
  weekday: 1,
  day: 1,
})

const weekdayOptions = [
  { label: '周日', value: 0 },
  { label: '周一', value: 1 },
  { label: '周二', value: 2 },
  { label: '周三', value: 3 },
  { label: '周四', value: 4 },
  { label: '周五', value: 5 },
  { label: '周六', value: 6 },
]

const expression = computed(() => {
  if (activeTab.value === 'minute') {
    return `0 */${config.minuteInterval} * * * *`
  }
  if (activeTab.value === 'hour') {
    return `0 0 */${config.hourInterval} * * *`
  }
  if (activeTab.value === 'day') {
    return `${config.second} ${config.minute} ${config.hour} * * *`
  }
  if (activeTab.value === 'week') {
    return `${config.second} ${config.minute} ${config.hour} * * ${config.weekday}`
  }
  if (activeTab.value === 'month') {
    return `${config.second} ${config.minute} ${config.hour} ${config.day} * *`
  }
  return customExpr.value.trim()
})

const summary = computed(() => describeCron(expression.value))
const nextRuns = computed(() => formatNextRunTimes(expression.value, 5))

watch(
  () => props.value,
  (value) => {
    customExpr.value = value || '0 */5 * * * *'
  },
)

function openPicker() {
  customExpr.value = props.value || '0 */5 * * * *'
  open.value = true
}

function updateValue(value: string) {
  emit('update:value', value)
}

function applyExpression() {
  emit('update:value', expression.value)
  open.value = false
}
</script>

<template>
  <div class="cron-picker">
    <a-input :value="value" class="mono" placeholder="0 */5 * * * *" @update:value="updateValue">
      <template #addonAfter>
        <a-tooltip title="配置 Cron">
          <a-button type="text" size="small" class="cron-edit-button" @click="openPicker">
            <template #icon><EditOutlined /></template>
          </a-button>
        </a-tooltip>
      </template>
    </a-input>

    <a-modal v-model:open="open" title="Cron 配置" width="760px" @ok="applyExpression">
      <a-tabs v-model:active-key="activeTab">
        <a-tab-pane key="minute" tab="分钟">
          <a-form layout="vertical">
            <a-form-item label="间隔分钟">
              <a-input-number v-model:value="config.minuteInterval" :min="1" :max="59" style="width: 180px" />
            </a-form-item>
          </a-form>
        </a-tab-pane>
        <a-tab-pane key="hour" tab="小时">
          <a-form layout="vertical">
            <a-form-item label="间隔小时">
              <a-input-number v-model:value="config.hourInterval" :min="1" :max="23" style="width: 180px" />
            </a-form-item>
          </a-form>
        </a-tab-pane>
        <a-tab-pane key="day" tab="日">
          <div class="time-grid">
            <a-form-item label="时"><a-input-number v-model:value="config.hour" :min="0" :max="23" /></a-form-item>
            <a-form-item label="分"><a-input-number v-model:value="config.minute" :min="0" :max="59" /></a-form-item>
            <a-form-item label="秒"><a-input-number v-model:value="config.second" :min="0" :max="59" /></a-form-item>
          </div>
        </a-tab-pane>
        <a-tab-pane key="week" tab="周">
          <a-form layout="vertical">
            <a-form-item label="星期">
              <a-select v-model:value="config.weekday" :options="weekdayOptions" style="width: 180px" />
            </a-form-item>
            <div class="time-grid">
              <a-form-item label="时"><a-input-number v-model:value="config.hour" :min="0" :max="23" /></a-form-item>
              <a-form-item label="分"><a-input-number v-model:value="config.minute" :min="0" :max="59" /></a-form-item>
              <a-form-item label="秒"><a-input-number v-model:value="config.second" :min="0" :max="59" /></a-form-item>
            </div>
          </a-form>
        </a-tab-pane>
        <a-tab-pane key="month" tab="月">
          <a-form layout="vertical">
            <a-form-item label="日期">
              <a-input-number v-model:value="config.day" :min="1" :max="31" style="width: 180px" />
            </a-form-item>
            <div class="time-grid">
              <a-form-item label="时"><a-input-number v-model:value="config.hour" :min="0" :max="23" /></a-form-item>
              <a-form-item label="分"><a-input-number v-model:value="config.minute" :min="0" :max="59" /></a-form-item>
              <a-form-item label="秒"><a-input-number v-model:value="config.second" :min="0" :max="59" /></a-form-item>
            </div>
          </a-form>
        </a-tab-pane>
        <a-tab-pane key="custom" tab="手动">
          <a-form layout="vertical">
            <a-form-item label="Cron 表达式">
              <a-input v-model:value="customExpr" class="mono" />
            </a-form-item>
          </a-form>
        </a-tab-pane>
      </a-tabs>

      <div class="cron-preview">
        <div>
          <span class="preview-label">表达式</span>
          <span class="mono">{{ expression }}</span>
        </div>
        <div>
          <span class="preview-label">说明</span>
          <span>{{ summary }}</span>
        </div>
        <div>
          <span class="preview-label">下次运行</span>
          <ol v-if="nextRuns.length" class="next-run-list">
            <li v-for="item in nextRuns" :key="item">{{ item }}</li>
          </ol>
          <span v-else>无法计算</span>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<style scoped>
.cron-picker {
  width: 100%;
}

.cron-edit-button {
  width: 28px;
  height: 24px;
}

.time-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 120px));
  gap: 12px;
}

.cron-preview {
  display: grid;
  gap: 8px;
  padding: 12px;
  margin-top: 12px;
  background: #f8fafc;
  border: 1px solid #d9e2f2;
  border-radius: 6px;
}

.preview-label {
  display: inline-block;
  width: 72px;
  color: #64748b;
}

.next-run-list {
  display: inline-grid;
  gap: 4px;
  padding-left: 20px;
  margin: 0;
  vertical-align: top;
}

@media (max-width: 640px) {
  .time-grid {
    grid-template-columns: 1fr;
  }
}
</style>
