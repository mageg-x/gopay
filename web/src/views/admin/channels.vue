<template>
  <div class="space-y-4">
    <div class="page-head">
      <div>
        <h1 class="page-title no-wrap">通道管理</h1>
        <p class="page-subtitle">配置支付通道和费率</p>
      </div>
      <button @click="showAddModal" class="btn btn-primary">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        添加通道
      </button>
    </div>

    <div class="table-shell">
      <div class="overflow-x-auto">
        <table class="table min-w-[860px] whitespace-nowrap">
          <thead>
            <tr>
              <th class="text-left">ID</th>
              <th class="text-left">通道名称</th>
              <th class="text-left">插件</th>
              <th class="text-left">支付类型</th>
              <th class="text-left">支付方式</th>
              <th class="text-right">费率</th>
              <th class="text-right">成本</th>
              <th>限额</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="ch in channels" :key="ch.id">
              <td class="text-left text-gray-900 font-medium">{{ ch.id }}</td>
              <td class="text-left text-gray-900">{{ ch.name }}</td>
              <td class="text-left text-gray-600">{{ ch.plugin_showname || ch.plugin }}</td>
              <td class="text-left">
                <div class="flex items-center gap-1.5">
                  <SvgIcon :name="typeIcon(ch.type)" :size="16" />
                  <span class="text-sm">{{ typeName(ch.type) }}</span>
                </div>
              </td>
              <td class="text-left text-gray-500 text-xs">
                <div v-if="ch.paymethod_names">{{ ch.paymethod_names }}</div>
                <div v-else class="text-gray-400">未配置</div>
              </td>
              <td class="text-right">
                <span class="font-semibold text-green-600">{{ ch.rate }}%</span>
              </td>
              <td class="text-right text-gray-500">{{ ch.costrate }}%</td>
              <td class="text-gray-500 text-xs">
                <div>￥{{ ch.paymin }} - ￥{{ ch.paymax }}</div>
                <div class="text-gray-400">日限 ￥{{ ch.daytop }}</div>
              </td>
              <td>
                <span :class="['badge', ch.status ? 'badge-success' : 'badge-info']">
                  {{ ch.status ? '开启' : '关闭' }}
                </span>
              </td>
              <td>
                <div class="inline-flex items-center gap-1">
                  <button @click="showEditModal(ch)" class="action-link action-link-primary">编辑</button>
                  <button @click="handleDelete(ch.id)" class="action-link action-link-danger">删除</button>
                  <button @click="toggleStatus(ch)" :class="['action-link', ch.status ? 'action-link-warning' : 'action-link-success']">
                    {{ ch.status ? '关闭' : '开启' }}
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="channels.length === 0">
              <td colspan="10" class="py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                  </svg>
                  <span>暂无通道配置</span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-if="showModal" class="dialog-backdrop">
      <div class="dialog-wrap">
        <div class="dialog-mask" @click="showModal = false"></div>
        <div class="dialog-panel max-w-2xl">
          <div class="dialog-header">
            <div>
              <h3 class="dialog-title">{{ isEdit ? '编辑通道' : '添加通道' }}</h3>
              <p class="dialog-subtitle">设置通道插件、费率和限额策略</p>
            </div>
            <button class="dialog-close" @click="showModal = false">✕</button>
          </div>
          <div class="dialog-body">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-x-6 gap-y-4">
            <div class="col-span-2">
              <label class="form-label">通道名称</label>
              <input v-model="form.name" type="text" placeholder="例如：支付宝通道A"
                class="form-input px-3" />
            </div>
            <div>
              <label class="form-label">支付类型</label>
              <select v-model="form.type" class="form-input px-3">
                <option v-for="pt in payTypes" :key="pt.id" :value="pt.id">
                  {{ pt.showname || pt.name || ('类型' + pt.id) }}
                </option>
              </select>
            </div>
            <div>
              <label class="form-label">插件</label>
              <select v-model="form.plugin" @change="handlePluginChange" class="form-input px-3">
                <option value="">请选择插件</option>
                <option v-for="p in plugins" :key="p.name" :value="p.name">{{ p.showname || p.name }}</option>
              </select>
            </div>
            <div class="col-span-2">
              <label class="form-label">支付方式</label>
              <div v-if="currentPluginMethods.length === 0"
                class="px-3 py-2 text-xs text-gray-500 border border-dashed border-slate-300 rounded-lg bg-slate-50">
                当前插件未提供可选支付方式，留空将按插件默认逻辑路由
              </div>
              <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-2 p-3 border border-slate-200 rounded-lg bg-slate-50">
                <label v-for="[code, label] in currentPluginMethods" :key="code"
                  class="inline-flex items-center gap-2 text-sm text-gray-700">
                  <input v-model="selectedPaymethods" type="checkbox" :value="code" class="rounded border-gray-300" />
                  <span>{{ code }} - {{ label }}</span>
                </label>
              </div>
            </div>
            <div>
              <label class="form-label">通道模式</label>
              <select v-model="form.mode" class="form-input px-3">
                <option :value="0">平台代收</option>
                <option :value="1">商户直清</option>
              </select>
            </div>
            <div>
              <label class="form-label">状态</label>
              <select v-model="form.status" class="form-input px-3">
                <option :value="1">开启</option>
                <option :value="0">关闭</option>
              </select>
            </div>
            <div>
              <label class="form-label">分成比例 (%)</label>
              <input v-model.number="form.rate" type="number" step="0.01"
                class="form-input px-3" />
            </div>
            <div>
              <label class="form-label">成本费率 (%)</label>
              <input v-model.number="form.costrate" type="number" step="0.01"
                class="form-input px-3" />
            </div>
            <div>
              <label class="form-label">单笔最小 (元)</label>
              <input v-model="form.paymin" type="number" class="form-input px-3" />
            </div>
            <div>
              <label class="form-label">单笔最大 (元)</label>
              <input v-model="form.paymax" type="number" class="form-input px-3" />
            </div>
            <div>
              <label class="form-label">单日限额 (元)</label>
              <input v-model.number="form.daytop" type="number" class="form-input px-3" />
            </div>
          </div>
          </div>
          <div class="dialog-footer">
            <button @click="showModal = false" class="btn btn-outline">取消</button>
            <button @click="handleSave" class="btn btn-primary">保存</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { getChannelList, channelOp, getPluginList, getPayTypeList } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'
