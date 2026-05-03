import axios, { type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { useAuthStore } from '@/stores/auth'

export interface ApiResponse<T = any> {
  code: number
  msg: string
  data: T
}

const api = axios.create({
  baseURL: '/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const authStore = useAuthStore()
    if (authStore.token) {
      config.headers.Authorization = `Bearer ${authStore.token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

api.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const { data } = response
    if (data.code >= 200 && data.code < 300) {
      return data.data
    }
    return Promise.reject(new Error(data.msg || 'Request failed'))
  },
  (error) => {
    if (error.response?.status === 401) {
      const authStore = useAuthStore()
      authStore.logout()
      window.location.href = '/login'
    }
    const msg = error.response?.data?.msg || error.message
    return Promise.reject(new Error(msg))
  }
)

export default api
