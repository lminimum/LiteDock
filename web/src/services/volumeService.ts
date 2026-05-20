import api from '@/utils/api'
import type { Volume } from '@/types'

export const volumeService = {
  async listVolumes(machineId: string): Promise<Volume[]> {
    const r: any = await api.get(`/machines/${machineId}/volumes`)
    return r.volumes ?? []
  },

  async createVolume(machineId: string, data: { name: string; driver?: string }): Promise<Volume> {
    return await api.post<Volume>(`/machines/${machineId}/volumes`, data)
  },

  async deleteVolume(machineId: string, volumeName: string): Promise<void> {
    await api.delete(`/machines/${machineId}/volumes/${volumeName}`)
  }
}
