<template>
  <el-container class="d">
    <app-aside /><el-container>
      <app-header crumb="举报处理" />
      <el-main>
        <div class="bar">
          <el-select v-model="statusFilter" placeholder="处理状态" clearable style="width:160px" @change="load">
            <el-option label="待处理" value="pending" />
            <el-option label="处理中" value="processing" />
            <el-option label="已处理" value="resolved" />
            <el-option label="忽略" value="ignored" />
          </el-select>
          <el-select v-model="typeFilter" placeholder="举报对象" clearable style="width:140px" @change="load">
            <el-option label="创业者" value="founder" />
            <el-option label="产品" value="product" />
          </el-select>
          <el-select v-model="reasonFilter" placeholder="原因分类" clearable style="width:160px" @change="load">
            <el-option v-for="o in reasonOptions" :key="o.value" :label="o.label" :value="o.value" />
          </el-select>
          <span class="spacer" />
          <el-button @click="load">刷新</el-button>
        </div>
        <el-table :data="rows" stripe v-loading="loading" style="margin-top:12px;">
          <el-table-column label="类型" width="100">
            <template #default="{row}">
              <el-tag size="small" :type="row.target_type==='founder' ? 'primary' : 'success'">{{ row.target_type === 'founder' ? '创业者' : '产品' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="目标" min-width="240">
            <template #default="{row}">
              <router-link
                v-if="row.target_type==='founder' && (row.target_meta as any)?.slug"
                :to="`/@${(row.target_meta as any)?.slug}`"
                target="_blank"
                class="target-link"
              >
                <el-icon><UserFilled /></el-icon>
                {{ row.target_display_name || '#' + row.target_id }}
              </router-link>
              <router-link
                v-else-if="row.target_type==='product' && (row.target_meta as any)?.slug && (row.target_meta as any)?.founder_slug"
                :to="`/@${(row.target_meta as any)?.founder_slug}/${(row.target_meta as any)?.slug}`"
                target="_blank"
                class="target-link"
              >
                <el-icon><Goods /></el-icon>
                {{ row.target_display_name || '#' + row.target_id }}
              </router-link>
              <span v-else class="target-link na">
                <el-icon><QuestionFilled /></el-icon>
                {{ row.target_display_name || '#' + row.target_id }}
              </span>
              <el-tag v-if="(row.target_meta as any)?.is_visible===0" size="small" type="info" style="margin-left:6px;">已隐藏</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="原因分类" width="140">
            <template #default="{row}">{{ reasonLabel(row.reason_type) }}</template>
          </el-table-column>
          <el-table-column label="详情" min-width="220" show-overflow-tooltip prop="reason_detail" />
          <el-table-column label="举报人" width="160">
            <template #default="{row}">
              {{ row.reporter_name || '匿名' }}
              <div v-if="row.contact_email" class="sub">{{ row.contact_email }}</div>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="110">
            <template #default="{row}">
              <el-tag size="small" :type="statusTagType(row.status)">{{ statusLabel(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="时间" width="170">
            <template #default="{row}">{{ fmt(row.created_at) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="260">
            <template #default="{row}">
              <el-button link type="primary" @click="openView(row)">详情</el-button>
              <el-button
                v-if="row.status !== 'resolved' && (row.target_meta as any)?.is_visible !== 0"
                link
                type="warning"
                @click="openHide(row)"
              >隐藏目标</el-button>
              <el-button
                v-if="row.status !== 'resolved' && row.status !== 'ignored'"
                link
                type="primary"
                @click="markResolved(row)"
              >标记已处理</el-button>
              <el-button
                v-if="row.status !== 'ignored'"
                link
                @click="markIgnored(row)"
              >忽略</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="pg">
          <el-pagination :current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="onPage" />
        </div>

        <el-dialog v-model="viewShow" width="600px" title="举报详情">
          <el-descriptions v-if="viewing && viewRow" :column="1" border>
            <el-descriptions-item label="举报类型">{{ viewRow.target_type==='founder' ? '创业者' : '产品' }}</el-descriptions-item>
            <el-descriptions-item label="对象ID / 显示名">#{{ viewRow.target_id }} · {{ viewRow.target_display_name || '—' }}</el-descriptions-item>
            <el-descriptions-item label="原因分类">{{ reasonLabel(viewRow.reason_type) }}</el-descriptions-item>
            <el-descriptions-item label="详细描述">
              <div style="white-space: pre-wrap; line-height:1.6; max-height: 200px; overflow: auto;">{{ viewRow.reason_detail || '—' }}</div>
            </el-descriptions-item>
            <el-descriptions-item label="举报人">{{ viewRow.reporter_name || '匿名' }} {{ viewRow.contact_email ? ' <' + viewRow.contact_email + '>' : '' }}</el-descriptions-item>
            <el-descriptions-item label="当前状态">{{ statusLabel(viewRow.status) }} · {{ fmt(viewRow.created_at) }}</el-descriptions-item>
            <el-descriptions-item label="对象元信息">
              <pre class="meta-json">{{ JSON.stringify(viewRow.target_meta || {}, null, 2) }}</pre>
            </el-descriptions-item>
          </el-descriptions>
        </el-dialog>

        <el-dialog v-model="hideShow" width="480px" title="隐藏目标">
          <el-alert type="warning" :closable="false" show-icon style="margin-bottom: 12px;">
            执行后会将被举报的 {{ (hideRow && hideRow.target_type) === 'founder' ? '创业者' : '产品' }} 设为「不显示」，访客端将无法访问。
          </el-alert>
          <el-form label-width="100px" :model="hideForm">
            <el-form-item label="处理备注"><el-input v-model="hideForm.admin_note" type="textarea" :rows="3" placeholder="用于内部记录，可选" /></el-form-item>
          </el-form>
          <template #footer>
            <el-button @click="hideShow=false">取消</el-button>
            <el-button type="danger" :loading="hideSaving" @click="confirmHide">确认隐藏</el-button>
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
import { UserFilled, Goods, QuestionFilled } from '@element-plus/icons-vue'
import {
  listReports,
  handleReport,
  type ReportVO,
} from '@/api/content'

const statusFilter = ref<string>('')
const typeFilter = ref<string>('')
const reasonFilter = ref<string>('')
const rows = ref<ReportVO[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const viewShow = ref(false)
const viewing = ref(false)
const viewRow = ref<ReportVO | null>(null)

const hideShow = ref(false)
const hideRow = ref<ReportVO | null>(null)
const hideSaving = ref(false)
const hideForm = reactive({ admin_note: '' })

const reasonOptions = [
  { value: 'spam', label: '垃圾/营销' },
  { value: 'inappropriate', label: '不当内容' },
  { value: 'illegal', label: '违法违规' },
  { value: 'fake', label: '虚假信息' },
  { value: 'infringement', label: '侵权' },
  { value: 'other', label: '其他原因' },
]
function reasonLabel(v: string) {
  return reasonOptions.find((o) => o.value === v)?.label || v
}
function statusLabel(s: string) {
  const m: Record<string, string> = { pending: '待处理', processing: '处理中', resolved: '已处理', ignored: '忽略' }
  return m[s] || s
}
function statusTagType(s: string) {
  const m: Record<string, any> = { pending: 'warning', processing: '', resolved: 'success', ignored: 'info' }
  return m[s] || 'info'
}
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
    if (statusFilter.value) params.status = statusFilter.value
    if (typeFilter.value) params.target_type = typeFilter.value
    if (reasonFilter.value) params.reason_type = reasonFilter.value
    const r = await listReports(params)
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

function openView(row: ReportVO) {
  viewRow.value = row
  viewing.value = true
  viewShow.value = true
}
function openHide(row: ReportVO) {
  hideRow.value = row
  hideForm.admin_note = ''
  hideShow.value = true
}
async function confirmHide() {
  if (!hideRow.value) return
  hideSaving.value = true
  try {
    await handleReport(hideRow.value.id, {
      action: 'hide_target',
      admin_note: hideForm.admin_note || undefined,
    })
    ElMessage.success('目标已隐藏')
    hideShow.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    hideSaving.value = false
  }
}
async function markResolved(row: ReportVO) {
  try {
    await handleReport(row.id, { action: 'resolve' })
    row.status = 'resolved'
    ElMessage.success('已标记处理完成')
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}
async function markIgnored(row: ReportVO) {
  try {
    await handleReport(row.id, { action: 'ignore' })
    row.status = 'ignored'
    ElMessage.success('已忽略')
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
.target-link{ display: inline-flex; gap:6px; align-items:center; color: #2563eb; text-decoration:none;}
.target-link.na{ color:#94A3B8; cursor:default;}
.target-link:hover{ text-decoration: underline; }
.sub{ font-size: 12px; color:#64748B; }
.pg{ margin-top: 20px; display:flex; justify-content: flex-end; }
.meta-json{ background:#F8FAFC; padding: 10px; border-radius: 6px; font-size: 12px; overflow: auto; max-height: 180px;}
</style>
