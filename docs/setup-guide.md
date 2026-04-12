# Antiochus Setup Guide

This guide walks you through setting up Antiochus for encrypted messaging between two people.

## What You Need

- The Antiochus binary (both you and your friend need a copy)
- A Telegram account (both of you)
- A shared passphrase (agreed on in person, phone call, or another secure channel)

## Overview

Antiochus uses Telegram as a transport layer. Messages are encrypted before they leave your machine and only decrypted on the other end. Telegram never sees the plaintext.

The setup has three parts:

1. Create a Telegram bot (the delivery mechanism)
2. Get your chat ID (your "address" on Telegram)
3. Exchange info with your friend and start chatting

## Step 1: Create a Telegram Bot

You need a Telegram bot to send and receive messages. Think of it as a private mail carrier — it moves encrypted envelopes between you and your friend.

1. Open Telegram and search for **@BotFather**
2. Send `/newbot`
3. Choose a name (e.g. "My Cipher Bot")
4. Choose a username (e.g. "my_cipher_1234_bot" — must end in `bot`)
5. BotFather gives you a **bot token** — a long string like `123456789:ABCdefGHIjklMNOpqrsTUVwxyz1234567890`
6. Save this token

**Important:** Both you and your friend need to **message this bot** on Telegram (just send "hi"). Telegram bots can't initiate conversations — a user must message the bot first before the bot can send messages to them.

You can either:
- **Share one bot** — you create the bot, give your friend the username, they message it too
- **Use two bots** — you each create your own bot. You message their bot, they message yours

One shared bot is simpler.

## Step 2: Get Your Chat ID

Your chat ID is your address on Telegram. Your friend needs it to send encrypted messages to you (and you need theirs).

### Option A: Use Antiochus

1. Run the binary: `./antiochus`
2. Open http://127.0.0.1:8080
3. Enter any passphrase and unlock
4. Click the gear icon (Settings)
5. Paste your bot token and save
6. Your chat ID appears automatically under "Your Chat IDs"
7. If nothing shows up, go to Telegram, send any message to your bot, then click **Refresh**

### Option B: Use @userinfobot

1. Open Telegram and search for **@userinfobot**
2. Send it any message
3. It replies with your chat ID

## Step 3: Exchange Info With Your Friend

You and your friend need to share three things **securely** (in person, phone call, or another trusted channel — not over unencrypted text):

| What | Example | Who needs it |
|------|---------|-------------|
| Bot token | `123456789:ABCdef...7890` | Both of you (if sharing one bot) |
| Your chat ID | `523841967` | Your friend adds this |
| Their chat ID | `891234567` | You add this |
| Passphrase | `correct horse battery staple` | Both of you (must be identical) |

**The passphrase must be exactly the same on both sides.** It's the shared secret that derives the encryption keys. If it's different by even one character, decryption fails.

## Step 4: Configure Antiochus

### Your machine

1. Run `./antiochus` and open http://127.0.0.1:8080
2. Enter the shared passphrase
3. Settings → paste the bot token → Save
4. Click **+** (Add Friend) → name: `alice`, chat ID: `891234567` (your friend's ID)

### Your friend's machine

1. They run `./antiochus` and open http://127.0.0.1:8080
2. They enter the **same** passphrase
3. Settings → paste the same bot token → Save
4. They click **+** → name: `you`, chat ID: `523841967` (your ID)

### Alternative: Edit the config file directly

Both of you can also edit `~/.ciphergram/ciphergram.yml` (or `internal/config/prod.yml`):

```yaml
telegram:
  bot_token: "YOUR_BOT_TOKEN_HERE"

friends:
  alice: "891234567"
```

## Step 5: Start Chatting

1. Select your friend in the sidebar
2. Type a message and hit Enter
3. The message is encrypted and sent through Telegram
4. Your friend sees it decrypted in their browser in real time

### Sending files

Click the paperclip icon next to the message input to send an encrypted file (up to 45 MB). The file arrives in Telegram as `ciphergram.enc` — unreadable without the passphrase. Your friend's app decrypts it and saves it to `~/.ciphergram/received/`.

## What Each Party Sees

| Where | What you see |
|-------|-------------|
| Your browser | Plaintext messages and files |
| Telegram chat | `🔐 CipherGram` + encrypted gibberish, or `ciphergram.enc` files |
| Your friend's browser | Plaintext messages and files |
| Anyone else looking at Telegram | Encrypted gibberish only |

## Troubleshooting

**"Bot token not configured"**
→ Open Settings (gear icon) and paste your bot token.

**Messages not appearing**
→ Make sure both sides are using the exact same passphrase. Even one character difference means decryption fails silently.

**"Send any message to @yourbot on Telegram"**
→ You (or your friend) haven't messaged the bot yet. Go to Telegram, find the bot by username, and send it any message. Then refresh.

**Friend's messages not arriving**
→ Check that the chat ID in your friends list matches their actual Telegram chat ID. They can verify theirs in Settings → Who Am I.

**File too large**
→ Max file size is 45 MB (Telegram's limit is 50 MB, we leave margin for encryption overhead). This is configurable in `ciphergram.yml` under `telegram.max_file_size_mb`.
