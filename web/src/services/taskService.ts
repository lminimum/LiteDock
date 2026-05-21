import api from '@/utils/api'
import type { Task } from '@/types'

export const taskService = {
  async list(): Promise<Task[]> {
    return await api.get('/tasks')
  },

  async get(id: string): Promise<Task> {
    return await api.get(`/tasks/${id}`)
  }
}
