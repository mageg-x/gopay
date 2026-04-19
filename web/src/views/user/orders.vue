<template>
  <div class="space-y-4">
    <div class="page-head">
      <div>
        <h2 class="page-title no-wrap">我的订单</h2>
        <p class="page-subtitle">查看订单状态与交易处理结果</p>
      </div>
    </div>

    <div class="panel-filter">
      <div class="card-body toolbar-wrap">
        <select v-model="filterStatus" class="form-input w-auto min-w-[128px] px-3">
          <option value="">全部状态</option>
          <option value="0">待支付</option>
          <option value="1">已支付</option>
          <option value="2">已退款</option>
          <option value="3">已冻结</option>
        </select>
        <input v-model="searchTradeNo" type="text" placeholder="订单号" class="form-input w-full md:w-64 px-3" />
        <button @click="page = 1; fetchOrders()" class="btn btn-primary">
          搜索
        </button>
      </div>
    </div>

    <div class="table-shell">
      <div class="table-shell-body">
        <div class="overflow-x-auto">
          <table class="table min-w-[760px] whitespace-nowrap">
          <thead>
            <tr>
              <th>订单号</th>
              <th>商品名称</th>
              <th>支付方式</th>
              <th>金额</th>
              <th>状态</th>
              <th>时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="order in orders" :key="order.trade_no">
              <td class="text-xs font-mono text-left">{{ order.trade_no }}</td>
              <td class="text-left">{{ order.name }}</td>
              <td>
                <div class="flex items-center justify-center gap-1.5">
                  <SvgIcon :name="payIcon(order)" :size="16" />
                  <span class="text-sm font-medium" :class="payTextClass(order)">{{ typeName(order) }}</span>
                </div>
              </td>
              <td class="font-semibold text-emerald-600 text-right">¥{{ order.money }}</td>
              <td>
                <span v-if="order.status === 1"
                  class="badge badge-success">
                  已支付
                </span>
                <span v-else-if="order.status === 0"
                  class="badge badge-warning">
                  待支付
                </span>
                <span v-else-if="order.status === 2"
                  class="badge badge-info">
                  已退款
                </span>
                <span v-else-if="order.status === 3"
                  class="badge badge-danger">
                  已冻结
                </span>
                <span v-else
                  class="badge">
                  未知
                </span>
              </td>
              <td class="text-left">{{ dayjs(order.addtime).format('YYYY-MM-DD HH:mm') }}</td>
              <td>
                <button @click="showDetail(order)" class="action-link action-link-primary">
                  详情
                </button>
              </td>
            </tr>
            <tr v-if="orders.length === 0">
              <td colspan="7" class="text-center text-gray-500 py-10">暂无订单</td>
            </tr>
          </tbody>
          </table>
        </div>

        <div class="flex flex-wrap items-center justify-between mt-4 gap-2">
          <div class="text-sm text-gray-500">共 {{ total }} 条</div>
          <div class="flex items-center gap-2">
            <button class="pagination-item no-wrap" :disabled="page === 1" @click="page--; fetchOrders()">上一页</button>
            <span class="px-4">{{ page }} / {{ totalPages }}</span>
            <button class="pagination-item no-wrap" :disabled="page >= totalPages" @click="page++; fetchOrders()">下一页</button>
          </div>
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
              <p class="dialog-subtitle">查看订单支付、回调和退款信息</p>
            </div>
            <button class="dialog-close" @click="detailVisible = false">✕</button>
          </div>

          <div class="dialog-body" v-if="currentOrder">
            <div class="grid grid-cols-2 gap-2">
              <div class="text-gray-500">订单号:</div>
              <div class="font-mono text-gray-900">{{ currentOrder.trade_no }}</div>
              <div class="text-gray-500">商户订单号:</div>
              <div class="font-mono text-gray-900">{{ currentOrder.out_trade_no || '-' }}</div>
              <div class="text-gray-500">商品名称:</div>
              <div class="text-gray-900">{{ currentOrder.name }}</div>
              <div class="text-gray-500">支付方式:</div>
              <div class="text-gray-900">{{ typeName(currentOrder) }}</div>
              <div class="text-gray-500">订单金额:</div>
              <div class="font-bold text-emerald-600">¥{{ currentOrder.money }}</div>
              <div class="text-gray-500">平台实收:</div>
              <div class="font-bold text-emerald-600">¥{{ currentOrder.realmoney || currentOrder.money }}</div>
              <div class="text-gray-500">商户所得:</div>
              <div class="text-blue-600">¥{{ currentOrder.getmoney || '-' }}</div>
              <div class="text-gray-500">状态:</div>
              <div>
                <span v-if="currentOrder.status === 1" class="text-emerald-600">已支付</span>
                <span v-else-if="currentOrder.status === 0" class="text-amber-600">待支付</span>
                <span v-else-if="currentOrder.status === 2" class="text-blue-600">已退款</span>
                <span v-else-if="currentOrder.status === 3" class="text-red-600">已冻结</span>
                <span v-else class="text-gray-500">未知</span>
              </div>
              <div class="text-gray-500">创建时间:</div>
              <div>{{ dayjs(currentOrder.addtime).format('YYYY-MM-DD HH:mm:ss') }}</div>
              <div class="text-gray-500">支付时间:</div>
              <div>{{ currentOrder.endtime ? dayjs(currentOrder.endtime).format('YYYY-MM-DD HH:mm:ss') : '-' }}</div>
              <div class="text-gray-500">回调状态:</div>
              <div>
                <span v-if="currentOrder.notify === 1" class="text-emerald-600">已回调</span>
                <span v-else class="text-amber-600">未回调</span>
              </div>
              <div class="text-gray-500">订单类型:</div>
              <div>{{ Number(currentOrder.tid || 0) === 2 ? '余额充值' : '普通订单' }}</div>
            </div>

            <div v-if="currentOrder.param" class="border-t pt-3 mt-3">
              <div class="text-gray-500 mb-1">订单备注:</div>
              <div class="text-gray-900">{{ currentOrder.param }}</div>
            </div>
          </div>

          <div class="dialog-footer">
            <button @click="handleNotify(currentOrder)" v-if="currentOrder?.status === 1" class="btn btn-primary">
              重新通知
            </button>
            <button @click="handleRefund(currentOrder)" v-if="currentOrder?.status === 1" class="btn btn-danger">
              退款
            </button>
            <button @click="detailVisible = false" class="btn btn-outline">
              关闭
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="refundVisible" class="dialog-backdrop">
      <div class="dialog-wrap">
        <div class="dialog-mask" @click="refundVisible = false"></div>
        <div class="dialog-panel max-w-md">
          <div class="dialog-header">
            <div>
              <h3 class="dialog-title">订单退款</h3>
              <p class="dialog-subtitle">请输入本次退款金额</p>
            </div>
            <button class="dialog-close" @click="refundVisible = false">✕</button>
          </div>
          <div class="dialog-body space-y-4">
            <div class="section-card">
              <p class="text-amber-800 text-sm">订单号: {{ refundForm.trade_no }}</p>
              <p class="text-amber-800 text-sm mt-1">订单金额: ¥{{ refundForm.money }}</p>
            </div>
            <div>
              <label class="form-label">退款金额</label>
              <input v-model="refundForm.amount" type="number" step="0.01"
                class="form-input px-3"
                placeholder="请输入退款金额" />
            </div>
          </div>
          <div class="dialog-footer">
            <button @click="refundVisible = false" class="btn btn-outline">取消</button>
            <button @click="submitRefund" class="btn btn-danger">确认退款</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getUserInfo, getUserOrders, userOrderOp } from '@/api/user'
