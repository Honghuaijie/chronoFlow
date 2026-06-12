export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
}

export interface PageResult<T> {
  items: T[]
  total: number
}

export interface PaginationState {
  page: number
  pageSize: number
  total: number
}

export type Id = string
