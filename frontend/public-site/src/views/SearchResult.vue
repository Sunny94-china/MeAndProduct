<template>
  <div class="search-page">
    <header class="page-head">
      <h1 class="title"><el-icon :size="24" color="#4F46E5"><Search /></el-icon>搜索结果</h1>
      <p class="sub" v-if="q">关键词：“{{ q }}”</p>
      <p class="sub" v-else-if="currentTag">#{{ currentTag?.name || '' }}</p>
    </header>

    <div class="filters">
      <el-input
        v-model="q"
        size="large"
        placeholder="搜索创业者、产品名称、关键词..."
        clearable
        @keyup.enter="onSearch"
      >
        <template #prefix><el-icon><Search /></el-icon></template>
        <template #append><el-button type="primary" @click="onSearch">搜索</el-button></template>
      </el-input>
      <el-tabs v-model="tab" @tab-change="onTab" class="tabs">
        <el-tab-pane label="创业者" name="founder" />
        <el-tab-pane label="产品" name="product" />
      </el-tabs>
    </div>

    <div v-if="loading" class="pad"><el-skeleton :rows="6" animated /></div>
    <template v-else>
      <el-row :gutter="20" v-if="tab === 'founder' && founderList.length">
        <el-col v-for="f in founderList" :key="f.id" :xs="24" :sm="12" :md="8" :lg="6">
          <router-link :to="`/@${f.slug}`" class="card-link">
            <el-card shadow="hover" class="founder-card">
              <template #header>
                <div class="card-head">
                  <el-avatar :size="52" :src="f.avatar_url">{{ (f.display_name || 'U').slice(0,1).toUpperCase() }}</el-avatar>
                  <div class="meta">
                    <div class="name">{{ f.display_name }}</div>
                    <div class="ttl">{{ f.title || '独立开发者' }}</div>
                  </div>
                </div>
              </template>
              <p class="bio">{{ f.bio || 'Ta 还没写介绍。' }}</p>
              <div class="tags">
                <el-tag v-for="t in (f.tags || []).slice(0,3)" :key="t.id" size="small" type="info" effect="plain" round>{{ t.name }}</el-tag>
              </div>
              <div class="stats"><span><el-icon><Star /></el-icon>{{ f.favorite_count }}</span><span><el-icon><View /></el-icon>{{ f.view_count }}</span></div>
            </el-card>
          </router-link>
        </el-col>
      </el-row>
      <el-row :gutter="20" v-else-if="tab === 'product' && productList.length">
        <el-col v-for="p in productList" :key="p.id" :xs="24" :sm="12" :md="8" :lg="6">
          <router-link :to="`/@${p.founder_slug || ''}/${p.slug}`" class="card-link">
            <el-card shadow="hover" class="product-card">
              <div class="logo" :style="pLogo(p)">
                <el-icon v-if="!p.logo_url" :size="26" color="#fff"><Goods /></el-icon>
                <img v-else :src="p.logo_url" alt="" class="p-thumb" />
              </div>
              <div class="name">{{ p.name }}</div>
              <div class="tagline">{{ p.tagline }}</div>
              <div class="tags">
                <el-tag size="small" type="success" effect="plain" round>{{ stageLabel(p.status) }}</el-tag>
              </div>
            </el-card>
          </router-link>
        </el-col>
      </el-row>
      <el-empty
        v-else
        :description="q || currentTag ? '没有找到相关结果，换个关键词试试吧' : '输入关键词开始搜索'"
      />

      <div class="pager" v-if="total > pageSize">
        <el-pagination
          background
          layout="prev, pager, next"
          :total="total"
          :page-size="pageSize"
          :current-page="page"
          @current-change="onPage"
        />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Search, Star, View, Goods } from '@element-plus/icons-vue'
import { listActiveTags, listFounders, listProducts, type FounderVO, type ProductVO, type TagVO } from '@/api/content'

const route = useRoute()
const router = useRouter()

const q = ref('')
const tab = ref<'founder' | 'product'>('founder')
const page = ref(1)
const pageSize = ref(24)
const total = ref(0)
const loading = ref(false)
const founderList = ref<FounderVO[]>([])
const productList = ref<ProductVO[]>([])

const tag = ref<string | null>(null)
const currentTag = computed<TagVO | null>(() => {
  if (!tag.value) return null
  return { id: Number(tag.value), name: '标签 #' + tag.value, slug: '', type: 'other' as any, founder_count: 0, product_count: 0, is_active: true }
})

function stageLabel(s: string) {
  const m: Record<string, string> = { idea: 'Idea', mvp: 'MVP', beta: 'Beta', launched: '已发布', growing: '成长中', mature: '成熟' }
  return m[s] || s
}

