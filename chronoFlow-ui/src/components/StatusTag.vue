<script setup lang="ts">
import { statusColor, statusText } from '@/utils/status'

const props = defineProps<{
  status: string
}>()
</script>

<template>
  <a-tag :color="statusColor(props.status)">{{ statusText(props.status) }}</a-tag>
</template>
