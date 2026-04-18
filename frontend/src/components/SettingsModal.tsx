import { useState, useEffect } from 'react'
import * as api from '../api'

interface Props {
  open: boolean
  onClose: () => void
  onSaveToken: (token: string) => Promise<void>
  botName: string
  tokenSet: boolean
}

export default function SettingsModal({ open, onClose, onSaveToken, botName, tokenSet }: Props) {
  const [token, setToken] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [whoAmI, setWhoAmI] = useState<{
    bot_username: string
    bot_id: string
    chat_ids: { chat_id: string; username: string; source: string }[]
    hint: string
  } | null>(null)

  // Auto-fetch whoami when modal opens and token is set
  useEffect(() => {
    if (open && tokenSet) {
      api.whoAmI().then(result => {
        if (!result.error) setWhoAmI(result)
      }).catch(() => {})
    }
    if (!open) setWhoAmI(null)
  }, [open, tokenSet])

  if (!open) return null

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!token.trim()) return
    setSaving(true)
    setError('')
    try {
      await onSaveToken(token.trim())
      setToken('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save token')
    } finally {
      setSaving(false)
    }
  }

  const handleRefresh = async () => {
    try {
      const result = await api.whoAmI()
      if (!result.error) setWhoAmI(result)
    } catch {
      // ignore
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60" onClick={onClose}>
      <div className="bg-cg-panel border border-cg-border rounded-xl p-6 w-full max-w-sm max-h-[90vh] overflow-y-auto" onClick={e => e.stopPropagation()}>
        <h2 className="font-display font-semibold text-lg mb-4">Settings</h2>

        {/* Bot status */}
        <div className="flex items-center gap-2 mb-4">
          <span className={`w-2 h-2 rounded-full ${tokenSet ? 'bg-cg-accent' : 'bg-cg-danger'}`} />
          <span className="text-sm">
            {tokenSet ? `Bot: @${botName}` : 'No bot token configured'}
          </span>
        </div>

        {/* Token form */}
        <form onSubmit={handleSave} className="space-y-3">
          <div>
            <label className="block text-xs text-cg-muted mb-1">Bot Token</label>
            <input
              type="password"
              value={token}
              onChange={e => setToken(e.target.value)}
              placeholder={tokenSet ? 'Update token...' : 'Enter bot token'}
              className="w-full px-3 py-2 bg-cg-bg border border-cg-border rounded-lg text-white
                         placeholder-cg-muted font-mono text-sm
                         focus:outline-none focus:border-cg-accent transition-colors"
            />
          </div>

          {error && <p className="text-cg-danger text-xs">{error}</p>}

          <div className="flex gap-2">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 py-2 text-sm border border-cg-border rounded-lg text-cg-muted
                         hover:bg-cg-border transition-colors"
            >
              Close
            </button>
            <button
              type="submit"
              disabled={saving || !token.trim()}
              className="flex-1 py-2 text-sm bg-cg-accent text-cg-bg font-semibold rounded-lg
                         hover:bg-cg-accent-dim disabled:opacity-50 transition-colors"
            >
              {saving ? 'Saving...' : 'Save'}
            </button>
          </div>
        </form>

        {/* Who Am I — shown automatically when token is configured */}
        {tokenSet && (
          <div className="mt-4 pt-4 border-t border-cg-border">
            <div className="flex items-center justify-between mb-2">
              <h3 className="text-xs font-semibold text-cg-muted uppercase tracking-wider">Your Chat IDs</h3>
              <button
                onClick={handleRefresh}
                className="text-xs text-cg-muted hover:text-cg-accent transition-colors"
              >
                Refresh
              </button>
            </div>

            {whoAmI ? (
              <>
                {whoAmI.chat_ids.length > 0 ? (
                  <div className="space-y-1.5">
                    {whoAmI.chat_ids.map(c => (
                      <div
                        key={c.chat_id}
                        className="flex items-center justify-between px-3 py-2 bg-cg-bg rounded-lg"
                      >
                        <div className="flex items-center gap-2 min-w-0">
                          <span className="font-mono text-sm text-cg-accent">{c.chat_id}</span>
                          {c.username && (
                            <span className="text-xs text-cg-muted truncate">@{c.username}</span>
                          )}
                        </div>
                        <button
                          onClick={() => navigator.clipboard.writeText(c.chat_id)}
                          className="text-xs text-cg-muted hover:text-cg-accent transition-colors flex-shrink-0 ml-2"
                        >
                          Copy
                        </button>
                      </div>
                    ))}
                  </div>
                ) : whoAmI.hint ? (
                  <p className="text-xs text-cg-accent">{whoAmI.hint}</p>
                ) : null}
              </>
            ) : (
              <p className="text-xs text-cg-muted">Loading...</p>
            )}
          </div>
        )}

        <div className="mt-4 pt-4 border-t border-cg-border text-center">
          <p className="text-xs text-cg-muted">
            Antiochus v2.0 &mdash; 512-bit cascade encryption
          </p>
        </div>
      </div>
    </div>
  )
}
