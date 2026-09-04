import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { projectApi } from '../api/projects'
import type { Project } from '../types/project'

const KEY = ['projects']

export function useProjects() {
  return useQuery<Project[]>({
    queryKey: KEY,
    queryFn: async () => {
      const res = await projectApi.list()
      return res.items
    },
  })
}

export function useCreateProject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ file, title }: { file: File; title?: string }) =>
      projectApi.create(file, title),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEY }),
  })
}

export function useRenameProject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, title }: { id: string; title: string }) =>
      projectApi.rename(id, title),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEY }),
  })
}

export function useDeleteProject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => projectApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEY }),
  })
}
