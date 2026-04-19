<template>
  <div class="space-y-4">
    <div class="page-head">
      <div>
        <h2 class="page-title no-wrap">商户管理</h2>
        <p class="page-subtitle">共 {{ total }} 个商户</p>
      </div>
      <button class="btn btn-primary" @click="openAddDialog">+ 添加商户</button>
    </div>

    <div class="panel-filter">
      <div class="card-body py-3">
        <div class="toolbar-wrap">
          <div class="relative w-full md:w-72">
            <input v-model="searchKeyword" type="text" class="form-input form-input-icon" placeholder="搜索商户 ID、姓名、账号..." />
            <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none"
              stroke="currentColor" viewBox="0 0 24 24">
              <circle cx="11" cy="11" r="8"></circle>
              <path d="m21 21-4.35-4.35"></path>
            </svg>
          </div>
          <span class="text-sm text-gray-500">筛选结果：{{ filteredUsers.length }} 条</span>
        </div>
      </div>
    </div>

    <div class="table-shell">
      <div class="table-shell-body overflow-x-auto">
        <div v-if="loading" class="flex items-center justify-center py-12 text-gray-500">
          <svg class="animate-spin h-6 w-6 mr-2" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor"
              d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
            </path>
          </svg>
          加载中...
        </div>

        <table v-else class="table table-fixed whitespace-nowrap min-w-[1120px]">
          <thead>
            <tr>
              <th class="pl-6 w-8">ID</th>
              <th class="w-32">商户信息</th>
              <th class="w-32">结算账号</th>
              <th class="w-24">余额</th>
              <th class="w-20">模式</th>
              <th class="w-20">支付</th>
              <th class="w-20">结算</th>
              <th class="w-20">状态</th>
              <th class="w-36">注册时间</th>
              <th class="pr-6 w-64">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in filteredUsers" :key="user.uid" class="align-middle">
              <td class="pl-6">
                <span
                  class="inline-flex items-center justify-center min-w-[32px] h-6 px-2 bg-blue-50 text-blue-600 text-xs font-semibold rounded">
                  {{ user.uid }}
                </span>
              </td>
              <td class="truncate">
                <div class="font-medium text-gray-900 truncate">{{ user.username || '-' }}</div>
                <div class="text-xs text-gray-500 truncate">
                  <span v-if="user.phone">📱{{ user.phone }}</span>
                  <span v-if="user.phone && user.email"> | </span>
                  <span v-if="user.email">✉️{{ user.email }}</span>
                </div>
              </td>
              <td class="truncate">
                <div class="text-gray-900 truncate">{{ user.account || '-' }}</div>
                <div class="text-xs text-gray-500 truncate" :title="user.url">{{ user.url || '-' }}</div>
              </td>
              <td class="whitespace-nowrap">
                <span
                  class="inline-flex items-center px-2 py-0.5 rounded-lg bg-emerald-50 text-emerald-700 font-semibold text-sm">
                  ¥{{ user.money }}
                </span>
              </td>
              <td>
                <span
                  :class="['inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium', user.mode === 1 ? 'bg-amber-100 text-amber-700' : 'bg-slate-100 text-slate-600']">
                  {{ user.mode === 1 ? '加费' : '减费' }}
                </span>
              </td>
              <td>
                <button @click="setStatus(user.uid, user.pay === 1 ? 0 : 1)" :class="[
                  'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium cursor-pointer transition-colors',
                  user.pay === 1 ? 'bg-emerald-100 text-emerald-700 hover:bg-emerald-200' : 'bg-rose-100 text-rose-700 hover:bg-rose-200'
                ]">
                  {{ payMap[user.pay] }}
                </button>
              </td>
              <td>
                <button @click="setStatus(user.uid, user.settle === 1 ? 0 : 1)" :class="[
                  'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium cursor-pointer transition-colors',
                  user.settle === 1 ? 'bg-emerald-100 text-emerald-700 hover:bg-emerald-200' : 'bg-rose-100 text-rose-700 hover:bg-rose-200'
                ]">
                  {{ settleMap[user.settle] }}
                </button>
              </td>
              <td>
                <span :class="['badge', statusMap[user.status]?.class]">
                  {{ statusMap[user.status]?.text || '未知' }}
                </span>
              </td>
              <td class="text-gray-500 text-sm whitespace-nowrap">{{ formatTime(user.addtime) }}</td>
              <td class="pr-6">
                <div class="flex items-center gap-1">
                  <button class="action-link action-link-primary" @click="openDetailDialog(user)">详情</button>
                  <button class="action-link action-link-primary" @click="openEditDialog(user)">编辑</button>
                  <button class="action-link action-link-primary" @click="setKey(user)">修改密钥</button>
                  <button class="action-link action-link-warning" @click="resetKey(user.uid)">重置密钥</button>
                  <button class="action-link action-link-success" @click="openMoneyDialog(user)">余额</button>
                  <button v-if="user.status === 0" class="action-link action-link-success" @click="setStatus(user.uid, 1)">
                    启用
                  </button>
                  <button v-else-if="user.status === 1" class="action-link action-link-warning" @click="setStatus(user.uid, 0)">
                    禁用
                  </button>
                  <button class="action-link action-link-danger" @click="deleteUser(user.uid)">删除</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>

        <div v-if="!loading && filteredUsers.length === 0" class="text-center py-12 text-gray-500">
          <svg class="w-16 h-16 mx-auto mb-4 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
              d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z">
            </path>
          </svg>
          <p class="mb-3">暂无商户数据</p>
          <button class="btn btn-primary" @click="openAddDialog">添加第一个商户</button>
        </div>

        <div v-if="!loading && filteredUsers.length > 0" class="pagination border-t text-sm border-gray-100">
          <button class="pagination-item" :disabled="page === 1" @click="page--; fetchUsers()">
            上一页
          </button>
          <span class="px-4 py-1 text-gray-600">
            第 {{ page }} / {{ Math.ceil(total / 20) || 1 }} 页，共 {{ total }} 条
          </span>
          <button class="pagination-item" :disabled="page * 20 >= total" @click="page++; fetchUsers()">
            下一页
          </button>
        </div>
      </div>
    </div>

    <Teleport to="body">
      <div v-if="dialogVisible" class="dialog-backdrop">
        <div class="dialog-wrap">
          <div class="dialog-mask" @click="dialogVisible = false"></div>
          <div class="dialog-panel max-w-[900px]">
            <div class="dialog-header">
            <div>
              <h3 class="dialog-title">{{ dialogTitle }}</h3>
              <p class="dialog-subtitle">{{ isEdit ? '修改商户信息' : '创建新商户账户' }}</p>
            </div>
            <button class="dialog-close" @click="dialogVisible = false">✕</button>
          </div>

          <div class="dialog-body">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4 md:gap-6">
              <div class="section-card">
                <h4 class="section-title">基本信息</h4>
                <div class="space-y-4">
                  <div>
                    <label class="form-label">用户组</label>
                    <select v-model="userForm.gid" class="form-input px-3">
                      <option v-for="g in groups" :key="g.gid" :value="g.gid">{{ g.name }}</option>
                    </select>
                  </div>
                  <div>
                    <label class="form-label">手机号</label>
                    <input v-model="userForm.phone" type="text" class="form-input px-3" placeholder="可留空">
                  </div>
                  <div>
                    <label class="form-label">邮箱</label>
                    <input v-model="userForm.email" type="email" class="form-input px-3" placeholder="可留空">
                  </div>
                  <div>
                    <label class="form-label">登录密码</label>
                    <input v-model="userForm.pwd" type="password" class="form-input px-3"
                      :placeholder="isEdit ? '留空则不修改' : '可留空'">
                  </div>
                  <div>
                    <label class="form-label">QQ</label>
                    <input v-model="userForm.qq" type="text" class="form-input px-3" placeholder="可留空">
                  </div>
                  <div>
                    <label class="form-label">网站域名</label>
                    <input v-model="userForm.url" type="text" class="form-input px-3" placeholder="可留空">
                  </div>
                </div>
              </div>

              <div class="section-card">
                <h4 class="section-title">商户资料（可写）</h4>
                <div class="space-y-4">
                  <div>
                    <label class="form-label">商户姓名</label>
                    <input v-model="userForm.username" type="text" class="form-input px-3" placeholder="必填">
                  </div>
                  <div>
                    <label class="form-label">
                      商户账号 <span class="text-red-500">*</span>
                    </label>
                    <input v-model="userForm.account" type="text" class="form-input px-3" placeholder="必填">
                  </div>
                  <div>
                    <label class="form-label">结算方式</label>
                    <select v-model="userForm.settle_id" class="form-input px-3">
                      <option :value="1">支付宝</option>
                      <option :value="2">微信</option>
                    </select>
                  </div>
                </div>
              </div>

              <div class="section-card">
                <h4 class="section-title">功能开关</h4>
                <div class="space-y-4">
                  <div>
                    <label class="form-label">手续费模式</label>
                    <select v-model="userForm.mode" class="form-input px-3">
                      <option :value="0">余额扣费</option>
                      <option :value="1">订单加费</option>
                    </select>
                  </div>
                  <div>
                    <label class="form-label">商户状态</label>
                    <select v-model="userForm.status" class="form-input px-3">
                      <option :value="1">正常</option>
                      <option :value="0">禁用</option>
                      <option :value="2">待审核</option>
                    </select>
                  </div>
                  <div>
                    <label class="form-label">支付权限</label>
                    <select v-model="userForm.pay" class="form-input px-3">
                      <option :value="1">开启</option>
                      <option :value="0">关闭</option>
                    </select>
                  </div>
                  <div>
                    <label class="form-label">结算权限</label>
                    <select v-model="userForm.settle" class="form-input px-3">
                      <option :value="1">开启</option>
                      <option :value="0">关闭</option>
                    </select>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="dialog-footer">
            <button class="btn btn-outline" @click="dialogVisible = false">取消</button>
            <button class="btn btn-primary" @click="submitForm">{{ isEdit ? '保存修改' : '创建商户' }}</button>
          </div>
        </div>
      </div>
      </div>

      <div v-if="moneyDialogVisible" class="dialog-backdrop">
        <div class="dialog-wrap">
          <div class="dialog-mask" @click="moneyDialogVisible = false"></div>
          <div class="dialog-panel max-w-[400px]">
            <div class="dialog-header">
            <div>
              <h3 class="dialog-title">余额操作</h3>
              <p class="dialog-subtitle">商户：{{ moneyForm.username || '-' }}（ID：{{ moneyForm.uid }}）</p>
            </div>
            <button class="dialog-close" @click="moneyDialogVisible = false">✕</button>
          </div>

          <div class="dialog-body">
            <div class="space-y-4">
              <div>
                <label class="form-label">操作类型</label>
                <select v-model="moneyForm.type" class="form-input px-3">
                  <option value="admin_add">充值（加款）</option>
                  <option value="admin_sub">扣除（减款）</option>
                </select>
              </div>
              <div>
                <label class="form-label">金额</label>
                <input v-model.number="moneyForm.money" type="number" step="0.01" min="0.01"
                  class="form-input px-3" placeholder="请输入金额" />
              </div>
              <div>
                <label class="form-label">备注</label>
                <input v-model="moneyForm.remark" type="text"
                  class="form-input px-3" placeholder="可选" />
              </div>
            </div>
          </div>

          <div class="dialog-footer">
            <button class="btn btn-outline" @click="moneyDialogVisible = false">取消</button>
            <button class="btn btn-primary" @click="submitMoneyOp">确定</button>
          </div>
        </div>
      </div>
      </div>

      <div v-if="detailDialogVisible" class="dialog-backdrop">
        <div class="dialog-wrap">
          <div class="dialog-mask" @click="detailDialogVisible = false"></div>
          <div class="dialog-panel max-w-[900px]">
            <div class="dialog-header">
            <div>
              <h3 class="dialog-title">商户详情</h3>
              <p class="dialog-subtitle">仅展示，不可编辑</p>
            </div>
            <button class="dialog-close" @click="detailDialogVisible = false">✕</button>
          </div>

          <div class="dialog-body" v-if="detailUser">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4 md:gap-6">
              <div class="section-card">
                <h4 class="section-title">账号信息</h4>
                <div class="space-y-3">
                  <div class="kv-row"><span class="kv-key">商户ID</span><span class="kv-value">{{ detailUser.uid }}</span></div>
                  <div class="kv-row"><span class="kv-key">用户组ID</span><span class="kv-value">{{ detailUser.gid ?? '-' }}</span></div>
                  <div class="kv-row"><span class="kv-key">上级ID</span><span class="kv-value">{{ detailUser.upid || '-' }}</span></div>
                  <div class="kv-row"><span class="kv-key">姓名</span><span class="kv-value">{{ detailUser.username || '-' }}</span></div>
                  <div class="kv-row"><span class="kv-key">账号</span><span class="kv-value">{{ detailUser.account || '-' }}</span></div>
                  <div class="kv-row"><span class="kv-key">手机号</span><span class="kv-value">{{ detailUser.phone || '-' }}</span></div>
                  <div class="kv-row"><span class="kv-key">邮箱</span><span class="kv-value">{{ detailUser.email || '-' }}</span></div>
                  <div class="kv-row"><span class="kv-key">QQ</span><span class="kv-value">{{ detailUser.qq || '-' }}</span></div>
                  <div class="kv-row"><span class="kv-key">站点域名</span><span class="kv-value">{{ detailUser.url || '-' }}</span></div>
                </div>
              </div>

              <div class="section-card">
                <h4 class="section-title">支付与结算</h4>
                <div class="space-y-3">
                  <div class="kv-row"><span class="kv-key">余额</span><span class="kv-value">¥{{ detailUser.money ?? 0 }}</span></div>
                  <div class="kv-row"><span class="kv-key">手续费模式</span><span class="kv-value">{{ detailUser.mode === 1 ? '订单加费' : '余额扣费' }}</span></div>
                  <div class="kv-row"><span class="kv-key">支付权限</span><span class="kv-value">{{ payMap[detailUser.pay] || '-' }}</span></div>
                  <div class="kv-row"><span class="kv-key">结算权限</span><span class="kv-value">{{ settleMap[detailUser.settle] || '-' }}</span></div>
                  <div class="kv-row"><span class="kv-key">结算方式</span><span class="kv-value">{{ settleMethodText(detailUser.settle_id) }}</span></div>
                  <div class="kv-row"><span class="kv-key">结算账号</span><span class="kv-value">{{ detailUser.account || '-' }}</span></div>
                  <div class="kv-row"><span class="kv-key">结算姓名</span><span class="kv-value">{{ detailUser.username || '-' }}</span></div>
                </div>
              </div>

              <div class="section-card">
                <h4 class="section-title">扩展与安全</h4>
                <div class="space-y-3">
                  <div class="kv-row"><span class="kv-key">商户状态</span><span class="kv-value">{{ statusMap[detailUser.status]?.text || '未知' }}</span></div>
                  <div class="kv-row"><span class="kv-key">实名状态</span><span class="kv-value">{{ certText(detailUser.cert) }}</span></div>
                  <div class="kv-row"><span class="kv-key">支付宝UID</span><span class="kv-value">{{ detailUser.alipay_uid || '-' }}</span></div>
                  <div class="kv-row"><span class="kv-key">微信UID</span><span class="kv-value">{{ detailUser.wx_uid || '-' }}</span></div>
                  <div class="kv-row"><span class="kv-key">QQ钱包UID</span><span class="kv-value">{{ detailUser.qq_uid || '-' }}</span></div>
                  <div class="kv-row"><span class="kv-key">API Key</span><span class="kv-value">{{ detailUser.key || '-' }}</span></div>
                  <div class="kv-row"><span class="kv-key">创建时间</span><span class="kv-value">{{ formatTime(detailUser.addtime) }}</span></div>
                  <div class="kv-row"><span class="kv-key">最后登录</span><span class="kv-value">{{ formatTime(detailUser.lasttime) }}</span></div>
                </div>
              </div>
            </div>
          </div>

          <div class="dialog-footer">
            <button class="btn btn-primary" @click="detailDialogVisible = false">关闭</button>
          </div>
        </div>
      </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import { getUserList, addUser, updateUser, userOp, getUserEdit, getGroupList } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'
