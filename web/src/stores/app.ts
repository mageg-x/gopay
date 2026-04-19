import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

// 用户信息接口
export interface User {
  uid: number
  username: string
  email: string
  phone: string
  money: number
  status: number
}

// 管理器状态
export const useAppStore = defineStore('app', () => {
  // 状态
  const adminToken = ref<string>(sessionStorage.getItem('admin_token') || '')
  const adminCSRFToken = ref<string>(sessionStorage.getItem('admin_csrf_token') || '')
  const userToken = ref<string>(sessionStorage.getItem('user_token') || '')
  const userCSRFToken = ref<string>(sessionStorage.getItem('user_csrf_token') || '')
  const adminUser = ref<string>('')
  const userInfo = ref<User | null>(null)
  const configs = ref<Record<string, string>>({})

  // 计算属性
  const isAdminLoggedIn = computed(() => !!adminToken.value)
  const isUserLoggedIn = computed(() => !!userToken.value)

  // 管理器登录
  function adminLogin(token: string, username: string, csrfToken = '') {
    adminToken.value = token
    adminCSRFToken.value = csrfToken
    adminUser.value = username
    sessionStorage.setItem('admin_token', token)
    if (csrfToken) {
      sessionStorage.setItem('admin_csrf_token', csrfToken)
    } else {
      sessionStorage.removeItem('admin_csrf_token')
    }
  }

  // 管理器登出
  function adminLogout() {
    adminToken.value = ''
    adminCSRFToken.value = ''
    adminUser.value = ''
    sessionStorage.removeItem('admin_token')
    sessionStorage.removeItem('admin_csrf_token')
  }

  // 商户登录
  function userLogin(token: string, info: User, csrfToken = '') {
    userToken.value = token
    userCSRFToken.value = csrfToken
    userInfo.value = info
    sessionStorage.setItem('user_token', token)
    if (csrfToken) {
      sessionStorage.setItem('user_csrf_token', csrfToken)
    } else {
      sessionStorage.removeItem('user_csrf_token')
    }
  }

  // 商户登出
  function userLogout() {
    userToken.value = ''
    userCSRFToken.value = ''
    userInfo.value = null
    sessionStorage.removeItem('user_token')
    sessionStorage.removeItem('user_csrf_token')
  }

  // 设置配置
  function setConfigs(data: Record<string, string>) {
    configs.value = data
  }

  // 获取配置
  function getConfig(key: string): string {
    return configs.value[key] || ''
  }

  return {
    adminToken,
    adminCSRFToken,
    userToken,
    userCSRFToken,
    adminUser,
    userInfo,
    configs,
    isAdminLoggedIn,
    isUserLoggedIn,
    adminLogin,
    adminLogout,
    userLogin,
    userLogout,
    setConfigs,
    getConfig
  }
})
