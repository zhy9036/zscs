import { useNavigate } from 'react-router-dom'
import { ConfigDropZone } from '../components/upload/ConfigDropZone'
import { useCreateProject } from '../hooks/useProjects'

export function NewProjectPage() {
  const navigate = useNavigate()
  const create = useCreateProject()

  return (
    <div className="flex flex-1 items-center justify-center p-8">
      <div className="w-full max-w-2xl">
        <ConfigDropZone
          onFile={(file) => {
            create.mutate(
              { file },
              {
                onSuccess: (project) => navigate(`/chat/${project.id}`),
              },
            )
          }}
        />

        {create.isPending && (
          <p className="mt-4 text-center text-sm text-gray-500">Uploading…</p>
        )}
        {create.isError && (
          <p className="mt-4 text-center text-sm text-red-600">
            {create.error instanceof Error ? create.error.message : 'Upload failed'}
          </p>
        )}
      </div>
    </div>
  )
}
