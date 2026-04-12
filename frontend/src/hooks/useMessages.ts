import { useState, useCallback } from 'react'
import { Message } from '../types'

export function useMessages() {
  const [messages, setMessages] = useState<Record<string, Message[]>>({})

  const addMessage = useCallback((friend: string, msg: Message) => {
    setMessages(prev => ({
      ...prev,
      [friend]: [...(prev[friend] || []), msg],
    }))
  }, [])

  const setFriendMessages = useCallback((friend: string, msgs: Message[]) => {
    setMessages(prev => ({
      ...prev,
      [friend]: msgs,
    }))
  }, [])

  const getMessages = useCallback((friend: string): Message[] => {
    return messages[friend] || []
  }, [messages])

  return { messages, addMessage, setFriendMessages, getMessages }
}
