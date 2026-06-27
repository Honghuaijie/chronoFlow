<script setup lang="ts">
import { computed, h, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  BarChartOutlined,
  ClockCircleOutlined,
  DatabaseOutlined,
  FileSearchOutlined,
  LogoutOutlined,
  QuestionCircleOutlined,
  SettingOutlined,
} from '@ant-design/icons-vue'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const selectedKeys = computed(() => [route.path.startsWith('/logs') ? '/logs' : route.path])

onMounted(() => {
  void authStore.fetchCurrent()
})

function logout() {
  authStore.logout()
  void router.push('/login')
}
</script>

<template>
  <a-layout class="admin-layout">
    <a-layout-sider class="admin-sider" :width="216" breakpoint="lg" collapsed-width="0">
      <div class="brand">
        <div class="brand-mark">CF</div>
        <div>
          <div class="brand-title">ChronoFlow</div>
          <div class="brand-subtitle">调度中心</div>
        </div>
      </div>
      <a-menu mode="inline" :selected-keys="selectedKeys">
        <a-menu-item key="/executors" @click="router.push('/executors')">
          <DatabaseOutlined />
          <span>执行器</span>
        </a-menu-item>
        <a-menu-item key="/jobs" @click="router.push('/jobs')">
          <ClockCircleOutlined />
          <span>任务</span>
        </a-menu-item>
        <a-menu-item key="/reports" @click="router.push('/reports')">
          <BarChartOutlined />
          <span>运行报表</span>
        </a-menu-item>
        <a-menu-item key="/logs" @click="router.push('/logs')">
          <FileSearchOutlined />
          <span>执行日志</span>
        </a-menu-item>
        <a-menu-item key="/settings" @click="router.push('/settings')">
          <SettingOutlined />
          <span>设置</span>
        </a-menu-item>
        <a-menu-item key="/help" @click="router.push('/help')">
          <QuestionCircleOutlined />
          <span>使用说明</span>
        </a-menu-item>
      </a-menu>
    </a-layout-sider>
    <a-layout>
      <a-layout-header class="admin-header">
        <div class="header-title">{{ route.meta.title || '调度中心' }}</div>
        <a-space>
          <span class="muted-text">{{ authStore.user?.username || 'admin' }}</span>
          <a-tooltip title="退出登录">
            <a-button aria-label="退出登录" :icon="h(LogoutOutlined)" @click="logout" />
          </a-tooltip>
        </a-space>
      </a-layout-header>
      <a-layout-content class="admin-content">
        <router-view />
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<style scoped>
.admin-layout {
  min-height: 100vh;
}

.admin-sider {
  background: #fff;
  border-right: 1px solid #d9e2f2;
}

.brand {
  display: flex;
  gap: 10px;
  align-items: center;
  height: 56px;
  padding: 0 16px;
  border-bottom: 1px solid #e5ecf6;
}

.brand-mark {
  display: grid;
  width: 32px;
  height: 32px;
  color: #fff;
  font-weight: 700;
  background: #1e40af;
  border-radius: 6px;
  place-items: center;
}

.brand-title {
  color: #172033;
  font-weight: 700;
  line-height: 1.1;
}

.brand-subtitle {
  margin-top: 2px;
  color: #64748b;
  font-size: 12px;
}

.admin-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 24px;
  background: #fff;
  border-bottom: 1px solid #d9e2f2;
}

.header-title {
  color: #172033;
  font-weight: 650;
}

.admin-content {
  min-width: 0;
  padding: 24px;
  background: #f8fafc;
}

@media (max-width: 768px) {
  .admin-header,
  .admin-content {
    padding-right: 16px;
    padding-left: 16px;
  }
}
</style>
