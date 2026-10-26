import { defineStore } from 'pinia'
import { ref } from 'vue'

export type FavoriteTarget = { type: 'founder' | 'product'; id: string | number; title: string; url: string }

export const useFavoriteStore = defineStore('favorite', () => {
  const items = ref<FavoriteTarget[]>([])

  function add(t: FavoriteTarget) {
    if (items.value.some((x) => x.type === t.type && x.id === t.id)) return
    items.value.push(t)
  }
  function remove(type: string, id: string | number) {
    items.value = items.value.filter((x) => !(x.type === type && x.id === id))
  }
  function has(type: string, id: string | number) {
    return items.value.some((x) => x.type === type && x.id === id)
  }

  return { items, add, remove, has }
})
