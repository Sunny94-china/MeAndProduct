<template>
  <el-container class="d">
    <app-aside /><el-container>
      <app-header crumb="创业者管理" />
      <el-main>
        <div class="bar">
          <el-input
            v-model="kw"
            placeholder="搜索 display_name / slug / email"
            clearable
            style="width:300px"
            :prefix-icon="Search"
            @clear="load"
            @keyup.enter="load"
          />
          <el-select v-model="visFilter" placeholder="可见性" clearable style="width:140px" @change="load">
            <el-option label="全部" value="all" />
            <el-option label="已上架" value="on" />
            <el-option label="已下架" value="off" />
          </el-select>
          <span class="spacer" />
          <el-button :icon="Refresh" type="default" @click="load">刷新</el-button>
        </div>
        <el-table :data="rows" style="margin-top:12px;" v-loading="loading" stripe>
          <el-table-column label="名片" min-width="260">
            <template #default="{row}">
              <div class="founder-row">
                <el-avatar :size="40" :src="row.avatar_url">
                  {{ (row.display_name || 'U').slice(0,1).toUpperCase() }}
                </el-avatar>
                <div class="meta">
                  <div class="name">{{ row.display_name }} <el-tag v-if="row.title" size="small" effect="plain">{{ row.title }}</el-tag></div>
                  <div class="sub">@{{ row.slug }} · {{ row.public_email || '—' }}</div>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="标签" min-width="180">
            <template #default="{row}">
              <el-tag v-for="t in (row.tags||[]).slice(0,6)" :key="t.id" size="small" effect="plain" style="margin-right:4px;">{{ t.name }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="浏览" width="90" prop="view_count" sortable />
          <el-table-column label="收藏" width="90" prop="favorite_count" sortable />
          <el-table-column label="显示/隐藏" width="130">
            <template #default="{row}">
              <el-switch
                :model-value="!!row.is_visible"
                :loading="row._loading"
                active-text="显示"
                inactive-text="隐藏"
                @change="(v:boolean)=>toggleVisible(row,v)"
              />
            </template>
          </el-table-column>
          <el-table-column label="创建时间" width="170">
            <template #default="{row}">{{ fmt(row.created_at) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="180">
            <template #default="{row}">
              <el-button link type="primary" @click="preview(row)">预览</el-button>
              <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="pg">
          <el-pagination
            :current-page="page"
            :page-size="pageSize"
            :total="total"
            layout="total, prev, pager, next"
            @current-change="onPage"
          />
        </div>

        <el-dialog v-model="showEdit" width="600px" title="编辑名片">
          <el-form v-if="editing" label-width="100px" :model="editing">
            <el-form-item label="显示名称"><el-input v-model="editing.display_name" /></el-form-item>
            <el-form-item label="Slug"><el-input v-model="editing.slug" /></el-form-item>
            <el-form-item label="一句话介绍"><el-input v-model="editing.bio" type="textarea" :rows="2" /></el-form-item>
            <el-form-item label="头衔"><el-input v-model="editing.title" /></el-form-item>
            <el-form-item label="公开邮箱"><el-input v-model="editing.public_email" /></el-form-item>
            <el-form-item label="地区"><el-input v-model="editing.location" /></el-form-item>
            <el-form-item label="是否显示"><el-switch v-model="editing.is_visible" active-text="显示" inactive-text="隐藏" /></el-form-item>
          </el-form>
          <template #footer>
            <el-button @click="showEdit=false">取消</el-button>
            <el-button type="primary" :loading="saving" @click="saveEdit">保存</el-button>
          </template>
        </el-dialog>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import AppAside from '@/components/AppAside.vue'
import AppHeader from '@/components/AppHeader.vue'
import { ElMessage } from 'element-plus'
import { Search, Refresh } from '@element-plus/icons-vue'
import {
  listFounders,
  adminUpdateFounder,
  type FounderProfileItemVO,
} from '@/api/content'

const kw = ref('')
const visFilter = ref<'all' | 'on' | 'off'>('all')
const rows = ref<(FounderProfileItemVO & { _loading?: boolean })[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const showEdit = ref(false)
const saving = ref(false)
const editing = ref<any>(null)

function fmt(s?: string) {
  if (!s) return '—'
  const d = new Date(s)
  if (Number.isNaN(d.getTime())) return s.slice(0, 10)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

async function load() {
  loading.value = true
  try {
    const r = await listFounders({
      page: page.value,
      page_size: pageSize.value,
      keyword: kw.value || undefined,
    })
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

async function toggleVisible(row: FounderProfileItemVO & { _loading?: boolean }, val: boolean) {
  row._loading = true
  try {
    await adminUpdateFounder(row.id, { is_visible: val })
    row.is_visible = val
    ElMessage.success(val ? '已上架显示' : '已下架隐藏')
  } catch (e: any) {
    row.is_visible = !val
    ElMessage.error(e?.message || '操作失败')
  } finally {
    row._loading = false
  }
}

function preview(row: FounderProfileItemVO) {
  window.open(`/@${row.slug}`, '_blank')
}
function openEdit(row: FounderProfileItemVO) {
  editing.value = { ...row }
  showEdit.value = true
}

async function saveEdit() {
  if (!editing.value) return
  saving.value = true
  try {
    await adminUpdateFounder(editing.value.id, {
      display_name: editing.value.display_name,
      slug: editing.value.slug,
      bio: editing.value.bio,
      title: editing.value.title,
      public_email: editing.value.public_email,
      location: editing.value.location,
      is_visible: !!editing.value.is_visible,
    })
    ElMessage.success('已保存')
    showEdit.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.d{ min-height:100vh;}
.bar{ display:flex; gap:10px; align-items:center;}
.spacer{ flex:1; }
.founder-row{ display:flex; align-items:center; gap:12px; }
.founder-row .name{ font-weight:600; color:#0F172A; }
.founder-row .sub{ font-size:12px; color:#64748B; margin-top: 4px; }
.pg{ margin-top: 20px; display:flex; justify-content: flex-end; }
</style>
