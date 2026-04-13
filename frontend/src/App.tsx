import { useState, useEffect, useCallback } from 'react'
import { Friend, WSEvent } from './types'
import * as api from './api'
import { useWebSocket } from './hooks/useWebSocket'
import { useMessages } from './hooks/useMessages'
import PassphraseGate from './components/PassphraseGate'
import StatusBar from './components/StatusBar'
import Sidebar from './components/Sidebar'
import ChatPanel from './components/ChatPanel'
import AddFriendModal from './components/AddFriendModal'
import SettingsModal from './components/SettingsModal'

export default function App() {
  const [passphrase, setPassphrase] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [friends, setFriends] = useState<Friend[]>([])
  const [selectedFriend, setSelectedFriend] = useState<string | null>(null)
  const [showAddFriend, setShowAddFriend] = useState(false)
  const [showSettings, setShowSettings] = useState(false)
  const [botName, setBotName] = useState('')
  const [tokenSet, setTokenSet] = useState(false)
  const [polling, setPolling] = useState(false)

  const { addMessage, setFriendMessages, getMessages } = useMessages()

  const handleWSEvent = useCallback((event: WSEvent) => {
    if (event.type === 'new_message' && event.friend && event.message) {
      addMessage(event.friend, event.message)
    } else if (event.type === 'poll_status') {
      setPolling(!!event.active)
    }
  }, [addMessage])

  const { connected } = useWebSocket(handleWSEvent, !!passphrase)

  // Load config on mount
  useEffect(() => {
    api.getConfig().then(cfg => {
      setTokenSet(cfg.token_set)
      setBotName(cfg.bot_username)
    }).catch(() => {})
  }, [])

  // Load friends list
  const loadFriends = useCallback(async () => {
    try {
      const f = await api.getFriends()
      setFriends(f)
    } catch {
      // ignore
    }
  }, [])

  useEffect(() => {
    if (passphrase) loadFriends()
  }, [passphrase, loadFriends])

  // Load messages when selecting a friend
  useEffect(() => {
    if (selectedFriend && passphrase) {
      api.getMessages(selectedFriend).then(msgs => {
        setFriendMessages(selectedFriend, msgs)
      }).catch(() => {})
    }
  }, [selectedFriend, passphrase, setFriendMessages])

  const handleUnlock = async (pp: string) => {
    setLoading(true)
    setError('')
    try {
      const cfg = await api.getConfig()

      if (!cfg.token_set) {
        // Let user through to configure the token via Settings
        setPassphrase(pp)
        setShowSettings(true)
        return
      }

      const result = await api.startPolling(pp)
      if (result.error) {
        setError(result.error)
        return
      }
      // Only set passphrase (pass the gate) after successful handshake
      setPassphrase(pp)
      setPolling(true)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Connection failed')
    } finally {
      setLoading(false)
    }
  }

  const handleSend = async (text: string) => {
    if (!selectedFriend || !passphrase) return
    const result = await api.sendMessage(selectedFriend, text)
    if (result.error) {
      alert(result.error)
    }
  }

  const handleSendFile = async (file: File) => {
    if (!selectedFriend || !passphrase) return
    const result = await api.sendFile(selectedFriend, file)
    if (result.error) {
      alert(result.error)
    }
  }

  const handleAddFriend = async (name: string, chatId: string) => {
    await api.addFriend(name, chatId)
    await loadFriends()
  }

  const handleRemoveFriend = async (name: string) => {
    await api.removeFriend(name)
    if (selectedFriend === name) setSelectedFriend(null)
    await loadFriends()
  }

  const handleSaveToken = async (token: string) => {
    const result = await api.setConfig(token)
    if (result.error) throw new Error(result.error)
    setTokenSet(true)
    setBotName(result.bot_username)

    // Start polling if we have a passphrase but weren't polling yet
    if (passphrase && !polling) {
      const pollResult = await api.startPolling(passphrase)
      if (!pollResult.error) {
        setPolling(true)
        await loadFriends()
      }
    }
  }

  const getLastMessage = (friend: string) => {
    const msgs = getMessages(friend)
    return msgs.length > 0 ? msgs[msgs.length - 1] : undefined
  }

  const selectedFriendObj = friends.find(f => f.name === selectedFriend) ?? null

  // Passphrase gate
  if (!passphrase) {
    return <PassphraseGate onSubmit={handleUnlock} loading={loading} error={error} />
  }

  return (
    <div className="h-screen flex flex-col bg-cg-bg">
      <StatusBar connected={connected} botName={botName} polling={polling} />

      <div className="flex flex-1 overflow-hidden">
        <Sidebar
          friends={friends}
          selected={selectedFriend}
          onSelect={setSelectedFriend}
          onAddClick={() => setShowAddFriend(true)}
          onSettingsClick={() => setShowSettings(true)}
          getLastMessage={getLastMessage}
        />

        <ChatPanel
          friend={selectedFriendObj}
          messages={selectedFriend ? getMessages(selectedFriend) : []}
          onSend={handleSend}
          onSendFile={handleSendFile}
          onRemoveFriend={handleRemoveFriend}
          disabled={!tokenSet}
        />
      </div>

      <AddFriendModal
        open={showAddFriend}
        onClose={() => setShowAddFriend(false)}
        onAdd={handleAddFriend}
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
