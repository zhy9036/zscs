import { useNavigate } from 'react-router-dom'

export function HomePage() {
  const navigate = useNavigate()
  return (
    <div className="flex flex-1 items-center justify-center p-8">
      <div className="max-w-md text-center">
        <h1 className="text-2xl font-semibold">Zscaler Migration</h1>
        <p className="mt-2 text-gray-500">
          Upload a Zscaler configuration to start a migration project.
        </p>
        <button
          onClick={() => navigate('/new')}
          className="mt-6 rounded-md bg-gray-900 px-5 py-2.5 text-sm font-medium text-white hover:bg-gray-700"
        >
          + New project
        </button>
      </div>
    </div>
  )
}
