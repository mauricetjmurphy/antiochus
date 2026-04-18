package handlers

import (
	"encoding/json"
	"net/http"

	"antiochus/internal/config"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	Cfg        *config.Config
	Session    *Session
	AddMessage func(room string, msg interface{})
	GetMsgs    func(room string) interface{}

	// Polling lifecycle callbacks — set by the server package
	OnStartPoll    func(passphrase string) error
	OnStopPoll     func()
	OnFetchUpdates func() (int, error)
	OnTokenChanged func() error

	// RoomCandidates returns group chats the bot has received messages
	// from that aren't yet configured as rooms.
	RoomCandidates func() []map[string]string
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
