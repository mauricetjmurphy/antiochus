export interface Room {
  name: string
  chat_id: string
  title?: string
  added: string
}

export interface RoomCandidate {
  chat_id: string
  title: string
  type: string
  first: string
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
  room?: string
  message?: Message
  active?: boolean
  error?: string
}
