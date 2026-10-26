<template>
  <div class="product-detail">
    <el-breadcrumb separator="/" class="crumb">
      <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
      <el-breadcrumb-item :to="{ path: `/@${slug}` }" v-if="founderName">{{ founderName }}</el-breadcrumb-item>
      <el-breadcrumb-item>{{ product?.name || '产品详情' }}</el-breadcrumb-item>
    </el-breadcrumb>

    <template v-if="notFound">
      <el-result icon="warning" title="产品不存在或已隐藏" sub-title="返回产品列表继续探索">
        <template #extra>
          <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
        </template>
      </el-result>
    </template>

    <template v-else>
      <el-card shadow="never" class="hero" v-loading="loading">
        <div class="hero-row">
          <div class="logo" :style="logoStyle">
            <el-icon v-if="!product?.logo_url" :size="44" color="#fff"><Goods /></el-icon>
            <img v-else :src="product.logo_url" class="logo-img" alt="" />
          </div>
          <div class="info">
            <div class="name-row">
              <h1 class="name">{{ product?.name || '加载中...' }}</h1>
              <el-tag size="large" type="success" effect="plain" round>{{ stageLabel(product?.status) }}</el-tag>
            </div>
            <div class="tagline">{{ product?.tagline }}</div>
            <div class="tags">
              <el-tag
                v-for="t in (product?.tags || [])"
                :key="t.id"
                type="info"
                effect="plain"
                round
              >{{ t.name }}</el-tag>
            </div>
            <div class="stats">
              <router-link v-if="founder" :to="`/@${founder}`" class="by-link">
                <el-avatar :size="24" :src="founderAvatar">{{ founderAvatarText }}</el-avatar>
                <span>by {{ founderName }}</span>
              </router-link>
              <span><el-icon><View /></el-icon>{{ product?.view_count || 0 }}</span>
              <span><el-icon><Star /></el-icon>{{ displayFavCount }}</span>
            </div>
          </div>
          <div class="actions">
            <a
              v-if="product?.official_url"
              :href="product.official_url"
              class="primary-btn"
              target="_blank"
              rel="noopener"
            >
              <el-icon><Link /></el-icon>访问官方网站
            </a>
            <el-button
              v-if="auth.isLoggedIn()"
              :type="favorited ? 'danger' : 'default'"
              size="large"
              :loading="togglingFav"
              @click="onToggleFav"
            >
              <el-icon v-if="favorited"><StarFilled /></el-icon>
              <el-icon v-else><Star /></el-icon>
              {{ favorited ? '已收藏（取消）' : '收藏产品' }}
            </el-button>
            <el-button size="large" @click="goReport">
              <el-icon><Warning /></el-icon>举报
            </el-button>
          </div>
        </div>
      </el-card>

      <div class="body-row" v-loading="loading">
        <aside class="side">
          <el-card shadow="never" class="side-card">
            <h3 class="side-title">📎 外部链接</h3>
            <div class="link-list">
              <a
                v-if="product?.official_url"
                target="_blank"
                rel="noopener"
                :href="product.official_url"
                class="link-row official"
              >
                <span class="type">🌐 官方网站</span>
                <span class="url">{{ truncate(product.official_url) }}</span>
              </a>
              <a
                v-for="(l, i) in (product?.custom_links || [])"
                :key="i"
                target="_blank"
                rel="noopener"
                class="link-row"
                :href="l.url"
                @click="onLink(l)"
              >
                <span class="type">{{ linkIcon(l.link_type) }} {{ l.name || linkTypeLabel(l.link_type) }}</span>
                <span class="url">{{ truncate(l.url) }}</span>
              </a>
              <p v-if="!product?.official_url && !(product?.custom_links?.length)" class="empty">
                该产品暂未添加外部链接。
              </p>
            </div>
          </el-card>
        </aside>

        <section class="main-col">
          <el-card shadow="never" class="desc-card">
            <h2 class="sec-title">产品介绍</h2>
            <div v-if="product?.description" class="desc-body" v-html="formatDesc(product.description)"></div>
            <p v-else class="empty-desc">该产品暂未填写详细介绍。</p>
          </el-card>
        </section>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Link, Star, StarFilled, View, Warning, Goods } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import {
  getFounderProductDetail,
  toggleFavorite,
  checkFavorite,
  trackLinkClick,
  getFounderDetail,
  type ProductVO,
  type FounderVO,
} from '@/api/content'

