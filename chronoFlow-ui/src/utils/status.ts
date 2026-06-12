import type { TagProps } from 'ant-design-vue'

export function statusColor(status: string): TagProps['color'] {
  const map: Record<string, TagProps['color']> = {
    online: 'success',
    offline: 'error',
    running: 'processing',
    killing: 'warning',
    stopped: 'default',
    success: 'success',
    failed: 'error',
    killed: 'warning',
    skipped: 'default',
  }
  return map[status] || 'default'
}

export function statusText(status: string): string {
  const map: Record<string, string> = {
    online: '在线',
    offline: '离线',
    running: '运行中',
    killing: '终止中',
    stopped: '已停止',
    success: '成功',
    failed: '失败',
    killed: '已终止',
    skipped: '已跳过',
    manual: '手动',
    cron: '定时',
  }
  return map[status] || status || '-'
}

export function isActiveLogStatus(status: string): boolean {
  return status === 'running' || status === 'killing'
}
