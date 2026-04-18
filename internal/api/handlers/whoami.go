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

	seen := make(map[string]bool)
	type chatInfo struct {
		ChatID   string `json:"chat_id"`
		Title    string `json:"title"`
		Type     string `json:"type"`
		Username string `json:"username"`
		Source   string `json:"source"`
	}
	chats := make([]chatInfo, 0)

	for name, r := range h.Cfg.Rooms {
		if !seen[r.ChatID] {
			seen[r.ChatID] = true
			chats = append(chats, chatInfo{
				ChatID: r.ChatID, Title: r.Title, Username: name, Source: "room",
			})
		}
	}

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
				ChatID:   cid,
				Title:    u.Message.Chat.Title,
				Type:     u.Message.Chat.Type,
				Username: username,
				Source:   "telegram",
			})
		}
	}

	hint := ""
	if len(chats) == 0 {
		hint = fmt.Sprintf("Add @%s to a Telegram group and post a message", me.Username)
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"bot_username": me.Username,
		"bot_id":       fmt.Sprintf("%d", me.ID),
		"chat_ids":     chats,
		"hint":         hint,
	})
}
