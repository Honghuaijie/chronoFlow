import type { Id } from '@/types/api'

export function toApiId(id: Id): number {
  return Number(id)
}

export function normalizeId(value: string | number | undefined | null): Id {
  if (value === undefined || value === null) {
    return ''
  }
  return String(value)
}
