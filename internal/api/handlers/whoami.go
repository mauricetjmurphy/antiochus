package handlers

import (
	"fmt"
	"net/http"

	"antiochus/internal/telegram"
)

func (h *Handler) WhoAmI(w http.ResponseWriter, r *http.Request) {
	if h.Cfg.Telegram.BotToken == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "bot token not configured"})
		return
	}

	client := telegram.NewClient(h.Cfg.Telegram.BotToken)

	me, err := client.GetMe()
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "getMe: " + err.Error()})
		return
	}

	// Collect chat IDs from multiple sources:
	// 1. The config's default chat_id
	// 2. All friends' chat IDs
	// 3. A short-poll of recent Telegram updates (offset=-1 gets the last update only)
	seen := make(map[string]bool)
	type chatInfo struct {
		ChatID   string `json:"chat_id"`
		Username string `json:"username"`
		Source   string `json:"source"`
	}
	chats := make([]chatInfo, 0)

	// From config default
	if h.Cfg.Telegram.ChatID != "" && !seen[h.Cfg.Telegram.ChatID] {
		seen[h.Cfg.Telegram.ChatID] = true
		chats = append(chats, chatInfo{
			ChatID: h.Cfg.Telegram.ChatID, Username: "", Source: "config",
		})
	}

	// From friends list
	for name, chatID := range h.Cfg.Friends {
		if !seen[chatID] {
			seen[chatID] = true
			chats = append(chats, chatInfo{
				ChatID: chatID, Username: name, Source: "friend",
			})
		}
	}

	// Try to get the latest update from Telegram (short poll, won't block)
	// Use offset -1 to get only the most recent update without consuming the queue
	updates, err := client.GetUpdates(-1, 0)
	if err == nil {
		for _, u := range updates {
			if u.Message == nil || u.Message.Chat == nil {
				continue
			}
			cid := fmt.Sprintf("%d", u.Message.Chat.ID)
			if seen[cid] {
				continue
			}
			seen[cid] = true
			username := ""
			if u.Message.From != nil {
				username = u.Message.From.Username
			}
			chats = append(chats, chatInfo{
				ChatID: cid, Username: username, Source: "telegram",
			})
		}
	}

	hint := ""
	if len(chats) == 0 {
		hint = fmt.Sprintf("Send any message to @%s on Telegram, then click again", me.Username)
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"bot_username": me.Username,
		"bot_id":       fmt.Sprintf("%d", me.ID),
		"chat_ids":     chats,
		"hint":         hint,
	})
}
