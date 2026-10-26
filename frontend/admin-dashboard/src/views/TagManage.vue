<template>
  <el-container class="d">
    <app-aside /><el-container>
      <app-header crumb="标签管理" />
      <el-main>
        <div class="bar">
          <el-tabs v-model="typeTab" class="type-tabs" @tab-change="onType">
            <el-tab-pane label="全部" value="all" />
            <el-tab-pane label="行业 industry" value="industry" />
            <el-tab-pane label="阶段 stage" value="stage" />
            <el-tab-pane label="角色 role" value="role" />
            <el-tab-pane label="其他 other" value="other" />
          </el-tabs>
          <el-input v-model="kw" placeholder="按名称搜索" clearable style="width:240px" :prefix-icon="Search" @clear="load" @keyup.enter="load" />
          <span class="spacer" />
          <el-button :icon="Plus" type="primary" @click="openNew">+ 新增标签</el-button>
        </div>
        <el-table :data="rows" stripe v-loading="loading" style="margin-top:12px;">
          <el-table-column label="名称" width="220">
            <template #default="{row}">
              <el-tag :type="typeTagColor(row.type)" effect="light" size="small" style="margin-right:8px;">
                {{ typeLabel(row.type) }}
              </el-tag>
              <strong>{{ row.name }}</strong>
              <span class="slug"> · {{ row.slug }}</span>
            </template>
          </el-table-column>
          <el-table-column label="描述" prop="description" show-overflow-tooltip />
          <el-table-column label="创业者数" width="100" prop="founder_count" sortable />
          <el-table-column label="产品数" width="100" prop="product_count" sortable />
          <el-table-column label="状态" width="110">
            <template #default="{row}">
              <el-tag :type="row.is_active ? 'success' : 'info'" size="small">
                {{ row.is_active ? '启用' : '已停用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="220">
            <template #default="{row}">
              <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
              <el-button
                v-if="row.is_active"
                link
                type="warning"
                @click="deactivate(row)"
              >停用</el-button>
              <el-button
                v-else
                link
                type="success"
                @click="reactivate(row)"
              >启用</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="pg">
          <el-pagination :current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="onPage" />
        </div>

        <el-dialog v-model="showDlg" :title="isNew ? '新增标签' : '编辑标签'" width="520px">
          <el-form ref="formRef" label-width="100px" :model="form" :rules="rules">
            <el-form-item label="名称" prop="name"><el-input v-model="form.name" placeholder="例：SaaS" /></el-form-item>
            <el-form-item label="Slug" prop="slug"><el-input v-model="form.slug" placeholder="例：saas" /></el-form-item>
            <el-form-item label="类型" prop="type">
              <el-select v-model="form.type" style="width:100%">
                <el-option label="行业 industry" value="industry" />
                <el-option label="阶段 stage" value="stage" />
                <el-option label="角色 role" value="role" />
                <el-option label="其他 other" value="other" />
              </el-select>
            </el-form-item>
            <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="3" /></el-form-item>
            <el-form-item label="启用">
              <el-switch v-model="form.is_active" active-text="启用" inactive-text="停用" />
            </el-form-item>
          </el-form>
          <template #footer>
            <el-button @click="showDlg=false">取消</el-button>
            <el-button type="primary" :loading="saving" @click="save">保存</el-button>
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
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Search, Plus } from '@element-plus/icons-vue'
import {
  listTags,
  createTag,
  updateTag,
  deactivateTag,
  type TagVO,
} from '@/api/content'

const kw = ref('')
const typeTab = ref<'all' | TagVO['type']>('all')
const rows = ref<TagVO[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(50)
const loading = ref(false)
const showDlg = ref(false)
const saving = ref(false)
const isNew = ref(true)
const editingId = ref<number | null>(null)
const formRef = ref<FormInstance>()

const form = reactive({
  name: '',
  slug: '',
  type: 'industry' as TagVO['type'],
  description: '',
  is_active: true,
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  slug: [{ required: true, message: '请输入 Slug', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
}

function typeLabel(t: string) {
  const m: Record<string, string> = { industry: '行业', stage: '阶段', role: '角色', other: '其他' }
  return m[t] || t
}
function typeTagColor(t: string): any {
  const m: Record<string, any> = { industry: '', stage: 'warning', role: 'success', other: 'info' }
  return m[t] || ''
}

async function load() {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: pageSize.value }
    if (typeTab.value !== 'all') params.type = typeTab.value
    if (kw.value) params.keyword = kw.value
    const r = await listTags(params)
    rows.value = r.items || []
    total.value = r.total || 0
  } catch (_e) {
    rows.value = []
  } finally {
    loading.value = false
  }
}
function onType() {
  page.value = 1
  load()
}
function onPage(p: number) {
  page.value = p
  load()
}
function openNew() {
  isNew.value = true
  editingId.value = null
  Object.assign(form, { name: '', slug: '', type: 'industry' as TagVO['type'], description: '', is_active: true })
  showDlg.value = true
}
function openEdit(row: TagVO) {
  isNew.value = false
  editingId.value = row.id
  Object.assign(form, { name: row.name, slug: row.slug, type: row.type, description: row.description || '', is_active: !!row.is_active })
  showDlg.value = true
}
async function save() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch (_e) { return }
  saving.value = true
  try {
    if (isNew.value) await createTag({ ...form })
    else await updateTag(editingId.value as number, { ...form })
    ElMessage.success('已保存')
    showDlg.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    saving.value = false
  }
}
async function deactivate(row: TagVO) {
  try {
    await deactivateTag(row.id)
    row.is_active = false
    ElMessage.success('已停用')
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}
async function reactivate(row: TagVO) {
  try {
    await updateTag(row.id, { is_active: true })
    row.is_active = true
    ElMessage.success('已启用')
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

onMounted(load)
</script>
<style scoped>
.d{ min-height:100vh;}
.bar{ display:flex; gap:10px; align-items:center;}
.spacer{ flex:1; }
.slug{ color:#64748B; font-size: 12px;}
.pg{ margin-top: 20px; display:flex; justify-content: flex-end; }
.type-tabs{ margin:0; }
</style>