import dayjs from 'dayjs'

interface User {
  uid: number
  gid: number
  upid?: number
  key?: string
  username: string
  email: string
  phone: string
  qq: string
  alipay_uid?: string
  wx_uid?: string
  qq_uid?: string
  cert?: number
  settle_id?: number
  url: string
  account: string
  money: number
  pay: number
  settle: number
  status: number
  mode: number
  addtime: string
}

interface Group {
  gid: number
  name: string
}

const users = ref<User[]>([])
const loading = ref(false)
const page = ref(1)
const total = ref(0)
const searchKeyword = ref('')

const dialogVisible = ref(false)
const dialogTitle = ref('添加商户')
const isEdit = ref(false)
const editingUser = ref<User | null>(null)
const moneyDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const detailUser = ref<User | null>(null)

const moneyForm = reactive({
  uid: 0,
  username: '',
  type: 'admin_add',
  money: 0,
  remark: ''
})

const userForm = reactive({
  gid: 1,
  phone: '',
  email: '',
  pwd: '',
  qq: '',
  url: '',
  settle_id: 1,
  account: '',
  username: '',
  mode: 0,
  pay: 1,
  settle: 1,
  status: 1
})

const groups = ref<Group[]>([])

const statusMap: Record<number, { text: string; class: string }> = {
  0: { text: '禁用', class: 'badge-danger' },
  1: { text: '正常', class: 'badge-success' },
  2: { text: '待审核', class: 'badge-warning' }
}

