import { ref } from 'vue'

const drawerOpen = ref(false)

export function useAdminNav() {
  return {
    drawerOpen,
    open() {
      drawerOpen.value = true
    },
    close() {
      drawerOpen.value = false
    },
    toggle() {
      drawerOpen.value = !drawerOpen.value
    },
  }
}
