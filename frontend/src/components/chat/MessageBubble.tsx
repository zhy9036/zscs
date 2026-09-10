import ReactMarkdown from 'react-markdown'
import type { Message } from '../../types/message'

export function MessageBubble({ message }: { message: Message }) {
  const isUser = message.role === 'user'
  const isSystem = message.role === 'system'

  return (
    <div className={`flex ${isUser ? 'justify-end' : 'justify-start'} px-4 py-2`}>
      <div
        className={`max-w-[80%] rounded-lg px-4 py-2 text-sm ${
          isUser
            ? 'bg-gray-900 text-white'
            : isSystem
              ? 'bg-yellow-50 text-yellow-900'
              : 'bg-white border border-gray-200 text-gray-900'
        }`}
      >
        {isUser ? (
          <p className="whitespace-pre-wrap break-words">{message.content}</p>
        ) : (
          <div className="max-w-none space-y-2 [&_p]:whitespace-pre-wrap [&_p]:break-words [&_pre]:overflow-x-auto [&_pre]:rounded-md [&_pre]:bg-gray-50 [&_pre]:p-3 [&_pre]:text-xs [&_code]:break-words [&_li]:break-words">
            <ReactMarkdown>{message.content}</ReactMarkdown>
          </div>
        )}
      </div>
    </div>
  )
}