const props = defineProps<{ slug: string; productSlug: string }>()
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const product = ref<ProductVO | null>(null)
const founder = ref<FounderVO | null>(null)
const loading = ref(false)
const notFound = ref(false)

const favorited = ref(false)
const displayFavCount = ref(0)
const togglingFav = ref(false)

const founderName = computed(() => founder.value?.display_name || product.value?.founder_display_name || '')
const founderAvatar = computed(() => founder.value?.avatar_url || '')
const founderAvatarText = computed(() => (founderName.value || 'F').slice(0, 1).toUpperCase())

const logoStyle = computed(() => {
  if (product.value?.logo_url) return {}
  const palette = [
    'linear-gradient(135deg,#6366F1,#A855F7)', 'linear-gradient(135deg,#EC4899,#F97316)',
    'linear-gradient(135deg,#10B981,#06B6D4)', 'linear-gradient(135deg,#F43F5E,#8B5CF6)',
    'linear-gradient(135deg,#F59E0B,#EF4444)', 'linear-gradient(135deg,#3B82F6,#14B8A6)',
  ]
  return { background: palette[(product.value?.id || 0) % palette.length] }
})

function stageLabel(s?: string) {
  const m: Record<string, string> = {
    idea: 'Idea', mvp: 'MVP', beta: 'Beta', launched: '已发布', growing: '成长中', mature: '成熟',
  }
  return m[s || ''] || s || '—'
}

function linkTypeLabel(t: string) {
  const m: Record<string, string> = {
    github: 'GitHub', twitter: 'Twitter/X', linkedin: 'LinkedIn', wechat: '微信',
    zhihu: '知乎', xiaohongshu: '小红书', producthunt: 'Product Hunt', official: '官网',
    blog: '博客', other: '其他链接',
  }
  return m[t] || t
}

function linkIcon(t: string) {
  const icons: Record<string, string> = {
    github: '🐙', twitter: '🐦', linkedin: '💼', wechat: '💬', zhihu: '🔖',
    xiaohongshu: '📕', producthunt: '🪄', official: '🔗', blog: '✍️', other: '📎',
  }
  return icons[t] || '🔗'
}

function truncate(s: string, n = 42) {
  return s.length > n ? s.slice(0, n) + '…' : s
}

function formatDesc(s: string) {
  const safe = s.replace(/</g, '&lt;').replace(/>/g, '&gt;')
  return safe.split(/\n\n+/).map((p) => `<p>${p.replace(/\n/g, '<br>')}</p>`).join('')
}

async function load() {
  loading.value = true
  notFound.value = false
  try {
    const [p, f] = await Promise.all([
      getFounderProductDetail(props.slug, props.productSlug),
      getFounderDetail(props.slug).catch(() => null as FounderVO | null),
    ])
    product.value = p
    founder.value = f
    displayFavCount.value = p.favorite_count
    if (auth.isLoggedIn()) {
      try {
        const r = await checkFavorite('product', p.id)
        favorited.value = !!r?.favorited
      } catch (_e) {}
    }
  } catch (e: any) {
    notFound.value = true
    ElMessage.error(e?.message || '产品不存在')
  } finally {
    loading.value = false
  }
}

async function onToggleFav() {
  if (!product.value || togglingFav.value) return
  togglingFav.value = true
  try {
    const r = await toggleFavorite({ target_type: 'product', target_id: Number(product.value.id) })
    favorited.value = !!r?.favorited
    displayFavCount.value = r?.favorite_count ?? displayFavCount.value
    ElMessage.success(favorited.value ? '已收藏该产品' : '已取消收藏')
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    togglingFav.value = false
  }
}

async function onLink(l: { id?: number; url: string }) {
  if (l.id) {
    try { await trackLinkClick(l.id) } catch (_e) {}
  }
}

