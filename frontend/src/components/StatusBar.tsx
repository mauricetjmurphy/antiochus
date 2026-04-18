interface Props {
  connected: boolean
  botName: string
  polling: boolean
}

export default function StatusBar({ connected, botName, polling }: Props) {
  return (
    <div className="flex items-center justify-between px-4 py-2 bg-cg-panel border-b border-cg-border text-xs">
      <div className="flex items-center gap-3">
        <span className="font-display font-bold text-cg-accent text-sm">Antiochus</span>
        {botName && (
          <span className="text-cg-muted">@{botName}</span>
        )}
      </div>
      <div className="flex items-center gap-3">
        {polling && (
          <span className="flex items-center gap-1 text-cg-accent">
            <span className="w-1.5 h-1.5 bg-cg-accent rounded-full animate-pulse" />
            Listening
          </span>
        )}
        <span className={`flex items-center gap-1 ${connected ? 'text-cg-accent' : 'text-cg-danger'}`}>
          <span className={`w-1.5 h-1.5 rounded-full ${connected ? 'bg-cg-accent' : 'bg-cg-danger'}`} />
          {connected ? 'WS' : 'Offline'}
        </span>
      </div>
    </div>
  )
}
