import axios, { type AxiosInstance, type AxiosResponse, type AxiosRequestConfig } from 'axios'
import router from '@/router'

// API方法类型
export interface ApiResponse<T = any> {
  code: number
  msg: string
  data: T
  token?: string
  count?: number
  user?: any
  uid?: number
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

// 自定义请求接口，返回 ApiResponse 而不是 AxiosResponse
interface RequestInstance {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>>
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>>
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>>
  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>>
}

function extractRequestId(raw: string): string {
  const text = String(raw || '')
  const m = text.match(/Request-Id=\[([^\]]+)\]/i)
  return m?.[1]?.trim() || ''
}

function normalizeErrorMessage(raw: string): string {
  const text = String(raw || '').trim()
  if (!text) return '请求失败'

  if (text.includes('APPID_MCHID_NOT_MATCH')) {
    const rid = extractRequestId(text)
    return rid
      ? `微信支付配置不匹配（appid/mchid），请检查绑定关系。Request-Id: ${rid}`
      : '微信支付配置不匹配（appid/mchid），请检查绑定关系'
  }

  if (text.includes('PARAM_ERROR')) {
    const rid = extractRequestId(text)
    return rid
      ? `微信支付参数错误，请检查支付方式与请求参数。Request-Id: ${rid}`
      : '微信支付参数错误，请检查支付方式与请求参数'
  }

  if (text.includes('unsupported protocol scheme')) {
    return '支付通道配置异常：上游接口地址无效'
  }

  if (text.length > 220) {
    const rid = extractRequestId(text)
    if (rid) return `请求失败，请查看后端日志。Request-Id: ${rid}`
    return '请求失败，请查看后端日志'
  }
  return text
}

// 创建axios实例
const axiosInstance: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
axiosInstance.interceptors.request.use(
  (config) => {
    // 添加admin token
    const adminToken = sessionStorage.getItem('admin_token')
    if (adminToken) {
      config.headers['Admin-Token'] = adminToken
    }
    const adminCSRF = sessionStorage.getItem('admin_csrf_token')
    if (adminCSRF) {
      config.headers['X-CSRF-Token'] = adminCSRF
    }

    // 添加user token
    const userToken = sessionStorage.getItem('user_token')
    if (userToken) {
      config.headers['User-Token'] = userToken
    }
    const userCSRF = sessionStorage.getItem('user_csrf_token')
    if (userCSRF && !config.headers['X-CSRF-Token']) {
      config.headers['X-CSRF-Token'] = userCSRF
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
axiosInstance.interceptors.response.use(
  (response: AxiosResponse) => {
    const res = response.data

    if (res.code !== 0 && res.code !== undefined) {
      return Promise.reject(new Error(normalizeErrorMessage(res.msg || '请求失败')))
    }

    return res
  },
  (error) => {
    // 401 未授权，登录过期（但登录接口除外，让登录页显示具体错误）
    if (error.response?.status === 401) {
      const url = error.config?.url || ''
      const isLoginPage = url.includes('/admin/login') || url.includes('/user/login') || url.includes('/user/reg')
      if (!isLoginPage) {
        const isAdminRoute = url.includes('/admin/')
        if (isAdminRoute) {
          sessionStorage.removeItem('admin_token')
          sessionStorage.removeItem('admin_csrf_token')
          router.push('/admin/login')
        } else {
          sessionStorage.removeItem('user_token')
          sessionStorage.removeItem('user_csrf_token')
          router.push('/user/login')
        }
        return Promise.reject(new Error('登录已过期'))
      }
    }

    const raw = error.response?.data?.msg || error.message || '网络错误'
    const msg = normalizeErrorMessage(raw)
    return Promise.reject(new Error(msg))
  }
)

const request = axiosInstance as RequestInstance

export default request
