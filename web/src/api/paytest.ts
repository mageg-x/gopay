import request from './request'

export interface TestNotifyEvent {
  received_at: number
  trade_no: string
  out_trade_no: string
  status: string
  money: string
  sign_type: string
  sign_valid: boolean
  verify_reason: string
  form: Record<string, string>
}

export interface TestNotifySession {
  token: string
  notify_url: string
  return_url: string
  created_at: number
  expires_at: number
  hit_count: number
  last_event?: TestNotifyEvent | null
}

// 测试回调：创建会话
export function createTestNotifySession() {
  return request.post<TestNotifySession>('/pay/test_notify_session')
}

// 测试回调：查询会话
export function getTestNotifySession(token: string) {
  return request.get<TestNotifySession>(`/pay/test_notify_session/${encodeURIComponent(token)}`)
}
