<template>
  <div class="space-y-4">
    <div class="page-head">
      <div>
        <h1 class="page-title no-wrap">接口下单测试</h1>
        <p class="page-subtitle">商户后台模拟调用平台 OpenAPI（PID + API Key + Sign）进行下单与查单</p>
      </div>
    </div>

    <div class="card">
      <div class="card-body space-y-4">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label class="form-label">当前商户ID</label>
          <input
            :value="pidText"
            type="text"
            readonly
            class="form-input bg-gray-50 text-gray-600"
          />
        </div>
        <div>
          <label class="form-label">支付类型</label>
          <select
            v-model.number="form.type"
            class="form-input"
          >
            <option :value="0">请选择支付类型</option>
            <option v-for="pt in payTypes" :key="pt.id" :value="Number(pt.id)">
              {{ pt.showname || pt.name || ('类型' + pt.id) }}
            </option>
          </select>
        </div>
        <div>
          <label class="form-label">API密钥（手动填写）</label>
          <input
            v-model.trim="apiKey"
            type="text"
            placeholder="请输入商户 API Key（仅本页签名使用）"
            class="form-input"
          />
        </div>
        <div>
          <label class="form-label">设备标识（可选）</label>
          <select
            v-model="form.device"
            class="form-input"
          >
            <option value="">自动</option>
            <option value="pc">PC</option>
            <option value="mobile">MOBILE</option>
          </select>
        </div>
        <div>
          <label class="form-label">金额 (元)</label>
          <input
            v-model.number="form.money"
            type="number"
            min="0.01"
            step="0.01"
            class="form-input"
          />
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label class="form-label">商户订单号</label>
          <div class="flex gap-2">
            <input
              v-model.trim="form.out_trade_no"
              type="text"
              class="form-input flex-1"
            />
            <button
              class="btn btn-outline text-xs px-3"
              @click="regenerateOutTradeNo"
            >
              重新生成
            </button>
          </div>
        </div>
        <div>
          <label class="form-label">商品名称</label>
          <input
            v-model.trim="form.name"
            type="text"
            class="form-input"
          />
        </div>
      </div>
      <div class="text-xs text-gray-500">
        回调地址已自动处理：`notify_url` 固定为空，`return_url` 固定为当前站点根路径。
      </div>

      <div class="toolbar-wrap pt-1">
        <button
          class="btn btn-outline"
          @click="loadPayTypes"
        >
          刷新支付类型
        </button>
        <button
          class="btn btn-primary disabled:opacity-60"
          :disabled="submitting || !pid || !apiKey"
          @click="submitTest"
        >
          {{ submitting ? '下单中...' : '提交接口测试' }}
        </button>
        <button
          class="btn btn-success disabled:opacity-60"
          :disabled="!hasPayAction"
          @click="openPayAction"
        >
          打开支付页
        </button>
        <button
          class="btn btn-warning disabled:opacity-60"
          :disabled="!tradeNo && !form.out_trade_no"
          @click="queryOrder"
        >
          查询订单状态
        </button>
      </div>
      </div>
    </div>

    <div v-if="tradeNo" class="card">
      <div class="card-body space-y-4">
      <div class="text-sm text-gray-600">
        平台订单号：
        <span class="font-mono text-gray-900">{{ tradeNo }}</span>
      </div>

      <div v-if="qrCodeUrl" class="text-center">
        <img :src="qrCodeUrl" alt="支付二维码" class="mx-auto w-56 h-56 border rounded-lg p-2 bg-white" />
      </div>

      <div v-if="htmlPayload" class="text-xs text-orange-600 bg-orange-50 border border-orange-200 rounded-lg px-3 py-2">
        当前为 HTML 表单支付，点击“打开支付页”可在新窗口自动提交。
      </div>

      <div v-if="payUrl" class="text-xs text-blue-600 break-all bg-blue-50 border border-blue-200 rounded-lg px-3 py-2">
        {{ payUrl }}
      </div>

      <div>
        <h3 class="text-sm font-semibold text-gray-700 mb-2">提交返回</h3>
        <pre class="bg-gray-900 text-gray-100 rounded-lg p-3 text-xs overflow-auto">{{ prettyResult }}</pre>
      </div>
      </div>
    </div>

    <div v-if="orderInfo" class="card">
      <div class="card-body space-y-2">
      <h3 class="text-sm font-semibold text-gray-700">订单查询结果</h3>
      <div class="text-sm text-gray-700">商户订单号：{{ orderInfo.out_trade_no }}</div>
      <div class="text-sm text-gray-700">金额：{{ orderInfo.money }}</div>
      <div class="text-sm text-gray-700">
        状态：
        <span class="font-semibold">{{ statusText(orderInfo.status) }}</span>
      </div>
      <div class="text-sm text-gray-700">支付类型：{{ orderInfo.type }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import QRCode from 'qrcode'
import { getPayTypes, payCreate, payQuery } from '@/api/pay'
import { getUserInfo } from '@/api/user'
import { makeOpenAPISign } from '@/utils/sign'

const pid = ref<number>(0)
const apiKey = ref('')
const payTypes = ref<any[]>([])
const submitting = ref(false)
const tradeNo = ref('')
const submitResult = ref<any>(null)
const orderInfo = ref<any>(null)
const qrCodeUrl = ref('')
const payUrl = ref('')
const htmlPayload = ref('')

const form = ref({
  type: 0,
  out_trade_no: '',
  name: '商户接口测试订单',
  money: 0.01,
  param: '',
  device: ''
})

const pidText = computed(() => (pid.value ? String(pid.value) : '未获取到'))
const prettyResult = computed(() => JSON.stringify(submitResult.value || {}, null, 2))
const hasPayAction = computed(() => Boolean(payUrl.value || htmlPayload.value))

function regenerateOutTradeNo() {
  const rand = Math.random().toString(36).slice(2, 8).toUpperCase()
  form.value.out_trade_no = `USRTEST_${Date.now()}_${rand}`
}

function clearPayDisplay() {
  qrCodeUrl.value = ''
  payUrl.value = ''
  htmlPayload.value = ''
}

function normalizePayHtml(raw: string) {
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

async function renderSubmitResult(result: any) {
  clearPayDisplay()
  if (!result) return

  const resultType = String(result.type || result.Type || '').toLowerCase()
  const resultURL = result.url || result.URL || ''
  const resultData = result.data ?? result.Data

  if (resultType === 'qrcode' && resultURL) {
    qrCodeUrl.value = await QRCode.toDataURL(resultURL, { width: 220, margin: 1 })
    return
  }

  if (resultType === 'jump' && resultURL) {
    payUrl.value = resultURL
    return
  }

  if (resultType === 'html' && typeof resultData === 'string') {
    htmlPayload.value = resultData
    return
  }

  if (resultURL) {
    payUrl.value = resultURL
  }
}

async function loadPayTypes() {
  if (!pid.value) return
  try {
    const res = await getPayTypes(pid.value)
    payTypes.value = res.data || []
    if (payTypes.value.length === 0) {
      form.value.type = 0
      return
    }

    if (!form.value.type) {
      form.value.type = Number(payTypes.value[0].id)
    }
  } catch (error: any) {
    ElMessage.error(error?.message || '加载支付类型失败')
  }
}

async function submitTest() {
  if (!form.value.type) {
    ElMessage.warning('请选择支付类型')
    return
  }
  if (!apiKey.value) {
    ElMessage.warning('请输入 API 密钥')
    return
  }
  if (!form.value.out_trade_no) {
    regenerateOutTradeNo()
  }
  if (!form.value.money || Number(form.value.money) <= 0) {
    ElMessage.warning('金额必须大于 0')
    return
  }

  submitting.value = true
  orderInfo.value = null
  clearPayDisplay()

  try {
    const createParams = {
      pid: pid.value,
      type: Number(form.value.type),
      out_trade_no: form.value.out_trade_no,
      name: form.value.name,
      money: Number(form.value.money),
      notify_url: '',
      return_url: `${window.location.origin}/`,
      clientip: '',
      device: form.value.device || '',
      param: form.value.param || ''
    }

    const sign = await makeOpenAPISign(createParams, apiKey.value)
    const res = await payCreate({
      ...createParams,
      sign,
      sign_type: 'HMAC-SHA256'
    })

    tradeNo.value = res.trade_no || ''
    submitResult.value = res.result || {
      Type: res.pay_type || '',
      URL: res.pay_info || '',
      Data: res.pay_data || '',
      Msg: ''
    }

    await renderSubmitResult(submitResult.value)
    ElMessage.success('接口测试下单成功')
  } catch (error: any) {
    ElMessage.error(error?.message || '下单失败')
  } finally {
    submitting.value = false
  }
}

function openPayAction() {
  if (payUrl.value) {
    const u = String(payUrl.value || '').trim()
    if (!/^https:\/\//i.test(u) && !/^http:\/\/localhost(?::\d+)?\//i.test(u) && !/^http:\/\/127\.0\.0\.1(?::\d+)?\//i.test(u)) {
      ElMessage.error('支付链接不安全，已阻止打开')
      return
    }
    window.open(u, '_blank')
    return
  }

  if (htmlPayload.value) {
    const win = window.open('', '_blank')
    if (!win) {
      ElMessage.error('浏览器阻止了新窗口，请允许弹窗后重试')
      return
    }
    const html = normalizePayHtml(htmlPayload.value)
    win.document.open()
    win.document.write(html)
    win.document.close()
    return
  }

  ElMessage.warning('当前返回结果没有可直接打开的支付链接')
}

async function queryOrder() {
  if (!pid.value) {
    ElMessage.warning('未获取到商户信息')
    return
  }
  if (!apiKey.value) {
    ElMessage.warning('请输入 API 密钥')
    return
  }

  try {
    const queryParams = {
      pid: pid.value,
      trade_no: tradeNo.value || undefined,
      out_trade_no: tradeNo.value ? undefined : form.value.out_trade_no
    }
    const sign = await makeOpenAPISign(queryParams, apiKey.value)
    const res = await payQuery({
      ...queryParams,
      sign,
      sign_type: 'HMAC-SHA256'
    })

    orderInfo.value = res
    ElMessage.success('查询成功')
  } catch (error: any) {
    ElMessage.error(error?.message || '查询失败')
  }
}

function statusText(status: number) {
  const map: Record<number, string> = {
    0: '待支付',
    1: '已支付',
    2: '已退款',
    3: '已冻结'
  }
  return map[status] || `未知(${status})`
}

async function initUser() {
  const res = await getUserInfo()
  if (res.code === 0 && res.data?.uid) {
    pid.value = Number(res.data.uid)
    return
  }
  throw new Error('未获取到当前商户ID')
}

onMounted(async () => {
  regenerateOutTradeNo()

  try {
    await initUser()
    await loadPayTypes()
  } catch (error: any) {
    ElMessage.error(error?.message || '初始化失败')
  }
})
</script>
