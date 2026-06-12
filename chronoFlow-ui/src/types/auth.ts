export interface LoginParams {
  username: string
  password: string
}

export interface LoginResult {
  token: string
  username: string
}

export interface CurrentUser {
  userId: number
  username: string
  role: string
}
