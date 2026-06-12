<script setup lang="ts">
const themeConfig = {
  token: {
    colorPrimary: '#1E40AF',
    colorInfo: '#3B82F6',
    colorSuccess: '#16A34A',
    colorWarning: '#D97706',
    colorError: '#DC2626',
    colorBgLayout: '#F8FAFC',
    colorBorder: '#D9E2F2',
    borderRadius: 6,
    fontSize: 14,
  },
}
</script>

<template>
  <a-config-provider :theme="themeConfig">
    <router-view />
  </a-config-provider>
</template>
