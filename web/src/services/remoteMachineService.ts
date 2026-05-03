import api from '@/utils/api'
import type {
  RemoteMachine,
  CreateMachineRequest,
  UpdateMachineRequest,
  RemoteContainer
} from '@/types'

export const remoteMachineService = {
  list(): Promise<RemoteMachine[]> {
    return api.get('/machines')
  },

  get(id: string): Promise<RemoteMachine> {
    return api.get(`/machines/${id}`)
  },

  create(data: CreateMachineRequest): Promise<RemoteMachine> {
    return api.post('/machines', data)
  },

  update(id: string, data: UpdateMachineRequest): Promise<RemoteMachine> {
    return api.put(`/machines/${id}`, data)
  },

  delete(id: string): Promise<void> {
    return api.delete(`/machines/${id}`)
  },

  testConnection(id: string): Promise<void> {
    return api.post(`/machines/${id}/test`)
  },

  listContainers(machineId: string): Promise<RemoteContainer[]> {
    return api.get(`/machines/${machineId}/containers`)
  },

  getContainerLogs(machineId: string, containerId: string, tail = '100'): Promise<string> {
    return api.get(`/machines/${machineId}/containers/${containerId}/logs`, {
      params: { tail }
    })
  },

  execContainer(machineId: string, containerId: string, cmd: string[]): Promise<string> {
    return api.post(`/machines/${machineId}/containers/${containerId}/exec`, { cmd })
  },

  startContainer(machineId: string, containerId: string): Promise<void> {
    return api.post(`/machines/${machineId}/containers/${containerId}/start`)
  },

  stopContainer(machineId: string, containerId: string): Promise<void> {
    return api.post(`/machines/${machineId}/containers/${containerId}/stop`)
  },

  restartContainer(machineId: string, containerId: string): Promise<void> {
    return api.post(`/machines/${machineId}/containers/${containerId}/restart`)
  },

  removeContainer(machineId: string, containerId: string, force = false): Promise<void> {
    return api.delete(`/machines/${machineId}/containers/${containerId}`, { params: { force } })
  },

  inspectContainer(machineId: string, containerId: string): Promise<any> {
    return api.get(`/machines/${machineId}/containers/${containerId}`)
  }
}
