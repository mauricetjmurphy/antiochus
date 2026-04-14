import { Room, Message } from '../types'
import RoomItem from './RoomItem'

interface Props {
  rooms: Room[]
  selected: string | null
  onSelect: (name: string) => void
  onAddClick: () => void
  onSettingsClick: () => void
  getLastMessage: (room: string) => Message | undefined
}

export default function Sidebar({ rooms, selected, onSelect, onAddClick, onSettingsClick, getLastMessage }: Props) {
  return (
    <div className="w-72 bg-cg-bg border-r border-cg-border flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-3 border-b border-cg-border">
        <h2 className="font-display font-semibold text-sm text-white">Rooms</h2>
        <div className="flex gap-1">
          <button
            onClick={onAddClick}
            className="w-9 h-9 flex items-center justify-center rounded-md text-cg-muted
                       hover:bg-cg-panel hover:text-cg-accent transition-colors text-xl"
            title="Add room"
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

      <div className="flex-1 overflow-y-auto">
        {rooms.length === 0 ? (
          <div className="px-4 py-8 text-center">
            <p className="text-cg-muted text-sm mb-2">No rooms yet</p>
            <button
              onClick={onAddClick}
              className="text-cg-accent text-sm hover:underline"
            >
              Create your first room
            </button>
          </div>
        ) : (
          rooms.map(r => (
            <RoomItem
              key={r.name}
              room={r}
              selected={selected === r.name}
              lastMessage={getLastMessage(r.name)}
              onClick={() => onSelect(r.name)}
            />
          ))
        )}
      </div>
    </div>
  )
}
