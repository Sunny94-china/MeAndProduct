<template>
  <div class="studio-page">
        <el-form ref="formRef" :model="form" :rules="rules" :label-width="formLabelWidth" :label-position="formLabelPos" @submit.prevent="save">
          <el-row :gutter="20">
            <el-col :md="16" :xs="24">
              <el-card shadow="never">
                <template #header>基本信息</template>
                <el-form-item label="显示名称" prop="display_name">
                  <el-input v-model="form.display_name" placeholder="你的名字或昵称" maxlength="100" show-word-limit />
                </el-form-item>
                <el-form-item label="Slug" prop="slug">
                  <el-input v-model="form.slug" placeholder="url-friendly，如 zhang-san" />
                  <div class="tip">名片 URL：/ @{{ form.slug || 'your-slug' }}</div>
                </el-form-item>
                <el-form-item label="头衔">
                  <el-input v-model="form.title" placeholder="独立开发者 / 产品经理" maxlength="100" show-word-limit />
                </el-form-item>
                <el-form-item label="一句话介绍" prop="bio">
                  <el-input
                    v-model="form.bio"
                    type="textarea"
                    :rows="2"
                    maxlength="200"
                    show-word-limit
                    placeholder="会显示在卡片的头像下面"
                  />
                </el-form-item>
                <el-form-item label="详细介绍">
                  <el-input
                    v-model="form.bio_long"
                    type="textarea"
                    :rows="8"
                    placeholder="支持 Markdown，会显示在“关于”Tab"
                    maxlength="20000"
                    show-word-limit
                  />
                </el-form-item>
                <el-form-item label="所在地区">
                  <el-input v-model="form.location" placeholder="北京 / Remote" />
                </el-form-item>
                <el-form-item label="公开邮箱" prop="public_email">
                  <el-input v-model="form.public_email" placeholder="example@mail.com" />
                </el-form-item>
                <el-form-item label="个人官网">
                  <el-input v-model="form.website_url" placeholder="https://your-site.com" />
                </el-form-item>
                <el-form-item label="标签">
                  <el-select
                    v-model="form.tag_ids"
                    multiple
                    filterable
                    placeholder="选择标签（从后台创建的标签中选择）"
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
              </el-card>

              <el-card shadow="never" class="mt20">
                <template #header>
                  <div class="card-head">
                    <span>社交链接 / 外链</span>
                    <el-button size="small" type="primary" plain @click="addLink">+ 新增链接</el-button>
                  </div>
                </template>
                <div v-for="(l, i) in form.custom_links" :key="i" class="link-row">
                  <el-select v-model="l.link_type" placeholder="类型" style="width: 150px; flex-shrink: 0;">
                    <el-option v-for="o in linkTypeOptions" :key="o.value" :label="o.label" :value="o.value" />
                  </el-select>
                  <el-input v-model="l.name" placeholder="显示名称（可选）" style="width: 180px; flex-shrink: 0;" />
                  <el-input v-model="l.url" placeholder="https://..." />
                  <el-button link type="danger" @click="removeLink(i)">删除</el-button>
                </div>
                <div v-if="!form.custom_links.length" class="empty-links">
                  暂无链接，点击右上角「新增链接」
                </div>
              </el-card>
            </el-col>

            <el-col :md="8" :xs="24">
              <el-card shadow="never">
                <template #header>素材</template>
                <el-form label-width="80px">
                  <el-form-item label="头像">
                    <div class="avatar-upload">
                      <el-avatar :size="96" :src="form.avatar_url">{{ avatarText }}</el-avatar>
                    </div>
                    <el-input v-model="form.avatar_url" placeholder="头像 URL（图片上传功能后续接入）" class="mt10" />
                  </el-form-item>
                  <el-form-item label="封面">
                    <div class="cover-preview" v-if="form.cover_url" :style="{ backgroundImage: `url(${form.cover_url})` }"></div>
                    <div class="cover-placeholder" v-else>封面图 预览区</div>
                    <el-input v-model="form.cover_url" placeholder="封面图 URL" class="mt10" />
                  </el-form-item>
                </el-form>
              </el-card>
              <el-card shadow="never" class="mt20">
                <template #header>名片效果预览</template>
                <div class="preview-card">
                  <div class="row">
                    <el-avatar :size="56" :src="form.avatar_url">{{ avatarText }}</el-avatar>
                    <div>
                      <div class="nm">{{ form.display_name || '你的名字' }}</div>
                      <div class="ttl">{{ form.title || '独立开发者' }}</div>
                    </div>
                  </div>
                  <p class="bio">{{ form.bio || '一句话介绍自己' }}</p>
                  <div class="link-tags"><el-tag v-for="l in (form.custom_links||[]).slice(0,4)" :key="l.link_type+l.url" size="small" effect="plain" round>{{ linkLabel(l) }}</el-tag></div>
                </div>
              </el-card>
              <div class="mt20 ta-right">
                <el-button v-if="!isNew" type="success" plain @click="goPreview">预览名片</el-button>
                <el-button type="primary" :loading="saving" @click="save">保存修改</el-button>
              </div>
            </el-col>
          </el-row>
        </el-form>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import {
  applyProfile,
  getMyProfile,
  updateMyProfile,
  listActiveTags,
  type FounderProfileItemVO,
  type TagVO,
} from '@/api/content'

