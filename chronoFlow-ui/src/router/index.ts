import { createRouter, createWebHistory } from 'vue-router'
import { getStoredToken } from '@/api/request'
import AdminLayout from '@/layouts/AdminLayout.vue'
import LoginView from '@/views/LoginView.vue'
import ExecutorListView from '@/views/executors/ExecutorListView.vue'
import JobListView from '@/views/jobs/JobListView.vue'
import JobLogListView from '@/views/logs/JobLogListView.vue'
import JobLogDetailView from '@/views/logs/JobLogDetailView.vue'
import SettingsView from '@/views/settings/SettingsView.vue'
import HelpView from '@/views/help/HelpView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: LoginView,
      meta: { public: true, title: '登录' },
    },
    {
      path: '/',
      component: AdminLayout,
      children: [
        {
          path: '',
          redirect: '/jobs',
        },
        {
          path: 'executors',
          name: 'executors',
          component: ExecutorListView,
          meta: { title: '执行器' },
        },
        {
          path: 'jobs',
          name: 'jobs',
          component: JobListView,
          meta: { title: '任务' },
        },
        {
          path: 'logs',
          name: 'logs',
          component: JobLogListView,
          meta: { title: '执行日志' },
        },
        {
          path: 'logs/:id',
          name: 'logDetail',
          component: JobLogDetailView,
          meta: { title: '日志详情' },
        },
        {
          path: 'settings',
          name: 'settings',
          component: SettingsView,
          meta: { title: '设置' },
        },
        {
          path: 'help',
          name: 'help',
          component: HelpView,
          meta: { title: '使用说明' },
        },
      ],
    },
  ],
})

router.beforeEach((to) => {
  if (!to.meta.public && !getStoredToken()) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (to.path === '/login' && getStoredToken()) {
    return '/jobs'
  }
  return true
})

export default router
