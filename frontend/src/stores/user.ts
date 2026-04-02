import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { User } from '@/types'

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(
    JSON.parse(localStorage.getItem('user') || 'null'),
  )
  const showLogin = ref(false)

  function setUser(u: User | null) {
    user.value = u
    if (u) {
      localStorage.setItem('user', JSON.stringify(u))
    } else {
      localStorage.removeItem('user')
    }
  }

  function setShowLogin(val: boolean) {
    showLogin.value = val
  }

  function logout() {
    setUser(null)
  }

  return { user, showLogin, setUser, setShowLogin, logout }
})
