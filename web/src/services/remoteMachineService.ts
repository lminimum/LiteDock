import api from '@/utils/api'
import type {
  RemoteMachine,
  CreateMachineRequest,
  UpdateMachineRequest,
  RemoteContainer
} from '@/types'

export const remoteMachineService = {
  list(): Promise<RemoteMachine[]> {
    return api.get('/machines').then(r => r.data.data ?? [])
  },

  get(id: string): Promise<RemoteMachine> {
    return api.get(`/machines/${id}`).then(r => r.data.data)
  },

  create(data: CreateMachineRequest): Promise<RemoteMachine> {
    return api.post('/machines', data).then(r => r.data.data)
  },

  update(id: string, data: UpdateMachineRequest): Promise<RemoteMachine> {
    return api.put(`/machines/${id}`, data).then(r => r.data.data)
  },

  delete(id: string): Promise<void> {
    return api.delete(`/machines/${id}`).then(() => {})
  },

  testConnection(id: string): Promise<void> {
    return api.post(`/machines/${id}/test`).then(() => {})
  },

  // Container operations
  listContainers(machineId: string): Promise<RemoteContainer[]> {
    return api.get(`/machines/${machineId}/containers`).then(r => r.data.containers ?? [])
  },

  getContainerLogs(machineId: string, containerId: string, tail = '100'): Promise<string> {
    return api.get(`/machines/${machineId}/containers/${containerId}/logs`, {
      params: { tail }
    }).then(r => r.data.logs ?? '')
  },

  execContainer(machineId: string, containerId: string, cmd: string[]): Promise<string> {
    return api.post(`/machines/${machineId}/containers/${containerId}/exec`, { cmd })
      .then(r => r.data.output ?? '')
  },

  startContainer(machineId: string, containerId: string): Promise<void> {
    return api.post(`/machines/${machineId}/containers/${containerId}/start`).then(() => {})
  },

  stopContainer(machineId: string, containerId: string): Promise<void> {
    return api.post(`/machines/${machineId}/containers/${containerId}/stop`).then(() => {})
  },

  restartContainer(machineId: string, containerId: string): Promise<void> {
    return api.post(`/machines/${machineId}/containers/${containerId}/restart`).then(() => {})
  },

  removeContainer(machineId: string, containerId: string, force = false): Promise<void> {
    return api.delete(`/machines/${machineId}/containers/${containerId}`, { params: { force } })
      .then(() => {})
  },

  inspectContainer(machineId: string, containerId: string): Promise<any> {
    return api.get(`/machines/${machineId}/containers/${containerId}`)
      .then(r => r.data.container)
  }
}
