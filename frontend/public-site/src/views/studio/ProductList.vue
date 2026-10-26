<template>
  <div class="studio-page">
        <div class="tool-bar">
          <el-input v-model="kw" class="kw" placeholder="搜索产品名称" clearable :prefix-icon="Search" />
          <div class="bar-actions">
            <el-tooltip content="拖拽调整排序后会自动保存">
              <el-button :icon="Rank" type="info" plain @click="toggleSortMode" :disabled="rows.length < 2">
                {{ sortMode ? '完成排序' : '拖拽排序' }}
              </el-button>
            </el-tooltip>
            <el-button type="primary" :icon="Plus" @click="$router.push('/studio/products/new')">新增产品</el-button>
          </div>
        </div>

        <el-alert
          v-if="sortMode"
          type="info"
          title="拖拽排序模式：拖动左侧 ⋮⋮ 调整顺序，完成后点击“完成排序”保存"
          show-icon
          :closable="false"
          style="margin: 12px 0;"
        />

        <div class="table-scroll">
        <el-table
          :data="sortMode ? rows : shown"
          stripe
          style="margin-top: 12px; width: 100%;"
          v-loading="loading"
          row-key="id"
        >
          <el-table-column v-if="sortMode" label="排序" width="60" align="center">
            <template #default>
              <el-icon class="sort-handle" :size="18" color="#94A3B8"><Rank /></el-icon>
            </template>
          </el-table-column>
          <el-table-column label="产品" min-width="260">
            <template #default="{row}">
              <router-link :to="`/@${founderSlug}/${row.slug}`" target="_blank" class="prod-link">
                <div class="prod-name">
                  <div class="p-logo-mini" :style="pLogo(row)"></div>
                  <div>
                    <div class="strong">{{ row.name }}</div>
                    <div class="sub">{{ row.tagline }}</div>
                  </div>
                </div>
              </router-link>
            </template>
          </el-table-column>
          <el-table-column label="阶段" width="120">
            <template #default="{row}">
              <el-tag size="small" :type="stageTagType(row.status)">{{ stageLabel(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="访问 / 收藏" width="140">
            <template #default="{row}">
              <span><el-icon><View /></el-icon>{{ row.view_count }}</span>
              <span style="margin-left: 10px;"><el-icon><Star /></el-icon>{{ row.favorite_count }}</span>
            </template>
          </el-table-column>
          <el-table-column label="排序号" width="100" v-if="!sortMode">
            <template #default="{row}">{{ row.sort_order }}</template>
          </el-table-column>
          <el-table-column label="显示" width="100">
            <template #default="{row}">
              <el-switch
                :model-value="!!row.is_visible"
                :loading="row._visLoading"
                @change="(val: boolean) => toggleVisible(row, val)"
              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="220" align="right">
            <template #default="{row}">
              <el-button link type="primary" @click="goEdit(row)">编辑</el-button>
              <el-button link type="primary" @click="goPreview(row)">预览</el-button>
              <el-popconfirm title="确定删除该产品吗？（可在后台恢复）" @confirm="remove(row)">
                <template #reference>
                  <el-button link type="danger">删除</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>
        </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Plus, View, Star, Rank } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
import {
  listMyProducts,
  updateMyProduct,
  deleteMyProduct,
  reorderProducts,
  getMyProfile,
  type ProductVO,
} from '@/api/content'

const router = useRouter()
const loading = ref(false)
const kw = ref('')
const shown = computed(() => {
  const q = kw.value.trim().toLowerCase()
  if (!q) return rows.value
  return rows.value.filter((r) => (r.name || '').toLowerCase().includes(q) || (r.tagline || '').toLowerCase().includes(q))
})
const rows = ref<(ProductVO & { _visLoading?: boolean })[]>([])
const sortMode = ref(false)
const founderSlug = ref('')

function stageLabel(s: string) {
  const m: Record<string, string> = { idea: '想法', beta: '内测', launched: '已上线', growing: '增长中', archived: '已归档' }
  return m[s] || s
}
function stageTagType(s: string) {
  const m: Record<string, any> = { idea: 'info', beta: 'warning', launched: 'success', growing: 'success', archived: 'info' }
  return m[s] || 'info'
}

function pLogo(p: ProductVO) {
  if (p.logo_url) return { background: `url(${p.logo_url}) center/cover no-repeat` }
  const palette = ['#6366F1', '#EC4899', '#10B981', '#F43F5E', '#F59E0B', '#3B82F6', '#8B5CF6', '#14B8A6']
  return { background: palette[(p.id || 0) % palette.length] }
}

async function load() {
  loading.value = true
  try {
    const r = await listMyProducts({ page: 1, page_size: 100, sort: 'sort_order' })
    rows.value = r?.items || []
  } catch (_e) {
    rows.value = []
  } finally {
    loading.value = false
  }
  try {
    const p = await getMyProfile()
    if (p) founderSlug.value = p.slug || ''
  } catch (_e) {}
}

function goEdit(row: ProductVO) {
  router.push(`/studio/products/${row.id}/edit`)
}
function goPreview(row: ProductVO) {
  if (!founderSlug.value) { ElMessage.warning('请先完善名片信息后再预览'); return }
  router.push(`/@${founderSlug.value}/${row.slug}`)
}

async function toggleVisible(row: ProductVO & { _visLoading?: boolean }, val: boolean) {
  row._visLoading = true
  try {
    await updateMyProduct(row.id, { is_visible: val })
    row.is_visible = val
    ElMessage.success(val ? '产品已显示' : '产品已隐藏')
  } catch (e: any) {
    row.is_visible = !val
    ElMessage.error(e?.message || '操作失败')
  } finally {
    row._visLoading = false
  }
}

async function remove(row: ProductVO) {
  try {
    await deleteMyProduct(row.id)
    ElMessage.success('已删除')
    load()
  } catch (e: any) {
    ElMessage.error(e?.message || '删除失败')
  }
}

function toggleSortMode() {
  sortMode.value = !sortMode.value
  if (!sortMode.value) saveReorder()
}

async function saveReorder() {
  const ids = rows.value.map((r) => r.id)
  if (ids.length < 2) return
  try {
    await reorderProducts(ids as any)
    ElMessage.success('排序已保存')
    load()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存排序失败')
  }
}

onMounted(load)
</script>

<style scoped>
.studio-page{ padding: 8px 4px 24px; }
.tool-bar{ display:flex; justify-content: space-between; align-items:center; gap: 12px; flex-wrap: wrap; }
.kw{ width: 260px; max-width: 100%; }
.bar-actions{ display: flex; gap: 10px; flex-wrap: wrap; }
.p-logo-mini{ width: 40px; height:40px; border-radius:8px; margin-right: 12px; flex-shrink: 0;}
.prod-name{ display:flex; align-items:center; }
.prod-name .strong{ font-weight: 600; color:#0F172A; }
.prod-name .sub{ font-size: 12px; color: #64748B; }
.prod-link{ color: inherit; text-decoration: none; }
.sort-handle{ cursor: grab; }
@media (max-width: 768px) {
  .kw { width: 100%; }
  .bar-actions { width: 100%; }
  .bar-actions .el-button { flex: 1; }
}
</style>
