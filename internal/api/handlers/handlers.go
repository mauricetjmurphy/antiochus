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
	AddMessage func(friend string, msg interface{})
	GetMsgs    func(friend string) interface{}

	// Polling lifecycle callbacks — set by the server package
	OnStartPoll func(passphrase string) error
	OnStopPoll  func()
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
