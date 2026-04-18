<template>
  <div class="space-y-4">
    <div class="page-head">
      <div>
        <h1 class="page-title no-wrap">支付方式管理</h1>
        <p class="page-subtitle">配置支付方式名称、图标和状态</p>
      </div>
      <button @click="showAddModal" class="btn btn-primary">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        添加类型
      </button>
    </div>

    <div class="table-shell">
      <div class="overflow-x-auto">
        <table class="table min-w-[860px] whitespace-nowrap">
          <thead>
            <tr>
              <th class="text-left">ID</th>
              <th class="text-left">标识</th>
              <th class="text-left">显示名称</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="pt in payTypes" :key="pt.id">
              <td class="text-left text-gray-900 font-medium">{{ pt.id }}</td>
              <td class="text-left text-gray-900">
                <div class="flex items-center gap-1.5">
                  <SvgIcon :name="getIconName(pt.name)" :size="16" />
                  <span class="font-medium">{{ pt.name }}</span>
                </div>
              </td>
              <td class="text-left text-gray-600">{{ pt.showname || '-' }}</td>
              <td>
                <span :class="['badge', pt.status ? 'badge-success' : 'badge-info']">
                  {{ pt.status ? '开启' : '关闭' }}
                </span>
              </td>
              <td>
                <div class="inline-flex items-center gap-1">
                  <button @click="showEditModal(pt)" class="action-link action-link-primary">编辑</button>
                  <button @click="toggleStatus(pt)" :class="['action-link', pt.status ? 'action-link-warning' : 'action-link-success']">
                    {{ pt.status ? '关闭' : '开启' }}
                  </button>
                  <button @click="handleDelete(pt.id)" class="action-link action-link-danger">删除</button>
                </div>
              </td>
            </tr>
            <tr v-if="payTypes.length === 0">
              <td colspan="5" class="py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
                  </svg>
                  <span>暂无支付类型配置</span>
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
        <div class="dialog-panel max-w-md">
          <div class="dialog-header">
            <div>
              <h3 class="dialog-title">{{ isEdit ? '编辑支付类型' : '添加支付类型' }}</h3>
              <p class="dialog-subtitle">配置标识名称与展示名称</p>
            </div>
            <button class="dialog-close" @click="showModal = false">✕</button>
          </div>
          <div class="dialog-body space-y-4">
            <div>
              <label class="form-label">标识名称</label>
              <input v-model="form.name" type="text" placeholder="如：alipay, wechatpay"
                class="form-input px-3" />
            </div>
            <div>
              <label class="form-label">显示名称</label>
              <input v-model="form.showname" type="text" placeholder="如：支付宝、微信支付"
                class="form-input px-3" />
            </div>
            <div>
              <label class="form-label">状态</label>
              <select v-model="form.status" class="form-input px-3">
                <option :value="1">开启</option>
                <option :value="0">关闭</option>
              </select>
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
import { ref, onMounted } from 'vue'
import { getPayTypeList, payTypeOp } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'
import SvgIcon from '@/components/svgicon.vue'

const payTypes = ref<any[]>([])
const showModal = ref(false)
const isEdit = ref(false)
const form = ref({
  id: 0,
  name: '',
  showname: '',
  status: 1
})

async function fetchPayTypes() {
  try {
    const res = await getPayTypeList()
    if (res.code === 0) {
      payTypes.value = res.data || []
    }
  } catch (error) {
    console.error('获取支付类型失败:', error)
  }
}

function showAddModal() {
  isEdit.value = false
  form.value = {
    id: 0,
    name: '',
    showname: '',
    status: 1
  }
  showModal.value = true
}

function showEditModal(pt: any) {
  isEdit.value = true
  form.value = {
    id: pt.id,
    name: pt.name,
    showname: pt.showname || '',
    status: pt.status
  }
  showModal.value = true
}

async function handleSave() {
  if (!form.value.name) {
    ElMessage.warning('请输入标识名称')
    return
  }
  try {
    const res = await payTypeOp({
      action: isEdit.value ? 'edit' : 'add',
      ...form.value
    })
    ElMessage.success(res.msg || '保存成功')
    showModal.value = false
    fetchPayTypes()
  } catch (error) {
    console.error('保存失败:', error)
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确定要删除这个支付类型吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    const res = await payTypeOp({ action: 'delete', id })
    ElMessage.success(res.msg || '删除成功')
    fetchPayTypes()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

async function toggleStatus(pt: any) {
  try {
    const res = await payTypeOp({
      action: 'set_status',
      id: pt.id,
      status: pt.status ? 0 : 1
    })
    ElMessage.success(res.msg || '状态已更新')
    fetchPayTypes()
  } catch (error) {
    console.error('更新状态失败:', error)
  }
}

function getIconName(name: string): string {
  const map: Record<string, string> = {
    alipay: 'alipay',
    wechatpay: 'wechatpay',
    qqpay: 'qqpay',
    unionpay: 'unionpay',
    jdpay: 'jdpay'
  }
  return map[name] || 'creditcard'
}

onMounted(() => {
  fetchPayTypes()
})
</script>
