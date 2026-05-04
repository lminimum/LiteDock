import type { AxiosInstance, AxiosRequestConfig } from 'axios'
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import { redirectToLogin } from '@/utils/redirect'

export interface ApiResponse<T = unknown> {
  code: number
  msg: string
  data: T
}

// HttpClient 类型：拦截器已将 response.data.data 解包，直接返回 T
export interface HttpClient {
  get<T>(url: string, config?: AxiosRequestConfig): Promise<T>
  post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T>
  put<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T>
  delete<T>(url: string, config?: AxiosRequestConfig): Promise<T>
}

const instance: AxiosInstance = axios.create({
  baseURL: '/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

instance.interceptors.request.use(
  (config) => {
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

instance.interceptors.response.use(
  (response): any => {
    const body = response.data as ApiResponse
    if (body.code >= 200 && body.code < 300) {
      return body.data
    }
    return Promise.reject(new Error(body.msg || 'Request failed'))
  },
  (error): any => {
    if (error.response?.status === 401) {
      const authStore = useAuthStore()
      authStore.logout()
      redirectToLogin()
    }
    const msg = error.response?.data?.msg || error.message
    return Promise.reject(new Error(msg))
  }
)

const api = instance as unknown as HttpClient

export default api
