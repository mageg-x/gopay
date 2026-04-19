<template>
  <div class="space-y-4">
    <div class="page-head">
      <div>
        <h1 class="page-title no-wrap">操作日志</h1>
        <p class="page-subtitle">查看商户操作记录</p>
      </div>
    </div>

    <div class="panel-filter">
      <div class="card-body">
        <div class="toolbar-wrap">
          <input v-model="searchUid" type="number" placeholder="商户ID" class="form-input w-[180px]"
            @keyup.enter="page = 1; fetchList()" />
          <button @click="page = 1; fetchList()" class="btn btn-primary">搜索</button>
        </div>
      </div>
    </div>

    <div class="table-shell">
      <div class="table-shell-body overflow-x-auto">
        <table class="table min-w-[860px] whitespace-nowrap">
          <thead>
            <tr>
              <th class="text-left">ID</th>
              <th>商户</th>
              <th class="text-left">操作</th>
              <th class="text-left">详情</th>
              <th class="text-left">IP</th>
              <th class="text-left">时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="l in list" :key="l.id">
              <td class="text-left text-gray-900">{{ l.id }}</td>
              <td>
                <span class="font-medium">{{ l.user_name || l.uid }}</span>
              </td>
              <td class="text-left">
                <span :class="typeClass(l.type)">
                  {{ typeName(l.type) }}
                </span>
              </td>
              <td class="text-left text-gray-600 max-w-xs truncate">{{ l.content }}</td>
              <td class="text-left text-gray-500 font-mono text-xs">{{ l.ip }}</td>
              <td class="text-left text-gray-500 text-xs">{{ formatTime(l.time) }}</td>
            </tr>
            <tr v-if="list.length === 0">
              <td colspan="6" class="py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                  </svg>
                  <span>暂无操作日志</span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="px-4 py-3 border-t border-slate-200/70 flex flex-wrap items-center justify-between gap-2">
        <div class="text-sm text-gray-500">共 {{ total }} 条</div>
        <div class="flex items-center gap-2">
          <button @click="page--; fetchList()" :disabled="page <= 1"
            class="pagination-item disabled:opacity-50 disabled:cursor-not-allowed">上一页</button>
          <span class="px-3 py-1 text-sm">{{ page }} / {{ totalPages }}</span>
          <button @click="page++; fetchList()" :disabled="page >= totalPages"
            class="pagination-item disabled:opacity-50 disabled:cursor-not-allowed">下一页</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { logList } from '@/api/admin'

const list = ref<any[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const searchUid = ref('')

const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

function typeName(type: string) {
  const map: Record<string, string> = {
    'login': '登录',
    'logout': '登出',
    'order': '订单操作',
    'settle': '结算操作',
    'edit': '资料修改',
    'recharge': '余额操作',
    'other': '其他'
  }
  return map[type] || type
}

function typeClass(type: string) {
  const map: Record<string, string> = {
    'login': 'inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-700',
    'logout': 'inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-700',
    'order': 'inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-100 text-green-700',
    'settle': 'inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-purple-100 text-purple-700',
    'edit': 'inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-orange-100 text-orange-700',
    'recharge': 'inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-100 text-red-700',
    'other': 'inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-700'
  }
  return map[type] || map['other']
}

function formatTime(time: string) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

async function fetchList() {
  try {
    const params: any = { page: page.value, limit: pageSize.value }
    if (searchUid.value) {
      params.uid = searchUid.value
    }
    const res = await logList(params)
    if (res.code === 0) {
      list.value = res.data || []
      total.value = res.count || 0
    }
  } catch (error) {
    console.error('获取日志列表失败:', error)
  }
}

onMounted(() => {
  fetchList()
})
</script>
