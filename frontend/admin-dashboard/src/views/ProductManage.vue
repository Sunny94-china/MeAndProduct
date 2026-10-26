<template>
  <el-container class="d">
    <app-aside /><el-container>
      <app-header crumb="产品管理" />
      <el-main>
        <div class="bar">
          <el-input v-model="kw" placeholder="搜索产品名 / 标语 / slug" clearable style="width:320px" :prefix-icon="Search" @clear="load" @keyup.enter="load" />
          <el-select v-model="statusFilter" placeholder="阶段" clearable style="width:140px" @change="load">
            <el-option label="Idea" value="idea" />
            <el-option label="MVP" value="mvp" />
            <el-option label="Beta" value="beta" />
            <el-option label="Launched 已上线" value="launched" />
            <el-option label="Growing 增长" value="growing" />
            <el-option label="Mature 成熟" value="mature" />
          </el-select>
          <el-select v-model="visFilter" placeholder="显示" clearable style="width:140px" @change="load">
            <el-option label="已显示" value="on" />
            <el-option label="已隐藏" value="off" />
            <el-option label="软删除" value="deleted" />
          </el-select>
          <span class="spacer" />
          <el-button :icon="Refresh" @click="load">刷新</el-button>
        </div>

        <el-table :data="rows" v-loading="loading" stripe style="margin-top:12px;">
          <el-table-column label="产品" min-width="260">
            <template #default="{row}">
              <div class="row">
                <div class="p-logo" :style="logoStyle(row)">
                  <img v-if="row.logo_url" :src="row.logo_url" alt="" />
                  <el-icon v-else :size="18" color="#fff"><Goods /></el-icon>
                </div>
                <div class="meta">
                  <div class="nm">{{ row.name }}
                    <el-tag v-if="row.deleted_at" size="small" type="danger" style="margin-left:6px;">已删</el-tag>
                  </div>
                  <div class="sub">@{{ row.founder_slug || '—' }} / {{ row.slug }}</div>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="标语" min-width="220" prop="tagline" show-overflow-tooltip />
          <el-table-column label="阶段" width="120">
            <template #default="{row}">
              <el-tag size="small" :type="stageType(row.status)">{{ stageLabel(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="标签" width="180">
            <template #default="{row}">
              <el-tag v-for="t in (row.tags||[]).slice(0,5)" :key="t.id" size="small" effect="plain" style="margin-right:4px;">{{ t.name }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="显示" width="120">
            <template #default="{row}">
              <el-switch
                :model-value="!!row.is_visible"
                :loading="row._visLoading"
                :disabled="!!row.deleted_at"
                active-text="显示"
                inactive-text="隐藏"
                @change="(v:boolean)=>setVisible(row,v)"
              />
            </template>
          </el-table-column>
          <el-table-column label="浏览" width="80" prop="view_count" sortable />
          <el-table-column label="收藏" width="80" prop="favorite_count" sortable />
          <el-table-column label="操作" width="260">
            <template #default="{row}">
              <el-button link type="primary" @click="preview(row)">预览</el-button>
              <el-button v-if="!row.deleted_at" link type="warning" @click="forceHide(row)">强制隐藏</el-button>
              <el-button v-if="!row.deleted_at" link type="danger" @click="softDel(row)">软删除</el-button>
              <el-button v-if="row.deleted_at" link type="success" @click="restore(row)">恢复</el-button>
            </template>
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
import { Search, Refresh, Goods } from '@element-plus/icons-vue'
import {
  listProducts,
  adminSetProductStatus,
  type ProductVO,
} from '@/api/content'

const kw = ref('')
const statusFilter = ref<string>('')
const visFilter = ref<string>('')
const rows = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

function stageLabel(s: string) {
  const m: Record<string, string> = { idea: 'Idea', mvp: 'MVP', beta: 'Beta', launched: '已发布', growing: '成长中', mature: '成熟' }
  return m[s] || s
}
function stageType(s: string) {
  const m: Record<string, any> = { idea: 'info', mvp: 'warning', beta: '', launched: 'success', growing: 'success', mature: 'success' }
  return m[s] || 'info'
}

function logoStyle(row: ProductVO) {
  if (row.logo_url) return {}
  const palette = [
    'linear-gradient(135deg,#6366F1,#A855F7)',
    'linear-gradient(135deg,#10B981,#06B6D4)',
    'linear-gradient(135deg,#F43F5E,#8B5CF6)',
    'linear-gradient(135deg,#F59E0B,#EF4444)',
  ]
  return { background: palette[row.id % palette.length] }
}

async function load() {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: pageSize.value }
    if (kw.value) params.keyword = kw.value
    if (statusFilter.value) params.status = statusFilter.value
    if (visFilter.value === 'on') params.is_visible = 1
    if (visFilter.value === 'off') params.is_visible = 0
    if (visFilter.value === 'deleted') params.include_deleted = 1
    const r = await listProducts(params)
    rows.value = r.items || []
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

async function setVisible(row: any, val: boolean) {
  row._visLoading = true
  try {
    await adminSetProductStatus(row.id, { is_visible: val })
    row.is_visible = val
    ElMessage.success(val ? '已显示' : '已隐藏')
  } catch (e: any) {
    row.is_visible = !val
    ElMessage.error(e?.message || '操作失败')
  } finally {
    row._visLoading = false
  }
}
function forceHide(row: any) {
  setVisible(row, false)
}
async function softDel(row: any) {
  try {
    await adminSetProductStatus(row.id, { action: 'soft_delete' })
    ElMessage.success('已软删除')
    load()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}
async function restore(row: any) {
  try {
    await adminSetProductStatus(row.id, { action: 'restore' })
    ElMessage.success('已恢复')
    load()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}
function preview(row: any) {
  const s = row.founder_slug || ''
  if (!s || !row.slug) { ElMessage.warning('Slug 不完整无法预览'); return }
  window.open(`/@${s}/${row.slug}`, '_blank')
}

onMounted(load)
</script>
<style scoped>
.d{ min-height:100vh;}
.bar{ display:flex; gap:10px; align-items:center;}
.spacer{ flex:1; }
.row{ display:flex; align-items:center; gap:12px; }
.p-logo{ width: 40px; height: 40px; border-radius: 8px; overflow: hidden; display: flex; align-items: center; justify-content: center; background:#334155;}
.p-logo img{ width:100%; height:100%; object-fit: cover;}
.meta{ min-width: 0; }
.nm{ font-weight:600; color:#0F172A; }
.sub{ font-size:12px; color:#64748B; margin-top:4px; }
.pg{ margin-top: 20px; display:flex; justify-content: flex-end; }
</style>
