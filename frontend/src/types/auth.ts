export interface User {
  id: string
  username: string
}

export interface LoginResponse {
  access_token: string
  token_type: string
  expires_in: number
  user: User
}
