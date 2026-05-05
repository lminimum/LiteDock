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

function snakeToCamelImage(raw: any): Image {
  return {
    id: raw.id ?? '',
    machineId: raw.machine_id || raw.machineId || '',
    repoTags: raw.repo_tags || raw.repoTags || [],
    repoDigests: raw.repo_digests || raw.repoDigests || [],
    size: raw.size ?? 0,
    createdAt: raw.created_at || raw.createdAt || '',
    cachedAt: raw.cached_at || raw.cachedAt || '',
    labels: raw.labels ?? {},
  }
}

export const imageService = {
  async list(machineId: string): Promise<Image[]> {
    const r: any = await api.get(`/machines/${machineId}/images`)
    return (r.images ?? []).map(snakeToCamelImage)
  },

  async inspect(machineId: string, imageId: string): Promise<Image> {
    const r: any = await api.get(`/machines/${machineId}/images/${imageId}`)
    return snakeToCamelImage(r)
  },

  async pull(machineId: string, data: PullImageRequest): Promise<Image> {
    const r: any = await api.post(`/machines/${machineId}/images/pull`, data)
    return snakeToCamelImage(r)
  },

  async delete(machineId: string, imageId: string): Promise<void> {
    await api.delete(`/machines/${machineId}/images/${imageId}`)
  },

  async prune(machineId: string): Promise<PruneResponse> {
    const r: any = await api.post(`/machines/${machineId}/images/prune`)
    return r
  },
}
