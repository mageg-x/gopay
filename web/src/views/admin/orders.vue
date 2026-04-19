<template>
  <div class="space-y-4">
    <div class="page-head">
      <div>
        <h1 class="page-title no-wrap">订单管理</h1>
        <p class="page-subtitle">查看和处理所有支付订单</p>
      </div>
      <select v-model="status" @change="page = 1; fetchOrders()"
        class="form-input w-auto min-w-[132px] px-3">
        <option :value="-1">全部状态</option>
        <option :value="0">待支付</option>
        <option :value="1">已支付</option>
        <option :value="2">已退款</option>
        <option :value="3">已冻结</option>
      </select>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 md:gap-4">
      <div class="card p-4">
        <div class="text-sm text-gray-500">全部订单</div>
        <div class="text-2xl font-bold text-slate-700 mt-1">{{ total }}</div>
      </div>
      <div class="card p-4">
        <div class="text-sm text-gray-500">待支付</div>
        <div class="text-2xl font-bold text-amber-600 mt-1">{{ statusCount(0) }}</div>
      </div>
      <div class="card p-4">
        <div class="text-sm text-gray-500">已支付</div>
        <div class="text-2xl font-bold text-emerald-600 mt-1">{{ statusCount(1) }}</div>
      </div>
      <div class="card p-4">
        <div class="text-sm text-gray-500">已退款/冻结</div>
        <div class="text-2xl font-bold text-rose-600 mt-1">{{ statusCount(2) + statusCount(3) }}</div>
      </div>
    </div>

    <div class="table-shell">
      <div class="overflow-x-auto">
        <table class="table min-w-[980px] whitespace-nowrap">
          <thead>
            <tr>
              <th class="text-left">订单号</th>
              <th class="text-left">商户订单号</th>
              <th class="text-left">商户ID</th>
              <th class="text-left">商品名称</th>
              <th class="text-right">金额</th>
              <th>支付方式</th>
              <th>状态</th>
              <th class="text-left">下单时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="order in orders" :key="order.trade_no">
              <td class="text-left text-gray-500 font-mono text-xs">{{ order.trade_no }}</td>
              <td class="text-left text-gray-500 font-mono text-xs">{{ order.out_trade_no || '-' }}</td>
              <td class="text-left text-gray-900">{{ order.uid }}</td>
              <td class="text-left text-gray-900">{{ order.name || '-' }}</td>
              <td class="text-right font-semibold text-emerald-600">￥{{ order.money }}</td>
              <td>
                <div class="flex items-center justify-center gap-1">
                  <SvgIcon :name="payIcon(order)" :size="16" />
                  <span class="text-xs font-medium" :class="payTextClass(order)">{{
                    order.typename || '未知' }}</span>
                </div>
              </td>
              <td>
                <span :class="['badge', statusClass(order.status)]">
                  {{ statusMap[order.status]?.text || '未知' }}
                </span>
              </td>
              <td class="text-left text-gray-500 text-xs">{{ formatTime(order.addtime) }}</td>
              <td>
                <button @click="showDetail(order)" class="action-link action-link-primary mr-1">详情</button>
                <template v-if="order.status === 1">
                  <button @click="handleOp('refund', order.trade_no)" class="action-link action-link-danger">退款</button>
                </template>
                <template v-if="order.status === 0">
                  <button @click="handleOp('refresh', order.trade_no)" class="action-link action-link-primary mr-1">刷新状态</button>
                  <button @click="handleOp('freeze', order.trade_no)" class="action-link action-link-warning">冻结</button>
                </template>
                <template v-if="order.status === 3">
                  <button @click="handleOp('unfreeze', order.trade_no)" class="action-link action-link-success">解冻</button>
                </template>
              </td>
            </tr>
            <tr v-if="orders.length === 0">
              <td colspan="9" class="py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                  </svg>
                  <span>暂无订单数据</span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="px-4 py-3 border-t border-slate-200/70 flex flex-wrap items-center justify-between gap-2">
        <div class="text-sm text-gray-500">共 {{ total }} 条</div>
        <div class="flex items-center gap-2">
          <button class="pagination-item disabled:opacity-50 disabled:cursor-not-allowed" :disabled="page <= 1"
            @click="page--; fetchOrders()">上一页</button>
          <span class="px-3 py-1 text-sm">{{ page }} / {{ totalPages }}</span>
          <button class="pagination-item disabled:opacity-50 disabled:cursor-not-allowed" :disabled="page >= totalPages"
            @click="page++; fetchOrders()">下一页</button>
        </div>
      </div>
    </div>

    <div v-if="detailVisible" class="dialog-backdrop">
      <div class="dialog-wrap">
        <div class="dialog-mask" @click="detailVisible = false"></div>
        <div class="dialog-panel max-w-lg">
          <div class="dialog-header">
            <div>
              <h3 class="dialog-title">订单详情</h3>
              <p class="dialog-subtitle">查看订单交易信息</p>
            </div>
            <button class="dialog-close" @click="detailVisible = false">✕</button>
          </div>

          <div class="dialog-body" v-if="currentOrder">
            <div class="grid grid-cols-2 gap-2">
              <div class="text-gray-500">平台订单号:</div>
              <div class="font-mono text-gray-900 break-all">{{ currentOrder.trade_no }}</div>
              <div class="text-gray-500">商户订单号:</div>
              <div class="font-mono text-gray-900 break-all">{{ currentOrder.out_trade_no || '-' }}</div>
              <div class="text-gray-500">商户ID:</div>
              <div class="text-gray-900">{{ currentOrder.uid }}</div>
              <div class="text-gray-500">支付方式:</div>
              <div class="text-gray-900">{{ currentOrder.typename || '未知' }}</div>
              <div class="text-gray-500">订单金额:</div>
              <div class="font-bold text-emerald-600">¥{{ currentOrder.money }}</div>
              <div class="text-gray-500">平台实收:</div>
              <div class="font-bold text-emerald-600">¥{{ currentOrder.realmoney || currentOrder.money }}</div>
              <div class="text-gray-500">商户所得:</div>
              <div class="text-blue-600">¥{{ currentOrder.getmoney || '-' }}</div>
              <div class="text-gray-500">回调状态:</div>
              <div>
                <span v-if="currentOrder.notify === 1" class="text-emerald-600">已回调</span>
                <span v-else class="text-amber-600">未回调</span>
              </div>
              <div class="text-gray-500">订单状态:</div>
              <div>
                <span :class="statusClass(currentOrder.status)" class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium">
                  {{ statusMap[currentOrder.status]?.text || '未知' }}
                </span>
              </div>
              <div class="text-gray-500">创建时间:</div>
              <div>{{ currentOrder.addtime ? dayjs(currentOrder.addtime).format('YYYY-MM-DD HH:mm:ss') : '-' }}</div>
              <div class="text-gray-500">支付时间:</div>
              <div>{{ currentOrder.endtime ? dayjs(currentOrder.endtime).format('YYYY-MM-DD HH:mm:ss') : '-' }}</div>
            </div>

            <div v-if="currentOrder.param" class="border-t pt-3 mt-3">
              <div class="text-gray-500 mb-1">订单备注:</div>
              <div class="text-gray-900 break-all">{{ currentOrder.param }}</div>
            </div>
          </div>

          <div class="dialog-footer">
            <button @click="detailVisible = false" class="btn btn-outline">关闭</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getOrderList, orderOp } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'
