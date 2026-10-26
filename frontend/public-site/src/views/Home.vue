<template>
  <div class="home">
    <header class="hero">
      <div class="hero-inner">
        <h1 class="brand">
          <el-icon :size="36" color="#4F46E5"><Link /></el-icon>
          MeAndProduct
        </h1>
        <p class="tagline">每个创业者的思考复盘理念都应该被记录，每个产品的诞生成长都值得被记录。</p>

        <el-form :inline="true" :model="searchForm" @submit.prevent="goSearch">
          <el-input
            v-model="searchForm.q"
            size="large"
            placeholder="搜索创业者、产品、关键词、标签..."
            class="hero-search"
            clearable
            @keyup.enter="goSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
            <template #append>
              <el-button type="primary" @click="goSearch">搜索</el-button>
            </template>
          </el-input>
        </el-form>

        <div class="chips" v-if="hotTags.length">
          <el-tag
            v-for="t in hotTags"
            :key="t.id"
            size="large"
            class="chip"
            effect="plain"
            round
            @click="goTagSearch(t)"
          >{{ t.name }}</el-tag>
        </div>
      </div>
    </header>

    <main class="container">
      <section>
        <div class="section-head">
          <h2><el-icon><UserFilled /></el-icon> 推荐创业者</h2>
          <el-button link type="primary" @click="goList('founder')">查看全部 →</el-button>
        </div>
        <el-row :gutter="20">
          <el-col v-for="f in founders" :key="f.id" :xs="24" :sm="12" :md="8" :lg="6">
            <router-link :to="`/@${f.slug}`" class="founder-card-link">
              <el-card shadow="hover" class="founder-card">
                <template #header>
                  <div class="card-head">
                    <el-avatar :size="56" :src="f.avatar_url">{{ avatarText(f.display_name) }}</el-avatar>
                    <div class="meta">
                      <div class="name">{{ f.display_name }}</div>
                      <div class="ttl">{{ f.title || '独立开发者 · 创业者' }}</div>
                    </div>
                  </div>
                </template>
                <p class="card-bio">{{ f.bio || '记录自己与产品的成长轨迹。' }}</p>
                <div class="tags">
                  <el-tag
                    v-for="t in (f.tags || []).slice(0, 3)"
                    :key="t.id"
                    size="small"
                    type="info"
                    effect="plain"
                    round
                  >{{ t.name }}</el-tag>
                </div>
                <div class="stats">
                  <span><el-icon><View /></el-icon> {{ f.view_count }}</span>
                  <span><el-icon><Star /></el-icon> {{ f.favorite_count }}</span>
                </div>
              </el-card>
            </router-link>
          </el-col>
          <el-col v-if="!foundersLoading && founders.length === 0" :xs="24">
            <el-empty description="暂无推荐创业者，去成为第一个吧！" />
          </el-col>
        </el-row>
      </section>

      <section style="margin-top:40px;">
        <div class="section-head">
          <h2><el-icon><Goods /></el-icon> 值得关注的产品</h2>
          <el-button link type="primary" @click="goList('product')">查看全部 →</el-button>
        </div>
        <el-row :gutter="20">
          <el-col
            v-for="p in products"
            :key="p.id"
            :xs="24"
            :sm="12"
            :md="8"
            :lg="6"
          >
            <router-link :to="`/@${p.founder_slug}/${p.slug}`" class="product-card-link">
              <el-card shadow="hover" class="product-card">
                <div class="p-logo" :style="logoStyle(p)">
                  <el-icon v-if="!p.logo_url" :size="28" color="#fff"><Box /></el-icon>
                  <img v-else :src="p.logo_url" class="p-logo-img" alt="" />
                </div>
                <div class="p-name">{{ p.name }}</div>
                <div class="p-tagline">{{ p.tagline }}</div>
                <div class="tags mt8">
                  <el-tag v-if="p.status" size="small" type="success" effect="plain" round>{{ stageLabel(p.status) }}</el-tag>
                  <el-tag
                    v-for="t in (p.tags || []).slice(0, 2)"
                    :key="t.id"
                    size="small"
                    type="info"
                    effect="plain"
                    round
                  >{{ t.name }}</el-tag>
                </div>
                <div class="p-foot">
                  <span class="by" v-if="p.founder_display_name">by {{ p.founder_display_name }}</span>
                  <span class="p-stats">
                    <el-icon><Star /></el-icon>{{ p.favorite_count }}
                  </span>
                </div>
              </el-card>
            </router-link>
          </el-col>
          <el-col v-if="!productsLoading && products.length === 0" :xs="24">
            <el-empty description="暂无产品，来发布第一个产品吧！" />
          </el-col>
        </el-row>
      </section>

      <footer class="foot">
        <div>© MeAndProduct · 链接的是人与产品的故事</div>
      </footer>
    </main>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Search, Link, UserFilled, Goods, View, Star, Box } from '@element-plus/icons-vue'
