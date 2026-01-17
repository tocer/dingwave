import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export const useAppStore = defineStore('app', () => {
  const isDark = ref(localStorage.getItem('isDark') === 'true')
  const collapsed = ref(localStorage.getItem('collapsed') === 'true')

  watch(isDark, (v) => localStorage.setItem('isDark', String(v)))
  watch(collapsed, (v) => localStorage.setItem('collapsed', String(v)))

  return { isDark, collapsed }
})
