import { useState, useRef } from 'react'

interface Props {
  onSend: (text: string) => void
  onSendFile: (file: File) => void
  disabled: boolean
}

export default function MessageInput({ onSend, onSendFile, disabled }: Props) {
  const [text, setText] = useState('')
  const fileRef = useRef<HTMLInputElement>(null)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (text.trim() && !disabled) {
      onSend(text.trim())
      setText('')
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit(e)
    }
  }

  const handleFile = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      if (file.size > 45 * 1024 * 1024) {
        alert('File too large (max 45 MB)')
        return
      }
      onSendFile(file)
    }
    e.target.value = ''
  }

  return (
    <form onSubmit={handleSubmit} className="border-t border-cg-border bg-cg-bg px-4 py-3">
      <div className="flex items-end gap-2">
        {/* File upload button */}
        <button
          type="button"
          onClick={() => fileRef.current?.click()}
          disabled={disabled}
          className="w-9 h-9 flex items-center justify-center rounded-lg text-cg-muted
                     hover:bg-cg-panel hover:text-cg-accent disabled:opacity-50 transition-colors
                     flex-shrink-0"
          title="Send file"
        >
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M21.44 11.05l-9.19 9.19a6 6 0 01-8.49-8.49l9.19-9.19a4 4 0 015.66 5.66l-9.2 9.19a2 2 0 01-2.83-2.83l8.49-8.48" />
          </svg>
        </button>
        <input
          ref={fileRef}
          type="file"
          onChange={handleFile}
          className="hidden"
        />

        {/* Text input */}
        <textarea
          value={text}
          onChange={e => setText(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Type a message..."
          disabled={disabled}
          rows={1}
          className="flex-1 px-4 py-2 bg-cg-panel border border-cg-border rounded-xl text-white
                     placeholder-cg-muted font-mono text-sm resize-none
                     focus:outline-none focus:border-cg-accent transition-colors
                     disabled:opacity-50 max-h-32"
        />

        {/* Send button */}
        <button
          type="submit"
          disabled={disabled || !text.trim()}
          className="w-9 h-9 flex items-center justify-center rounded-lg bg-cg-accent text-cg-bg
                     hover:bg-cg-accent-dim disabled:opacity-50 transition-colors flex-shrink-0"
        >
          <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
            <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z" />
          </svg>
        </button>
      </div>
    </form>
  )
}
