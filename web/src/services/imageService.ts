import api from '@/utils/api'
import type { Image } from '@/types'

export interface PullImageRequest {
  repository: string
  tag?: string
}

export interface PruneResponse {
  imagesDeleted: number
  spaceReclaimed: number
}

export const imageService = {
  list(machineId: string): Promise<Image[]> {
    return api.get(`/machines/${machineId}/images`).then(r => r.images ?? [])
  },

  inspect(machineId: string, imageId: string): Promise<Image> {
    return api.get(`/machines/${machineId}/images/${imageId}`)
  },

  pull(machineId: string, data: PullImageRequest): Promise<Image> {
    return api.post(`/machines/${machineId}/images/pull`, data)
  },

  delete(machineId: string, imageId: string): Promise<void> {
    return api.delete(`/machines/${machineId}/images/${imageId}`).then(() => {})
  },

  prune(machineId: string): Promise<PruneResponse> {
    return api.post(`/machines/${machineId}/images/prune`)
  },
}