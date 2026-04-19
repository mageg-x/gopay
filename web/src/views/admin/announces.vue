<template>
  <div class="space-y-4">
    <div class="page-head">
      <div>
        <h1 class="page-title no-wrap">公告管理</h1>
        <p class="page-subtitle">管理网站公告</p>
      </div>
      <button @click="openAddDialog" class="btn btn-primary">添加公告</button>
    </div>

    <div class="table-shell">
      <div class="table-shell-body overflow-x-auto">
        <table class="table min-w-[860px] whitespace-nowrap">
          <thead>
            <tr>
              <th class="text-left">ID</th>
              <th class="text-left">内容</th>
              <th>颜色</th>
              <th>排序</th>
              <th>状态</th>
              <th class="text-left">添加时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="a in list" :key="a.id">
              <td class="text-left text-gray-900">{{ a.id }}</td>
              <td class="text-left">
                <div :style="{ color: a.color || '#333' }" class="max-w-xs truncate">{{ a.content }}</div>
              </td>
              <td>
                <div class="flex items-center justify-center gap-2">
                  <div class="w-4 h-4 rounded" :style="{ backgroundColor: a.color || '#333' }"></div>
                  <span class="text-xs text-gray-500">{{ a.color || '#333' }}</span>
                </div>
              </td>
              <td class="text-gray-500">{{ a.sort }}</td>
              <td>
                <button @click="toggleStatus(a)" :class="[
                  'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium cursor-pointer transition-colors',
                  a.status === 1 ? 'bg-green-100 text-green-700 hover:bg-green-200' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                ]">
                  {{ a.status === 1 ? '显示' : '隐藏' }}
                </button>
              </td>
              <td class="text-left text-gray-500 text-xs">{{ formatTime(a.addtime) }}</td>
              <td>
                <button @click="openEditDialog(a)" class="action-link action-link-primary mr-1">编辑</button>
                <button @click="handleDelete(a.id)" class="action-link action-link-danger">删除</button>
              </td>
            </tr>
            <tr v-if="list.length === 0">
              <td colspan="7" class="py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" />
                  </svg>
                  <span>暂无公告</span>
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
              <h3 class="dialog-title">{{ isEdit ? '编辑公告' : '添加公告' }}</h3>
            </div>
            <button class="dialog-close" @click="dialogVisible = false">✕</button>
          </div>
          <div class="dialog-body space-y-4">
            <div>
              <label class="form-label">内容</label>
              <textarea v-model="form.content" rows="3"
                class="form-input"
                placeholder="公告内容"></textarea>
            </div>
            <div>
              <label class="form-label">颜色</label>
              <div class="flex items-center gap-3">
                <input v-model="form.color" type="color"
                  class="w-10 h-10 border border-gray-200 rounded cursor-pointer" />
                <input v-model="form.color" type="text"
                  class="form-input flex-1"
                  placeholder="#333333" />
              </div>
            </div>
            <div>
              <label class="form-label">排序</label>
              <input v-model="form.sort" type="number"
                class="form-input"
                placeholder="数值越大越靠前" />
            </div>
            <div>
              <label class="form-label">状态</label>
              <select v-model="form.status"
                class="form-input">
                <option :value="1">显示</option>
                <option :value="0">隐藏</option>
              </select>
            </div>
          </div>
          <div class="dialog-footer">
            <button @click="dialogVisible = false" class="btn btn-outline">取消</button>
            <button @click="handleSave" class="btn btn-primary">{{ isEdit ? '保存' : '添加' }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { announceList, announceOp } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'

const list = ref<any[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const dialogVisible = ref(false)
const isEdit = ref(false)

const form = ref({
  id: 0,
  content: '',
  color: '#333333',
  sort: 0,
  status: 1
})

const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

function formatTime(time: string) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

async function fetchList() {
  try {
    const res = await announceList({ page: page.value, limit: pageSize.value })
    if (res.code === 0) {
      list.value = res.data || []
      total.value = res.count || 0
    }
  } catch (error) {
    console.error('获取公告列表失败:', error)
  }
}

function openAddDialog() {
  isEdit.value = false
  form.value = { id: 0, content: '', color: '#333333', sort: 0, status: 1 }
  dialogVisible.value = true
}

function openEditDialog(a: any) {
  isEdit.value = true
  form.value = {
    id: a.id,
    content: a.content,
    color: a.color || '#333333',
    sort: a.sort,
    status: a.status
  }
  dialogVisible.value = true
}

async function handleSave() {
  if (!form.value.content.trim()) {
    ElMessage.warning('请输入公告内容')
    return
  }
  try {
    const action = isEdit.value ? 'edit' : 'add'
    const res = await announceOp({
      action,
      id: form.value.id,
      content: form.value.content,
      color: form.value.color,
      sort: form.value.sort,
      status: form.value.status
    })
    if (res.code === 0) {
      ElMessage.success(isEdit.value ? '保存成功' : '添加成功')
      dialogVisible.value = false
      fetchList()
    } else {
      ElMessage.error(res.msg || '操作失败')
    }
  } catch (error) {
    console.error('操作失败:', error)
  }
}

async function toggleStatus(a: any) {
  try {
    const newStatus = a.status === 1 ? 0 : 1
    await announceOp({ action: 'edit', id: a.id, content: a.content, color: a.color, sort: a.sort, status: newStatus })
    ElMessage.success(newStatus === 1 ? '已显示' : '已隐藏')
    a.status = newStatus
  } catch (error) {
    console.error('操作失败:', error)
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确定要删除该公告吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    const res = await announceOp({ action: 'delete', id })
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
