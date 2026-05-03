import api from '@/utils/api'
import type { Volume } from '@/types'

export const volumeService = {
  listVolumes(machineId: string): Promise<Volume[]> {
    return api.get(`/machines/${machineId}/volumes`).then(r => r.volumes ?? [])
  },
  createVolume(machineId: string, data: { name: string; driver?: string }): Promise<Volume> {
    return api.post(`/machines/${machineId}/volumes`, data)
  },
  deleteVolume(machineId: string, volumeName: string): Promise<void> {
    return api.delete(`/machines/${machineId}/volumes/${volumeName}`).then(() => {})
  }
}