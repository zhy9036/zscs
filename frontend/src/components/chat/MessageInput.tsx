import { useState, type KeyboardEvent } from 'react'

export function MessageInput({
  onSend,
  disabled,
  error,
}: {
  onSend: (content: string) => void
  disabled: boolean
  error?: string
}) {
  const [value, setValue] = useState('')

  const submit = () => {
    const text = value.trim()
    if (!text || disabled) return
    onSend(text)
    setValue('')
  }

  const onKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      submit()
    }
  }

  return (
    <div className="border-t border-gray-200 bg-white p-4">
      {error && (
        <p className="mb-2 text-sm text-red-600">{error}</p>
      )}
      <div className="flex items-end gap-2">
        <textarea
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onKeyDown={onKeyDown}
          rows={1}
          placeholder="Send a message…  (Enter to send, Shift+Enter for newline)"
          className="flex-1 resize-none rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-400 focus:outline-none"
        />
        <button
          onClick={submit}
          disabled={disabled || !value.trim()}
          className="rounded-md bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-700 disabled:opacity-50"
        >
          Send
        </button>
      </div>
    </div>
  )
}
