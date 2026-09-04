export type MessageRole = 'user' | 'assistant' | 'system'

export interface Message {
  id: string
  role: MessageRole
  content: string
  created_at: string
}

export interface SendResult {
  user_message: Message
  assistant_message: Message | null
  status: string
}
