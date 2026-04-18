<template>
  <div class="space-y-4">
    <div class="page-head">
      <div>
        <h2 class="page-title no-wrap">资金记录</h2>
        <p class="page-subtitle">查看余额变动流水</p>
      </div>
    </div>

    <div class="table-shell">
      <div class="table-shell-body overflow-x-auto">
        <table class="table min-w-[760px] whitespace-nowrap">
          <thead>
            <tr>
              <th class="text-left">类型</th>
              <th class="text-right">金额</th>
              <th class="text-right">余额</th>
              <th class="text-left">关联订单</th>
              <th class="text-left">时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in records" :key="r.id">
              <td class="text-left">{{ actionMap[r.action] || '未知' }}</td>
              <td class="text-right" :class="r.money >= 0 ? 'text-emerald-600' : 'text-rose-600'">
                {{ r.money >= 0 ? '+' : '' }}{{ r.money }}
              </td>
              <td class="text-right">¥{{ r.newmoney }}</td>
              <td class="text-left font-mono text-xs">{{ r.trade_no || '-' }}</td>
              <td class="text-left text-xs">{{ dayjs(r.date).format('YYYY-MM-DD HH:mm') }}</td>
            </tr>
            <tr v-if="records.length === 0">
              <td colspan="5" class="text-center text-gray-500 py-8">暂无资金记录</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getUserRecords } from '@/api/user'
import dayjs from 'dayjs'

const records = ref<any[]>([])
const loading = ref(false)

const actionMap: Record<number, string> = {
  1: '订单收入',
  2: '结算扣款',
  3: '转账',
  4: '退款',
  5: '后台加款',
  6: '后台扣款',
  7: '邀请返现',
  8: '结算返还',
  9: '转账退款'
}

async function fetchRecords() {
  loading.value = true
  try {
    const res = await getUserRecords({ page: 1, limit: 50 })
    if (res.code === 0) {
      records.value = res.data || []
    }
  } catch (error) {
    console.error('获取资金记录失败:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchRecords()
})
</script>
