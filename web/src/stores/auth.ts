import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/utils/api'

export interface User {
  id: string
  username: string
  email?: string
  role: string
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('litedock-token'))
  
  const isAuthenticated = computed(() => !!token.value && !!user.value)
  
  const login = async (credentials: { username: string; password: string }) => {
    try {
      const response = await api.post('/auth/login', credentials)
      const { token: authToken, user: userData } = response.data
      
      token.value = authToken
      user.value = userData
      
      localStorage.setItem('litedock-token', authToken)
      
      return { success: true }
    } catch (error: any) {
      return { 
        success: false, 
        message: error.response?.data?.message || '登录失败' 
      }
    }
  }
  
  const logout = () => {
    token.value = null
    user.value = null
    localStorage.removeItem('litedock-token')
  }
  
  const checkAuth = async () => {
    if (!token.value) return false
    
    try {
      const response = await api.get('/auth/me')
      user.value = response.data
      return true
    } catch (error) {
      logout()
      return false
    }
  }
  
  return {
    user,
    token,
    isAuthenticated,
    login,
    logout,
    checkAuth
  }
})