package server

import (
	authmw "antiochus/internal/api/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) setupRoutes() {
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Logger)

	h := s.handler
	requireSession := authmw.RequireSession(h.Session.HasPassphrase)

	// ── Public (no session required) ─────────────────────
	s.router.Get("/api/config", h.GetConfig)
	s.router.Get("/api/session/key", h.GetSessionKey)
	s.router.Post("/api/poll/start", h.StartPoll)
	s.router.Get("/guide", s.handleGuide)

	// ── Authenticated (session required) ─────────────────
	s.router.Group(func(r chi.Router) {
		r.Use(requireSession)

		// Config
		r.Put("/api/config", h.PutConfig)

		// Friends
		r.Get("/api/friends", h.GetFriends)
		r.Post("/api/friends", h.PostFriend)
		r.Delete("/api/friends/{name}", h.DeleteFriend)

		// Messages
		r.Post("/api/messages/send", h.SendMessage)
		r.Post("/api/messages/send-file", h.SendFile)
		r.Get("/api/messages/{friend}", h.GetMessages)

		// Decrypt
		r.Post("/api/decrypt", h.Decrypt)

		// Who am I
		r.Get("/api/whoami", h.WhoAmI)

		// Polling control
		r.Post("/api/poll/stop", h.StopPoll)
	})

	// ── WebSocket (authenticated via session token query param) ──
	s.router.Get("/ws", s.handleWSUpgrade)
}
