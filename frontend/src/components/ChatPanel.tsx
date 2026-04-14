import { useEffect, useRef, useState } from 'react'
import { Room, Message } from '../types'
import MessageBubble from './MessageBubble'
import MessageInput from './MessageInput'

interface Props {
  room: Room | null
  messages: Message[]
  onSend: (text: string) => void
  onSendFile: (file: File) => void
  onRemoveRoom: (name: string) => void
  onRefresh: () => Promise<number>
  disabled: boolean
}

export default function ChatPanel({ room, messages, onSend, onSendFile, onRemoveRoom, onRefresh, disabled }: Props) {
  const [refreshing, setRefreshing] = useState(false)
  const [lastRefresh, setLastRefresh] = useState<string>('')

  const handleRefresh = async () => {
    setRefreshing(true)
    try {
      const count = await onRefresh()
      setLastRefresh(`${new Date().toLocaleTimeString()} — ${count} new`)
    } catch (err) {
      setLastRefresh('refresh failed')
    } finally {
      setRefreshing(false)
    }
  }
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  if (!room) {
    return (
      <div className="flex-1 flex items-center justify-center bg-cg-bg">
        <div className="text-center">
          <p className="text-cg-muted text-lg mb-1">Select a room to start chatting</p>
          <p className="text-cg-muted/50 text-sm">
            Messages are encrypted with AES-256-GCM + ChaCha20-Poly1305
          </p>
        </div>
      </div>
    )
  }

  const displayTitle = room.title || room.name

  return (
    <div className="flex-1 flex flex-col bg-cg-bg h-full">
      {/* Chat header */}
      <div className="flex items-center justify-between px-5 py-3 border-b border-cg-border bg-cg-panel/50">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-full bg-cg-accent flex items-center justify-center text-cg-bg font-bold text-sm">
            {displayTitle[0]?.toUpperCase() ?? '#'}
          </div>
          <div>
            <h3 className="font-display font-semibold text-sm text-white">{displayTitle}</h3>
            <p className="text-xs text-cg-muted">ID: {room.chat_id}</p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          {lastRefresh && (
            <span className="text-xs text-cg-muted hidden sm:inline">{lastRefresh}</span>
          )}
          <button
            onClick={handleRefresh}
            disabled={refreshing}
            className="text-xs text-cg-muted hover:text-cg-accent transition-colors disabled:opacity-50"
            title="Fetch new messages"
          >
            {refreshing ? 'Refreshing…' : 'Refresh'}
          </button>
          <button
            onClick={() => {
              if (confirm(`Remove room "${room.name}"?`)) {
                onRemoveRoom(room.name)
              }
            }}
            className="text-xs text-cg-muted hover:text-cg-danger transition-colors"
            title="Remove room"
          >
            Remove
          </button>
        </div>
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
