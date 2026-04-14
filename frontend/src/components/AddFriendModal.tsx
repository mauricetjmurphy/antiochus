import { useState } from 'react'

interface Props {
  open: boolean
  onClose: () => void
  onAdd: (name: string, chatId: string, sendBotToken: string) => Promise<void>
}

export default function AddFriendModal({ open, onClose, onAdd }: Props) {
  const [name, setName] = useState('')
  const [chatId, setChatId] = useState('')
  const [sendBotToken, setSendBotToken] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  if (!open) return null

  const reset = () => {
    setName('')
    setChatId('')
    setSendBotToken('')
    setError('')
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim() || !chatId.trim()) return
    setSubmitting(true)
    setError('')
    try {
      await onAdd(name.trim().toLowerCase(), chatId.trim(), sendBotToken.trim())
      reset()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add friend')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60" onClick={onClose}>
      <div className="bg-cg-panel border border-cg-border rounded-xl p-6 w-full max-w-sm max-h-[90vh] overflow-y-auto" onClick={e => e.stopPropagation()}>
        <h2 className="font-display font-semibold text-lg mb-4">Add Friend</h2>
        <form onSubmit={handleSubmit} className="space-y-3">
          <div>
            <label className="block text-xs text-cg-muted mb-1">Name</label>
            <input
              type="text"
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder="alice"
              autoFocus
              className="w-full px-3 py-2 bg-cg-bg border border-cg-border rounded-lg text-white
                         placeholder-cg-muted font-mono text-sm
                         focus:outline-none focus:border-cg-accent transition-colors"
            />
          </div>
          <div>
            <label className="block text-xs text-cg-muted mb-1">Their Telegram Chat ID</label>
            <input
              type="text"
              value={chatId}
              onChange={e => setChatId(e.target.value)}
              placeholder="123456789"
              className="w-full px-3 py-2 bg-cg-bg border border-cg-border rounded-lg text-white
                         placeholder-cg-muted font-mono text-sm
                         focus:outline-none focus:border-cg-accent transition-colors"
            />
            <p className="text-xs text-cg-muted mt-1">Their chat ID with their own bot</p>
          </div>
          <div>
            <label className="block text-xs text-cg-muted mb-1">Their Bot Token</label>
            <input
              type="password"
              value={sendBotToken}
              onChange={e => setSendBotToken(e.target.value)}
              placeholder="123456789:ABCdef..."
              className="w-full px-3 py-2 bg-cg-bg border border-cg-border rounded-lg text-white
                         placeholder-cg-muted font-mono text-sm
                         focus:outline-none focus:border-cg-accent transition-colors"
            />
            <p className="text-xs text-cg-muted mt-1">
              Used to send messages to them. Leave blank to use your own bot token.
            </p>
          </div>

          {error && <p className="text-cg-danger text-xs">{error}</p>}

          <div className="flex gap-2 pt-2">
            <button
              type="button"
              onClick={() => { reset(); onClose() }}
              className="flex-1 py-2 text-sm border border-cg-border rounded-lg text-cg-muted
                         hover:bg-cg-border transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting || !name.trim() || !chatId.trim()}
              className="flex-1 py-2 text-sm bg-cg-accent text-cg-bg font-semibold rounded-lg
                         hover:bg-cg-accent-dim disabled:opacity-50 transition-colors"
            >
              {submitting ? 'Adding...' : 'Add'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
