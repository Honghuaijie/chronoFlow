<script setup lang="ts">
defineProps<{
  active: boolean
}>()
</script>

<template>
  <span class="polling-indicator" :class="{ active }">
    <span class="dot" />
    {{ active ? '自动刷新中' : '未轮询' }}
  </span>
</template>

<style scoped>
.polling-indicator {
  display: inline-flex;
  gap: 6px;
  align-items: center;
  color: #64748b;
  font-size: 13px;
}

.dot {
  width: 7px;
  height: 7px;
  background: #94a3b8;
  border-radius: 999px;
}

.active .dot {
  background: #16a34a;
}
</style>