const payMap: Record<number, string> = { 0: '关闭', 1: '开启' }
const settleMap: Record<number, string> = { 0: '关闭', 1: '开启' }

const filteredUsers = computed(() => {
  if (!searchKeyword.value) return users.value
  const kw = searchKeyword.value.toLowerCase()
  return users.value.filter(u =>
    u.username?.toLowerCase().includes(kw) ||
    u.account?.toLowerCase().includes(kw) ||
    u.uid.toString().includes(kw)
  )
})

async function fetchUsers() {
  loading.value = true
  try {
    const res = await getUserList({ page: page.value, limit: 20 })
    if (res.code === 0) {
      users.value = (res.data || []).map((u: any) => ({
        ...u,
        uid: Number(u.uid || 0),
        gid: Number(u.gid || 0),
        settle_id: Number(u.settle_id || 1)
      }))
      total.value = res.count || 0
    }
  } catch (error) {
    console.error('获取商户列表失败:', error)
  } finally {
    loading.value = false
  }
}

async function fetchGroups() {
  try {
    const res = await getGroupList({ page: 1, limit: 500 })
    groups.value = (res.data || []).map((g: any) => ({
      gid: Number(g.gid),
      name: g.name
    }))
    if (groups.value.length === 0) {
      groups.value = [{ gid: 1, name: '默认组' }]
    }
    if (!groups.value.some(g => Number(g.gid) === Number(userForm.gid))) {
      userForm.gid = groups.value[0].gid
    }
  } catch (error: any) {
    console.error('获取用户组失败:', error)
    ElMessage.error(error?.message || '获取用户组失败')
    if (groups.value.length === 0) {
      groups.value = [{ gid: 1, name: '默认组' }]
    }
  }
}

