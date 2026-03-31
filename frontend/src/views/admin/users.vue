<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getUserList } from '@/api/admin'
import dayjs from 'dayjs'

interface User {
  uid: number
  username: string
  email: string
  phone: string
  money: number
  status: number
  addtime: string
}

const users = ref<User[]>([])
const loading = ref(false)
const page = ref(1)
const total = ref(0)

const statusMap: Record<number, string> = {
  0: { text: '正常', class: 'badge-success' },
  1: { text: '禁用', class: 'badge-danger' }
}

async function fetchUsers() {
  loading.value = true
  try {
    const res = await getUserList({ page: page.value, limit: 20 })
    if (res.code === 0) {
      users.value = res.data || []
      total.value = res.count || 0
    }
  } catch (error) {
    console.error('获取商户列表失败:', error)
  } finally {
    loading.value = false
  }
}

function formatTime(time: string) {
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

onMounted(() => {
  fetchUsers()
})
</script>

<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-2xl font-bold text-gray-800">商户管理</h2>
    </div>

    <div class="card">
      <div class="card-body">
        <table class="table">
          <thead>
            <tr>
              <th>商户ID</th>
              <th>用户名</th>
              <th>邮箱</th>
              <th>手机</th>
              <th>余额</th>
              <th>状态</th>
              <th>注册时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in users" :key="user.uid">
              <td>{{ user.uid }}</td>
              <td>{{ user.username || '-' }}</td>
              <td>{{ user.email || '-' }}</td>
              <td>{{ user.phone || '-' }}</td>
              <td class="text-primary-600 font-medium">¥{{ user.money }}</td>
              <td>
                <span :class="['badge', statusMap[user.status]?.class]">
                  {{ statusMap[user.status]?.text || '未知' }}
                </span>
              </td>
              <td>{{ formatTime(user.addtime) }}</td>
              <td>
                <button class="text-primary-600 hover:text-primary-800 mr-3">编辑</button>
                <button class="text-warning hover:text-warning mr-3">重置密钥</button>
                <button class="text-danger hover:text-danger">禁用</button>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- 分页 -->
        <div class="pagination">
          <button
            class="pagination-item"
            :disabled="page === 1"
            @click="page--; fetchUsers()"
          >
            上一页
          </button>
          <span class="px-4 py-1">第 {{ page }} / {{ Math.ceil(total / 20) }} 页</span>
          <button
            class="pagination-item"
            :disabled="page * 20 >= total"
            @click="page++; fetchUsers()"
          >
            下一页
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
