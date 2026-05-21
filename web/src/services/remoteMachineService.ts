import api from '@/utils/api'
import type {
  RemoteMachine,
  CreateMachineRequest,
  UpdateMachineRequest,
  RemoteContainer
} from '@/types'

export const remoteMachineService = {
  async list(): Promise<RemoteMachine[]> {
    return await api.get('/machines')
  },

  async get(id: string): Promise<RemoteMachine> {
    return await api.get(`/machines/${id}`)
  },

  async create(data: CreateMachineRequest): Promise<RemoteMachine> {
    return await api.post('/machines', data)
  },

  async update(id: string, data: UpdateMachineRequest): Promise<RemoteMachine> {
    return await api.put(`/machines/${id}`, data)
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/machines/${id}`)
  },

  async testConnection(id: string): Promise<void> {
    await api.post(`/machines/${id}/test`)
  },

  async listContainers(machineId: string): Promise<RemoteContainer[]> {
    return await api.get(`/machines/${machineId}/containers`)
  },

  async getContainerLogs(machineId: string, containerId: string, tail = '100'): Promise<string> {
    return await api.get(`/machines/${machineId}/containers/${containerId}/logs`, {
      params: { tail }
    })
  },

  async execContainer(machineId: string, containerId: string, cmd: string[]): Promise<string> {
    return await api.post(`/machines/${machineId}/containers/${containerId}/exec`, { cmd })
  },

  async startContainer(machineId: string, containerId: string): Promise<void> {
    await api.post(`/machines/${machineId}/containers/${containerId}/start`)
  },

  async stopContainer(machineId: string, containerId: string): Promise<void> {
    await api.post(`/machines/${machineId}/containers/${containerId}/stop`)
  },

  async restartContainer(machineId: string, containerId: string): Promise<void> {
    await api.post(`/machines/${machineId}/containers/${containerId}/restart`)
  },

  async pauseContainer(machineId: string, containerId: string): Promise<void> {
    await api.post(`/machines/${machineId}/containers/${containerId}/pause`)
  },

  async resumeContainer(machineId: string, containerId: string): Promise<void> {
    await api.post(`/machines/${machineId}/containers/${containerId}/unpause`)
  },

  async killContainer(machineId: string, containerId: string): Promise<void> {
    await api.post(`/machines/${machineId}/containers/${containerId}/kill`)
  },

  async removeContainer(machineId: string, containerId: string, force = false): Promise<void> {
    await api.delete(`/machines/${machineId}/containers/${containerId}`, { params: { force } })
  },

  async inspectContainer(machineId: string, containerId: string): Promise<any> {
    return await api.get(`/machines/${machineId}/containers/${containerId}`)
  },

  async createContainer(machineId: string, data: CreateContainerRequest): Promise<{ taskId: string }> {
    return await api.post(`/machines/${machineId}/containers/create`, data)
  }
}

export interface CreateContainerRequest {
  name?: string
  image: string
  env?: string[]
  ports?: string[]
  volumes?: string[]
  network?: string
  cmd?: string[]
  labels?: Record<string, string>
}
