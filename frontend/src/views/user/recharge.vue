<template>
  <div class="space-y-4">
    <div class="page-head">
      <div>
        <h1 class="page-title no-wrap">余额充值</h1>
        <p class="page-subtitle">查看账户余额和充值记录</p>
      </div>
    </div>

    <!-- 余额卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-3 md:gap-4">
      <div class="bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl p-4 md:p-6 text-white">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 bg-white/20 rounded-lg flex items-center justify-center">
            <Wallet class="w-5 h-5" />
          </div>
          <div>
            <p class="text-blue-100 text-sm">当前余额</p>
            <p class="text-2xl font-bold">¥{{ userInfo?.money?.toFixed(2) || '0.00' }}</p>
          </div>
        </div>
      </div>

      <div class="card p-4 md:p-6">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center">
            <TrendingUp class="w-5 h-5 text-green-600" />
          </div>
          <div>
            <p class="text-gray-500 text-sm">累计充值</p>
            <p class="text-xl font-bold text-gray-800">¥{{ stats.totalRecharge?.toFixed(2) || '0.00' }}</p>
          </div>
        </div>
      </div>

      <div class="card p-4 md:p-6">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 bg-orange-100 rounded-lg flex items-center justify-center">
            <ArrowDownCircle class="w-5 h-5 text-orange-600" />
          </div>
          <div>
            <p class="text-gray-500 text-sm">累计支出</p>
            <p class="text-xl font-bold text-gray-800">¥{{ stats.totalExpense?.toFixed(2) || '0.00' }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 充值说明 -->
    <div class="bg-blue-50 rounded-xl p-4 border border-blue-100">
      <div class="flex gap-3">
        <Info class="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
        <div class="text-sm text-blue-800">
          <p class="font-medium mb-1">充值说明</p>
          <ul class="list-disc list-inside space-y-1 text-blue-700">
            <li>余额用于支付交易手续费和开通增值服务</li>
            <li>充值后余额可在结算时抵扣手续费</li>
            <li>如有疑问请联系客服</li>
          </ul>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="card-body">
      <div class="flex items-center justify-between mb-4">
        <h3 class="font-semibold text-gray-700">在线充值</h3>
        <span class="text-xs text-gray-500">创建充值订单后可直接支付</span>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
        <div>
          <label class="form-label">充值金额</label>
          <input
            v-model.number="rechargeForm.money"
            type="number"
            min="0.01"
            step="0.01"
            class="form-input"
          />
        </div>
        <div>
          <label class="form-label">支付方式</label>
          <select
            v-model.number="rechargeForm.type"
            class="form-input"
          >
            <option :value="0">请选择支付方式</option>
            <option v-for="pt in payTypes" :key="pt.id" :value="Number(pt.id)">
              {{ pt.showname || pt.name || ('类型' + pt.id) }}
            </option>
          </select>
        </div>
        <div class="flex items-end">
          <button
            @click="submitRecharge"
            :disabled="rechargeLoading"
            class="btn btn-primary w-full disabled:opacity-50"
          >
            {{ rechargeLoading ? '创建中...' : '去充值' }}
          </button>
        </div>
      </div>
      <div v-if="rechargeTradeNo" class="mt-3 text-xs text-gray-600">
        充值订单号：<span class="font-mono">{{ rechargeTradeNo }}</span>
      </div>
      </div>
    </div>

    <div class="table-shell">
      <div class="card-header flex flex-wrap items-center justify-between gap-2">
        <h3 class="font-semibold text-gray-700">充值记录</h3>
        <div class="flex flex-wrap items-center gap-2">
          <select v-model="filterType"
            class="form-input w-[132px] min-h-[34px] py-1">
            <option value="">全部类型</option>
            <option value="1">充值</option>
            <option value="2">退款</option>
            <option value="3">提现</option>
            <option value="4">消费</option>
          </select>
          <button @click="page = 1; fetchRecords()" class="btn btn-primary py-1.5 px-3 text-sm">筛选</button>
        </div>
      </div>
      <div class="table-shell-body overflow-x-auto">
        <table class="table min-w-[760px] whitespace-nowrap">
          <thead>
            <tr>
              <th class="text-left">时间</th>
              <th>类型</th>
              <th class="text-right">金额</th>
              <th class="text-right">余额</th>
              <th class="text-left">备注</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in records" :key="r.id">
              <td class="text-left text-gray-500 text-xs">{{ formatTime(r.date) }}</td>
              <td>
                <span :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                  typeClass(r.action)]">
                  {{ typeName(r.action) }}
                </span>
              </td>
              <td class="text-right">
                <span :class="['font-medium', r.money >= 0 ? 'text-green-600' : 'text-red-600']">
                  {{ r.money >= 0 ? '+' : '' }}{{ r.money.toFixed(2) }}
                </span>
              </td>
              <td class="text-right text-gray-600">¥{{ r.newmoney.toFixed(2) }}</td>
              <td class="text-left text-gray-500 text-xs">{{ r.trade_no || '-' }}</td>
            </tr>
            <tr v-if="records.length === 0">
              <td colspan="5" class="py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <Receipt class="w-10 h-10 text-gray-300 mb-2" />
                  <span>暂无充值记录</span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="px-4 py-3 border-t border-slate-200/70 flex flex-wrap items-center justify-between gap-2">
        <div class="text-sm text-gray-500">共 {{ total }} 条</div>
        <div class="flex items-center gap-2">
          <button @click="page--; fetchRecords()" :disabled="page <= 1"
            class="pagination-item disabled:opacity-50 disabled:cursor-not-allowed">上一页</button>
          <span class="px-3 py-1 text-sm">{{ page }} / {{ totalPages }}</span>
          <button @click="page++; fetchRecords()" :disabled="page >= totalPages"
            class="pagination-item disabled:opacity-50 disabled:cursor-not-allowed">下一页</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { createRechargeOrder, getUserRecords } from '@/api/user'
