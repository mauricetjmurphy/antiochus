# Antiochus Setup Guide

Antiochus turns a Telegram group into an encrypted chat room. Messages are encrypted before they leave your machine and only decrypted by people who share the passphrase. Telegram never sees the plaintext.

## What You Need

- The Antiochus binary (each participant needs a copy)
- A Telegram account
- A shared passphrase, agreed on out of band (in person, phone call, trusted channel)

## How It Works

Each participant runs their own Telegram bot. Both bots live in the same Telegram group. When you send a message from the Antiochus app, your bot posts an encrypted blob to the group. The other participant's bot — also in the group — sees that post via `getUpdates` and their app decrypts it.

Why each user needs their own bot: Telegram only allows one `getUpdates` consumer per bot token, and bot-sent messages never appear in the sender's own `getUpdates`. A shared bot cannot work for two pollers.

## Step 1: Create Your Bot

1. Open Telegram and message **@BotFather**
2. Send `/newbot`, pick a name and a `_bot`-suffixed username
3. Copy the **bot token** BotFather gives you

## Step 2: Configure Your Bot in BotFather (critical)

Two settings must be toggled before your bot will work as a relay. Message **@BotFather** and run both commands:

**1. Disable privacy mode**
By default a bot in a group only sees messages that @mention it. For Antiochus the bot must see every message.
- Send `/setprivacy` → pick your bot → **Disable**

**2. Enable Bot-to-Bot mode**
By default a bot cannot see messages from other bots in a group. Antiochus relies on the other participant's bot posting encrypted blobs, so your bot needs to be able to see them.
- Send `/setbot2bot` → pick your bot → **Enable**

**Re-add the bot to any group after changing these settings.** Both options only take effect on fresh group memberships — if the bot is already in a group, remove it and add it back.

## Step 3: Configure the App

1. Run the binary: `./antiochus`
2. Open http://127.0.0.1:8080
3. Enter a passphrase (the same passphrase every participant will use) and unlock
4. Click the gear icon (Settings) and paste your bot token

## Step 4: Create a Room

1. Click **+** in the sidebar to open the New Room dialog
2. Click **Open Telegram →**. Telegram opens and asks which group to add your bot to — pick an existing group or create a new one
3. Invite the other participant to the group (via Telegram's invite link)
4. Ask them to:
   - Create their own bot (Step 1)
   - Disable privacy mode on it (Step 2)
   - Add their bot to the same group
5. Send any message in the group. The group appears under **Detected Groups** in the Antiochus dialog
6. Click the detected group, give it a local name (e.g. `project-alpha`), and save

Each participant repeats step 6 on their own machine to add the same room.

## Step 5: Chat

1. Select the room in the sidebar
2. Type a message and hit Enter

The message is encrypted, posted to the group by your bot, seen by the other bot, and decrypted on the other side. Click **Refresh** on the chat header to pull new messages.

### Sending files

Click the paperclip icon next to the message input to send an encrypted file (up to 45 MB). The file arrives in the group as `antiochus.enc` — unreadable without the passphrase. The recipient's app decrypts it and saves it to `~/.antiochus/received/`.

## What Each Party Sees

| Where | What you see |
|-------|-------------|
| Your browser | Plaintext messages and files |
| Telegram group | `🔐 Antiochus` + encrypted gibberish, or `antiochus.enc` files |
| Other participants' browsers | Plaintext messages and files (with matching passphrase) |
| Anyone else in the group | Encrypted gibberish only |

## Troubleshooting

**"Bot token not configured"**
→ Open Settings (gear icon) and paste your bot token.

**No groups detected in the New Room dialog**
→ Post a message in the group after both bots have been added. The app pulls updates automatically every couple of seconds while the dialog is open.

**Messages from other participants not appearing**
- Make sure every participant uses the exact same passphrase. One character off = silent decryption failure.
- On every participant's bot, both `/setprivacy` (Disable) and `/setbot2bot` (Enable) must be set in @BotFather.
- Each bot must be removed and re-added to the group after toggling those settings.
- Check that both bots are still members of the group.

**File too large**
→ Max file size is 45 MB (Telegram's hard limit is 50 MB — we leave margin for encryption overhead). Configurable in `prod.yml` under `telegram.max_file_size_mb`.
