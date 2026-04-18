import { Message } from '../types'

interface Props {
  message: Message
}

function basename(path: string): string {
  const parts = path.split(/[\\/]/)
  return parts[parts.length - 1] || path
}

export default function MessageBubble({ message }: Props) {
  const isMe = message.sender === 'you'
  const isFile = message.type === 'file'
  const downloadName = message.filePath ? basename(message.filePath) : ''
  const downloadable = isFile && !isMe && downloadName.length > 0

  return (
    <div className={`flex ${isMe ? 'justify-end' : 'justify-start'} mb-2`}>
      <div className={`max-w-[70%] rounded-xl px-4 py-2 ${
        isMe
          ? 'bg-cg-accent/15 border border-cg-accent/20'
          : 'bg-cg-panel border border-cg-border'
      }`}>
        {!isMe && (
          <p className="text-xs text-cg-accent font-semibold mb-1">
            {message.sender}
          </p>
        )}
        {downloadable ? (
          <a
            href={`/api/files/${encodeURIComponent(downloadName)}`}
            download={message.filename || downloadName}
            className="text-sm break-words text-cg-accent underline hover:no-underline"
          >
            {message.content}
          </a>
        ) : (
          <p className={`text-sm break-words ${isFile ? 'text-cg-accent' : 'text-white'}`}>
            {message.content}
          </p>
        )}
        <p className={`text-xs mt-1 ${isMe ? 'text-cg-accent/50' : 'text-cg-muted'}`}>
          {message.time}
        </p>
      </div>
    </div>
  )
}
