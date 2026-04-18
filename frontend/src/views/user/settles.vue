<template>
  <div class="space-y-4">
    <div class="page-head">
      <div>
        <h2 class="page-title no-wrap">结算管理</h2>
        <p class="page-subtitle">统一管理结算记录、发起结算与结算账号绑定</p>
      </div>
    </div>

    <div class="card">
      <div class="px-3 md:px-5 pt-3 md:pt-4 border-b border-gray-100">
        <div class="flex flex-wrap gap-2">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            :class="[
              'menu-link px-3 md:px-4 py-2 rounded-t-lg text-xs md:text-sm font-medium transition-colors no-wrap',
              activeTab === tab.key
                ? 'menu-link-active border border-b-0 border-primary-100'
                : ''
            ]"
            @click="switchTab(tab.key)"
          >
            {{ tab.name }}
          </button>
        </div>
      </div>

      <div class="p-3 md:p-5">
        <div v-if="activeTab === 'records'" class="space-y-4">
          <div class="flex items-center justify-between">
            <h3 class="text-base font-semibold text-gray-800">结算记录</h3>
            <button
              class="btn btn-outline !px-3 !py-1.5 !min-h-0 text-xs"
              :disabled="loadingSettles"
              @click="fetchSettles"
            >
              {{ loadingSettles ? '刷新中...' : '刷新记录' }}
            </button>
          </div>

          <div class="overflow-auto">
            <table class="table min-w-[760px] whitespace-nowrap">
              <thead>
                <tr>
                  <th>结算方式</th>
                  <th>账号</th>
                  <th>申请金额</th>
                  <th>实际到账</th>
                  <th>状态</th>
                  <th>时间</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="s in settles" :key="s.id">
                  <td class="text-left">
                    <div class="flex items-center gap-1.5">
                      <SvgIcon :name="settleIconName(s.type)" :size="16" />
                      <span>{{ settleTypeText(s.type) }}</span>
                    </div>
                  </td>
                  <td class="text-left">{{ s.account || '-' }}</td>
                  <td class="text-right text-amber-600 font-semibold">¥{{ s.money }}</td>
                  <td class="text-right text-emerald-600 font-semibold">¥{{ s.realmoney }}</td>
                  <td>
                    <span :class="['badge', settleStatusClass(s.status)]">
                      {{ settleStatusText(s.status) }}
                    </span>
                  </td>
                  <td class="text-left">{{ dayjs(s.addtime).format('YYYY-MM-DD HH:mm') }}</td>
                </tr>
                <tr v-if="settles.length === 0">
                  <td colspan="6" class="text-center text-gray-500 py-8">暂无结算记录</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div v-if="activeTab === 'apply'" class="space-y-4 max-w-xl">
          <div class="section-card text-sm text-blue-700">
            当前绑定结算账号：{{ bindingForm.account || '未绑定' }}（{{ settleTypeText(bindingForm.settle_id) }}）
          </div>
          <form class="space-y-4" @submit.prevent="handleApplySettle">
            <div>
              <label class="form-label">结算方式</label>
              <select v-model.number="applyForm.type" class="form-input px-3">
                <option :value="1">支付宝</option>
                <option :value="2">微信</option>
                <option :value="3">QQ钱包</option>
                <option :value="4">银行卡</option>
              </select>
            </div>
            <div>
              <label class="form-label">结算账号</label>
              <input v-model.trim="applyForm.account" type="text" class="form-input px-3" placeholder="请输入结算账号" />
            </div>
            <div>
              <label class="form-label">收款姓名</label>
              <input v-model.trim="applyForm.username" type="text" class="form-input px-3" placeholder="请输入收款姓名" />
            </div>
            <div>
              <label class="form-label">结算金额</label>
              <input v-model.number="applyForm.money" type="number" min="0.01" step="0.01" class="form-input px-3" />
            </div>
            <div class="flex items-center gap-2">
              <button type="submit" class="btn btn-primary" :disabled="applying">
                {{ applying ? '提交中...' : '提交结算申请' }}
              </button>
              <button type="button" class="btn btn-outline" @click="fillApplyFromBinding">
                使用绑定账号填充
              </button>
            </div>
          </form>
        </div>

        <div v-if="activeTab === 'binding'" class="space-y-4 max-w-xl">
          <div class="section-card text-sm text-emerald-700">
            在此维护默认结算账号，发起结算时可一键带入。
          </div>
          <form class="space-y-4" @submit.prevent="handleSaveBinding">
            <div>
              <label class="form-label">默认结算方式</label>
              <select v-model.number="bindingForm.settle_id" class="form-input px-3">
                <option :value="1">支付宝</option>
                <option :value="2">微信</option>
                <option :value="3">QQ钱包</option>
                <option :value="4">银行卡</option>
              </select>
            </div>
            <div>
              <label class="form-label">结算账号</label>
              <input v-model.trim="bindingForm.account" type="text" class="form-input px-3" placeholder="请输入账号（如支付宝账号/微信OpenID）" />
            </div>
            <div>
              <label class="form-label">收款姓名</label>
              <input v-model.trim="bindingForm.username" type="text" class="form-input px-3" placeholder="请输入收款人姓名" />
            </div>
            <div>
              <button type="submit" class="btn btn-primary" :disabled="savingBinding">
                {{ savingBinding ? '保存中...' : '保存结算账号' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'
import SvgIcon from '@/components/svgicon.vue'
import { applySettle, getUserInfo, getUserSettles, updateProfile } from '@/api/user'

type TabKey = 'records' | 'apply' | 'binding'

const route = useRoute()
const router = useRouter()

const tabs: Array<{ key: TabKey; name: string }> = [
  { key: 'records', name: '结算记录' },
  { key: 'apply', name: '发起结算' },
  { key: 'binding', name: '结算账号绑定' }
]

const activeTab = ref<TabKey>('records')
const settles = ref<any[]>([])
const loadingSettles = ref(false)
const applying = ref(false)
const savingBinding = ref(false)

const bindingForm = reactive({
  settle_id: 1,
  account: '',
  username: ''
})

const applyForm = reactive({
  type: 1,
  account: '',
  username: '',
  money: 0.01
})

function settleTypeText(type: number) {
  const map: Record<number, string> = {
    1: '支付宝',
    2: '微信',
    3: 'QQ钱包',
    4: '银行卡'
  }
  return map[type] || `类型${type}`
}

function settleIconName(type: number) {
  if (type === 1) return 'alipay'
  if (type === 2) return 'wechatpay'
  return 'bankcard'
}

function settleStatusText(status: number) {
  const map: Record<number, string> = {
    0: '待处理',
    1: '已完成',
    2: '处理中',
    3: '失败'
  }
  return map[status] || `状态${status}`
}

function settleStatusClass(status: number) {
  if (status === 1) return 'badge-success'
  if (status === 3) return 'badge-danger'
  return 'badge-warning'
}

function switchTab(tab: TabKey) {
  activeTab.value = tab
  router.replace({ query: { ...route.query, tab } })
}

function fillApplyFromBinding() {
  applyForm.type = Number(bindingForm.settle_id || 1)
  applyForm.account = (bindingForm.account || '').trim()
  applyForm.username = (bindingForm.username || '').trim()
}

async function fetchSettles() {
  loadingSettles.value = true
  try {
    const res = await getUserSettles({ page: 1, limit: 20 })
    if (res.code === 0) {
      settles.value = Array.isArray(res.data) ? res.data : []
    } else {
      ElMessage.error(res.msg || '获取结算记录失败')
    }
  } catch (error: any) {
    console.error('获取结算记录失败:', error)
    ElMessage.error(error?.message || '获取结算记录失败')
  } finally {
    loadingSettles.value = false
  }
}

async function loadBindingFromProfile() {
  try {
    const res = await getUserInfo()
    if (res.code !== 0 || !res.data) return
    const data = res.data as any
    bindingForm.settle_id = Number(data.settle_id || 1) || 1
    bindingForm.account = data.account || ''
    bindingForm.username = data.username || ''
    fillApplyFromBinding()
  } catch (error) {
    console.error('获取用户资料失败:', error)
  }
}

async function handleSaveBinding() {
  if (!bindingForm.account.trim()) {
    ElMessage.warning('请输入结算账号')
    return
  }
  if (!bindingForm.username.trim()) {
    ElMessage.warning('请输入收款姓名')
    return
  }

  savingBinding.value = true
  try {
    const res = await updateProfile({
      settle_id: Number(bindingForm.settle_id),
      account: bindingForm.account.trim(),
      username: bindingForm.username.trim()
    })
    if (res.code !== 0) {
      ElMessage.error(res.msg || '保存失败')
      return
    }
    fillApplyFromBinding()
    ElMessage.success('结算账号已保存')
  } catch (error: any) {
    console.error('保存结算账号失败:', error)
    ElMessage.error(error?.message || '保存失败')
  } finally {
    savingBinding.value = false
  }
}

async function handleApplySettle() {
  if (!applyForm.account.trim()) {
    ElMessage.warning('请输入结算账号')
    return
  }
  if (!applyForm.username.trim()) {
    ElMessage.warning('请输入收款姓名')
    return
  }
  if (!applyForm.money || Number(applyForm.money) <= 0) {
    ElMessage.warning('请输入正确的结算金额')
    return
  }

  applying.value = true
  try {
    const res = await applySettle({
      type: Number(applyForm.type),
      account: applyForm.account.trim(),
      username: applyForm.username.trim(),
      money: Number(applyForm.money)
    })
    if (res.code !== 0) {
      ElMessage.error(res.msg || '申请失败')
      return
    }
    ElMessage.success('结算申请已提交')
    await fetchSettles()
    switchTab('records')
  } catch (error: any) {
    console.error('申请结算失败:', error)
    ElMessage.error(error?.message || '申请失败')
  } finally {
    applying.value = false
  }
}

onMounted(async () => {
  const qTab = String(route.query.tab || '')
  if (qTab === 'apply' || qTab === 'binding' || qTab === 'records') {
    activeTab.value = qTab
  }
  await Promise.all([loadBindingFromProfile(), fetchSettles()])
})
</script>
