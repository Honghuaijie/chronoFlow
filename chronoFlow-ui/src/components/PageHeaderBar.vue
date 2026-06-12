<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <div class="page-header-bar">
    <div>
      <h1>{{ title }}</h1>
      <p v-if="description">{{ description }}</p>
    </div>
    <div class="page-header-actions">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.page-header-bar {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  justify-content: space-between;
}

h1 {
  margin: 0;
  color: #172033;
  font-size: 20px;
  font-weight: 650;
  line-height: 1.3;
}

p {
  margin: 6px 0 0;
  color: #64748b;
  line-height: 1.5;
}

.page-header-actions {
  display: flex;
  flex-shrink: 0;
  gap: 8px;
  align-items: center;
}

@media (max-width: 768px) {
  .page-header-bar {
    flex-direction: column;
  }
}
</style>