const router = useRouter()
const formRef = ref<FormInstance>()
const saving = ref(false)
const isNew = ref(true)
const loaded = ref(false)
const winW = ref(typeof window !== 'undefined' ? window.innerWidth : 1024)
const formLabelPos = computed(() => (winW.value < 768 ? 'top' : 'right'))
const formLabelWidth = computed(() => (winW.value < 768 ? 'auto' : '100px'))
function onResize() {
  winW.value = window.innerWidth
}
onMounted(() => window.addEventListener('resize', onResize))
onUnmounted(() => window.removeEventListener('resize', onResize))

const tagOptions = ref<TagVO[]>([])
const loadingTags = ref(false)

function typeLabel(t: string) {
  const m: Record<string, string> = { industry: '行业', stage: '阶段', role: '角色', other: '其他' }
  return m[t] || t
}
function linkLabel(l: { link_type: string; name: string }) {
  const m: Record<string, string> = { github: 'GitHub', twitter: 'X/Twitter', linkedin: 'LinkedIn', wechat: '微信', zhihu: '知乎', xiaohongshu: '小红书', producthunt: 'ProductHunt', official: '官网', blog: '博客', other: '链接' }
  return l.name || m[l.link_type] || l.link_type
}
const linkTypeOptions = [
  { value: 'github', label: 'GitHub' },
  { value: 'twitter', label: 'X / Twitter' },
  { value: 'linkedin', label: 'LinkedIn' },
  { value: 'wechat', label: '微信' },
  { value: 'zhihu', label: '知乎' },
  { value: 'xiaohongshu', label: '小红书' },
  { value: 'producthunt', label: 'Product Hunt' },
  { value: 'official', label: '个人官网' },
  { value: 'blog', label: '博客' },
  { value: 'other', label: '其他链接' },
]

const form = reactive({
  display_name: '',
  slug: '',
  title: '',
  bio: '',
  bio_long: '',
  avatar_url: '',
  cover_url: '',
  location: '',
  public_email: '',
  website_url: '',
  tag_ids: [] as number[],
  custom_links: [] as { id?: number; link_type: string; name: string; url: string }[],
})

const rules: FormRules = {
  display_name: [
    { required: true, message: '请输入显示名称', trigger: 'blur' },
    { min: 2, max: 100, message: '长度需在 2~100 之间', trigger: 'blur' },
  ],
  bio: [{ max: 500, message: '最多 500 字', trigger: 'blur' }],
  public_email: [{ type: 'email', message: '请输入正确的邮箱', trigger: 'blur' }],
}

const avatarText = computed(() => (form.display_name || 'U').slice(0, 1).toUpperCase())

function addLink() {
  form.custom_links.push({ link_type: 'github', name: '', url: '' })
}
function removeLink(i: number) {
  form.custom_links.splice(i, 1)
}

