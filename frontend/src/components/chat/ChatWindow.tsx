import { useMessages, useSendMessage } from '../../hooks/useMessages'
import { MessageList } from './MessageList'
import { MessageInput } from './MessageInput'

export function ChatWindow({ projectId }: { projectId: string }) {
  const { data: messages = [], isLoading } = useMessages(projectId)
  const send = useSendMessage(projectId)

  return (
    <div className="flex h-full flex-col">
      <MessageList
        messages={messages}
        loading={send.isPending}
        onSuggestion={(text) => send.mutate(text)}
      />
      <MessageInput
        onSend={(content) => send.mutate(content)}
        disabled={send.isPending}
        error={
          send.isError
            ? send.error instanceof Error
              ? send.error.message
              : 'Failed to send'
            : undefined
        }
      />
      {isLoading && messages.length === 0 && (
        <div className="p-4 text-sm text-gray-400">Loading messages…</div>
      )}
    </div>
  )
}
