import api from '@/utils/api'
import type { Network } from '@/types'

export const networkService = {
  async listNetworks(machineId: string): Promise<Network[]> {
    const r: any = await api.get(`/machines/${machineId}/networks`)
    return r.networks ?? []
  },

  async createNetwork(machineId: string, data: { name: string; driver?: string }): Promise<void> {
    await api.post(`/machines/${machineId}/networks`, data)
  },

  async deleteNetwork(machineId: string, networkId: string): Promise<void> {
    await api.delete(`/machines/${machineId}/networks/${networkId}`)
  }
}
