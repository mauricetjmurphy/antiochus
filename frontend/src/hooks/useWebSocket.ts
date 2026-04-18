import { useEffect, useRef, useState, useCallback } from 'react'
import { WSEvent } from '../types'
import { decryptWSMessage, getSessionKey } from '../api'

export function useWebSocket(onEvent: (event: WSEvent) => void, enabled: boolean) {
  const [connected, setConnected] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)
  const onEventRef = useRef(onEvent)
  onEventRef.current = onEvent

  const connect = useCallback(() => {
    const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const ws = new WebSocket(`${proto}//${window.location.host}/ws`)

    ws.onopen = () => setConnected(true)

    ws.onmessage = async (e) => {
      try {
        const raw = JSON.parse(e.data)

        if (raw.encrypted && getSessionKey()) {
          const decrypted = await decryptWSMessage(raw.encrypted)
          const event: WSEvent = JSON.parse(decrypted)
          onEventRef.current(event)
        } else {
          const event: WSEvent = raw
          onEventRef.current(event)
        }
      } catch {
        // ignore malformed events
      }
    }

    ws.onclose = () => {
      setConnected(false)
      // Only reconnect if still enabled
      setTimeout(() => {
        if (wsRef.current === ws) {
          connect()
        }
      }, 2000)
    }

    ws.onerror = () => ws.close()

    wsRef.current = ws
  }, [])

  useEffect(() => {
    if (!enabled) {
      // Close any existing connection when disabled
      if (wsRef.current) {
        const ws = wsRef.current
        wsRef.current = null
        ws.close()
      }
      setConnected(false)
      return
    }

    connect()
    return () => {
      const ws = wsRef.current
      wsRef.current = null
      ws?.close()
    }
  }, [enabled, connect])

  return { connected }
}
