<template>
  <div class="space-y-4">
    <div class="page-head">
      <div>
        <h1 class="page-title no-wrap">计划任务</h1>
        <p class="page-subtitle">管理系统定时任务</p>
      </div>
    </div>

    <div class="table-shell">
      <div class="table-shell-body overflow-x-auto">
        <table class="table min-w-[860px] whitespace-nowrap">
          <thead>
            <tr>
              <th class="text-left">任务名称</th>
              <th>执行周期</th>
              <th>下次执行</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="task in taskList" :key="task.name">
              <td class="text-left">
                <div class="flex items-center gap-2">
                  <span class="w-2 h-2 rounded-full" :class="task.running ? 'bg-green-500' : 'bg-gray-300'"></span>
                  <span class="font-medium text-gray-900">{{ taskName(task.name) }}</span>
                </div>
              </td>
              <td>
                <span class="font-mono text-xs bg-gray-100 px-2 py-1 rounded">{{ task.spec || '默认组' }}</span>
              </td>
              <td class="text-gray-500 text-xs">{{ task.next || '-' }}</td>
              <td>
                <button @click="toggleTask(task)" :class="[
                  'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium cursor-pointer transition-colors',
                  task.enabled ? 'bg-green-100 text-green-700 hover:bg-green-200' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                ]">
                  {{ task.enabled ? '启用' : '禁用' }}
                </button>
              </td>
              <td>
                <button @click="runTask(task)" class="action-link action-link-primary mr-1">立即执行</button>
                <button @click="openEditDialog(task)" class="action-link action-link-warning">编辑</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-if="dialogVisible" class="dialog-backdrop">
      <div class="dialog-wrap">
        <div class="dialog-mask" @click="dialogVisible = false"></div>
        <div class="dialog-panel max-w-md">
          <div class="dialog-header">
            <div>
              <h3 class="dialog-title">编辑执行周期</h3>
            </div>
            <button class="dialog-close" @click="dialogVisible = false">✕</button>
          </div>
          <div class="dialog-body space-y-4">
            <div>
              <label class="form-label">任务名称</label>
              <input :value="currentTask.name" type="text" disabled class="form-input bg-gray-50" />
            </div>
            <div>
              <label class="form-label">Cron表达式</label>
              <input v-model="form.spec" type="text" class="form-input" placeholder="0 * * * * ?" />
              <p class="text-xs text-gray-500 mt-1">
                格式: 秒 分 时 日 月 周(可选)
                <br />常用: <span class="cursor-pointer text-blue-500" @click="form.spec = '0 */5 * * * ?'">每5分钟</span> |
                <span class="cursor-pointer text-blue-500" @click="form.spec = '0 0 * * * ?'">每小时</span> |
                <span class="cursor-pointer text-blue-500" @click="form.spec = '0 0 0 * * ?'">每天</span>
              </p>
            </div>
          </div>
          <div class="dialog-footer">
            <button @click="dialogVisible = false" class="btn btn-outline">取消</button>
            <button @click="saveTask" class="btn btn-primary">保存</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { cronList, cronOp } from '@/api/admin'
import { ElMessage } from 'element-plus'

const taskList = ref<any[]>([])
const dialogVisible = ref(false)
const currentTask = ref<any>({})
const form = ref({
  spec: ''
})

const defaultTasks = [
  { name: 'auto_settle', desc: '自动结算', defaultSpec: '0 0 * * * ?' },
  { name: 'retry_notify', desc: '回调重试', defaultSpec: '0 */5 * * * ?' },
  { name: 'order_query', desc: '订单状态刷新', defaultSpec: '0 */3 * * * ?' },
  { name: 'risk_check', desc: '风控检查', defaultSpec: '0 */30 * * * ?' },
  { name: 'cleanup', desc: '清理过期数据', defaultSpec: '0 0 0 * * ?' },
  { name: 'db_backup', desc: '数据库备份', defaultSpec: '0 0 2 * * ?' }
]

function taskName(name: string) {
  const task = defaultTasks.find(t => t.name === name)
  return task ? task.desc : name
}

async function fetchTasks() {
  try {
    const res = await cronList()
    if (res.code === 0) {
      // 合并任务列表
      const serverTasks = res.data || []
      taskList.value = defaultTasks.map(t => {
        const serverTask = serverTasks.find((s: any) => s.name === t.name)
        return {
          name: t.name,
          desc: t.desc,
          spec: serverTask?.spec || t.defaultSpec,
          next: serverTask?.next || '-',
          running: serverTask?.running || false,
          enabled: serverTask?.next !== undefined
        }
      })
    }
  } catch (error) {
    console.error('获取任务列表失败:', error)
  }
}

async function toggleTask(task: any) {
  try {
    await cronOp({
      action: 'set',
      name: task.name,
      enable: !task.enabled
    })
    ElMessage.success(task.enabled ? '已禁用' : '已启用')
    task.enabled = !task.enabled
  } catch (error) {
    console.error('操作失败:', error)
  }
}

async function runTask(task: any) {
  try {
    await cronOp({ action: 'run', name: task.name })
    ElMessage.success('任务已触发')
  } catch (error) {
    console.error('操作失败:', error)
  }
}

function openEditDialog(task: any) {
  currentTask.value = task
  form.value.spec = task.spec
  dialogVisible.value = true
}

async function saveTask() {
  if (!form.value.spec.trim()) {
    ElMessage.warning('请输入执行周期')
    return
  }
  try {
    await cronOp({
      action: 'set',
      name: currentTask.value.name,
      spec: form.value.spec
    })
    ElMessage.success('保存成功')
    dialogVisible.value = false
    fetchTasks()
  } catch (error) {
    console.error('保存失败:', error)
  }
}

onMounted(() => {
  fetchTasks()
})
</script>
