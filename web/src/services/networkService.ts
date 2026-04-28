import api from '@/utils/api'
import type { Network } from '@/types'

export const networkService = {
  listNetworks(machineId: string): Promise<Network[]> {
    return api.get(`/machines/${machineId}/networks`).then(r => r.data.networks ?? [])
  },

  createNetwork(machineId: string, data: { name: string; driver?: string }): Promise<Network> {
    return api.post(`/machines/${machineId}/networks`, data).then(r => r.data.data)
  },

  deleteNetwork(machineId: string, networkId: string): Promise<void> {
    return api.delete(`/machines/${machineId}/networks/${networkId}`).then(() => {})
  }
}