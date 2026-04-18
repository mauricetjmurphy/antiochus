import { useEffect, useState } from 'react'
import * as api from '../api'
import { RoomCandidate } from '../types'

interface Props {
  open: boolean
  botUsername: string
  onClose: () => void
  onAdded: () => Promise<void>
}

export default function AddRoomModal({ open, botUsername, onClose, onAdded }: Props) {
  const [candidates, setCandidates] = useState<RoomCandidate[]>([])
  const [picked, setPicked] = useState<RoomCandidate | null>(null)
  const [name, setName] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [polling, setPolling] = useState(false)
  const [manualOpen, setManualOpen] = useState(false)
  const [manualName, setManualName] = useState('')
  const [manualChatID, setManualChatID] = useState('')
  const [manualTitle, setManualTitle] = useState('')

  useEffect(() => {
    if (!open) return
    setCandidates([])
    setPicked(null)
    setName('')
    setError('')
    setManualOpen(false)
    setManualName('')
    setManualChatID('')
    setManualTitle('')
  }, [open])

  // Poll for candidates every 2s while modal is open and no candidate picked
  useEffect(() => {
    if (!open || picked) return

    let cancelled = false
    const tick = async () => {
      try {
        setPolling(true)
        const cs = await api.getRoomCandidates()
        if (!cancelled) setCandidates(cs)
      } catch {
        // ignore
      } finally {
        if (!cancelled) setPolling(false)
      }
    }
    tick()
    const id = window.setInterval(tick, 2500)
    return () => {
      cancelled = true
      window.clearInterval(id)
    }
  }, [open, picked])

  if (!open) return null

  const groupLink = botUsername
    ? `https://t.me/${botUsername}?startgroup=true`
    : ''

  const handleSave = async () => {
    if (!picked || !name.trim()) return
    setSubmitting(true)
    setError('')
    try {
      const res = await api.addRoom(name.trim().toLowerCase(), picked.chat_id, picked.title)
      if (res.error) throw new Error(res.error)
      await onAdded()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add room')
    } finally {
      setSubmitting(false)
    }
  }

  const handleManualSave = async () => {
    const n = manualName.trim().toLowerCase()
    const cid = manualChatID.trim()
    const title = manualTitle.trim()
    if (!n || !cid) return
    setSubmitting(true)
    setError('')
    try {
      const res = await api.addRoom(n, cid, title)
      if (res.error) throw new Error(res.error)
      await onAdded()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add room')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60" onClick={onClose}>
      <div
        className="bg-cg-panel border border-cg-border rounded-xl p-6 w-full max-w-md max-h-[90vh] overflow-y-auto"
        onClick={e => e.stopPropagation()}
      >
        <h2 className="font-display font-semibold text-lg mb-4">New Room</h2>

        {!picked && (
          <>
            <ol className="space-y-3 text-sm text-cg-muted mb-4">
              <li>
                <span className="text-white font-semibold">1.</span> Create a Telegram group with your bot:
                {groupLink ? (
                  <a
                    href={groupLink}
                    target="_blank"
                    rel="noreferrer"
                    className="ml-2 inline-block px-2 py-1 bg-cg-accent text-cg-bg rounded text-xs font-semibold hover:bg-cg-accent-dim"
                  >
                    Open Telegram →
                  </a>
                ) : (
                  <span className="ml-2 text-cg-danger text-xs">Configure bot token first</span>
                )}
              </li>
              <li>
                <span className="text-white font-semibold">2.</span> Invite your friend to the group and have them add <em>their</em> own bot too.
              </li>
              <li>
                <span className="text-white font-semibold">3.</span> In{' '}
                <a href="https://t.me/BotFather" target="_blank" rel="noreferrer" className="text-cg-accent hover:underline">
                  @BotFather
                </a>{' '}
                on <em>both</em> bots:
                <ul className="mt-1 ml-4 list-disc space-y-0.5 text-xs">
                  <li><code>/setprivacy</code> → Disable (lets the bot see all group messages)</li>
                  <li><code>/setbot2bot</code> → Enable (lets the bot see the other bot's messages)</li>
                </ul>
                Then remove and re-add each bot to the group — settings only apply to fresh memberships.
              </li>
              <li>
                <span className="text-white font-semibold">4.</span> Send any message in the group. It'll appear below.
              </li>
            </ol>

            <div className="border border-cg-border rounded-lg p-3 bg-cg-bg">
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs text-cg-muted font-semibold uppercase tracking-wide">Detected Groups</span>
                {polling && <span className="text-xs text-cg-muted">Listening…</span>}
              </div>
              {candidates.length === 0 ? (
                <p className="text-xs text-cg-muted py-2">Waiting for a message in a group…</p>
              ) : (
                <ul className="space-y-1">
                  {candidates.map(c => (
                    <li key={c.chat_id}>
                      <button
                        onClick={() => { setPicked(c); setName(c.title.toLowerCase().replace(/\s+/g, '-')) }}
                        className="w-full text-left px-2 py-2 rounded hover:bg-cg-panel transition-colors"
                      >
                        <div className="text-sm text-white">{c.title || '(untitled group)'}</div>
                        <div className="text-xs text-cg-muted font-mono">{c.chat_id} · {c.type}</div>
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>

            <div className="mt-4 border-t border-cg-border pt-3">
              <button
                onClick={() => setManualOpen(o => !o)}
                className="text-xs text-cg-muted hover:text-cg-accent transition-colors"
              >
                {manualOpen ? '▾ Hide manual entry' : '▸ Enter chat ID manually'}
              </button>

              {manualOpen && (
                <div className="mt-3 space-y-2">
                  <p className="text-xs text-cg-muted">
                    If detection isn't working, paste the group's chat_id directly. Get it from the server log
                    (<code>[poll] incoming message: ... chatID=...</code>) or from @userinfobot in the group.
                  </p>
                  <div>
                    <label className="block text-xs text-cg-muted mb-1">Chat ID</label>
                    <input
                      type="text"
                      value={manualChatID}
                      onChange={e => setManualChatID(e.target.value)}
                      placeholder="-1001234567890"
                      className="w-full px-3 py-2 bg-cg-bg border border-cg-border rounded-lg text-white
                                 placeholder-cg-muted font-mono text-sm
                                 focus:outline-none focus:border-cg-accent transition-colors"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-cg-muted mb-1">Title (optional)</label>
                    <input
                      type="text"
                      value={manualTitle}
                      onChange={e => setManualTitle(e.target.value)}
                      placeholder="Antiochus Room"
                      className="w-full px-3 py-2 bg-cg-bg border border-cg-border rounded-lg text-white
                                 placeholder-cg-muted text-sm
                                 focus:outline-none focus:border-cg-accent transition-colors"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-cg-muted mb-1">Local name</label>
                    <input
                      type="text"
                      value={manualName}
                      onChange={e => setManualName(e.target.value)}
                      placeholder="project-alpha"
                      className="w-full px-3 py-2 bg-cg-bg border border-cg-border rounded-lg text-white
                                 placeholder-cg-muted font-mono text-sm
                                 focus:outline-none focus:border-cg-accent transition-colors"
                    />
                  </div>
                  <button
                    onClick={handleManualSave}
                    disabled={submitting || !manualName.trim() || !manualChatID.trim()}
                    className="w-full py-2 text-sm bg-cg-accent text-cg-bg font-semibold rounded-lg
                               hover:bg-cg-accent-dim disabled:opacity-50 transition-colors"
                  >
                    {submitting ? 'Saving…' : 'Save Room'}
                  </button>
                </div>
              )}
            </div>

            {error && <p className="text-cg-danger text-xs mt-2">{error}</p>}

            <div className="flex justify-end mt-4">
              <button
                onClick={onClose}
                className="px-4 py-2 text-sm border border-cg-border rounded-lg text-cg-muted
                           hover:bg-cg-border transition-colors"
              >
                Cancel
              </button>
            </div>
          </>
        )}

        {picked && (
          <>
            <div className="mb-4 p-3 bg-cg-bg border border-cg-border rounded-lg">
              <div className="text-xs text-cg-muted uppercase tracking-wide mb-1">Selected</div>
              <div className="text-white text-sm">{picked.title || '(untitled group)'}</div>
              <div className="text-xs text-cg-muted font-mono">{picked.chat_id}</div>
            </div>

            <div className="mb-3">
              <label className="block text-xs text-cg-muted mb-1">Local name</label>
              <input
                type="text"
                value={name}
                onChange={e => setName(e.target.value)}
                placeholder="project-alpha"
                autoFocus
                className="w-full px-3 py-2 bg-cg-bg border border-cg-border rounded-lg text-white
                           placeholder-cg-muted font-mono text-sm
                           focus:outline-none focus:border-cg-accent transition-colors"
              />
              <p className="text-xs text-cg-muted mt-1">Lowercase, no spaces. Shown in your sidebar.</p>
            </div>

            {error && <p className="text-cg-danger text-xs mb-2">{error}</p>}

            <div className="flex gap-2">
              <button
                onClick={() => setPicked(null)}
                className="flex-1 py-2 text-sm border border-cg-border rounded-lg text-cg-muted
                           hover:bg-cg-border transition-colors"
              >
                Back
              </button>
              <button
                onClick={handleSave}
                disabled={submitting || !name.trim()}
                className="flex-1 py-2 text-sm bg-cg-accent text-cg-bg font-semibold rounded-lg
                           hover:bg-cg-accent-dim disabled:opacity-50 transition-colors"
              >
                {submitting ? 'Saving…' : 'Save Room'}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
