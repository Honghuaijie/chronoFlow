import dayjs from 'dayjs'

export function formatDateTime(value: string): string {
  if (!value) {
    return '-'
  }
  const parsed = dayjs(value)
  return parsed.isValid() ? parsed.format('YYYY-MM-DD HH:mm:ss') : value
}

export function formatDuration(ms: number): string {
  if (!ms || ms < 0) {
    return '-'
  }
  if (ms < 1000) {
    return `${ms} ms`
  }
  const seconds = ms / 1000
  if (seconds < 60) {
    return `${seconds.toFixed(1)} s`
  }
  return `${Math.floor(seconds / 60)}m ${Math.floor(seconds % 60)}s`
}

export function formatBytes(bytes: number): string {
  if (!bytes || bytes < 0) {
    return '0 B'
  }
  if (bytes < 1024) {
    return `${bytes} B`
  }
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`
  }
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}
