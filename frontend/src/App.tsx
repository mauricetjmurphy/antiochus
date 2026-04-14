import { useState, useEffect, useCallback } from 'react'
import { Room, WSEvent } from './types'
import * as api from './api'
import { useWebSocket } from './hooks/useWebSocket'
import { useMessages } from './hooks/useMessages'
import PassphraseGate from './components/PassphraseGate'
import StatusBar from './components/StatusBar'
import Sidebar from './components/Sidebar'
import ChatPanel from './components/ChatPanel'
import AddRoomModal from './components/AddRoomModal'
import SettingsModal from './components/SettingsModal'

export default function App() {
  const [passphrase, setPassphrase] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [rooms, setRooms] = useState<Room[]>([])
  const [selectedRoom, setSelectedRoom] = useState<string | null>(null)
  const [showAddRoom, setShowAddRoom] = useState(false)
  const [showSettings, setShowSettings] = useState(false)
  const [botName, setBotName] = useState('')
  const [tokenSet, setTokenSet] = useState(false)
  const [polling, setPolling] = useState(false)

  const { addMessage, setRoomMessages, getMessages } = useMessages()

  const handleWSEvent = useCallback((event: WSEvent) => {
    if (event.type === 'new_message' && event.room && event.message) {
      addMessage(event.room, event.message)
    } else if (event.type === 'poll_status') {
      setPolling(!!event.active)
    }
  }, [addMessage])

  const { connected } = useWebSocket(handleWSEvent, !!passphrase)

  useEffect(() => {
    api.getConfig().then(cfg => {
      setTokenSet(cfg.token_set)
      setBotName(cfg.bot_username)
    }).catch(() => {})
  }, [])

  const loadRooms = useCallback(async () => {
    try {
      const r = await api.getRooms()
      setRooms(r)
    } catch {
      // ignore
    }
  }, [])

  useEffect(() => {
    if (passphrase) loadRooms()
  }, [passphrase, loadRooms])

  useEffect(() => {
    if (selectedRoom && passphrase) {
      api.getMessages(selectedRoom).then(msgs => {
        setRoomMessages(selectedRoom, msgs)
      }).catch(() => {})
    }
  }, [selectedRoom, passphrase, setRoomMessages])

  const handleUnlock = async (pp: string) => {
    setLoading(true)
    setError('')
    try {
      const cfg = await api.getConfig()

      if (!cfg.token_set) {
        setPassphrase(pp)
        setShowSettings(true)
        return
      }

      const result = await api.startPolling(pp)
      if (result.error) {
        setError(result.error)
        return
      }
      setPassphrase(pp)
      setPolling(true)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Connection failed')
    } finally {
      setLoading(false)
    }
  }

  const handleSend = async (text: string) => {
    if (!selectedRoom || !passphrase) return
    const result = await api.sendMessage(selectedRoom, text)
    if (result.error) {
      alert(result.error)
    }
  }

  const handleRefresh = async (): Promise<number> => {
    if (!selectedRoom) return 0
    const before = getMessages(selectedRoom).length
    const msgs = await api.getMessages(selectedRoom)
    setRoomMessages(selectedRoom, msgs)
    return Math.max(0, msgs.length - before)
  }

  const handleSendFile = async (file: File) => {
    if (!selectedRoom || !passphrase) return
    const result = await api.sendFile(selectedRoom, file)
    if (result.error) {
      alert(result.error)
    }
  }

  const handleRemoveRoom = async (name: string) => {
    await api.removeRoom(name)
    if (selectedRoom === name) setSelectedRoom(null)
    await loadRooms()
  }

  const handleSaveToken = async (token: string) => {
    const result = await api.setConfig(token)
    if (result.error) throw new Error(result.error)
    setTokenSet(true)
    setBotName(result.bot_username)

    if (passphrase && !polling) {
      const pollResult = await api.startPolling(passphrase)
      if (!pollResult.error) {
        setPolling(true)
        await loadRooms()
      }
    }
  }

  const getLastMessage = (room: string) => {
    const msgs = getMessages(room)
    return msgs.length > 0 ? msgs[msgs.length - 1] : undefined
  }

  const selectedRoomObj = rooms.find(r => r.name === selectedRoom) ?? null

  if (!passphrase) {
    return <PassphraseGate onSubmit={handleUnlock} loading={loading} error={error} />
  }

  return (
    <div className="h-screen flex flex-col bg-cg-bg">
      <StatusBar connected={connected} botName={botName} polling={polling} />

      <div className="flex flex-1 overflow-hidden">
        <Sidebar
          rooms={rooms}
          selected={selectedRoom}
          onSelect={setSelectedRoom}
          onAddClick={() => setShowAddRoom(true)}
          onSettingsClick={() => setShowSettings(true)}
          getLastMessage={getLastMessage}
        />

        <ChatPanel
          room={selectedRoomObj}
          messages={selectedRoom ? getMessages(selectedRoom) : []}
          onSend={handleSend}
          onSendFile={handleSendFile}
          onRemoveRoom={handleRemoveRoom}
          onRefresh={handleRefresh}
          disabled={!tokenSet}
        />
      </div>

      <AddRoomModal
        open={showAddRoom}
        botUsername={botName}
        onClose={() => setShowAddRoom(false)}
        onAdded={loadRooms}
      />

      <SettingsModal
        open={showSettings}
        onClose={() => setShowSettings(false)}
        onSaveToken={handleSaveToken}
        botName={botName}
        tokenSet={tokenSet}
      />
    </div>
  )
}
