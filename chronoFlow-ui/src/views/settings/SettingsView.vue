<script setup lang="ts">
import PageHeaderBar from '@/components/PageHeaderBar.vue'
</script>

<template>
  <div class="page-body">
    <PageHeaderBar title="设置" description="V1 配置由后端配置文件和环境变量管理，页面仅展示约定。" />
    <a-alert
      type="info"
      show-icon
      message="配置入口"
      description="调度时区、callback_token、日志目录、保留天数、执行器健康检查和启动恢复策略暂不在前端修改。环境变量优先覆盖配置文件。"
    />
  </div>
</template>
