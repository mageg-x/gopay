<template>
  <div class="space-y-4">
    <div class="page-head">
      <div>
        <h1 class="page-title no-wrap">数据导出</h1>
        <p class="page-subtitle">导出订单数据为CSV文件</p>
      </div>
    </div>

    <div class="card">
      <div class="card-body">
      <h3 class="text-lg font-semibold text-gray-800 mb-4 no-wrap">导出配置</h3>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <label class="form-label">日期范围</label>
          <div class="flex items-center gap-2">
            <input v-model="form.start_date" type="date"
              class="form-input flex-1" />
            <span class="text-gray-400">至</span>
            <input v-model="form.end_date" type="date"
              class="form-input flex-1" />
          </div>
        </div>

        <div>
          <label class="form-label">商户筛选</label>
          <input v-model="form.uid" type="number" placeholder="输入商户ID，不填则导出全部"
            class="form-input" />
        </div>

        <div>
          <label class="form-label">订单状态</label>
          <select v-model="form.status"
            class="form-input">
            <option value="">全部状态</option>
            <option value="0">待支付</option>
            <option value="1">已支付</option>
            <option value="2">已退款</option>
            <option value="3">已冻结</option>
          </select>
        </div>

        <div>
          <label class="form-label">支付方式</label>
          <select v-model="form.type"
            class="form-input">
            <option value="">全部方式</option>
            <option value="1">支付宝</option>
            <option value="2">微信支付</option>
            <option value="3">QQ钱包</option>
            <option value="4">银行卡</option>
          </select>
        </div>
      </div>

      <div class="mt-6 pt-4 border-t flex items-center gap-4">
        <button @click="handleExport" class="btn btn-primary">
          <Download class="w-4 h-4" />
          导出CSV
        </button>
        <span class="text-sm text-gray-500">导出文件格式为 .csv，每次最多导出10万条记录</span>
      </div>
      </div>
    </div>

    <div class="table-shell">
      <div class="card-header">
        <h3 class="font-semibold text-gray-700">导出记录</h3>
      </div>
      <div class="table-shell-body overflow-x-auto">
        <table class="table min-w-[860px] whitespace-nowrap">
          <thead>
            <tr>
              <th class="text-left">文件名</th>
              <th class="text-left">记录数</th>
              <th class="text-left">时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="record in records" :key="record.id">
              <td class="text-left text-gray-900">{{ record.filename }}</td>
              <td class="text-left text-gray-600">{{ record.count }} 条</td>
              <td class="text-left text-gray-500 text-xs">{{ record.time }}</td>
              <td>
                <button @click="downloadFile(record)" class="action-link action-link-primary">下载</button>
              </td>
            </tr>
            <tr v-if="records.length === 0">
              <td colspan="4" class="py-8 text-center text-gray-400">
                暂无导出记录
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { exportOrders } from '@/api/admin'
import { ElMessage } from 'element-plus'
import { Download } from 'lucide-vue-next'

const form = ref({
  start_date: '',
  end_date: '',
  uid: '',
  status: '',
  type: ''
})

const records = ref<any[]>([])

async function handleExport() {
  if (!form.value.start_date || !form.value.end_date) {
    ElMessage.warning('请选择日期范围')
    return
  }
  if (form.value.start_date > form.value.end_date) {
    ElMessage.warning('开始日期不能晚于结束日期')
    return
  }

  try {
    ElMessage.info('正在生成导出文件，请稍候...')

    const res = await exportOrders({
      start_date: form.value.start_date,
      end_date: form.value.end_date,
      uid: form.value.uid || undefined,
      status: form.value.status || undefined,
      type: form.value.type || undefined,
      limit: 100000
    })

    if (res.code === 0) {
      const data = res.data || []
      const csv = generateCSV(data)
      const filename = `订单导出_${form.value.start_date}_${form.value.end_date}.csv`
      downloadCSV(csv, filename)
      ElMessage.success('导出成功')

      // 添加到记录
      records.value.unshift({
        id: Date.now(),
        filename,
        count: data.length,
        time: new Date().toLocaleString('zh-CN'),
        content: csv
      })
    }
  } catch (error) {
    console.error('导出失败:', error)
    ElMessage.error('导出失败')
  }
}

function generateCSV(data: any[]) {
  const headers = ['订单号', '商户订单号', '商户ID', '商品名称', '支付方式', '金额', '实付金额', '商户所得', '状态', '创建时间', '支付时间', '回调状态', 'IP']
  const rows = data.map((o: any) => [
    o.trade_no,
    o.out_trade_no || '',
    o.uid,
    o.name,
    typeName(o.type),
    o.money,
    o.realmoney || o.money,
    o.getmoney || '0',
    statusName(o.status),
    o.addtime,
    o.endtime || '',
    o.notify === 1 ? '已回调' : '未回调',
    o.ip || ''
  ])

  const escapeCell = (value: unknown) => {
    const text = String(value ?? '')
    return `"${text.replace(/"/g, '""')}"`
  }

  return [headers.map(escapeCell).join(','), ...rows.map(r => r.map(escapeCell).join(','))].join('\n')
}

function downloadCSV(content: string, filename: string) {
  const BOM = '\uFEFF'
  const blob = new Blob([BOM + content], { type: 'text/csv;charset=utf-8' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = filename
  link.click()
  URL.revokeObjectURL(link.href)
}

function downloadFile(record: any) {
  if (!record?.content) {
    ElMessage.warning('该记录无可下载内容')
    return
  }
  downloadCSV(record.content, record.filename)
}

function typeName(type: number) {
  const map: Record<number, string> = {
    1: '支付宝',
    2: '微信支付',
    3: 'QQ钱包',
    4: '银行卡'
  }
  return map[type] || '其他'
}

function statusName(status: number) {
  const map: Record<number, string> = {
    0: '待支付',
    1: '已支付',
    2: '已退款',
    3: '已冻结'
  }
  return map[status] || '未知'
}
</script>
