package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"antiochus/internal/models"
	"antiochus/internal/telegram"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	tokenSet := h.Cfg.Telegram.BotToken != ""
	botUsername := ""
	if tokenSet {
		client := telegram.NewClient(h.Cfg.Telegram.BotToken)
		if me, err := client.GetMe(); err == nil {
			botUsername = me.Username
		}
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"token_set":    tokenSet,
		"bot_username": botUsername,
	})
}

func (h *Handler) PutConfig(w http.ResponseWriter, r *http.Request) {
	var req models.ConfigUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if req.Token == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "token required"})
		return
	}

	client := telegram.NewClient(req.Token)
	me, err := client.GetMe()
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid token: " + err.Error()})
		return
	}

	h.Cfg.SetToken(req.Token)
	if err := h.Cfg.Save(); err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "save config: " + err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"ok":           true,
		"bot_username": me.Username,
	})
}

func (h *Handler) GetRooms(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, h.Cfg.RoomsWithMeta())
}

// GetRoomCandidates returns group chats the bot has seen that aren't yet
// saved as rooms. The frontend polls this during the add-room flow.
func (h *Handler) GetRoomCandidates(w http.ResponseWriter, r *http.Request) {
	if h.RoomCandidates == nil {
		WriteJSON(w, http.StatusOK, []map[string]string{})
		return
	}
	WriteJSON(w, http.StatusOK, h.RoomCandidates())
}

func (h *Handler) PostRoom(w http.ResponseWriter, r *http.Request) {
	var req models.RoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	name := strings.ToLower(strings.TrimSpace(req.Name))
	if name == "" || req.ChatID == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "name and chat_id required"})
		return
	}

	h.Cfg.AddRoom(name, req.ChatID, strings.TrimSpace(req.Title))
	if err := h.Cfg.Save(); err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "save config: " + err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

func (h *Handler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	name = strings.ToLower(strings.TrimSpace(name))

	if !h.Cfg.RemoveRoom(name) {
		WriteJSON(w, http.StatusNotFound, map[string]string{"error": "room not found"})
		return
	}

	if err := h.Cfg.Save(); err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "save config: " + err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}