import {
  listFounders,
  listProducts,
  getTagsGrouped,
  type FounderVO,
  type ProductVO,
  type TagVO,
  type TagGroupedVO,
} from '@/api/content'

const router = useRouter()
const searchForm = reactive({ q: '' })

const founders = ref<FounderVO[]>([])
const products = ref<ProductVO[]>([])
const hotTags = ref<TagVO[]>([])
const foundersLoading = ref(false)
const productsLoading = ref(false)

function avatarText(s: string) {
  return (s || 'U').slice(0, 1).toUpperCase()
}

function stageLabel(s: string) {
  const map: Record<string, string> = {
    idea: 'Idea',
    mvp: 'MVP',
    beta: 'Beta',
    launched: '已发布',
    growing: '成长中',
    mature: '成熟',
  }
  return map[s] || s
}

function logoStyle(p: ProductVO) {
  if (p.logo_url) return {}
  const palette = [
    'linear-gradient(135deg,#6366F1,#A855F7)',
    'linear-gradient(135deg,#EC4899,#F97316)',
    'linear-gradient(135deg,#10B981,#06B6D4)',
    'linear-gradient(135deg,#F43F5E,#8B5CF6)',
    'linear-gradient(135deg,#F59E0B,#EF4444)',
    'linear-gradient(135deg,#3B82F6,#14B8A6)',
  ]
  return { background: palette[p.id % palette.length] }
}

async function loadFounders() {
  foundersLoading.value = true
  try {
    const r = await listFounders({ page: 1, page_size: 6, sort: 'favorite_count_desc' })
    founders.value = (r && r.items) || []
  } catch (_e) {
    founders.value = []
  } finally {
    foundersLoading.value = false
  }
}

async function loadProducts() {
  productsLoading.value = true
  try {
    const r = await listProducts({ page: 1, page_size: 6, sort: 'favorite_count_desc' })
    products.value = (r && r.items) || []
  } catch (_e) {
    products.value = []
  } finally {
    productsLoading.value = false
  }
}

async function loadHotTags() {
  try {
    const r0: any = (await getTagsGrouped()) || {}
    const r: TagGroupedVO = r0 && typeof r0 === 'object' && 'groups' in r0 && r0.groups ? r0.groups : r0
    const flat: TagVO[] = []
    for (const k of Object.keys(r)) {
      const arr = (r as any)[k] as TagVO[] | undefined
      if (Array.isArray(arr)) flat.push(...arr)
    }
    flat.sort((a, b) => (b.product_count + b.founder_count) - (a.product_count + a.founder_count))
    hotTags.value = flat.slice(0, 6)
  } catch (_e) {
    hotTags.value = []
  }
}

onMounted(() => {
  loadFounders()
  loadProducts()
  loadHotTags()
})

function goSearch() {
  router.push({ path: '/search', query: { q: searchForm.q } })
}

function goList(t: string) {
  router.push({ path: '/search', query: { type: t } })
}

function goTagSearch(t: TagVO) {
  router.push({ path: '/search', query: { tag: t.id.toString() } })
}
</script>

<style scoped>
.home {
  margin: -24px calc(50% - 50vw) 0;
  width: 100vw;
}

