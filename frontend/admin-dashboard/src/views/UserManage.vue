<template>
  <el-container class="d">
    <app-aside /><el-container>
      <app-header crumb="用户与角色" />
      <el-main>
        <div class="bar">
          <el-input v-model="kw" placeholder="搜索昵称 / 邮箱" clearable style="width:300px" :prefix-icon="Search" @clear="load" @keyup.enter="load" />
          <el-select v-model="roleFilter" placeholder="角色" clearable style="width:140px" @change="load">
            <el-option label="admin" value="admin" />
            <el-option label="founder" value="founder" />
            <el-option label="guest" value="guest" />
          </el-select>
          <el-select v-model="statusFilter" placeholder="状态" clearable style="width:140px" @change="load">
            <el-option label="活跃 active" value="active" />
            <el-option label="已禁用 banned" value="banned" />
            <el-option label="待审核 pending" value="pending" />
          </el-select>
          <span class="spacer" />
          <el-button @click="load">刷新</el-button>
        </div>
        <el-table :data="rows" stripe v-loading="loading" style="margin-top:12px;">
          <el-table-column label="账号" min-width="240">
            <template #default="{row}">
              <div class="user-row">
                <el-avatar :size="36">
                  {{ (row.display_name || row.username || 'U').slice(0,1).toUpperCase() }}
                </el-avatar>
                <div class="meta">
                  <div class="nm">{{ row.display_name || row.username }}
                    <el-tag v-if="row.id === myId" size="small" type="success" style="margin-left:6px;">我</el-tag>
                  </div>
                  <div class="sub">{{ row.email || row.phone || '—' }}</div>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="角色" width="180">
            <template #default="{row}">
              <el-select
                v-model="row._role"
                size="small"
                :disabled="!!row._loading_role || row.id === myId"
                @change="(v: string) => saveRole(row, v)"
                style="width: 160px;"
              >
                <el-option label="guest" value="guest" />
                <el-option label="founder" value="founder" />
                <el-option label="admin" value="admin" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="180">
            <template #default="{row}">
              <el-select
                v-model="row._status"
                size="small"
                :disabled="!!row._loading_status || row.id === myId"
                @change="(v: string) => saveStatus(row, v)"
                style="width: 160px;"
              >
                <el-option label="活跃 active" value="active" />
                <el-option label="待审核 pending" value="pending" />
                <el-option label="已禁用 banned" value="banned" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="注册时间" width="170">
            <template #default="{row}">{{ fmt(row.created_at) }}</template>
          </el-table-column>
        </el-table>
        <div class="pg">
          <el-pagination :current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="onPage" />
        </div>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import AppAside from '@/components/AppAside.vue'
import AppHeader from '@/components/AppHeader.vue'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import {
  listUsers,
  adminUpdateUser,
  type UserVO,
} from '@/api/content'

const kw = ref('')
const roleFilter = ref('')
const statusFilter = ref('')
const rows = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const myId = ref<number | null>(null)

function fmt(s?: string) {
  if (!s) return '—'
  const d = new Date(s)
  if (Number.isNaN(d.getTime())) return s.slice(0, 10)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

async function load() {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: pageSize.value }
    if (kw.value) params.keyword = kw.value
    if (roleFilter.value) params.role = roleFilter.value
    if (statusFilter.value) params.status = statusFilter.value
    const r = await listUsers(params)
    rows.value = (r.items || []).map((u: UserVO) => ({ ...u, _role: u.role, _status: u.status, _loading_role: false, _loading_status: false }))
    total.value = r.total || 0
  } catch (_e) {
    rows.value = []
  } finally {
    loading.value = false
  }
}
function onPage(p: number) {
  page.value = p
  load()
}

async function saveRole(row: any, v: string) {
  row._loading_role = true
  try {
    await adminUpdateUser(row.id, { role: v as UserVO['role'] })
    row.role = v
    ElMessage.success('角色已更新')
  } catch (e: any) {
    row._role = row.role
    ElMessage.error(e?.message || '操作失败')
  } finally {
    row._loading_role = false
  }
}
async function saveStatus(row: any, v: string) {
  row._loading_status = true
  try {
    await adminUpdateUser(row.id, { status: v as UserVO['status'] })
    row.status = v
    ElMessage.success('状态已更新')
  } catch (e: any) {
    row._status = row.status
    ElMessage.error(e?.message || '操作失败')
  } finally {
    row._loading_status = false
  }
}

onMounted(() => {
  try {
    const raw = localStorage.getItem('mp_admin_user')
    if (raw) {
      const me = JSON.parse(raw)
      if (me?.id) myId.value = Number(me.id)
    }
  } catch (_e) {}
  load()
})
</script>
<style scoped>
.d{ min-height:100vh;}
.bar{ display:flex; gap:10px; align-items:center;}
.spacer{ flex:1; }
.user-row{ display:flex; align-items:center; gap:12px; }
.nm{ font-weight:600; color:#0F172A;}
.sub{ font-size:12px; color:#64748B; margin-top:4px;}
.pg{ margin-top: 20px; display:flex; justify-content: flex-end; }
</style>