function pLogo(p: ProductVO) {
  if (p.logo_url) return {}
  const palette = [
    'linear-gradient(135deg,#6366F1,#A855F7)', 'linear-gradient(135deg,#EC4899,#F97316)',
    'linear-gradient(135deg,#10B981,#06B6D4)', 'linear-gradient(135deg,#F43F5E,#8B5CF6)',
  ]
  return { background: palette[p.id % palette.length] }
}

function readFromRoute() {
  q.value = (route.query.q as string) || ''
  tab.value = (route.query.type as 'founder' | 'product') || 'founder'
  tag.value = (route.query.tag as string) || null
  page.value = Number(route.query.page) || 1
  load()
}

async function load() {
  loading.value = true
  try {
    const params: Record<string, any> = { page: page.value, page_size: pageSize.value }
    if (q.value) params.keyword = q.value
    if (tag.value) params.tags = tag.value
    const kw = q.value.trim().toLowerCase()
    const hit = (...parts: Array<string | undefined>) => !kw || parts.some((p) => (p || '').toLowerCase().includes(kw))
    if (tab.value === 'founder') {
      params.sort = 'favorite_count_desc'
      const r = await listFounders(params)
      const items = (r?.items || []).filter((f) => hit(f.display_name, f.title, f.bio, f.slug))
      founderList.value = items
      total.value = kw ? items.length : r?.total || 0
    } else {
      params.sort = 'favorite_count_desc'
      const r = await listProducts(params)
      const items = (r?.items || []).filter((p) => hit(p.name, p.tagline, p.description, p.slug))
      productList.value = items
      total.value = kw ? items.length : r?.total || 0
    }
  } catch (_e) {} finally {
    loading.value = false
  }
}

function onSearch() {
  router.replace({ path: '/search', query: { q: q.value || undefined, type: tab.value, tag: tag.value || undefined, page: '1' } })
}

function onTab(t: string | number) {
  tab.value = t as any
  page.value = 1
  router.replace({ path: '/search', query: { q: q.value || undefined, type: tab.value, tag: tag.value || undefined, page: '1' } })
  load()
}

function onPage(p: number) {
  page.value = p
  router.replace({ path: '/search', query: { q: q.value || undefined, type: tab.value, tag: tag.value || undefined, page: String(p) } })
  load()
}

onMounted(readFromRoute)
watch(() => route.query, readFromRoute)
</script>

<style scoped>
.search-page { max-width: 1200px; margin: 0 auto; }
.page-head { margin-bottom: 16px; }
.title { margin: 0; font-size: clamp(1.2rem, 4vw, 1.5rem); display: inline-flex; align-items: center; gap: 10px; }
.sub { color:#64748b; margin: 6px 0 0; }
.filters { background: #fff; padding: clamp(14px, 3vw, 20px); border: 1px solid var(--mp-border); border-radius: var(--mp-radius); margin-bottom: 20px; box-shadow: var(--mp-shadow); }
.tabs { margin-top: 16px; }
.pad { padding: 28px; background: #fff; border: 1px solid var(--mp-border); border-radius: var(--mp-radius); }
.card-link { text-decoration: none; color: inherit; display: block; }
.founder-card, .product-card { margin-bottom: 16px; border-radius: var(--mp-radius) !important; border: 1px solid var(--mp-border); box-shadow: var(--mp-shadow); transition: transform .2s, box-shadow .2s; height: 100%; }
.founder-card:hover, .product-card:hover { transform: translateY(-3px); box-shadow: var(--mp-shadow-hover); }
.card-head { display: flex; gap: 12px; align-items: center; }
.card-head .name { font-size: 15px; font-weight: 600; color: #0F172A; }
.card-head .ttl { font-size: 12px; color: #64748B; margin-top: 2px; }
.bio { color: #475569; font-size: 14px; margin: 4px 0 12px; min-height: 2.4em; display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden; line-height: 1.5;}
.tags { display: flex; flex-wrap: wrap; gap: 6px; }
.stats { margin-top: 10px; display: flex; gap: 16px; color: #64748B; font-size: 12px; }
.stats span { display: inline-flex; align-items: center; gap: 4px; }
.logo { width: 50px; height: 50px; border-radius: 10px; overflow: hidden; display: flex; align-items: center; justify-content: center; margin-bottom: 10px; }
.p-thumb { width: 100%; height: 100%; object-fit: cover; }
.name { font-weight: 700; color: #0F172A; font-size: 16px; }
.tagline { font-size: 13px; color: #475569; margin: 4px 0 10px; line-height: 1.5; min-height: 2em; display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden; }
.pager { margin: 24px 0; text-align: center; }
</style>
