import { apiClient } from './client'
import type { Project } from '../types/project'

export const projectApi = {
  list: () => apiClient.get<{ items: Project[] }>('/projects'),
  get: (id: string) => apiClient.get<Project>(`/projects/${id}`),
  create: (file: File, title?: string) => {
    const form = new FormData()
    form.append('file', file)
    if (title) form.append('title', title)
    return apiClient.upload<Project>('/projects', form)
  },
  rename: (id: string, title: string) =>
    apiClient.patch<Project>(`/projects/${id}`, { title }),
  delete: (id: string) => apiClient.delete<void>(`/projects/${id}`),
}
