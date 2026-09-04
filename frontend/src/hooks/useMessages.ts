import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { messageApi } from '../api/messages'
import type { Message } from '../types/message'

export function useMessages(projectId: string | undefined) {
  return useQuery<Message[]>({
    queryKey: ['messages', projectId],
    queryFn: async () => {
      if (!projectId) return []
      const res = await messageApi.list(projectId)
      return res.items
    },
    enabled: !!projectId,
  })
}

export function useSendMessage(projectId: string | undefined) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (content: string) => {
      if (!projectId) throw new Error('no project')
      return messageApi.send(projectId, content)
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['messages', projectId] })
      qc.invalidateQueries({ queryKey: ['projects'] })
    },
  })
}
