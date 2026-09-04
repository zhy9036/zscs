import { useEffect, useRef } from 'react'
import type { Message } from '../../types/message'
import { MessageBubble } from './MessageBubble'

const SUGGESTIONS = [
  'Analyze this configuration',
  'What migration issues should I know about?',
  'Summarize the configuration',
]

export function MessageList({
  messages,
  loading,
  onSuggestion,
}: {
  messages: Message[]
  loading: boolean
  onSuggestion: (text: string) => void
}) {
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, loading])

  if (messages.length === 0) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center p-8 text-center">
        <h2 className="text-lg font-semibold">Zscaler configuration loaded</h2>
        <p className="mt-2 text-gray-500">Ask a question to get started.</p>
        <div className="mt-6 flex flex-wrap justify-center gap-2">
          {SUGGESTIONS.map((s) => (
            <button
              key={s}
              onClick={() => onSuggestion(s)}
              className="rounded-full border border-gray-300 px-4 py-1.5 text-sm hover:bg-gray-50"
            >
              {s}
            </button>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="flex-1 overflow-y-auto py-4">
      {messages.map((m) => (
        <MessageBubble key={m.id} message={m} />
      ))}
      {loading && (
        <div className="px-4 py-2 text-sm text-gray-400">Assistant is typing…</div>
      )}
      <div ref={bottomRef} />
    </div>
  )
}
