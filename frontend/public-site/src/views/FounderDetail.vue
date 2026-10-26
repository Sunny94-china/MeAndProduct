<template>
  <div class="founder-detail">
    <template v-if="notFound">
      <el-result
        icon="warning"
        title="名片不存在或已被隐藏"
        sub-title="你可以返回首页继续浏览其他创业者"
      >
        <template #extra>
          <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
        </template>
      </el-result>
    </template>

    <template v-else>
      <header class="hero" :class="{ 'has-cover': !!profile?.cover_url }" v-loading="loading">
        <div class="cover" v-if="profile?.cover_url" :style="{ backgroundImage: `url(${profile.cover_url})` }"></div>
        <div class="hero-inner">
          <div class="main-row">
            <el-avatar :size="88" :src="profile?.avatar_url" class="avatar">
              {{ avatarText }}
            </el-avatar>
            <div class="info">
              <div class="top-row">
                <h1 class="name">{{ profile?.display_name || '加载中...' }}</h1>
                <div class="tags">
                  <el-tag
                    v-for="t in (profile?.tags || []).slice(0, 4)"
                    :key="t.id"
                    size="large"
                    type="info"
                    effect="plain"
                    round
                  >{{ t.name }}</el-tag>
                </div>
              </div>
              <div class="title-row">{{ profile?.title || '独立开发者 · 产品人' }}</div>
              <p class="bio">{{ profile?.bio || 'Ta 还没写一句话介绍，去看看 Ta 的产品吧。' }}</p>
              <div class="counts">
                <span><el-icon><View /></el-icon> {{ profile?.view_count || 0 }} 次查看</span>
                <span><el-icon><Star /></el-icon> {{ displayFavCount }} 人收藏</span>
                <span v-if="profile?.location"><el-icon><Location /></el-icon> {{ profile.location }}</span>
              </div>
            </div>
            <div class="actions">
              <el-button
                v-if="auth.isLoggedIn()"
                :type="favorited ? 'danger' : 'primary'"
                size="large"
                :loading="togglingFav"
                @click="onToggleFav"
              >
                <el-icon v-if="!favorited"><Star /></el-icon>
                <el-icon v-else><StarFilled /></el-icon>
                {{ favorited ? '已收藏（点击取消）' : '收藏这位创业者' }}
              </el-button>
              <el-button size="large" @click="goReport('founder')">
                <el-icon><Warning /></el-icon>举报
              </el-button>
            </div>
          </div>

          <div class="links" v-if="customLinks.length || hasNativeLinks">
            <a
              v-for="(l, i) in customLinks"
              :key="'c'+i"
              class="link-btn"
              target="_blank"
              rel="noopener"
              @click="onLink(l)"
              :href="l.url"
            >
              <span class="link-type">{{ linkIcon(l.link_type) }} {{ l.name || linkTypeLabel(l.link_type) }}</span>
              <span class="link-url">{{ truncate(l.url, 48) }}</span>
            </a>
            <a
              v-if="profile?.website_url"
              class="link-btn"
              target="_blank"
              rel="noopener"
              :href="profile.website_url"
            >
              <span class="link-type">🌐 个人官网</span>
              <span class="link-url">{{ truncate(profile.website_url, 48) }}</span>
            </a>
            <a
              v-if="profile?.public_email"
              class="link-btn"
              :href="`mailto:${profile.public_email}`"
            >
              <span class="link-type">✉️ 公开邮箱</span>
              <span class="link-url">{{ profile.public_email }}</span>
            </a>
          </div>
        </div>
      </header>

      <el-tabs v-model="tab" class="body-tabs" v-loading="loading">
        <el-tab-pane label="关于" name="about">
          <div class="about-card">
            <template v-if="profile?.bio_long">
              <div class="bio-long" v-html="formatLong(profile.bio_long)"></div>
            </template>
            <template v-else>
              <p class="bio-empty">{{ profile?.bio || 'Ta 还没有填写详细介绍。' }}</p>
            </template>
          </div>
        </el-tab-pane>

        <el-tab-pane label="产品" name="products">
          <div v-if="productsLoading" class="pad"><el-skeleton :rows="4" animated /></div>
          <el-row v-else-if="products.length" :gutter="20">
            <el-col v-for="p in products" :key="p.id" :xs="24" :sm="12" :md="8" :lg="6">
              <router-link :to="`/@${slug}/${p.slug}`" class="p-link">
                <el-card shadow="hover" class="p-card">
                  <div class="p-logo" :style="pLogo(p)">
                    <el-icon v-if="!p.logo_url" :size="28" color="#fff"><Goods /></el-icon>
                    <img v-else :src="p.logo_url" class="p-thumb" alt="" />
                  </div>
                  <div class="p-name">{{ p.name }}</div>
                  <div class="p-tagline">{{ p.tagline }}</div>
                  <div class="p-foot">
                    <el-tag size="small" type="success" effect="plain" round>{{ stageLabel(p.status) }}</el-tag>
                    <span class="p-stats"><el-icon><Star /></el-icon>{{ p.favorite_count }}</span>
                  </div>
                </el-card>
              </router-link>
            </el-col>
          </el-row>
          <el-empty v-else description="Ta 还没发布产品，敬请期待！" />
        </el-tab-pane>

        <el-tab-pane label="联系" name="contact">
          <div class="about-card">
            <ul class="contact-list">
              <li v-if="profile?.public_email">
                <span class="k">📮 公开邮箱</span>
                <a :href="`mailto:${profile.public_email}`">{{ profile.public_email }}</a>
              </li>
              <li v-if="profile?.website_url">
                <span class="k">🌐 个人官网</span>
                <a :href="profile.website_url" target="_blank" rel="noopener">{{ profile.website_url }}</a>
              </li>
              <li v-for="(l, i) in customLinks" :key="'ct'+i">
                <span class="k">{{ linkIcon(l.link_type) }} {{ l.name || linkTypeLabel(l.link_type) }}</span>
                <a :href="l.url" target="_blank" rel="noopener" @click="onLink(l)">{{ l.url }}</a>
              </li>
              <li v-if="!profile?.public_email && !profile?.website_url && customLinks.length === 0" class="empty">
                Ta 还没有留下任何联系方式。
              </li>
            </ul>
          </div>
        </el-tab-pane>
      </el-tabs>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  View, Star, StarFilled, Warning, Location, Goods,
} from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import {
  getFounderDetail,
  listFounderProducts,
  toggleFavorite,
  checkFavorite,
  trackLinkClick,
  type FounderVO,
  type ProductVO,
} from '@/api/content'

