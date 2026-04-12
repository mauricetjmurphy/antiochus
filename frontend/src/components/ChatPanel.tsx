import { useEffect, useRef } from 'react'
import { Friend, Message } from '../types'
import MessageBubble from './MessageBubble'
import MessageInput from './MessageInput'

interface Props {
  friend: Friend | null
  messages: Message[]
  onSend: (text: string) => void
  onSendFile: (file: File) => void
  onRemoveFriend: (name: string) => void
  disabled: boolean
}

export default function ChatPanel({ friend, messages, onSend, onSendFile, onRemoveFriend, disabled }: Props) {
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  if (!friend) {
    return (
      <div className="flex-1 flex items-center justify-center bg-cg-bg">
        <div className="text-center">
          <p className="text-cg-muted text-lg mb-1">Select a friend to start chatting</p>
          <p className="text-cg-muted/50 text-sm">
            Messages are encrypted with AES-256-GCM + ChaCha20-Poly1305
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex-1 flex flex-col bg-cg-bg h-full">
      {/* Chat header */}
      <div className="flex items-center justify-between px-5 py-3 border-b border-cg-border bg-cg-panel/50">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-full bg-cg-accent flex items-center justify-center text-cg-bg font-bold text-sm">
            {friend.name[0].toUpperCase()}
          </div>
          <div>
            <h3 className="font-display font-semibold text-sm text-white">{friend.name}</h3>
            <p className="text-xs text-cg-muted">ID: {friend.chat_id}</p>
          </div>
        </div>
        <button
          onClick={() => {
            if (confirm(`Remove ${friend.name}?`)) {
              onRemoveFriend(friend.name)
            }
          }}
          className="text-xs text-cg-muted hover:text-cg-danger transition-colors"
          title="Remove friend"
        >
          Remove
        </button>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto px-5 py-4">
        {messages.length === 0 ? (
          <div className="flex items-center justify-center h-full">
            <p className="text-cg-muted text-sm">
              No messages yet. Say hello!
            </p>
          </div>
        ) : (
          messages.map((msg, i) => (
            <MessageBubble key={`${msg.time}-${i}`} message={msg} />
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <MessageInput onSend={onSend} onSendFile={onSendFile} disabled={disabled} />
    </div>
  )
}
