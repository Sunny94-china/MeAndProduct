<template>
  <div class="studio-page">
        <el-card shadow="never" v-loading="loading">
          <el-form ref="formRef" label-width="110px" :model="f" :rules="rules" @submit.prevent="save">
            <el-row :gutter="20">
              <el-col :md="16" :xs="24">
                <el-form-item label="产品名称" prop="name">
                  <el-input v-model="f.name" maxlength="200" show-word-limit placeholder="你的产品名称" />
                </el-form-item>
                <el-form-item label="Slug" prop="slug">
                  <el-input v-model="f.slug" placeholder="URL-friendly，如 my-app" />
                  <div class="tip">访问路径：/@名片Slug/{{ f.slug || 'product-slug' }}</div>
                </el-form-item>
                <el-form-item label="一句话标语" prop="tagline">
                  <el-input v-model="f.tagline" maxlength="200" show-word-limit placeholder="让访客一眼记住你的产品" />
                </el-form-item>
                <el-form-item label="产品官网" prop="official_url">
                  <el-input v-model="f.official_url" placeholder="https://..." />
                </el-form-item>
                <el-form-item label="详细描述">
                  <el-input
                    v-model="f.description"
                    type="textarea"
                    :rows="8"
                    maxlength="10000"
                    show-word-limit
                    placeholder="支持 Markdown，介绍产品功能、目标用户、故事等"
                  />
                </el-form-item>
                <el-form-item label="阶段" prop="status">
                  <el-select v-model="f.status" style="width: 100%">
                    <el-option label="想法阶段" value="idea" />
                    <el-option label="内测中" value="beta" />
                    <el-option label="已上线" value="launched" />
                    <el-option label="增长期" value="growing" />
                    <el-option label="已归档" value="archived" />
                  </el-select>
                </el-form-item>
                <el-form-item label="标签">
                  <el-select
                    v-model="f.tag_ids"
                    multiple
                    filterable
                    placeholder="选择标签（最多 20 个）"
                    style="width:100%"
                    remote
                    :remote-method="searchTags"
                    :loading="loadingTags"
                  >
                    <el-option
                      v-for="t in tagOptions"
                      :key="t.id"
                      :label="`${t.name} (${typeLabel(t.type)})`"
                      :value="t.id"
                    />
                  </el-select>
                </el-form-item>
              </el-col>

              <el-col :md="8" :xs="24">
                <el-form-item label="产品 Logo">
                  <div class="logo-preview">
                    <div class="logo-box" :style="logoStyle">
                      <el-icon v-if="!f.logo_url" :size="28" color="#fff"><Goods /></el-icon>
                      <img v-else :src="f.logo_url" class="logo-img" alt="" />
                    </div>
                  </div>
                  <el-input v-model="f.logo_url" placeholder="Logo 图片 URL（上传功能后续接入）" class="mt10" />
                </el-form-item>
                <el-form-item label="封面图">
                  <div class="cover-preview" v-if="f.cover_url" :style="{ backgroundImage: `url(${f.cover_url})` }"></div>
                  <div class="cover-placeholder" v-else>封面图 预览区</div>
                  <el-input v-model="f.cover_url" placeholder="封面图 URL" class="mt10" />
                </el-form-item>
                <el-form-item label="显示状态" prop="is_visible">
                  <el-switch v-model="f.is_visible" active-text="公开显示" inactive-text="暂时隐藏" />
                </el-form-item>
              </el-col>
            </el-row>

            <el-divider content-position="left">🔗 自定义外链</el-divider>

            <div class="links-bar">
              <span class="hint">GitHub、Product Hunt、博客、文档、演示视频等链接</span>
              <el-button size="small" type="primary" plain @click="addLink">+ 新增链接</el-button>
            </div>
            <div v-for="(l, i) in f.custom_links" :key="i" class="link-row">
              <el-select v-model="l.link_type" placeholder="类型" style="width: 160px; flex-shrink: 0;">
                <el-option v-for="o in linkTypeOptions" :key="o.value" :label="o.label" :value="o.value" />
              </el-select>
              <el-input v-model="l.name" placeholder="显示名称（可选）" style="width: 180px; flex-shrink: 0;" />
              <el-input v-model="l.url" placeholder="https://..." />
              <el-button link type="danger" @click="removeLink(i)">删除</el-button>
            </div>
            <div v-if="!f.custom_links.length" class="empty-links">暂无链接</div>

            <div class="actions">
              <el-button @click="$router.push('/studio/products')">返回列表</el-button>
              <el-button type="primary" :loading="saving" @click="save">保存</el-button>
            </div>
          </el-form>
        </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Goods } from '@element-plus/icons-vue'
import {
  createProduct,
  getMyProduct,
  updateMyProduct,
  listActiveTags,
  type ProductVO,
  type TagVO,
} from '@/api/content'

const route = useRoute()
const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)
const saving = ref(false)
const isNew = computed(() => !route.params.id)

const tagOptions = ref<TagVO[]>([])
const loadingTags = ref(false)

function typeLabel(t: string) {
  const m: Record<string, string> = { industry: '行业', stage: '阶段', role: '角色', other: '其他' }
  return m[t] || t
}

const linkTypeOptions = [
  { value: 'github', label: 'GitHub' },
  { value: 'producthunt', label: 'Product Hunt' },
  { value: 'official', label: '官方站点' },
  { value: 'blog', label: '博客' },
  { value: 'docs', label: '文档' },
  { value: 'demo', label: '演示视频' },
  { value: 'twitter', label: 'X / Twitter' },
  { value: 'other', label: '其他' },
]