const props = defineProps<{ slug: string }>()
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const profile = ref<FounderVO | null>(null)
const products = ref<ProductVO[]>([])
const loading = ref(false)
const productsLoading = ref(false)
const notFound = ref(false)
const tab = ref('about')

const favorited = ref(false)
const displayFavCount = ref(0)
const togglingFav = ref(false)

const customLinks = computed(() => profile.value?.custom_links || [])
const hasNativeLinks = computed(() => !!(profile.value?.website_url || profile.value?.public_email))

const avatarText = computed(() =>
  (profile.value?.display_name || 'U').slice(0, 1).toUpperCase(),
)

function stageLabel(s: string) {
  const m: Record<string, string> = {
    idea: 'Idea', mvp: 'MVP', beta: 'Beta', launched: '已发布', growing: '成长中', mature: '成熟',
  }
  return m[s] || s
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

function truncate(s: string, n: number) {
  return s.length > n ? s.slice(0, n) + '…' : s
}

function pLogo(p: ProductVO) {
  if (p.logo_url) return {}
  const palette = [
    'linear-gradient(135deg,#6366F1,#A855F7)', 'linear-gradient(135deg,#EC4899,#F97316)',
    'linear-gradient(135deg,#10B981,#06B6D4)', 'linear-gradient(135deg,#F43F5E,#8B5CF6)',
    'linear-gradient(135deg,#F59E0B,#EF4444)', 'linear-gradient(135deg,#3B82F6,#14B8A6)',
  ]
  return { background: palette[p.id % palette.length] }
}

function formatLong(s: string) {
  const safe = s.replace(/</g, '&lt;').replace(/>/g, '&gt;')
  return safe.split(/\n\n+/).map((p) => `<p>${p.replace(/\n/g, '<br>')}</p>`).join('')
}

async function load() {
  if (!props.slug) {
    notFound.value = true
    return
  }
  loading.value = true
  notFound.value = false
  try {
    const p = await getFounderDetail(props.slug)
    profile.value = p
    displayFavCount.value = p.favorite_count
    tab.value = 'about'
    loadProducts()
    if (auth.isLoggedIn()) {
      try {
        const r = await checkFavorite('founder', p.id)
        favorited.value = !!r?.favorited
      } catch (_e) {}
    }
  } catch (e: any) {
    notFound.value = true
    ElMessage.error(e?.message || '名片不存在')
  } finally {
    loading.value = false
  }
}

async function loadProducts() {
  productsLoading.value = true
  try {
    const r = await listFounderProducts(props.slug, { page: 1, page_size: 50, sort: 'sort_order' })
    products.value = r?.items || []
  } catch (_e) {
    products.value = []
  } finally {
    productsLoading.value = false
  }
}

async function onToggleFav() {
  if (!profile.value || togglingFav.value) return
  togglingFav.value = true
  try {
    const r = await toggleFavorite({ target_type: 'founder', target_id: Number(profile.value.id) })
    favorited.value = !!r?.favorited
    displayFavCount.value = r?.favorite_count ?? displayFavCount.value
    ElMessage.success(favorited.value ? '已加入收藏夹' : '已取消收藏')
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

function goReport(type: string) {
  const target = profile.value?.id
  router.push({
    path: '/report',
    query: {
      target_type: type,
      target_id: target ? String(target) : '',
      back: route.fullPath,
    },
  })
}

onMounted(load)
watch(() => props.slug, load)
</script>

<style scoped>
.founder-detail { max-width: 1100px; margin: 0 auto; }
.hero {
  position: relative;
  background: #fff;
  border-radius: 14px;
  border: 1px solid #EEF2F7;
  overflow: hidden;
  margin-bottom: 24px;
}
.cover {
  height: 160px;
  background-size: cover;
  background-position: center;
  filter: saturate(1.1);
}
.hero-inner { padding: 24px 28px; position: relative; }
.main-row {
  display: flex;
  gap: 20px;
  align-items: flex-start;
  flex-wrap: wrap;
}
.hero.has-cover .hero-inner { padding: 0 28px 24px; }
.hero.has-cover .main-row { margin-top: -44px; }
.avatar {
  position: relative;
  z-index: 1;
  flex: none;
  border: 4px solid #fff;
  box-shadow: 0 4px 18px rgba(15,23,42,.08);
  background: linear-gradient(135deg,#E0E7FF,#F5D0FE);
  font-weight: 700;
  font-size: 28px;
  color: #4F46E5;
}
.info { flex: 1; min-width: 260px; }
.hero.has-cover .info,
.hero.has-cover .actions { padding-top: 36px; }
.top-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.name { margin: 0; font-size: 26px; font-weight: 700; color: #0F172A; }
.tags { display: flex; flex-wrap: wrap; gap: 6px; }
.title-row { margin-top: 6px; color: #4F46E5; font-weight: 600; }
.bio { margin: 10px 0 14px; color: #475569; line-height: 1.6; max-width: 780px; }
.counts { display: flex; flex-wrap: wrap; gap: 16px; color: #64748B; font-size: 13px; }
.counts span { display: inline-flex; align-items: center; gap: 5px; }
.actions { display: flex; gap: 10px; flex-wrap: wrap; }
.links {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 10px;
  margin-top: 24px;
}
.link-btn {
  display: flex;
  flex-direction: column;
  padding: 12px 14px;
  background: #F8FAFC;
  border: 1px solid #E2E8F0;
  border-radius: 10px;
  color: inherit;
  text-decoration: none;
  transition: all .15s;
}
.link-btn:hover { background: #EEF2FF; border-color: #C7D2FE; }
.link-type { font-weight: 600; font-size: 14px; color: #0F172A; margin-bottom: 2px; }
.link-url { font-size: 12px; color: #64748B; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.body-tabs {
  background: #fff;
  border: 1px solid #EEF2F7;
  border-radius: 14px;
  padding: 4px 20px;
}
.about-card { padding: 24px 4px; }
.bio-long { color: #1E293B; line-height: 1.8; }
.bio-long p { margin: 0 0 14px; }
.bio-empty { color: #64748B; }
.pad { padding: 18px; }
.p-link { text-decoration: none; color: inherit; }
.p-card { margin-bottom: 20px; transition: transform .15s; }
.p-card:hover { transform: translateY(-2px); }
.p-logo {
  width: 52px; height: 52px; border-radius: 10px; overflow: hidden;
  display: flex; align-items: center; justify-content: center; margin-bottom: 12px;
}
.p-thumb { width: 100%; height: 100%; object-fit: cover; }
.p-name { font-weight: 700; font-size: 16px; color: #0F172A; }
.p-tagline { color: #475569; font-size: 13px; margin: 4px 0 14px; line-height: 1.5; min-height: 2em; display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden; }
.p-foot { display: flex; justify-content: space-between; align-items: center; }
.p-stats { display: inline-flex; align-items: center; gap: 4px; color: #64748B; font-size: 12px; }

.contact-list { list-style: none; padding: 0; margin: 0; display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 14px; }
.contact-list li {
  padding: 14px; background: #F8FAFC; border: 1px solid #E2E8F0; border-radius: 10px;
  display: flex; flex-direction: column; gap: 6px;
}
.contact-list .k { font-weight: 600; color: #0F172A; }
.contact-list a { color: #4F46E5; word-break: break-all; }
.contact-list .empty { background: none; border: none; color: #64748B; grid-column: 1/-1; text-align: center; padding: 28px; }

@media (max-width: 768px) {
  .hero-inner { padding: 18px 16px; }
  .hero.has-cover .hero-inner { padding: 0 16px 18px; }
  .hero.has-cover .main-row { margin-top: -36px; }
  .hero.has-cover .info,
  .hero.has-cover .actions { padding-top: 28px; }
  .name { font-size: 22px; }
  .info { min-width: 0; width: 100%; }
  .actions { width: 100%; }
  .actions :deep(.el-button) { flex: 1; }
  .body-tabs { padding: 2px 12px; border-radius: 12px; }
  .links { grid-template-columns: 1fr; }
  .contact-list { grid-template-columns: 1fr; }
}
</style>
