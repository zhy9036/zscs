import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { projectApi } from '../api/projects'
import { ChatWindow } from '../components/chat/ChatWindow'

export function ProjectPage() {
  const { projectId } = useParams<{ projectId: string }>()
  const { data: project, isLoading } = useQuery({
    queryKey: ['project', projectId],
    queryFn: () => projectApi.get(projectId!),
    enabled: !!projectId,
  })

  if (!projectId) return null

  return (
    <div className="flex h-full flex-col">
      <header className="border-b border-gray-200 bg-white px-6 py-3">
        <h1 className="truncate text-base font-semibold">
          {isLoading ? 'Loading…' : project?.title ?? 'Project'}
        </h1>
      </header>
      <div className="flex-1 overflow-hidden">
        <ChatWindow projectId={projectId} />
      </div>
    </div>
  )
}
