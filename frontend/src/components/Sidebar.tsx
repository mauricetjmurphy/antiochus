import { Friend, Message } from '../types'
import FriendItem from './FriendItem'

interface Props {
  friends: Friend[]
  selected: string | null
  onSelect: (name: string) => void
  onAddClick: () => void
  onSettingsClick: () => void
  getLastMessage: (friend: string) => Message | undefined
}

export default function Sidebar({ friends, selected, onSelect, onAddClick, onSettingsClick, getLastMessage }: Props) {
  return (
    <div className="w-72 bg-cg-bg border-r border-cg-border flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-cg-border">
        <h2 className="font-display font-semibold text-sm text-white">Friends</h2>
        <div className="flex gap-1">
          <button
            onClick={onAddClick}
            className="w-9 h-9 flex items-center justify-center rounded-md text-cg-muted
                       hover:bg-cg-panel hover:text-cg-accent transition-colors text-xl"
            title="Add friend"
          >
            +
          </button>
          <button
            onClick={onSettingsClick}
            className="w-9 h-9 flex items-center justify-center rounded-md text-cg-muted
                       hover:bg-cg-panel hover:text-cg-accent transition-colors text-lg"
            title="Settings"
          >
            &#9881;
          </button>
        </div>
      </div>

      {/* Friend list */}
      <div className="flex-1 overflow-y-auto">
        {friends.length === 0 ? (
          <div className="px-4 py-8 text-center">
            <p className="text-cg-muted text-sm mb-2">No friends yet</p>
            <button
              onClick={onAddClick}
              className="text-cg-accent text-sm hover:underline"
            >
              Add your first friend
            </button>
          </div>
        ) : (
          friends.map(f => (
            <FriendItem
              key={f.name}
              friend={f}
              selected={selected === f.name}
              lastMessage={getLastMessage(f.name)}
              onClick={() => onSelect(f.name)}
            />
          ))
        )}
      </div>
    </div>
  )
}
