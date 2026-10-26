<template>
  <div class="fav-page">
    <header class="page-head">
      <h1 class="page-title">
        <el-icon :size="24" color="#4F46E5"><Star /></el-icon>我的收藏
      </h1>
      <p class="page-sub">收藏了有价值的创业者和产品，随时回看</p>
    </header>

    <template v-if="!auth.isLoggedIn()">
      <el-result
        icon="warning"
        title="登录后查看你的收藏夹"
        sub-title="我们会帮你按时间顺序记录每一份灵感"
      >
        <template #extra>
          <el-button type="primary" @click="$emit('needLogin')">立即登录</el-button>
        </template>
      </el-result>
    </template>

    <template v-else>
      <el-tabs v-model="tab" class="fav-tabs">
        <el-tab-pane label="创业者" name="founder">
          <div v-if="loadingF && founderList.length === 0" class="load-skel">
            <el-skeleton :rows="4" animated />
          </div>
          <el-row v-else-if="founderList.length" :gutter="20">
            <el-col v-for="it in founderList" :key="it.id" :xs="24" :sm="12" :md="8" :lg="6">
              <router-link
                v-if="it.target_snapshot?.slug"
                :to="`/@${it.target_snapshot.slug}`"
                class="card-link"
              >
                <el-card shadow="hover" class="fav-card">
                  <div class="row">
                    <el-avatar :size="44" :src="it.target_snapshot?.avatar_url">
                      {{ (it.target_snapshot?.display_name || 'U').slice(0,1).toUpperCase() }}
                    </el-avatar>
                    <div class="meta">
                      <div class="name">{{ it.target_snapshot?.display_name || '创业者' }}</div>
                      <div class="time">收藏于 {{ fmt(it.created_at) }}</div>
                    </div>
                    <el-tooltip content="取消收藏">
                      <el-button
                        :icon="StarFilled"
                        circle
                        type="danger"
                        size="small"
                        @click.stop="toggle(it, 'founder')"
                      />
                    </el-tooltip>
                  </div>
                </el-card>
              </router-link>
            </el-col>
          </el-row>
          <el-empty v-else-if="!loadingF" description="还没收藏任何创业者，去首页逛逛吧" />
        </el-tab-pane>

        <el-tab-pane label="产品" name="product">
          <div v-if="loadingP && productList.length === 0" class="load-skel">
            <el-skeleton :rows="4" animated />
          </div>
          <el-row v-else-if="productList.length" :gutter="20">
            <el-col v-for="it in productList" :key="it.id" :xs="24" :sm="12" :md="8" :lg="6">
              <router-link
                v-if="it.target_snapshot?.slug"
                :to="`/@${it.target_snapshot?.founder_slug || ''}/${it.target_snapshot.slug}`"
                class="card-link"
              >
                <el-card shadow="hover" class="fav-card product-card">
                  <div class="row">
                    <div class="p-logo" :style="logoStyle(it)">
                      <el-icon v-if="!it.target_snapshot?.logo_url" :size="22" color="#fff"><Goods /></el-icon>
                      <img v-else :src="it.target_snapshot.logo_url" alt="" class="p-thumb" />
                    </div>
                    <div class="meta flex1">
                      <div class="name">{{ it.target_snapshot?.name || '产品' }}</div>
                      <div class="tagline">{{ it.target_snapshot?.tagline || '收藏于 ' + fmt(it.created_at) }}</div>
                    </div>
                    <el-tooltip content="取消收藏">
                      <el-button
                        :icon="StarFilled"
                        circle
                        type="danger"
                        size="small"
                        @click.stop="toggle(it, 'product')"
                      />
                    </el-tooltip>
                  </div>
                </el-card>
              </router-link>
            </el-col>
          </el-row>
          <el-empty v-else-if="!loadingP" description="还没收藏任何产品，去发现值得关注的产品吧" />
        </el-tab-pane>
      </el-tabs>
    </template>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Star, StarFilled, Goods } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { listMyFavorites, toggleFavorite, type FavoriteItemVO } from '@/api/content'

