import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'
import { useDeleteProject, useProjects, useRenameProject } from '../../hooks/useProjects'
import type { Project } from '../../types/project'

export function Sidebar() {
  const navigate = useNavigate()
  const { user, logout } = useAuthStore()
  const { data: projects, isLoading } = useProjects()
  const rename = useRenameProject()
  const del = useDeleteProject()

  return (
    <aside className="flex w-[260px] shrink-0 flex-col border-r border-gray-200 bg-white">
      <div className="p-3">
        <button
          onClick={() => navigate('/new')}
          className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm font-medium hover:bg-gray-50"
        >
          + New project
        </button>
      </div>

      <div className="px-3 pb-1 text-xs font-semibold uppercase tracking-wide text-gray-400">
        Recent
      </div>

      <nav className="flex-1 overflow-y-auto px-2">
        {isLoading && <p className="p-2 text-sm text-gray-400">Loading…</p>}
        {!isLoading && projects?.length === 0 && (
          <p className="p-2 text-sm text-gray-400">No projects yet.</p>
        )}
        <ul>
          {projects?.map((p) => (
            <SidebarItem
              key={p.id}
              project={p}
              onRename={(title) => rename.mutate({ id: p.id, title })}
              onDelete={() => del.mutate(p.id)}
            />
          ))}
        </ul>
      </nav>

      <div className="border-t border-gray-200 p-3">
        <div className="mb-2 truncate text-sm text-gray-600">
          {user?.username ?? '—'}
        </div>
        <button
          onClick={() => {
            logout()
            navigate('/login')
          }}
          className="w-full rounded-md px-3 py-2 text-left text-sm hover:bg-gray-50"
        >
          Logout
        </button>
      </div>
    </aside>
  )
}

function SidebarItem({
  project,
  onRename,
  onDelete,
}: {
  project: Project
  onRename: (title: string) => void
  onDelete: () => void
}) {
  const navigate = useNavigate()
  const [editing, setEditing] = useState(false)
  const [draft, setDraft] = useState(project.title)

  const commit = () => {
    const t = draft.trim()
    if (t && t !== project.title) onRename(t)
    else setDraft(project.title)
    setEditing(false)
  }

  return (
    <li className="group flex items-center gap-1 rounded-md hover:bg-gray-50">
      {editing ? (
        <input
          autoFocus
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onBlur={commit}
          onKeyDown={(e) => {
            if (e.key === 'Enter') commit()
            if (e.key === 'Escape') {
              setDraft(project.title)
              setEditing(false)
            }
          }}
          className="m-1 flex-1 rounded border border-blue-400 px-2 py-1 text-sm"
        />
      ) : (
        <button
          onClick={() => navigate(`/chat/${project.id}`)}
          className="flex-1 truncate px-2 py-2 text-left text-sm"
        >
          {project.title}
        </button>
      )}

      {!editing && (
        <div className="hidden pr-1 group-hover:flex">
          <button
            onClick={() => {
              setDraft(project.title)
              setEditing(true)
            }}
            className="rounded p-1 text-xs text-gray-400 hover:text-gray-700"
            aria-label="Rename"
          >
            ✎
          </button>
          <button
            onClick={onDelete}
            className="rounded p-1 text-xs text-gray-400 hover:text-red-600"
            aria-label="Delete"
          >
            🗑
          </button>
        </div>
      )}
    </li>
  )
}
