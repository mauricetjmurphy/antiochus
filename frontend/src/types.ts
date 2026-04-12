export interface Friend {
  name: string
  chat_id: string
  added: string
}

export interface Message {
  time: string
  sender: string
  content: string
  type: 'text' | 'file'
  filename?: string
  filePath?: string
}

export interface WSEvent {
  type: 'new_message' | 'poll_status' | 'error'
  friend?: string
  message?: Message
  active?: boolean
  error?: string
}