import SvgIcon from '@/components/svgicon.vue'

const channels = ref<any[]>([])
const plugins = ref<any[]>([])
const payTypes = ref<any[]>([])
const selectedPaymethods = ref<string[]>([])
const showModal = ref(false)
const isEdit = ref(false)
const form = ref({
  id: 0,
  name: '',
  plugin: '',
  type: 0,
  mode: 0,
  rate: 0.5,
  costrate: 0.3,
  daytop: 100000,
  paymin: 10,
  paymax: 5000,
  paymethod: '',
  status: 1
})

const currentPluginMethods = computed<[string, string][]>(() => {
  const current = plugins.value.find((p: any) => p.name === form.value.plugin)
  const selectMap: Record<string, string> = (current?.select || {}) as Record<string, string>
  return Object.entries(selectMap).sort((a, b) => Number(a[0]) - Number(b[0]))
})

function parsePaymethod(value: string): string[] {
  if (!value) return []
  return value
    .split(',')
    .map((item) => item.trim())
    .filter((item) => item.length > 0)
}

function handlePluginChange() {
  if (!form.value.plugin) {
    selectedPaymethods.value = []
    return
  }
  const available = new Set(currentPluginMethods.value.map(([code]) => code))
  selectedPaymethods.value = selectedPaymethods.value.filter((code) => available.has(code))
}

async function fetchChannels() {
  try {
    const res = await getChannelList()
    if (res.code === 0) {
      console.log('channels response:', res.data)
      channels.value = res.data || []
    }
  } catch (error) {
    console.error('获取通道列表失败:', error)
  }
}