const f = reactive({
  name: '',
  slug: '',
  tagline: '',
  description: '',
  official_url: '',
  logo_url: '',
  cover_url: '',
  status: 'idea',
  is_visible: true,
  tag_ids: [] as number[],
  custom_links: [] as { id?: number; link_type: string; name: string; url: string }[],
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入产品名称', trigger: 'blur' },
    { min: 2, max: 200, message: '长度在 2~200 之间', trigger: 'blur' },
  ],
  tagline: [
    { required: true, message: '请填写一句话标语', trigger: 'blur' },
    { max: 200, message: '最多 200 字', trigger: 'blur' },
  ],
  official_url: [
    { required: true, message: '请填写产品官网（外部落地页）', trigger: 'blur' },
    { type: 'url', message: '请以 http(s):// 开头', trigger: 'blur' },
  ],
  status: [{ required: true, message: '请选择阶段', trigger: 'change' }],
}

const logoStyle = computed(() => {
  if (f.logo_url) return { backgroundImage: `url(${f.logo_url})`, backgroundSize: 'cover', backgroundPosition: 'center' }
  const palette = [
    'linear-gradient(135deg,#6366F1,#A855F7)',
    'linear-gradient(135deg,#EC4899,#F97316)',
    'linear-gradient(135deg,#10B981,#06B6D4)',
    'linear-gradient(135deg,#F43F5E,#8B5CF6)',
    'linear-gradient(135deg,#F59E0B,#EF4444)',
    'linear-gradient(135deg,#3B82F6,#14B8A6)',
  ]
  return { background: palette[(route.params.id ? Number(route.params.id) : 0) % palette.length] }
})

function fallbackSlug(name: string) {
  const s = (name || '').toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '').slice(0, 40)
  return s.length >= 3 ? s : `product-${Date.now().toString(36)}`
}

function addLink() { f.custom_links.push({ link_type: 'github', name: '', url: '' }) }
function removeLink(i: number) { f.custom_links.splice(i, 1) }

async function searchTags(q: string) {
  loadingTags.value = true
  try {
    const list = await listActiveTags()
    const kw = (q || '').toLowerCase()
    tagOptions.value = (list || []).filter((t) => !kw || t.name.toLowerCase().includes(kw) || t.slug.toLowerCase().includes(kw))
  } catch (_e) {} finally {
    loadingTags.value = false
  }
}

async function loadForEdit() {
  if (isNew.value) return
  loading.value = true
  try {
    const p: ProductVO = await getMyProduct(route.params.id as any)
    f.name = p.name
    f.slug = p.slug
    f.tagline = p.tagline
    f.description = p.description || ''
    f.official_url = p.official_url || ''
    f.logo_url = p.logo_url || ''
    f.cover_url = p.cover_url || ''
    f.status = p.status
    f.is_visible = !!p.is_visible
    f.tag_ids = (p.tags || []).map((t: any) => Number(t.id))
    f.custom_links = (p.custom_links || []).map((l: any) => ({ id: l.id, link_type: l.link_type, name: l.name, url: l.url }))
  } catch (e: any) {
    ElMessage.error(e?.message || '加载产品信息失败')
  } finally {
    loading.value = false
  }
}

async function save() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch (_e) { return }
  saving.value = true
  try {
    if (!f.slug) f.slug = fallbackSlug(f.name)
    const payload: any = {
      name: f.name,
      tagline: f.tagline,
      official_url: f.official_url,
      slug: f.slug,
      description: f.description || undefined,
      logo_url: f.logo_url || undefined,
      cover_url: f.cover_url || undefined,
      status: f.status,
      is_visible: f.is_visible,
      tag_ids: f.tag_ids.length ? f.tag_ids : undefined,
      custom_links: f.custom_links.filter((l) => !!l.url).length ? f.custom_links.filter((l) => !!l.url) : undefined,
    }
    if (isNew.value) await createProduct(payload)
    else await updateMyProduct(route.params.id as any, payload)
    ElMessage.success(isNew.value ? '产品已创建！' : '产品已保存')
    router.push('/studio/products')
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  await searchTags('')
  loadForEdit()
})
</script>

<style scoped>
.studio-page{ padding: 8px 4px 24px; }
.mt10{ margin-top: 10px;}
.tip{ font-size: 12px; color: #64748B; margin-top: 4px;}
.logo-preview{ display:flex; }
.logo-box{ width: 88px; height: 88px; border-radius: 14px; overflow: hidden; display:flex; align-items:center; justify-content:center; background:#E0E7FF; }
.logo-img{ width: 100%; height: 100%; object-fit: cover; }
.cover-placeholder{ height: 120px; border:1px dashed #CBD5E1; border-radius: 8px; display:flex; align-items:center; justify-content:center; color:#94A3B8; background:#F8FAFC;}
.cover-preview{ height: 120px; background-size: cover; background-position: center; border-radius: 8px; }
.links-bar{ display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.hint{ color: #64748B; font-size: 13px; }
.link-row{ display: flex; gap: 10px; align-items: center; margin-bottom: 10px; }
.empty-links{ text-align: center; padding: 18px; color: #94A3B8; font-size: 13px; }
.actions{ text-align: right; margin-top: 20px; }
</style>