.hero {
  position: relative;
  overflow: hidden;
  background:
    radial-gradient(ellipse 80% 60% at 15% 10%, rgba(99, 102, 241, 0.18), transparent 55%),
    radial-gradient(ellipse 70% 50% at 90% 20%, rgba(168, 85, 247, 0.12), transparent 50%),
    linear-gradient(180deg, #EEF2FF 0%, #F4F6FB 72%);
  padding: clamp(48px, 8vw, 88px) var(--mp-pad) clamp(40px, 6vw, 64px);
  animation: heroIn 0.55s ease both;
}

.hero-inner {
  max-width: 960px;
  margin: 0 auto;
  text-align: center;
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-size: clamp(1.75rem, 5vw, 2.35rem);
  margin: 0 0 12px;
  font-weight: 800;
  letter-spacing: -0.03em;
}

.tagline {
  font-size: clamp(0.95rem, 2.4vw, 1.125rem);
  color: #475569;
  max-width: 680px;
  margin: 0 auto 28px;
  line-height: 1.65;
}

.hero-search {
  width: 100%;
  max-width: 640px;
  margin: 0 auto;
}

.hero-search :deep(.el-input-group) {
  box-shadow: var(--mp-shadow);
  border-radius: 14px;
  overflow: hidden;
}

.chips {
  margin-top: 18px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: center;
}

.chip {
  cursor: pointer;
  transition: transform 0.15s;
}

.chip:active {
  transform: scale(0.97);
}

.container {
  max-width: var(--mp-max);
  margin: 0 auto;
  padding: clamp(28px, 5vw, 44px) var(--mp-pad) 48px;
}

.section-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 18px;
  flex-wrap: wrap;
}

.section-head h2 {
  margin: 0;
  font-size: clamp(1.15rem, 2.8vw, 1.35rem);
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.founder-card-link,
.product-card-link {
  text-decoration: none;
  color: inherit;
  display: block;
}

.founder-card,
.product-card {
  margin-bottom: 16px;
  border-radius: var(--mp-radius) !important;
  border: 1px solid var(--mp-border);
  box-shadow: var(--mp-shadow);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  height: 100%;
}

.founder-card:hover,
.product-card:hover {
  transform: translateY(-3px);
  box-shadow: var(--mp-shadow-hover);
}

.card-head {
  display: flex;
  gap: 14px;
  align-items: center;
}

.card-head .name {
  font-size: 16px;
  font-weight: 650;
  color: #0f172a;
}

.card-head .ttl {
  font-size: 12px;
  color: #64748b;
  margin-top: 2px;
}

.card-bio {
  color: #475569;
  font-size: 14px;
  margin: 6px 0 12px;
  line-height: 1.5;
  min-height: 2.5em;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.mt8 {
  margin-top: 8px;
}

.stats {
  margin-top: 12px;
  display: flex;
  gap: 16px;
  color: #64748b;
  font-size: 12px;
  align-items: center;
}

.stats span {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.product-card {
  text-align: left;
}

.p-logo {
  width: 52px;
  height: 52px;
  border-radius: 12px;
  background: linear-gradient(135deg, #6366f1, #a855f7);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12px;
  overflow: hidden;
}

.p-logo-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.p-name {
  font-weight: 700;
  font-size: 17px;
  color: #0f172a;
}

.p-tagline {
  color: #475569;
  font-size: 13px;
  margin: 4px 0 10px;
  line-height: 1.5;
  min-height: 2em;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.p-foot {
  margin-top: 14px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  color: #64748b;
  gap: 8px;
}

.p-foot .by {
  opacity: 0.85;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.p-stats {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  flex: none;
}

.foot {
  text-align: center;
  margin-top: 40px;
  color: #94a3b8;
  font-size: 13px;
}

@keyframes heroIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: none;
  }
}

@media (max-width: 768px) {
  .home {
    margin: -16px calc(50% - 50vw) 0;
  }

  .hero-search :deep(.el-input-group__append) {
    padding: 0 10px;
  }

  .founder-card :deep(.el-card__header),
  .founder-card :deep(.el-card__body),
  .product-card :deep(.el-card__body) {
    padding: 14px;
  }
}
</style>
