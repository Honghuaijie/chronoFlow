<script setup lang="ts">
import { reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { LockOutlined, UserOutlined } from '@ant-design/icons-vue'
import { useAuthStore } from '@/stores/auth'
import type { LoginParams } from '@/types/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const form = reactive<LoginParams>({
  username: 'admin',
  password: '',
})

async function submit() {
  await authStore.login(form)
  const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/jobs'
  await router.push(redirect)
}
</script>

<template>
  <main class="login-page">
    <section class="login-panel">
      <div class="login-brand">
        <div class="brand-mark">CF</div>
        <div>
          <h1>ChronoFlow</h1>
          <p>内网定时任务调度中心</p>
        </div>
      </div>
      <a-form layout="vertical" :model="form" @finish="submit">
        <a-form-item name="username" label="用户名" :rules="[{ required: true, message: '请输入用户名' }]">
          <a-input v-model:value="form.username" autocomplete="username">
            <template #prefix>
              <UserOutlined />
            </template>
          </a-input>
        </a-form-item>
        <a-form-item name="password" label="密码" :rules="[{ required: true, message: '请输入密码' }]">
          <a-input-password v-model:value="form.password" autocomplete="current-password">
            <template #prefix>
              <LockOutlined />
            </template>
          </a-input-password>
        </a-form-item>
        <a-button type="primary" html-type="submit" block :loading="authStore.loading">登录</a-button>
      </a-form>
    </section>
  </main>
</template>

<style scoped>
.login-page {
  display: grid;
  min-height: 100vh;
  padding: 24px;
  background: #f8fafc;
  place-items: center;
}

.login-panel {
  width: min(100%, 380px);
  padding: 28px;
  background: #fff;
  border: 1px solid #d9e2f2;
  border-radius: 8px;
  box-shadow: 0 16px 40px rgb(15 23 42 / 8%);
}

.login-brand {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 24px;
}

.brand-mark {
  display: grid;
  width: 42px;
  height: 42px;
  color: #fff;
  font-weight: 700;
  background: #1e40af;
  border-radius: 8px;
  place-items: center;
}

h1 {
  margin: 0;
  font-size: 22px;
  line-height: 1.2;
}

p {
  margin: 4px 0 0;
  color: #64748b;
}
</style>
