import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchHome, type User, type HomeResponse } from '@/api'

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)
  const homeData = ref<HomeResponse | null>(null)
  const homeLoading = ref(false)

  const fetchUser = async () => {
    homeLoading.value = true
    const res = await fetchHome()
    user.value = res.current_user
    homeData.value = res
    homeLoading.value = false
  }

  return { user, homeData, homeLoading, fetchUser }
})
