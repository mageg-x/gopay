import request from './request'

// 管理员登录
export function adminLogin(data: { username: string; password: string }) {
  return request.post('/admin/login', data)
}

// 管理员登出
export function adminLogout() {
  return request.post('/admin/logout')
}

// 获取统计数据
export function getAdminStats() {
  return request.get('/admin/stats')
}

// 获取商户列表
export function getUserList(params: { page?: number; limit?: number }) {
  return request.get('/admin/users', { params })
}

// 获取订单列表
export function getOrderList(params: { page?: number; limit?: number; status?: number }) {
  return request.get('/admin/orders', { params })
}

// 订单操作
export function orderOp(data: { action: string; trade_no: string; [key: string]: any }) {
  return request.post('/admin/order/op', data)
}

// 获取结算列表
export function getSettleList(params: { page?: number; limit?: number }) {
  return request.get('/admin/settles', { params })
}

// 结算操作
export function settleOp(data: { action: string; id: number; [key: string]: any }) {
  return request.post('/admin/settle/op', data)
}

// 获取转账列表
export function getTransferList(params: { page?: number; limit?: number }) {
  return request.get('/admin/transfer', { params })
}

// 获取通道列表
export function getChannelList() {
  return request.get('/admin/channel')
}

// 获取插件列表
export function getPluginList() {
  return request.get('/admin/plugin')
}

// 获取系统配置
export function getConfig(mod?: string) {
  return request.get('/admin/config', { params: { mod } })
}

// 保存系统配置
export function saveConfig(data: Record<string, string>) {
  return request.post('/admin/config', data)
}