async function fetchPlugins() {
  try {
    const res = await getPluginList()
    if (res.code === 0) {
      plugins.value = res.data || []
    }
  } catch (error) {
    console.error('获取插件列表失败:', error)
  }
}

async function fetchPayTypes() {
  try {
    const res = await getPayTypeList()
    if (res.code === 0) {
      payTypes.value = (res.data || []).map((item: any) => ({
        ...item,
        id: Number(item.id)
      }))
    }
  } catch (error) {
    console.error('获取支付类型失败:', error)
  }
}

function showAddModal() {
  if (payTypes.value.length === 0) {
    ElMessage.warning('请先在支付方式管理中添加支付方式')
    return
  }
  isEdit.value = false
  selectedPaymethods.value = []
  form.value = {
    id: 0,
    name: '',
    plugin: '',
    type: Number(payTypes.value[0].id),
    mode: 0,
    rate: 0.5,
    costrate: 0.3,
    daytop: 100000,
    paymin: 10,
    paymax: 5000,
    paymethod: '',
    status: 1
  }
  showModal.value = true
}

function showEditModal(ch: any) {
  isEdit.value = true
  selectedPaymethods.value = parsePaymethod(ch.paymethod || '')
  form.value = {
    id: ch.id,
    name: ch.name,
    plugin: ch.plugin,
    type: ch.type,
    mode: ch.mode,
    rate: ch.rate,
    costrate: ch.costrate,
    daytop: ch.daytop,
    paymin: Number(ch.paymin) || 10,
    paymax: Number(ch.paymax) || 5000,
    paymethod: ch.paymethod || '',
    status: ch.status
  }
  // 确保插件列表已加载
  if (plugins.value.length === 0) {
    fetchPlugins()
  }
  console.log('edit modal - ch.plugin:', ch.plugin, 'plugins:', plugins.value)
  showModal.value = true
}

async function handleSave() {
  if (!form.value.name) {
    ElMessage.warning('请输入通道名称')
    return
  }
  if (!form.value.plugin) {
    ElMessage.warning('请选择插件')
    return
  }
  const paymethod = [...new Set(selectedPaymethods.value)].join(',')
  if (currentPluginMethods.value.length > 0 && !paymethod) {
    ElMessage.warning('请至少选择一种支付方式')
    return
  }
  try {
    const res = await channelOp({
      action: isEdit.value ? 'edit' : 'add',
      ...form.value,
      paymethod
    })
    ElMessage.success(res.msg || '保存成功')
    showModal.value = false
    fetchChannels()
  } catch (error) {
    console.error('保存失败:', error)
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确定要删除这个通道吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    const res = await channelOp({ action: 'delete', id })
    ElMessage.success(res.msg || '删除成功')
    fetchChannels()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

async function toggleStatus(ch: any) {
  try {
    const res = await channelOp({
      action: 'set_status',
      id: ch.id,
      status: ch.status ? 0 : 1
    })
    ElMessage.success(res.msg || '状态已更新')
    fetchChannels()
  } catch (error) {
    console.error('更新状态失败:', error)
  }
}

function typeName(type: number) {
  const pt = payTypes.value.find((item: any) => Number(item.id) === Number(type))
  if (!pt) return '未知'
  return pt.showname || pt.name || `类型${type}`
}

function typeIcon(type: number) {
  const pt = payTypes.value.find((item: any) => Number(item.id) === Number(type))
  const key = (pt?.name || '').toLowerCase()
  const map: Record<string, string> = {
    alipay: 'alipay',
    wechatpay: 'wechatpay',
    wxpay: 'wechatpay',
    qqpay: 'qqpay',
    unionpay: 'unionpay',
    jdpay: 'jdpay'
  }
  return map[key] || 'creditcard'
}

onMounted(() => {
  fetchPayTypes()
  fetchChannels()
  fetchPlugins()
})
</script>
