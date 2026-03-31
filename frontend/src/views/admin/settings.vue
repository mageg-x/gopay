<script setup lang="ts">
import { ref, reactive } from 'vue'
import { saveConfig } from '@/api/admin'

const activeTab = ref('site')

const tabs = [
  { id: 'site', name: '网站设置' },
  { id: 'pay', name: '支付设置' },
  { id: 'settle', name: '结算设置' },
  { id: 'transfer', name: '转账设置' },
  { id: 'oauth', name: '快捷登录' },
  { id: 'notice', name: '通知设置' }
]

const form = reactive({
  sitename: '',
  localurl: '',
  apiurl: '',
  kfqq: '',
  reg_open: '1',
  settle_money: '30',
  settle_alipay: '1',
  settle_wxpay: '1'
})

async function handleSave() {
  try {
    const res = await saveConfig(form)
    if (res.code === 0) {
      alert('保存成功')
    }
  } catch (error) {
    console.error('保存失败:', error)
  }
}
</script>

<template>
  <div>
    <h2 class="text-2xl font-bold text-gray-800 mb-6">系统设置</h2>

    <div class="flex gap-6">
      <!-- 标签导航 -->
      <div class="w-48">
        <div class="bg-white rounded-lg border p-2">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            :class="[
              'w-full text-left px-4 py-2 rounded-lg text-sm transition-colors',
              activeTab === tab.id
                ? 'bg-primary-50 text-primary-700'
                : 'text-gray-600 hover:bg-gray-50'
            ]"
            @click="activeTab = tab.id"
          >
            {{ tab.name }}
          </button>
        </div>
      </div>

      <!-- 设置表单 -->
      <div class="flex-1">
        <div class="card">
          <div class="card-body">
            <form @submit.prevent="handleSave" class="space-y-4">
              <div v-if="activeTab === 'site'">
                <div class="mb-4">
                  <label class="form-label">网站名称</label>
                  <input v-model="form.sitename" type="text" class="form-input" />
                </div>
                <div class="mb-4">
                  <label class="form-label">本地地址</label>
                  <input v-model="form.localurl" type="text" class="form-input" />
                </div>
                <div class="mb-4">
                  <label class="form-label">API地址</label>
                  <input v-model="form.apiurl" type="text" class="form-input" />
                </div>
                <div class="mb-4">
                  <label class="form-label">客服QQ</label>
                  <input v-model="form.kfqq" type="text" class="form-input" />
                </div>
                <div class="mb-4">
                  <label class="form-label">开放注册</label>
                  <select v-model="form.reg_open" class="form-input">
                    <option value="0">关闭</option>
                    <option value="1">开放</option>
                    <option value="2">需要邀请码</option>
                  </select>
                </div>
              </div>

              <div v-if="activeTab === 'settle'">
                <div class="mb-4">
                  <label class="form-label">最低结算金额</label>
                  <input v-model="form.settle_money" type="text" class="form-input" />
                </div>
                <div class="mb-4">
                  <label class="form-label">支付宝结算</label>
                  <select v-model="form.settle_alipay" class="form-input">
                    <option value="1">开启</option>
                    <option value="0">关闭</option>
                  </select>
                </div>
                <div class="mb-4">
                  <label class="form-label">微信结算</label>
                  <select v-model="form.settle_wxpay" class="form-input">
                    <option value="1">开启</option>
                    <option value="0">关闭</option>
                  </select>
                </div>
              </div>

              <button type="submit" class="btn btn-primary">保存设置</button>
            </form>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
