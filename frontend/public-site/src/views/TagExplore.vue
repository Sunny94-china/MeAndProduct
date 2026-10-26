<template>
  <div class="tag-page">
    <header class="page-head">
      <h1 class="title"><el-icon :size="24" color="#4F46E5"><PriceTag /></el-icon>发现标签</h1>
      <p class="sub">用行业、产品阶段、角色标签定位你感兴趣的创业者和产品</p>
    </header>

    <div v-if="loading" class="pad"><el-skeleton :rows="6" animated /></div>
    <template v-else>
      <el-tabs v-model="tab">
        <el-tab-pane label="全部分组" name="all">
          <section v-for="(group, type) in grouped" :key="type" class="group">
            <h2 class="group-title">
              {{ typeLabel(type as string) }}
              <span class="group-total">共 {{ (group as any[]).length }} 个标签</span>
            </h2>
            <div class="tag-grid">
              <div
                v-for="t in (group as any[])"
                :key="t.id"
                class="tag-card"
                @click="goTag(t)"
              >
                <div class="name">{{ t.name }}</div>
                <div class="counts">
                  <span><el-icon><User /></el-icon>{{ t.founder_count || 0 }}</span>
                  <span><el-icon><Goods /></el-icon>{{ t.product_count || 0 }}</span>
                </div>
              </div>
            </div>
            <el-empty v-if="(group as any[]).length === 0" description="暂无此类型标签" />
          </section>
        </el-tab-pane>

        <el-tab-pane
          v-for="t in types"
          :key="t.key"
          :label="`${t.label} (${groupCount(t.key)})`"
          :name="t.key"
        >
          <div class="tag-grid">
            <div
              v-for="tag in (grouped as any)[t.key] || []"
              :key="tag.id"
              class="tag-card big"
              @click="goTag(tag)"
            >
              <div class="name">{{ tag.name }}</div>
              <div class="slug">#{{ tag.slug }}</div>
              <div class="counts">
                <span><el-icon><User /></el-icon>{{ tag.founder_count || 0 }} 位创业者</span>
                <span><el-icon><Goods /></el-icon>{{ tag.product_count || 0 }} 款产品</span>
              </div>
            </div>
          </div>
          <el-empty v-if="!((grouped as any)[t.key] || []).length" description="该类型暂无标签" />
        </el-tab-pane>
      </el-tabs>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { PriceTag, User, Goods } from '@element-plus/icons-vue'
import { getTagsGrouped, type TagGroupedVO, type TagVO } from '@/api/content'

const router = useRouter()
const loading = ref(false)
const tab = ref('all')
const grouped = reactive<TagGroupedVO>({})

const types = [
  { key: 'industry', label: '行业' },
  { key: 'stage', label: '产品阶段' },
  { key: 'role', label: '角色' },
  { key: 'other', label: '其他' },
]

function typeLabel(k: string) {
  const t = types.find((x) => x.key === k)
  return t ? t.label : k
}

function groupCount(k: string) {
  return ((grouped as any)[k] || []).length
}

async function load() {
  loading.value = true
  try {
    const r = await getTagsGrouped()
    const data = r && typeof r === 'object' && 'groups' in r && typeof (r as any).groups === 'object' && (r as any).groups ? (r as any).groups : r
    Object.keys(grouped).forEach((k) => delete (grouped as any)[k])
    Object.assign(grouped, data || {})
    for (const k of Object.keys(grouped)) {
      const arr = (grouped as any)[k] as TagVO[] | undefined
      if (Array.isArray(arr)) arr.sort((a, b) => (b.product_count + b.founder_count) - (a.product_count + a.founder_count))
    }
  } catch (_e) {} finally {
    loading.value = false
  }
}

function goTag(t: TagVO) {
  router.push({ path: '/search', query: { type: 'founder', tag: t.id.toString() } })
}

onMounted(load)
</script>

<style scoped>
.tag-page { max-width: 1200px; margin: 0 auto; }
.page-head { margin-bottom: 18px; }
.title { margin: 0; font-size: clamp(1.25rem, 4vw, 1.6rem); display: inline-flex; align-items: center; gap: 10px; }
.sub { color:#64748b; margin: 6px 0 0; }
.pad { padding: 28px; background: #fff; border: 1px solid var(--mp-border); border-radius: var(--mp-radius); }
.group { margin-bottom: 20px; background: #fff; border: 1px solid var(--mp-border); border-radius: var(--mp-radius); padding: clamp(14px, 3vw, 22px); box-shadow: var(--mp-shadow); }
.group-title { margin: 0 0 14px; font-size: 17px; color: #0F172A; display: flex; align-items: baseline; gap: 10px; flex-wrap: wrap; }
.group-total { font-size: 12px; color: #64748B; font-weight: 400; }
.tag-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(min(100%, 180px), 1fr)); gap: 12px; }
.tag-card {
  padding: 16px 18px;
  border-radius: 12px;
  background: #F8FAFC;
  border: 1px solid #E2E8F0;
  cursor: pointer;
  transition: all .15s;
}
.tag-card:hover {
  background: #EEF2FF;
  border-color: #C7D2FE;
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(79, 70, 229, 0.08);
}
.tag-card .name { font-weight: 700; color: #0F172A; font-size: 15px; }
.tag-card .slug { color: #64748B; font-size: 12px; margin-top: 2px; }
.tag-card .counts {
  margin-top: 10px;
  display: flex;
  gap: 16px;
  color: #64748B;
  font-size: 12px;
}
.tag-card .counts span { display: inline-flex; align-items: center; gap: 4px; }
.tag-card.big { padding: 20px; }
.tag-card.big .name { font-size: 17px; }
</style>
