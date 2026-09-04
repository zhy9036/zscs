import { apiClient } from './client'
import type { Message, SendResult } from '../types/message'

export const messageApi = {
  list: (projectId: string) =>
    apiClient.get<{ items: Message[] }>(`/projects/${projectId}/messages`),
  send: (projectId: string, content: string) =>
    apiClient.post<SendResult>(`/projects/${projectId}/messages`, { content }),
}
