import api from '@/utils/api'
import type { ComposeProject } from '@/types'

function snakeToCamelProject(raw: any): ComposeProject {
  return {
    id: raw.id || '',
    machineId: raw.machine_id || raw.machineId || '',
    name: raw.name || '',
    filePath: raw.file_path || raw.filePath || '',
    projectName: raw.project_name || raw.projectName || '',
    status: raw.status || 'unknown',
    services: (raw.services || []).map((s: any) => ({
      name: s.name || '',
      serviceName: s.service_name || s.serviceName || '',
      image: s.image || '',
      status: s.status || 'unknown',
      health: s.health || 'none',
      replicas: s.replicas ?? 1,
      publishers: (s.publishers || []).map((p: any) => ({
        url: p.url || '',
        targetPort: p.target_port || p.targetPort || 0,
        publishedPort: p.published_port || p.publishedPort || 0,
      })),
    })),
    createdAt: raw.created_at || raw.createdAt || '',
    updatedAt: raw.updated_at || raw.updatedAt || '',
    cachedAt: raw.cached_at || raw.cachedAt || '',
  }
}

export const composeService = {
  async listProjects(machineId: string): Promise<ComposeProject[]> {
    const r = await api.get(`/machines/${machineId}/compose`)
    return (r.projects ?? []).map(snakeToCamelProject)
  },

  async getProject(machineId: string, projectName: string): Promise<ComposeProject> {
    const r = await api.get(`/machines/${machineId}/compose/${projectName}`)
    return snakeToCamelProject(r)
  },

  async createProject(machineId: string, data: { name: string; content: string; file_path?: string }): Promise<ComposeProject> {
    const r = await api.post(`/machines/${machineId}/compose`, data)
    return snakeToCamelProject(r)
  },

  async updateProject(machineId: string, projectName: string, content: string): Promise<void> {
    await api.put(`/machines/${machineId}/compose/${projectName}`, { content })
  },

  async deleteProject(machineId: string, projectName: string): Promise<void> {
    await api.delete(`/machines/${machineId}/compose/${projectName}`)
  },

  async up(machineId: string, projectName: string): Promise<void> {
    await api.post(`/machines/${machineId}/compose/${projectName}/up`)
  },

  async down(machineId: string, projectName: string, volumes?: boolean): Promise<void> {
    const params = volumes ? { volumes: 'true' } : undefined
    await api.post(`/machines/${machineId}/compose/${projectName}/down`, undefined, { params })
  },

  async build(machineId: string, projectName: string): Promise<void> {
    await api.post(`/machines/${machineId}/compose/${projectName}/build`)
  },

  async start(machineId: string, projectName: string): Promise<void> {
    await api.post(`/machines/${machineId}/compose/${projectName}/start`)
  },

  async stop(machineId: string, projectName: string): Promise<void> {
    await api.post(`/machines/${machineId}/compose/${projectName}/stop`)
  },

  async restart(machineId: string, projectName: string): Promise<void> {
    await api.post(`/machines/${machineId}/compose/${projectName}/restart`)
  },

  async getLogs(machineId: string, projectName: string): Promise<string> {
    const r = await api.get(`/machines/${machineId}/compose/${projectName}/logs`)
    return r.logs ?? ''
  },
}

export default composeService