async function searchTags(q: string) {
  loadingTags.value = true
  try {
    const list = await listActiveTags()
    const kw = (q || '').toLowerCase()
    tagOptions.value = (list || []).filter((t: TagVO) => !kw || t.name.toLowerCase().includes(kw) || t.slug.toLowerCase().includes(kw))
  } catch (_e) {} finally {
    loadingTags.value = false
  }
}

async function loadProfile() {
  try {
    const p = await getMyProfile()
    isNew.value = !p
    if (p) fillFromVO(p)
  } catch (_e) {
    isNew.value = true
  } finally {
    loaded.value = true
  }
}

function fallbackSlug(name: string) {
  const s = (name || '').toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '').slice(0, 40)
  return s.length >= 3 ? s : `founder-${Date.now().toString(36)}`
}

function fillFromVO(p: FounderProfileItemVO) {
  form.display_name = p.display_name || ''
  form.slug = p.slug || ''
  form.title = p.title || ''
  form.bio = p.bio || ''
  form.bio_long = p.bio_long || ''
  form.avatar_url = p.avatar_url || ''
  form.cover_url = p.cover_url || ''
  form.location = p.location || ''
  form.public_email = p.contact_email || p.public_email || ''
  form.website_url = p.website_url || ''
  form.tag_ids = (p.tags || []).map((x: any) => Number(x.id))
  form.custom_links = (p.custom_links || []).map((l: any) => ({ id: l.id, link_type: l.link_type, name: l.name, url: l.url }))
}

async function save() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch (_e) {
    return
  }
  saving.value = true
  try {
    if (!form.slug) form.slug = fallbackSlug(form.display_name)
    const links = form.custom_links.filter((l) => !!l.url).map((l) => ({ link_type: l.link_type, url: l.url }))
    const payload: any = {
      display_name: form.display_name,
      title: form.title || undefined,
      bio: form.bio || undefined,
      bio_long: form.bio_long || undefined,
      slug: form.slug,
      avatar_url: form.avatar_url || undefined,
      cover_url: form.cover_url || undefined,
      location: form.location || undefined,
      contact_email: form.public_email || undefined,
      website_url: form.website_url || undefined,
      tag_ids: form.tag_ids,
      custom_links: links,
    }
    const result = isNew.value
      ? await applyProfile(payload)
      : await updateMyProfile(payload)
    ElMessage.success(isNew.value ? '名片已创建！' : '名片已保存')
    if (result && !isNew.value === false) {
      isNew.value = false
    } else {
      isNew.value = false
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

function goPreview() {
  if (form.slug) router.push(`/@${form.slug}`)
  else ElMessage.warning('先设置 Slug 再预览')
}

onMounted(async () => {
  await searchTags('')
  loadProfile()
})
</script>

<style scoped>
.studio-page{ padding: 8px 4px 24px; }
.mt10{ margin-top: 10px; }
.mt20{ margin-top: 20px; }
.ta-right{ text-align: right; }
.tip{ font-size: 12px; color: #64748B; margin-top: 4px; }
.card-head{ display: flex; justify-content: space-between; align-items: center; }
.link-row{ display: flex; gap: 10px; align-items: center; margin-bottom: 10px; }
.empty-links{ text-align: center; padding: 20px; color: #94A3B8; }
.avatar-upload{ text-align: left; }
.cover-placeholder{ height: 120px; border:1px dashed #CBD5E1; border-radius: 8px; display:flex; align-items:center; justify-content:center; color:#94A3B8; background:#F8FAFC;}
.cover-preview{ height: 120px; background-size: cover; background-position: center; border-radius: 8px; }
.preview-card{ padding: 16px; background: #F8FAFC; border: 1px solid #E2E8F0; border-radius: 10px; }
.preview-card .row{ display: flex; gap: 12px; align-items: center; margin-bottom: 10px; }
.preview-card .nm{ font-weight: 700; font-size: 16px; }
.preview-card .ttl{ font-size: 12px; color: #64748B; }
.preview-card .bio{ font-size: 13px; color: #334155; line-height: 1.5; margin: 10px 0; }
.link-tags{ display: flex; flex-wrap: wrap; gap: 6px; }
@media (max-width: 768px) {
  .link-row{ flex-wrap: wrap; }
  .ta-right{ text-align: left; }
  .ta-right .el-button{ width: 100%; }
}
</style>