defineEmits<{ (e: 'needLogin'): void }>()

const auth = useAuthStore()
const tab = ref<'founder' | 'product'>('founder')
const founderList = ref<FavoriteItemVO[]>([])
const productList = ref<FavoriteItemVO[]>([])
const loadingF = ref(false)
const loadingP = ref(false)

function fmt(s?: string) {
  if (!s) return '—'
  const d = new Date(s)
  if (Number.isNaN(d.getTime())) return s.slice(0, 10)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function logoStyle(it: FavoriteItemVO) {
  if (it.target_snapshot?.logo_url) return {}
  const palette = [
    'linear-gradient(135deg,#6366F1,#A855F7)',
    'linear-gradient(135deg,#10B981,#06B6D4)',
    'linear-gradient(135deg,#F43F5E,#8B5CF6)',
    'linear-gradient(135deg,#F59E0B,#EF4444)',
  ]
  return { background: palette[it.id % palette.length] }
}

async function loadFounder() {
  loadingF.value = true
  try {
    const r = await listMyFavorites({ target_type: 'founder', page: 1, page_size: 50 })
    founderList.value = r?.items || []
  } catch (_e) {
    founderList.value = []
  } finally {
    loadingF.value = false
  }
}

async function loadProduct() {
  loadingP.value = true
  try {
    const r = await listMyFavorites({ target_type: 'product', page: 1, page_size: 50 })
    productList.value = r?.items || []
  } catch (_e) {
    productList.value = []
  } finally {
    loadingP.value = false
  }
}

async function toggle(it: FavoriteItemVO, type: 'founder' | 'product') {
  try {
    await toggleFavorite({ target_type: type, target_id: Number(it.target_id) })
    ElMessage.success('已取消收藏')
    if (type === 'founder') {
      founderList.value = founderList.value.filter((x) => x.id !== it.id)
    } else {
      productList.value = productList.value.filter((x) => x.id !== it.id)
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

onMounted(() => {
  if (!auth.isLoggedIn()) return
  loadFounder()
  loadProduct()
})

watch(tab, (t) => {
  if (!auth.isLoggedIn()) return
  if (t === 'founder') loadFounder()
  else loadProduct()
})
</script>

<style scoped>
.fav-page { max-width: 1200px; margin: 0 auto; }
.page-head { margin-bottom: 20px; }
.page-title { margin: 0; font-size: clamp(1.25rem, 4vw, 1.6rem); display: inline-flex; align-items: center; gap: 10px; }
.page-sub { color:#64748b; margin: 6px 0 0; }
.fav-tabs { margin-top: 20px; }
.fav-card { margin-bottom: 16px; border-radius: var(--mp-radius) !important; border: 1px solid var(--mp-border); box-shadow: var(--mp-shadow); transition: transform .2s, box-shadow .2s; }
.fav-card:hover { transform: translateY(-3px); box-shadow: var(--mp-shadow-hover); }
.card-link { text-decoration: none; color: inherit; }
.fav-card .row { display: flex; align-items: center; gap: 14px; }
.fav-card .meta { flex: 1; min-width: 0; }
.fav-card .name { font-size: 15px; font-weight: 600; color:#0F172A; margin-bottom: 4px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.fav-card .time, .fav-card .tagline { font-size: 12px; color:#64748b; overflow:hidden; text-overflow:ellipsis; white-space: nowrap; }
.p-logo { width: 44px; height: 44px; border-radius: 8px; overflow: hidden; display: flex; align-items: center; justify-content: center; }
.p-thumb { width: 100%; height: 100%; object-fit: cover; }
.load-skel { padding: 20px; background:#fff; border:1px solid #EEF2F7; border-radius: 10px; }
.flex1 { flex:1; }
</style>
