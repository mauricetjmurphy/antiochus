import { Room, RoomCandidate, Message } from './types'

const BASE = ''

// ── Helpers ────────────────────────────────────────────────────

function toBase64(buf: ArrayBuffer): string {
  return btoa(String.fromCharCode(...new Uint8Array(buf)))
}

function fromBase64(b64: string): Uint8Array {
  return Uint8Array.from(atob(b64), c => c.charCodeAt(0))
}

// ── Session Key (ECDH shared secret) ───────────────────────────

let sessionKey: CryptoKey | null = null

export function getSessionKey(): CryptoKey | null {
  return sessionKey
}

// ── ECDH Key Exchange ──────────────────────────────────────────

async function encryptPassphrase(passphrase: string): Promise<{
  client_public_key: string
  encrypted_passphrase: string
  iv: string
}> {
  const keyRes = await fetch(`${BASE}/api/session/key`)
  const { public_key: serverPubB64 } = await keyRes.json()

  const clientKeyPair = await crypto.subtle.generateKey(
    { name: 'ECDH', namedCurve: 'P-256' },
    true,
    ['deriveKey']
  )

  const serverPubBytes = fromBase64(serverPubB64)
  const serverPubKey = await crypto.subtle.importKey(
    'raw',
    serverPubBytes.buffer as ArrayBuffer,
    { name: 'ECDH', namedCurve: 'P-256' },
    false,
    []
  )

  const sharedKey = await crypto.subtle.deriveKey(
    { name: 'ECDH', public: serverPubKey },
    clientKeyPair.privateKey,
    { name: 'AES-GCM', length: 256 },
    false,
    ['encrypt', 'decrypt']
  )

  sessionKey = sharedKey

  const iv = crypto.getRandomValues(new Uint8Array(12))
  const encoded = new TextEncoder().encode(passphrase)
  const encrypted = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv },
    sharedKey,
    encoded
  )

  const clientPubRaw = await crypto.subtle.exportKey('raw', clientKeyPair.publicKey)

  return {
    client_public_key: toBase64(clientPubRaw as ArrayBuffer),
    encrypted_passphrase: toBase64(encrypted),
    iv: toBase64(iv.buffer as ArrayBuffer),
  }
}

export async function decryptWSMessage(encryptedPayload: string): Promise<string> {
  if (!sessionKey) {
    throw new Error('No session key')
  }

  const parts = encryptedPayload.split('.')
  if (parts.length !== 2) {
    throw new Error('Invalid encrypted payload format')
  }

  const iv = fromBase64(parts[0])
  const ciphertext = fromBase64(parts[1])

  const plaintext = await crypto.subtle.decrypt(
    { name: 'AES-GCM', iv: iv.buffer as ArrayBuffer },
    sessionKey,
    ciphertext.buffer as ArrayBuffer
  )

  return new TextDecoder().decode(plaintext)
}

// ── API Functions ──────────────────────────────────────────────

export async function getConfig(): Promise<{ token_set: boolean; bot_username: string }> {
  const res = await fetch(`${BASE}/api/config`)
  return res.json()
}

export async function setConfig(token: string): Promise<{ ok: boolean; bot_username: string; error?: string }> {
  const res = await fetch(`${BASE}/api/config`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token }),
  })
  return res.json()
}

export async function getRooms(): Promise<Room[]> {
  const res = await fetch(`${BASE}/api/rooms`)
  if (!res.ok) throw new Error('Failed to load rooms')
  return res.json()
}

export async function getRoomCandidates(): Promise<RoomCandidate[]> {
  const res = await fetch(`${BASE}/api/rooms/candidates`)
  if (!res.ok) throw new Error('Failed to load room candidates')
  return res.json()
}

export async function addRoom(name: string, chatId: string, title: string): Promise<{ ok?: string; error?: string }> {
  const res = await fetch(`${BASE}/api/rooms`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, chat_id: chatId, title }),
  })
  return res.json()
}

export async function removeRoom(name: string): Promise<{ ok?: string; error?: string }> {
  const res = await fetch(`${BASE}/api/rooms/${name}`, { method: 'DELETE' })
  return res.json()
}

export async function sendMessage(to: string, text: string): Promise<{ ok?: boolean; error?: string }> {
  const res = await fetch(`${BASE}/api/messages/send`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ to, text }),
  })
  return res.json()
}

export async function sendFile(to: string, file: File): Promise<{ ok?: boolean; error?: string }> {
  const form = new FormData()
  form.append('to', to)
  form.append('file', file)
  const res = await fetch(`${BASE}/api/messages/send-file`, { method: 'POST', body: form })
  return res.json()
}

export async function getMessages(room: string): Promise<Message[]> {
  const res = await fetch(`${BASE}/api/messages/${room}`)
  if (!res.ok) throw new Error('Failed to load messages')
  return res.json()
}

export async function startPolling(passphrase: string): Promise<{ ok?: string; error?: string }> {
  const payload = await encryptPassphrase(passphrase)
  const res = await fetch(`${BASE}/api/poll/start`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })
  return res.json()
}

export async function stopPolling(): Promise<void> {
  await fetch(`${BASE}/api/poll/stop`, { method: 'POST' })
  sessionKey = null
}

export async function fetchUpdates(): Promise<{ processed?: number; error?: string }> {
  const res = await fetch(`${BASE}/api/poll/fetch`, { method: 'POST' })
  return res.json()
}

export async function whoAmI(): Promise<{
  bot_username: string
  bot_id: string
  chat_ids: { chat_id: string; title: string; type: string; username: string; source: string }[]
  hint: string
  error?: string
}> {
  const res = await fetch(`${BASE}/api/whoami`)
  return res.json()
}
