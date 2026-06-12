<script setup lang="ts">
defineProps<{
  content: string
  loading?: boolean
}>()
</script>

<template>
  <a-skeleton v-if="loading" active :paragraph="{ rows: 10 }" />
  <pre v-else class="code-box mono">{{ content || '暂无日志内容' }}</pre>
</template>
