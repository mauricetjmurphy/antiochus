import { useState, useCallback } from 'react'
import { Message } from '../types'

export function useMessages() {
  const [messages, setMessages] = useState<Record<string, Message[]>>({})

  const addMessage = useCallback((room: string, msg: Message) => {
    setMessages(prev => ({
      ...prev,
      [room]: [...(prev[room] || []), msg],
    }))
  }, [])

  const setRoomMessages = useCallback((room: string, msgs: Message[]) => {
    setMessages(prev => ({
      ...prev,
      [room]: msgs,
    }))
  }, [])

  const getMessages = useCallback((room: string): Message[] => {
    return messages[room] || []
  }, [messages])

  return { messages, addMessage, setRoomMessages, getMessages }
}