function goReport() {
  router.push({
    path: '/report',
    query: {
      target_type: 'product',
      target_id: product.value ? String(product.value.id) : '',
      back: route.fullPath,
    },
  })
}

onMounted(load)
watch([() => props.slug, () => props.productSlug], load)
</script>

<style scoped>
.product-detail { max-width: 1100px; margin: 0 auto; }
.crumb { margin: 0 0 20px; }
.hero {
  border-radius: 14px;
  border: 1px solid #EEF2F7;
  overflow: hidden;
}
.hero-row { display: flex; gap: 24px; align-items: flex-start; flex-wrap: wrap; padding: 8px; }
.logo {
  width: 120px; height: 120px; flex-shrink: 0; border-radius: 16px; overflow: hidden;
  display: flex; align-items: center; justify-content: center;
}
.logo-img { width: 100%; height: 100%; object-fit: cover; }
.info { flex: 1; min-width: 260px; padding-top: 10px; }
.name-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; margin-bottom: 10px; }
.name { margin: 0; font-size: 26px; font-weight: 700; color: #0F172A; }
.tagline { color: #475569; font-size: 15px; line-height: 1.6; margin-bottom: 14px; max-width: 780px; }
.tags { display: flex; flex-wrap: wrap; gap: 6px; margin-bottom: 14px; }
.stats { display: flex; flex-wrap: wrap; gap: 16px; align-items: center; color: #64748B; font-size: 13px; }
.stats span { display: inline-flex; align-items: center; gap: 5px; }
.by-link { display: inline-flex; align-items: center; gap: 6px; color: #4F46E5; text-decoration: none; font-weight: 500; }
.actions { display: flex; flex-direction: column; gap: 10px; padding-top: 10px; }
.primary-btn {
  display: inline-flex; align-items: center; gap: 8px;
  padding: 10px 18px; border-radius: 10px; background: #4F46E5; color: #fff;
  font-weight: 600; text-decoration: none; transition: all .15s;
}
.primary-btn:hover { background: #4338CA; transform: translateY(-1px); }

.body-row { display: flex; gap: 24px; margin-top: 24px; align-items: flex-start; flex-wrap: wrap; }
.side { width: 320px; flex-shrink: 0; }
.main-col { flex: 1; min-width: 0; }
.side-card { border-radius: 14px; border: 1px solid #EEF2F7; }
.side-title { margin: 0 0 14px; font-size: 16px; }
.link-list { display: flex; flex-direction: column; gap: 8px; }
.link-row {
  display: flex; flex-direction: column; gap: 2px; padding: 10px 12px;
  background: #F8FAFC; border: 1px solid #E2E8F0; border-radius: 10px;
  color: inherit; text-decoration: none; transition: all .15s;
}
.link-row:hover { background: #EEF2FF; border-color: #C7D2FE; }
.link-row.official { background: #EEF2FF; border-color: #C7D2FE; }
.link-row .type { font-weight: 600; font-size: 13px; color: #0F172A; }
.link-row .url { font-size: 12px; color: #64748B; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.link-list .empty { color: #64748B; text-align: center; padding: 20px 10px; }

.desc-card { border-radius: 14px; border: 1px solid #EEF2F7; }
.sec-title { margin: 0 0 16px; font-size: 18px; color: #0F172A; }
.desc-body { color: #1E293B; line-height: 1.8; font-size: 15px; }
.desc-body p { margin: 0 0 14px; }
.empty-desc { color: #64748B; text-align: center; padding: 40px 0; }

@media (max-width: 768px) {
  .hero-row { padding: 4px; gap: 16px; }
  .logo { width: 88px; height: 88px; border-radius: 14px; }
  .info { min-width: 0; width: 100%; padding-top: 0; }
  .name { font-size: 22px; }
  .actions { width: 100%; flex-direction: row; flex-wrap: wrap; }
  .actions :deep(.el-button),
  .primary-btn { flex: 1; justify-content: center; text-align: center; }
  .side { width: 100%; order: 2; }
  .main-col { width: 100%; order: 1; }
  .body-row { gap: 16px; margin-top: 16px; }
}
</style>
