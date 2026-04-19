<template>
  <div class="space-y-4">
    <div class="page-head">
      <div>
        <h1 class="page-title no-wrap">域名授权管理</h1>
        <p class="page-subtitle">管理商户域名授权</p>
      </div>
      <button @click="openAddDialog" class="btn btn-primary">添加授权</button>
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
              <th class="text-left">域名</th>
              <th>状态</th>
              <th class="text-left">添加时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="d in list" :key="d.id">
              <td class="text-left text-gray-900">{{ d.id }}</td>
              <td>
                <div class="flex items-center gap-2">
                  <span class="font-medium">{{ d.user_name || d.uid }}</span>
                </div>
              </td>
              <td class="text-left font-mono text-gray-900">{{ d.domain }}</td>
              <td>
                <button @click="toggleStatus(d)" :class="[
                  'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium cursor-pointer transition-colors',
                  d.status === 1 ? 'bg-green-100 text-green-700 hover:bg-green-200' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                ]">
                  {{ d.status === 1 ? '启用' : '禁用' }}
                </button>
              </td>
              <td class="text-left text-gray-500 text-xs">{{ formatTime(d.addtime) }}</td>
              <td>
                <button @click="handleDelete(d.id)" class="action-link action-link-danger">删除</button>
              </td>
            </tr>
            <tr v-if="list.length === 0">
              <td colspan="6" class="py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
                  </svg>
                  <span>暂无域名授权记录</span>
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
              <h3 class="dialog-title">添加域名授权</h3>
            </div>
            <button class="dialog-close" @click="dialogVisible = false">✕</button>
          </div>
          <div class="dialog-body space-y-4">
            <div>
              <label class="form-label">商户ID</label>
              <input v-model="form.uid" type="number" class="form-input" placeholder="请输入商户ID" />
            </div>
            <div>
              <label class="form-label">域名</label>
              <input v-model="form.domain" type="text" class="form-input" placeholder="例如: example.com" />
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
import { ref, computed, onMounted } from 'vue'
import { domainList, domainOp } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'

const list = ref<any[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const searchUid = ref('')
const dialogVisible = ref(false)

const form = ref({
  uid: '',
  domain: ''
})

const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

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
    const res = await domainList(params)
    if (res.code === 0) {
      list.value = res.data || []
      total.value = res.count || 0
    }
  } catch (error) {
    console.error('获取域名列表失败:', error)
  }
}

function openAddDialog() {
  form.value.uid = ''
  form.value.domain = ''
  dialogVisible.value = true
}

async function handleAdd() {
  if (!form.value.uid) {
    ElMessage.warning('请输入商户ID')
    return
  }
  if (!form.value.domain) {
    ElMessage.warning('请输入域名')
    return
  }
  try {
    const res = await domainOp({ action: 'add', uid: parseInt(form.value.uid), domain: form.value.domain })
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

async function toggleStatus(d: any) {
  try {
    const newStatus = d.status === 1 ? 0 : 1
    await domainOp({ action: 'set_status', id: d.id, status: newStatus })
    ElMessage.success(newStatus === 1 ? '已启用' : '已禁用')
    d.status = newStatus
  } catch (error) {
    console.error('操作失败:', error)
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确定要删除该域名授权吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    const res = await domainOp({ action: 'delete', id })
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