function formatTime(time: string) {
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

function openAddDialog() {
  isEdit.value = false
  dialogTitle.value = '添加商户'
  editingUser.value = null
  resetForm()
  if (!groups.value.some(g => Number(g.gid) === Number(userForm.gid))) {
    userForm.gid = groups.value[0]?.gid || 1
  }
  dialogVisible.value = true
}

async function openEditDialog(user: User) {
  isEdit.value = true
  dialogTitle.value = '编辑商户'
  editingUser.value = user

  try {
    const res = await getUserEdit(user.uid)
    if (res.code === 0) {
      if (Array.isArray(res.groups) && res.groups.length > 0) {
        groups.value = res.groups.map((g: any) => ({
          gid: Number(g.gid),
          name: g.name
        }))
      }
      const editUser = res.user
      user.key = editUser.key || user.key
      userForm.gid = editUser.gid ?? (groups.value[0]?.gid || 1)
      userForm.phone = editUser.phone || ''
      userForm.email = editUser.email || ''
      userForm.qq = editUser.qq || ''
      userForm.url = editUser.url || ''
      userForm.settle_id = editUser.settle_id ?? 1
      userForm.account = editUser.account || ''
      userForm.username = editUser.username || ''
      userForm.mode = editUser.mode ?? 0
      userForm.pay = editUser.pay ?? 1
      userForm.settle = editUser.settle ?? 1
      userForm.status = editUser.status ?? 1
      userForm.pwd = ''
    }
  } catch (error) {
    console.error('获取商户信息失败:', error)
    ElMessage.error('获取商户信息失败')
  }

  dialogVisible.value = true
}

async function openDetailDialog(user: User) {
  try {
    const res = await getUserEdit(user.uid)
    if (res.code !== 0 || !res.user) {
      ElMessage.error('获取商户详情失败')
      return
    }
    detailUser.value = {
      ...user,
      ...res.user
    }
    detailDialogVisible.value = true
  } catch (error: any) {
    ElMessage.error(error?.message || '获取商户详情失败')
  }
}

function resetForm() {
  userForm.gid = 1
  userForm.phone = ''
  userForm.email = ''
  userForm.pwd = ''
  userForm.qq = ''
  userForm.url = ''
  userForm.settle_id = 1
  userForm.account = ''
  userForm.username = ''
  userForm.mode = 0
  userForm.pay = 1
  userForm.settle = 1
  userForm.status = 1
}

function settleMethodText(settleID: number | undefined) {
  if (Number(settleID) === 2) return '微信'
  return '支付宝'
}

function certText(cert: number | undefined) {
  if (Number(cert) === 1) return '已实名'
  if (Number(cert) === 2) return '审核中'
  return '未实名'
}

async function submitForm() {
  if (!userForm.account || !userForm.username) {
    ElMessage.warning('结算账号和姓名不能为空')
    return
  }

  try {
    if (isEdit.value && editingUser.value) {
      await updateUser({
        uid: editingUser.value.uid,
        gid: userForm.gid,
        phone: userForm.phone,
        email: userForm.email,
        pwd: userForm.pwd,
        qq: userForm.qq,
        url: userForm.url,
        settle_id: userForm.settle_id,
        account: userForm.account,
        username: userForm.username,
        mode: userForm.mode,
        pay: userForm.pay,
        settle: userForm.settle,
        status: userForm.status
      })
      ElMessage.success('更新成功')
    } else {
      await addUser({
        gid: userForm.gid,
        phone: userForm.phone,
        email: userForm.email,
        pwd: userForm.pwd,
        qq: userForm.qq,
        url: userForm.url,
        settle_id: userForm.settle_id,
        account: userForm.account,
        username: userForm.username,
        mode: userForm.mode,
        pay: userForm.pay,
        settle: userForm.settle,
        status: userForm.status
      })
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    fetchUsers()
  } catch (error) {
    console.error('操作失败:', error)
  }
}

async function resetKey(uid: number) {
  try {
    const res = await userOp({ action: 'reset_key', uid })
    if (res.key) {
      await ElMessageBox.alert(`新密钥：${res.key}\n注意：旧密码登录将失效。`, '密钥已重置', {
        confirmButtonText: '我已复制'
      })
    } else {
      ElMessage.success('密钥已重置')
    }
    fetchUsers()
  } catch (error) {
    console.error('重置密钥失败:', error)
  }
}

async function setKey(user: User) {
  try {
    const ret = await ElMessageBox.prompt('请输入新的 API 密钥（至少 8 位）', `修改商户 ${user.uid} 密钥`, {
      confirmButtonText: '保存',
      cancelButtonText: '取消',
      inputValue: user.key || '',
      inputValidator: (val: string) => {
        if (!val || !val.trim()) return '密钥不能为空'
        if (val.trim().length < 8) return '密钥长度至少 8 位'
        return true
      }
    })
    const newKey = ret.value.trim()
    await userOp({ action: 'set_key', uid: user.uid, key: newKey })
    ElMessage.success('密钥已修改，旧密码登录将失效')
    fetchUsers()
  } catch (error: any) {
    if (error !== 'cancel' && error !== 'close') {
      console.error('修改密钥失败:', error)
    }
  }
}

async function deleteUser(uid: number) {
  try {
    await ElMessageBox.confirm('确定要删除该商户吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    await userOp({ action: 'delete', uid })
    ElMessage.success('删除成功')
    fetchUsers()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

async function setStatus(uid: number, status: number) {
  try {
    await userOp({ action: 'set_status', uid, status })
    ElMessage.success('状态已更新')
    fetchUsers()
  } catch (error) {
    console.error('更新状态失败:', error)
  }
}

function openMoneyDialog(user: User) {
  moneyForm.uid = user.uid
  moneyForm.username = user.username || ''
  moneyForm.type = 'admin_add'
  moneyForm.money = 0
  moneyForm.remark = ''
  moneyDialogVisible.value = true
}

async function submitMoneyOp() {
  if (moneyForm.money <= 0) {
    ElMessage.warning('请输入正确的金额')
    return
  }
  try {
    await userOp({
      action: 'recharge',
      uid: moneyForm.uid,
      type: moneyForm.type,
      money: moneyForm.money
    })
    ElMessage.success('余额操作成功')
    moneyDialogVisible.value = false
    fetchUsers()
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  }
}

onMounted(() => {
  fetchUsers()
  fetchGroups()
})
</script>