import dayjs from 'dayjs'
import SvgIcon from '@/components/svgicon.vue'

interface Order {
  trade_no: string
  out_trade_no: string
  uid: number
  name: string
  money: number
  status: number
  type: number
  typename: string
  addtime: string
  endtime: string
}

const orders = ref<Order[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const status = ref(-1)
const detailVisible = ref(false)
const currentOrder = ref<any>(null)

const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

function statusCount(s: number) {
  return orders.value.filter(o => o.status === s).length
}

const statusMap: Record<number, { text: string }> = {
  0: { text: '待支付' },
  1: { text: '已支付' },
  2: { text: '已退款' },
  3: { text: '已冻结' }
}

function statusClass(s: number) {
  const map: Record<number, string> = {
    0: 'bg-yellow-100 text-yellow-700',
    1: 'bg-green-100 text-green-700',
    2: 'bg-blue-100 text-blue-700',
    3: 'bg-red-100 text-red-700'
  }
  return map[s] || 'bg-gray-100 text-gray-700'
}

function payIcon(order: Order) {
  const name = String(order.typename || '')
  if (name.includes('支付宝')) return 'alipay'
  if (name.includes('微信')) return 'wechatpay'
  if (name.includes('银行卡')) return 'bankcard'
  if (Number(order.type) === 1) return 'alipay'
  if (Number(order.type) === 2) return 'wechatpay'
  if (Number(order.type) === 4) return 'bankcard'
  return 'bankcard'
}

function payTextClass(order: Order) {
  const name = String(order.typename || '')
  if (name.includes('支付宝')) return 'text-blue-600'
  if (name.includes('微信')) return 'text-green-600'
  if (name.includes('银行卡')) return 'text-gray-700'
  if (Number(order.type) === 1) return 'text-blue-600'
  if (Number(order.type) === 2) return 'text-green-600'
  if (Number(order.type) === 4) return 'text-gray-700'
  return 'text-gray-600'
}

function showDetail(order: Order) {
  currentOrder.value = order
  detailVisible.value = true
}

async function fetchOrders() {
  loading.value = true
  try {
    const res = await getOrderList({ page: page.value, limit: pageSize.value, status: status.value })
    if (res.code === 0) {
      orders.value = res.data || []
      total.value = res.count || 0
    }
  } catch (error) {
    console.error('获取订单列表失败:', error)
  } finally {
    loading.value = false
  }
}

async function handleOp(action: string, tradeNo: string) {
  const actionText = { refund: '退款', freeze: '冻结', unfreeze: '解冻', refresh: '刷新状态' }[action] || '操作'
  let refundMoney: number | undefined
  const targetOrder = orders.value.find(o => o.trade_no === tradeNo)

  if (action === 'refund') {
    const remain = Math.max(0, Number((targetOrder?.money || 0) - ((targetOrder as any)?.refundmoney || 0)))
    if (remain <= 0) {
      ElMessage.warning('可退款金额为0')
      return
    }
    try {
      const { value } = await ElMessageBox.prompt(
        `请输入退款金额（最大 ${remain.toFixed(2)} 元）`,
        '退款确认',
        {
          confirmButtonText: '确认退款',
          cancelButtonText: '取消',
          inputValue: remain.toFixed(2),
          inputPattern: /^(?:0|[1-9]\d*)(?:\.\d{1,2})?$/,
          inputErrorMessage: '请输入合法金额（最多2位小数）'
        }
      )
      refundMoney = Number(value)
      if (!refundMoney || refundMoney <= 0) {
        ElMessage.warning('退款金额必须大于0')
        return
      }
      if (refundMoney > remain) {
        ElMessage.warning(`退款金额不能超过 ${remain.toFixed(2)} 元`)
        return
      }
    } catch {
      return
    }
  }

  const needConfirm = action !== 'refresh' && action !== 'refund'
  if (needConfirm) {
    try {
      await ElMessageBox.confirm(`确定要${actionText}该订单吗？`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }
  }
  try {
    const payload: Record<string, any> = { action, trade_no: tradeNo }
    if (action === 'refund' && refundMoney !== undefined) {
      payload.money = refundMoney
    }
    const res = await orderOp(payload)
    ElMessage.success(res.msg || `${actionText}成功`)
    fetchOrders()
  } catch (error) {
    console.error('操作失败:', error)
  }
}

function formatTime(time: string) {
  if (!time) return '-'
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

onMounted(() => {
  fetchOrders()
})
</script>
