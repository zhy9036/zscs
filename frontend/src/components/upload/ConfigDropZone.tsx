import { useRef, useState, type DragEvent } from 'react'

const ALLOWED = ['.conf', '.txt', '.xml', '.json', '.csv']

export function ConfigDropZone({ onFile }: { onFile: (file: File) => void }) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [dragging, setDragging] = useState(false)
  const [error, setError] = useState('')

  const accept = ALLOWED.join(',')

  const validate = (file: File | undefined): file is File => {
    if (!file) return false
    const name = file.name.toLowerCase()
    const ok = ALLOWED.some((ext) => name.endsWith(ext))
    if (!ok) {
      setError(`Unsupported file type. Allowed: ${ALLOWED.join(', ')}`)
      return false
    }
    setError('')
    return true
  }

  const onDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    setDragging(false)
    const file = e.dataTransfer.files?.[0]
    if (validate(file)) onFile(file)
  }

  return (
    <div
      onDragOver={(e) => {
        e.preventDefault()
        setDragging(true)
      }}
      onDragLeave={() => setDragging(false)}
      onDrop={onDrop}
      className={`flex flex-col items-center justify-center rounded-xl border-2 border-dashed p-12 text-center transition ${
        dragging ? 'border-blue-500 bg-blue-50' : 'border-gray-300 bg-white'
      }`}
    >
      <h2 className="text-lg font-semibold">Start a Zscaler Migration</h2>
      <p className="mt-2 text-gray-500">Drag and drop your Zscaler config here</p>
      <p className="my-3 text-gray-400">or</p>
      <button
        onClick={() => inputRef.current?.click()}
        className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium hover:bg-gray-50"
      >
        Browse Files
      </button>
      <p className="mt-4 text-xs text-gray-400">
        Supported: {ALLOWED.join(', ')}
      </p>

      {error && (
        <p className="mt-3 text-sm text-red-600">{error}</p>
      )}

      <input
        ref={inputRef}
        type="file"
        accept={accept}
        className="hidden"
        onChange={(e) => {
          const file = e.target.files?.[0]
          if (validate(file)) onFile(file)
          e.target.value = ''
        }}
      />
    </div>
  )
}
