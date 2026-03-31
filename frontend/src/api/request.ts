import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'

// 创建axios实例
const request: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    // 添加admin token
    const adminToken = localStorage.getItem('admin_token')
    if (adminToken) {
      config.headers['Admin-Token'] = adminToken
    }

    // 添加user token
    const userToken = localStorage.getItem('user_token')
    if (userToken) {
      config.headers['User-Token'] = userToken
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse) => {
    const res = response.data

    if (res.code !== 0 && res.code !== undefined) {
      ElMessage.error(res.msg || '请求失败')
      return Promise.reject(new Error(res.msg || '请求失败'))
    }

    return res
  },
  (error) => {
    ElMessage.error(error.message || '网络错误')
    return Promise.reject(error)
  }
)

export default request

// API方法类型
export interface ApiResponse<T = any> {
  code: number
  msg: string
  data: T
}

// 分页参数
export interface PageParams {
  page?: number
  limit?: number
}

// 分页响应
export interface PageResponse<T> {
  list: T[]
  total: number
  page: number
  limit: number
}
