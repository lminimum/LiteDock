import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import api from '@/utils/api'
import { useAuthStore } from '../auth'

// Setup is already mocked in setup.ts

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    localStorage.clear()
  })

  describe('checkSetupStatus', () => {
    it('returns true when setup_complete is false (need setup)', async () => {
      vi.mocked(api.get).mockResolvedValueOnce({ setup_complete: false })

      const store = useAuthStore()
      const result = await store.checkSetupStatus()

      expect(result).toBe(false)
      expect(api.get).toHaveBeenCalledWith('/auth/setup-status')
    })

    it('returns true when setup_complete is true (setup done)', async () => {
      vi.mocked(api.get).mockResolvedValueOnce({ setup_complete: true })

      const store = useAuthStore()
      const result = await store.checkSetupStatus()

      expect(result).toBe(true)
      expect(api.get).toHaveBeenCalledWith('/auth/setup-status')
    })

    it('returns false on API error', async () => {
      vi.mocked(api.get).mockRejectedValueOnce(new Error('Network error'))

      const store = useAuthStore()
      const result = await store.checkSetupStatus()

      expect(result).toBe(false)
    })

    it('caches result and does not call API again on second call', async () => {
      vi.mocked(api.get).mockResolvedValueOnce({ setup_complete: true })

      const store = useAuthStore()
      const first = await store.checkSetupStatus()
      const second = await store.checkSetupStatus()

      expect(first).toBe(true)
      expect(second).toBe(true)
      expect(api.get).toHaveBeenCalledTimes(1)
    })

    it('caches error result and does not retry', async () => {
      vi.mocked(api.get).mockRejectedValueOnce(new Error('Server error'))

      const store = useAuthStore()
      const first = await store.checkSetupStatus()
      const second = await store.checkSetupStatus()

      expect(first).toBe(false)
      expect(second).toBe(false)
      expect(api.get).toHaveBeenCalledTimes(1)
    })

    it('refreshSetupStatus clears cache and re-fetches', async () => {
      vi.mocked(api.get)
        .mockRejectedValueOnce(new Error('Server error'))
        .mockResolvedValueOnce({ setup_complete: true })

      const store = useAuthStore()
      const first = await store.checkSetupStatus()
      const refreshed = await store.refreshSetupStatus()

      expect(first).toBe(false)
      expect(refreshed).toBe(true)
      expect(api.get).toHaveBeenCalledTimes(2)
    })
  })

  describe('login', () => {
    it('stores token and user on success', async () => {
      const mockUser = { id: '1', username: 'admin', email: 'admin@test.com', role: 'admin' }
      const mockResponse = { token: 'test-token-123', user: mockUser }
      vi.mocked(api.post).mockResolvedValueOnce(mockResponse)

      const store = useAuthStore()
      const result = await store.login({ username: 'admin', password: 'password123' })

      expect(result).toEqual({ success: true })
      expect(store.token).toBe('test-token-123')
      expect(store.user).toEqual(mockUser)
      expect(localStorage.getItem('litedock-token')).toBe('test-token-123')
      expect(api.post).toHaveBeenCalledWith('/auth/login', { username: 'admin', password: 'password123' })
    })

    it('returns error message on failure', async () => {
      vi.mocked(api.post).mockRejectedValueOnce(new Error('Invalid credentials'))

      const store = useAuthStore()
      const result = await store.login({ username: 'admin', password: 'wrong' })

      expect(result).toEqual({ success: false, message: 'Invalid credentials' })
      expect(store.token).toBeNull()
      expect(store.user).toBeNull()
    })
  })

  describe('logout', () => {
    it('clears user and token', () => {
      const store = useAuthStore()
      store.token = 'some-token'
      store.user = { id: '1', username: 'admin', role: 'admin' }
      localStorage.setItem('litedock-token', 'some-token')
      localStorage.setItem('litedock-user', '{"id":"1"}')

      store.logout()

      expect(store.token).toBeNull()
      expect(store.user).toBeNull()
      expect(localStorage.getItem('litedock-token')).toBeNull()
      expect(localStorage.getItem('litedock-user')).toBeNull()
    })
  })

  describe('checkAuth', () => {
    it('returns false with no token', async () => {
      const store = useAuthStore()
      store.token = null

      const result = await store.checkAuth()

      expect(result).toBe(false)
      expect(api.get).not.toHaveBeenCalled()
    })

    it('returns true with valid token and stores user', async () => {
      const mockUser = { id: '1', username: 'admin', email: 'admin@test.com', role: 'admin' }
      vi.mocked(api.get).mockResolvedValueOnce(mockUser)

      const store = useAuthStore()
      store.token = 'valid-token'

      const result = await store.checkAuth()

      expect(result).toBe(true)
      expect(store.user).toEqual(mockUser)
      expect(api.get).toHaveBeenCalledWith('/auth/me')
    })

    it('calls logout on API error', async () => {
      vi.mocked(api.get).mockRejectedValueOnce(new Error('Unauthorized'))

      const store = useAuthStore()
      store.token = 'invalid-token'
      store.user = { id: '1', username: 'admin', role: 'admin' }

      const result = await store.checkAuth()

      expect(result).toBe(false)
      expect(store.token).toBeNull()
      expect(store.user).toBeNull()
    })
  })
})