import { useAppStore } from '@/stores/app'
import { Wallet, TrendingUp, ArrowDownCircle, Info, Receipt } from 'lucide-vue-next'
import { getPayTypes } from '@/api/pay'
import { ElMessage } from 'element-plus'

const appStore = useAppStore()
const userInfo = computed(() => appStore.userInfo)

const records = ref<any[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const filterType = ref('')
const payTypes = ref<any[]>([])
const rechargeLoading = ref(false)
const rechargeTradeNo = ref('')
const rechargeForm = ref({
  money: 10,
  type: 0
})

const stats = ref({
  totalRecharge: 0,
  totalExpense: 0
})

const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

function typeName(action: number) {
  const map: Record<number, string> = {
    1: '充值',
    2: '退款',
    3: '提现',
    4: '消费',
    5: '结算',
    6: '返现'
  }
  return map[action] || '其他'
}

function typeClass(action: number) {
  const map: Record<number, string> = {
    1: 'bg-green-100 text-green-700',
    2: 'bg-blue-100 text-blue-700',
    3: 'bg-orange-100 text-orange-700',
    4: 'bg-red-100 text-red-700',
    5: 'bg-purple-100 text-purple-700',
    6: 'bg-yellow-100 text-yellow-700'
  }
  return map[action] || 'bg-gray-100 text-gray-700'
}

function formatTime(time: string) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

async function fetchRecords() {
  try {
    const params: any = { page: page.value, limit: pageSize.value }
    if (filterType.value) {
      params.action = filterType.value
    }
    const res = await getUserRecords(params)
    if (res.code === 0) {
      records.value = res.data || []
      total.value = res.count || 0

      // 计算统计
      let recharge = 0
      let expense = 0
      records.value.forEach(r => {
        if (r.action === 1 && r.money > 0) recharge += r.money
        if (r.action === 4 && r.money < 0) expense += Math.abs(r.money)
      })
      stats.value.totalRecharge = recharge
      stats.value.totalExpense = expense
    }
  } catch (error) {
    console.error('获取记录失败:', error)
  }
}

function buildRechargePageHtml(raw: string) {
  if (!raw) return raw
  let html = raw
  html = html.replace(/<form\b([^>]*)>/i, (all, attrs) => {
    if (/accept-charset\s*=/i.test(attrs)) return all
    return `<form${attrs} accept-charset="UTF-8">`
  })
  html = html.replace(
    /<script>\s*document\.getElementById\('payform'\)\.submit\(\);\s*<\/script>/i,
    "<script>(function(){var f=document.getElementById('payform');if(f){f.acceptCharset='UTF-8';f.submit();}})();<\\/script>"
  )
  return html
}

async function submitRecharge() {
  const uid = Number(appStore.userInfo?.uid || 0)
  if (!uid) {
    ElMessage.error('未获取到商户ID')
    return
  }
  if (!rechargeForm.value.type) {
    ElMessage.warning('请选择支付方式')
    return
  }
  if (!rechargeForm.value.money || Number(rechargeForm.value.money) <= 0) {
    ElMessage.warning('请输入有效金额')
    return
  }

  rechargeLoading.value = true
  try {
    const res = await createRechargeOrder({
      type: Number(rechargeForm.value.type),
      money: Number(rechargeForm.value.money),
      notify_url: '',
      return_url: ''
    })
    rechargeTradeNo.value = res.trade_no || ''

    const result = res.result || {
      Type: res.pay_type || '',
      URL: res.pay_info || '',
      Data: res.pay_data || ''
    }
    const t = String(result.Type || result.type || '').toLowerCase()
    const u = String(result.URL || result.url || '')
    const d = result.Data ?? result.data

    if (t === 'jump' && u) {
      if (!/^https:\/\//i.test(u) && !/^http:\/\/localhost(?::\d+)?\//i.test(u) && !/^http:\/\/127\.0\.0\.1(?::\d+)?\//i.test(u)) {
        ElMessage.error('支付链接不安全，已阻止打开')
        return
      }
      window.open(u, '_blank')
    } else if (t === 'html' && typeof d === 'string') {
      const win = window.open('', '_blank')
      if (!win) {
        ElMessage.error('浏览器阻止了新窗口，请允许弹窗后重试')
      } else {
        win.document.open()
        win.document.write(buildRechargePageHtml(d))
        win.document.close()
      }
    } else if (u) {
      if (!/^https:\/\//i.test(u) && !/^http:\/\/localhost(?::\d+)?\//i.test(u) && !/^http:\/\/127\.0\.0\.1(?::\d+)?\//i.test(u)) {
        ElMessage.error('支付链接不安全，已阻止打开')
        return
      }
      window.open(u, '_blank')
    }

    ElMessage.success('充值订单创建成功')
  } catch (error: any) {
    ElMessage.error(error?.message || '创建充值订单失败')
  } finally {
    rechargeLoading.value = false
  }
}

onMounted(() => {
  fetchRecords()
  const uid = Number(appStore.userInfo?.uid || 0)
  if (uid) {
    getPayTypes(uid).then((res) => {
      payTypes.value = Array.isArray(res.data) ? res.data : []
      if (!rechargeForm.value.type && payTypes.value.length > 0) {
        rechargeForm.value.type = Number(payTypes.value[0].id)
      }
    }).catch((error) => {
      console.error('获取支付方式失败:', error)
    })
  }
})
</script>
