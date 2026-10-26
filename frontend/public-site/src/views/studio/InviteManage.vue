<template>
  <div class="studio-page">
    <div class="tool-bar">
      <div class="hint">把邀请码发给朋友，对方注册后就是创业者。</div>
      <el-button type="primary" :loading="creating" @click="gen">生成邀请码</el-button>
    </div>
    <div class="table-scroll">
    <el-table :data="rows" v-loading="loading" style="margin-top: 12px;">
      <el-table-column prop="code" label="邀请码" min-width="160" />
      <el-table-column label="次数" width="120">
        <template #default="{ row }">{{ row.used_count }}/{{ row.max_uses }}</template>
      </el-table-column>
      <el-table-column label="过期时间" min-width="180">
        <template #default="{ row }">{{ fmt(row.expires_at) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }"><el-tag size="small">{{ row.status }}</el-tag></template>
      </el-table-column>
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button link type="primary" @click="copy(row.code)">复制</el-button>
          <el-button v-if="row.status === 'active'" link type="danger" @click="revoke(row)">作废</el-button>
        </template>
      </el-table-column>
    </el-table>
    </div>
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { createInvite, listMyInvites, revokeInvite } from '@/api/content'

const rows = ref<any[]>([])
const loading = ref(false)
const creating = ref(false)

function fmt(v?: string) {
  if (!v) return '不过期'
  const d = new Date(v)
  return Number.isNaN(d.getTime()) ? v : d.toLocaleString()
}

async function load() {
  loading.value = true
  try {
    const data = await listMyInvites()
    rows.value = data?.items || []
  } finally {
    loading.value = false
  }
}

async function gen() {
  creating.value = true
  try {
    await createInvite()
    ElMessage.success('邀请码已生成')
    await load()
  } finally {
    creating.value = false
  }
}

async function copy(code: string) {
  await navigator.clipboard.writeText(code)
  ElMessage.success('已复制')
}

async function revoke(row: any) {
  await revokeInvite(row.id)
  ElMessage.success('已作废')
  await load()
}

onMounted(load)
</script>
<style scoped>
.studio-page{ padding: 8px 4px 24px; }
.tool-bar{ display:flex; justify-content: space-between; align-items: center; gap: 12px; flex-wrap: wrap;}
.hint{ color:#64748B; flex: 1; min-width: 200px; line-height: 1.5; }
</style>
