import request from './request'

// 商户登录
export function userLogin(data: { uid: number; password: string } | { uid: number; key: string }) {
  return request.post('/user/login', data)
}

// 商户注册
export function userRegister(data: {
  email: string
  phone?: string
  password: string
  invite_code?: string
}) {
  return request.post('/user/reg', data)
}

// 登出
export function userLogout() {
  return request.post('/user/logout')
}

// 获取用户信息
export function getUserInfo() {
  return request.get('/user/index')
}

// 获取订单列表
export function getUserOrders(params: { page?: number; limit?: number; status?: number }) {
  return request.get('/user/orders', { params })
}

// 获取结算列表
export function getUserSettles(params: { page?: number; limit?: number }) {
  return request.get('/user/settles', { params })
}

// 申请结算
export function applySettle(data: {
  account: string
  username: string
  money: number
  type: number
}) {
  return request.post('/user/settle/apply', data)
}

// 获取资金记录
export function getUserRecords(params: { page?: number; limit?: number; action?: number }) {
  return request.get('/user/records', { params })
}

// 更新资料
export function updateProfile(data: { username?: string; phone?: string; qq?: string }) {
  return request.post('/user/editinfo', data)
}

// 实名认证
export function submitCertificate(data: { certname: string; certno: string; certtype: number }) {
  return request.post('/user/certificate', data)
}