import { ElMessage, ElMessageBox } from 'element-plus'
import dayjs from 'dayjs'
import { useAppStore } from '@/stores/app'
import SvgIcon from '@/components/svgicon.vue'

const appStore = useAppStore()

const orders = ref<any[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const filterStatus = ref('')
const searchTradeNo = ref('')
const detailVisible = ref(false)
const refundVisible = ref(false)
const currentOrder = ref<any>(null)

const refundForm = ref({
  trade_no: '',
  money: 0,
  amount: ''
})

const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

function typeName(order: any) {
  const typename = String(order?.typename || '').trim()
  if (typename) return typename
  const map: Record<number, string> = {
    1: '支付宝',
    2: '微信支付',
    3: 'QQ钱包',
    4: '银行卡'
  }
  return map[Number(order?.type)] || '其他'
}

function payIcon(order: any) {
  const name = typeName(order)
  if (name.includes('支付宝')) return 'alipay'
  if (name.includes('微信')) return 'wechatpay'
  if (name.includes('银行卡')) return 'bankcard'
  return 'bankcard'
}

function payTextClass(order: any) {
  const name = typeName(order)
  if (name.includes('支付宝')) return 'text-blue-600'
  if (name.includes('微信')) return 'text-green-600'
  if (name.includes('银行卡')) return 'text-gray-700'
  return 'text-gray-600'
}

async function fetchOrders() {
  try {
    const params: any = { page: page.value, limit: pageSize.value }
    if (filterStatus.value !== '') {
      params.status = filterStatus.value
    }
    if (searchTradeNo.value) {
      params.trade_no = searchTradeNo.value.trim()
    }
    const res = await getUserOrders(params)
    orders.value = Array.isArray(res.data) ? res.data : []
    total.value = res.count || 0
  } catch (error) {
    console.error('获取订单列表失败:', error)
    ElMessage.error((error as any)?.message || '获取订单列表失败')
  }
}

async function initUser() {
  const res = await getUserInfo()
  if (!res.data?.uid) {
    throw new Error('未获取到当前商户信息')
  }
  appStore.userInfo = {
    ...(appStore.userInfo || {
      uid: Number(res.data.uid),
      username: '',
      email: '',
      phone: '',
      money: 0,
      status: 1
    }),
    uid: Number(res.data.uid)
  }
}

function showDetail(order: any) {
  currentOrder.value = order
  detailVisible.value = true
}

async function handleNotify(order: any) {
  try {
    await ElMessageBox.confirm('确定要重新通知该订单吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }

  try {
    const res = await userOrderOp({ action: 'notify', trade_no: order.trade_no })
    ElMessage.success(res.msg || '已触发重新通知')
    fetchOrders()
  } catch (error) {
    console.error('重新通知失败:', error)
  }
}

function handleRefund(order: any) {
  const remain = Math.max(0, Number(order.money || 0) - Number(order.refundmoney || 0))
  if (remain <= 0) {
    ElMessage.warning('可退款金额为0')
    return
  }
  refundForm.value = {
    trade_no: order.trade_no,
    money: order.money,
    amount: remain.toFixed(2)
  }
  detailVisible.value = false
  refundVisible.value = true
}

async function submitRefund() {
  const amount = parseFloat(refundForm.value.amount)
  const remain = Math.max(0, Number(refundForm.value.money || 0) - Number(currentOrder.value?.refundmoney || 0))
  if (isNaN(amount) || amount <= 0) {
    ElMessage.warning('请输入有效的退款金额')
    return
  }
  if (amount > remain) {
    ElMessage.warning(`退款金额不能超过可退金额 ${remain.toFixed(2)} 元`)
    return
  }
  try {
    await ElMessageBox.confirm(`确定要退款 ¥${amount} 吗？`, '退款确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await userOrderOp({
      action: 'refund',
      trade_no: refundForm.value.trade_no,
      money: amount
    })
    ElMessage.success(res.msg || '退款成功')
    refundVisible.value = false
    fetchOrders()
  } catch {
    return
  }
}

onMounted(async () => {
  try {
    await initUser()
    await fetchOrders()
  } catch (error: any) {
    ElMessage.error(error?.message || '初始化失败')
  }
})
</script>
