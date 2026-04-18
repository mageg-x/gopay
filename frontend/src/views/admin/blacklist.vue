<template>
  <div class="space-y-4">
    <div class="page-head">
      <div>
        <h1 class="page-title no-wrap">黑名单管理</h1>
        <p class="page-subtitle">管理 IP 和账号黑名单</p>
      </div>
      <button @click="openAddDialog" class="btn btn-primary">添加黑名单</button>
    </div>

    <div class="panel-filter">
      <div class="card-body">
        <div class="toolbar-wrap">
          <select v-model="filterType" class="form-input w-[150px]">
          <option value="">全部类型</option>
          <option value="1">IP黑名单</option>
          <option value="2">账号黑名单</option>
        </select>
          <button @click="page = 1; fetchList()" class="btn btn-primary">筛选</button>
        </div>
      </div>
    </div>

    <div class="table-shell">
      <div class="table-shell-body overflow-x-auto">
        <table class="table min-w-[860px] whitespace-nowrap">
          <thead>
            <tr>
              <th class="text-left">ID</th>
              <th>类型</th>
              <th class="text-left">内容</th>
              <th class="text-left">备注</th>
              <th class="text-left">添加时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="b in list" :key="b.id">
              <td class="text-left text-gray-900">{{ b.id }}</td>
              <td>
                <span :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                  b.type === 1 ? 'bg-red-100 text-red-700' : 'bg-orange-100 text-orange-700']">
                  {{ b.type === 1 ? 'IP黑名单' : '账号黑名单' }}
                </span>
              </td>
              <td class="text-left font-mono text-gray-900">{{ b.content }}</td>
              <td class="text-left text-gray-500 max-w-xs truncate">{{ b.remark || '-' }}</td>
              <td class="text-left text-gray-500 text-xs">{{ formatTime(b.addtime) }}</td>
              <td>
                <button @click="handleDelete(b.id)" class="action-link action-link-danger">删除</button>
              </td>
            </tr>
            <tr v-if="list.length === 0">
              <td colspan="6" class="py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
                  </svg>
                  <span>暂无黑名单记录</span>
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

    <div v-if="dialogVisible" class="dialog-backdrop">
      <div class="dialog-wrap">
        <div class="dialog-mask" @click="dialogVisible = false"></div>
        <div class="dialog-panel max-w-md">
          <div class="dialog-header">
            <div>
              <h3 class="dialog-title">添加黑名单</h3>
            </div>
            <button class="dialog-close" @click="dialogVisible = false">✕</button>
          </div>
          <div class="dialog-body space-y-4">
            <div>
              <label class="form-label">类型</label>
              <select v-model="form.type" class="form-input">
                <option :value="1">IP黑名单</option>
                <option :value="2">账号黑名单</option>
              </select>
            </div>
            <div>
              <label class="form-label">内容</label>
              <input v-model="form.content" type="text" class="form-input" placeholder="IP或账号" />
            </div>
            <div>
              <label class="form-label">备注</label>
              <input v-model="form.remark" type="text" class="form-input" placeholder="可选" />
            </div>
          </div>
          <div class="dialog-footer">
            <button @click="dialogVisible = false" class="btn btn-outline">取消</button>
            <button @click="handleAdd" class="btn btn-primary">添加</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, reactive } from 'vue'
import { blacklistList, blacklistOp } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'

const list = ref<any[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const filterType = ref('')
const dialogVisible = ref(false)

const form = reactive({
  type: 1,
  content: '',
  remark: ''
})

const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

function formatTime(time: string) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

async function fetchList() {
  try {
    const params: any = { page: page.value, limit: pageSize.value }
    if (filterType.value) {
      params.type = filterType.value
    }
    const res = await blacklistList(params)
    if (res.code === 0) {
      list.value = res.data || []
      total.value = res.count || 0
    }
  } catch (error) {
    console.error('获取黑名单失败:', error)
  }
}

function openAddDialog() {
  form.type = 1
  form.content = ''
  form.remark = ''
  dialogVisible.value = true
}

async function handleAdd() {
  if (!form.content.trim()) {
    ElMessage.warning('请输入内容')
    return
  }
  try {
    const res = await blacklistOp({ action: 'add', type: form.type, content: form.content, remark: form.remark })
    if (res.code === 0) {
      ElMessage.success('添加成功')
      dialogVisible.value = false
      fetchList()
    } else {
      ElMessage.error(res.msg || '添加失败')
    }
  } catch (error) {
    console.error('添加失败:', error)
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确定要删除该黑名单吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    const res = await blacklistOp({ action: 'delete', id })
    if (res.code === 0) {
      ElMessage.success('删除成功')
      fetchList()
    } else {
      ElMessage.error(res.msg || '删除失败')
    }
  } catch (error) {
    console.error('删除失败:', error)
  }
}

onMounted(() => {
  fetchList()
})
</script>
