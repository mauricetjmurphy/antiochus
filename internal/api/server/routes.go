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
	s.router.Put("/api/config", h.PutConfig)
	s.router.Get("/api/session/key", h.GetSessionKey)
	s.router.Post("/api/poll/start", h.StartPoll)
	s.router.Get("/guide", s.handleGuide)

	// Rooms — public so they can be managed during initial setup
	s.router.Get("/api/rooms", h.GetRooms)
	s.router.Post("/api/rooms", h.PostRoom)
	s.router.Delete("/api/rooms/{name}", h.DeleteRoom)

	// ── Authenticated (session required) ─────────────────
	s.router.Group(func(r chi.Router) {
		r.Use(requireSession)

		r.Post("/api/messages/send", h.SendMessage)
		r.Post("/api/messages/send-file", h.SendFile)
		r.Get("/api/messages/{room}", h.GetMessages)
		r.Get("/api/files/{filename}", h.DownloadFile)

		r.Post("/api/decrypt", h.Decrypt)
		r.Get("/api/whoami", h.WhoAmI)

		r.Post("/api/poll/stop", h.StopPoll)
		r.Post("/api/poll/fetch", h.FetchUpdates)
		r.Get("/api/rooms/candidates", h.GetRoomCandidates)
	})

	s.router.Get("/ws", s.handleWSUpgrade)
}
