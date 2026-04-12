import { useState } from 'react'

interface Props {
  onSubmit: (passphrase: string) => void
  loading: boolean
  error: string
}

export default function PassphraseGate({ onSubmit, loading, error }: Props) {
  const [passphrase, setPassphrase] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (passphrase.trim()) {
      onSubmit(passphrase)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-cg-bg">
      <div className="w-full max-w-md p-8">
        <div className="text-center mb-8">
          <h1 className="font-display text-3xl font-bold text-cg-accent mb-2">
            Antiochus
          </h1>
          <p className="text-cg-muted text-sm">
            AES-256-GCM + ChaCha20-Poly1305
          </p>
          <p className="text-cg-muted text-xs mt-1">
            512-bit cascade encryption
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm text-cg-muted mb-2">
              Enter your passphrase to start
            </label>
            <input
              type="password"
              value={passphrase}
              onChange={e => setPassphrase(e.target.value)}
              placeholder="Passphrase"
              autoFocus
              className="w-full px-4 py-3 bg-cg-panel border border-cg-border rounded-lg
                         text-white placeholder-cg-muted font-mono
                         focus:outline-none focus:border-cg-accent focus:ring-1 focus:ring-cg-accent
                         transition-colors"
            />
          </div>

          {error && (
            <p className="text-cg-danger text-sm">{error}</p>
          )}

          <button
            type="submit"
            disabled={loading || !passphrase.trim()}
            className="w-full py-3 bg-cg-accent text-cg-bg font-display font-semibold rounded-lg
                       hover:bg-cg-accent-dim disabled:opacity-50 disabled:cursor-not-allowed
                       transition-colors"
          >
            {loading ? 'Connecting...' : 'Unlock'}
          </button>
        </form>

        <p className="text-center text-cg-muted text-xs mt-6">
          Your passphrase is never stored on disk
        </p>
        <p className="text-center mt-3">
          <a
            href="/guide"
            target="_blank"
            rel="noopener noreferrer"
            className="text-cg-accent text-xs hover:underline"
          >
            Setup Guide
          </a>
        </p>
      </div>
    </div>
  )
}
