import { Friend, Message } from '../types'

interface Props {
  friend: Friend
  selected: boolean
  lastMessage?: Message
  onClick: () => void
}

export default function FriendItem({ friend, selected, lastMessage, onClick }: Props) {
  return (
    <button
      onClick={onClick}
      className={`w-full text-left px-4 py-3 flex items-center gap-3 transition-colors
        ${selected
          ? 'bg-cg-accent/10 border-l-2 border-cg-accent'
          : 'hover:bg-cg-panel/50 border-l-2 border-transparent'
        }`}
    >
      <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold
        ${selected ? 'bg-cg-accent text-cg-bg' : 'bg-cg-border text-cg-muted'}`}>
        {friend.name[0].toUpperCase()}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between">
          <span className={`font-medium text-sm truncate ${selected ? 'text-cg-accent' : 'text-white'}`}>
            {friend.name}
          </span>
          {lastMessage && (
            <span className="text-xs text-cg-muted ml-2 flex-shrink-0">
              {lastMessage.time}
            </span>
          )}
        </div>
        {lastMessage && (
          <p className="text-xs text-cg-muted truncate mt-0.5">
            {lastMessage.sender === 'you' ? 'You: ' : ''}{lastMessage.content}
          </p>
        )}
      </div>
    </button>
  )
}
