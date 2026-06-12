import { getData, postData } from './request'
import type { CurrentUser, LoginParams, LoginResult } from '@/types/auth'

interface CurrentUserPayload {
  userId?: number
  user_id?: number
  username: string
  role: string
}

export async function login(params: LoginParams): Promise<LoginResult> {
  return postData<LoginResult, LoginParams>('/v1/public/auth/login', params)
}

export async function currentUser(): Promise<CurrentUser> {
  const data = await getData<CurrentUserPayload>('/v1/admin/auth/current')
  return {
    userId: data.userId ?? data.user_id ?? 0,
    username: data.username,
    role: data.role,
  }
}
