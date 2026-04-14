package server

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"antiochus/internal/api/handlers"
	"antiochus/internal/config"
	"antiochus/internal/models"
	"antiochus/internal/telegram"

	"github.com/go-chi/chi/v5"
)

type candidateRoom struct {
	ChatID string
	Title  string
	Type   string
	First  string
}

type Server struct {
	cfg            *config.Config
	hub            *WSHub
	router         chi.Router
	handler        *handlers.Handler
	poller         *telegram.Poller
	messages       map[string][]*models.Message
	msgMu          sync.RWMutex
	candidates     map[string]candidateRoom // chat_id -> candidate room
	candMu         sync.RWMutex
	frontendFS     fs.FS
	setupGuideHTML string
}

func New(cfg *config.Config, frontendFS fs.FS, setupGuideMD string) *Server {
	s := &Server{
		cfg:            cfg,
		hub:            NewWSHub(),
		router:         chi.NewRouter(),
		messages:       make(map[string][]*models.Message),
		candidates:     make(map[string]candidateRoom),
		frontendFS:     frontendFS,
		setupGuideHTML: renderMarkdownToHTML(setupGuideMD),
	}

	session := handlers.NewSession()

	s.handler = &handlers.Handler{
		Cfg:        cfg,
		Session:    session,
		AddMessage: func(room string, msg interface{}) { s.addMessage(room, msg.(*models.Message)) },
		GetMsgs:    func(room string) interface{} { return s.getMessages(room) },
		OnStartPoll: func(passphrase string) error {
			s.hub.EncryptFunc = session.EncryptJSON
			return s.startPolling(passphrase)
		},
		OnStopPoll: func() {
			s.stopPolling()
			s.hub.EncryptFunc = nil
		},
		OnFetchUpdates: func() (int, error) {
			return s.fetchUpdates()
		},
		RoomCandidates: s.listCandidates,
	}

	s.setupRoutes()
	s.setupFrontend()

	go s.hub.Run()

	return s
}

func (s *Server) setupFrontend() {
	if s.frontendFS == nil {
		return
	}

	fileServer := http.FileServer(http.FS(s.frontendFS))

	s.router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if _, err := fs.Stat(s.frontendFS, path); err != nil {
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	})
}

func (s *Server) ListenAndServe(addr string) error {
	log.Printf("Antiochus v2.0 listening on %s", addr)
	if strings.HasPrefix(addr, "0.0.0.0") || strings.HasPrefix(addr, ":") {
		log.Printf("WARNING: server is accessible from external hosts")
	}
	return http.ListenAndServe(addr, s.router)
}

func (s *Server) handleWSUpgrade(w http.ResponseWriter, r *http.Request) {
	if !s.handler.Session.HasPassphrase() {
		http.Error(w, `{"error":"session not established"}`, http.StatusUnauthorized)
		return
	}
	s.hub.HandleUpgrade(w, r)
}

func (s *Server) addMessage(room string, msg *models.Message) {
	s.msgMu.Lock()
	s.messages[room] = append(s.messages[room], msg)
	if len(s.messages[room]) > s.cfg.Storage.MaxMessagesPerRoom {
		s.messages[room] = s.messages[room][len(s.messages[room])-s.cfg.Storage.MaxMessagesPerRoom:]
	}
	s.msgMu.Unlock()

	log.Printf("[addMessage] room=%s sender=%s type=%s content=%q", room, msg.Sender, msg.Type, msg.Content)

	s.hub.Broadcast(models.WSEvent{
		Type:    "new_message",
		Room:    room,
		Message: msg,
	})
}

func (s *Server) getMessages(room string) []*models.Message {
	s.msgMu.RLock()
	msgs := s.messages[room]
	s.msgMu.RUnlock()
	if msgs == nil {
		return []*models.Message{}
	}
	return msgs
}

func (s *Server) recordCandidate(chatID, title, chatType, first string) {
	if chatID == "" || chatID == "0" {
		return
	}
	// Only groups and supergroups are valid rooms
	if chatType != "group" && chatType != "supergroup" {
		return
	}
	s.candMu.Lock()
	defer s.candMu.Unlock()
	if _, ok := s.candidates[chatID]; ok {
		return
	}
	s.candidates[chatID] = candidateRoom{ChatID: chatID, Title: title, Type: chatType, First: first}
}

func (s *Server) forgetCandidate(chatID string) {
	s.candMu.Lock()
	defer s.candMu.Unlock()
	delete(s.candidates, chatID)
}

func (s *Server) listCandidates() []map[string]string {
	// Build a set of chat_ids already saved as rooms so we don't return them.
	saved := make(map[string]bool)
	for _, r := range s.cfg.Rooms {
		saved[r.ChatID] = true
	}

	s.candMu.RLock()
	defer s.candMu.RUnlock()
	out := make([]map[string]string, 0, len(s.candidates))
	for _, c := range s.candidates {
		if saved[c.ChatID] {
			continue
		}
		out = append(out, map[string]string{
			"chat_id": c.ChatID,
			"title":   c.Title,
			"type":    c.Type,
			"first":   c.First,
		})
	}
	return out
}

func (s *Server) startPolling(passphrase string) error {
	if s.cfg.Telegram.BotToken == "" {
		return fmt.Errorf("bot token not configured")
	}

	if s.poller != nil && s.poller.IsRunning() {
		s.poller.Stop()
		time.Sleep(100 * time.Millisecond)
	}

	client := telegram.NewClient(s.cfg.Telegram.BotToken)
	s.poller = telegram.NewPoller(client, passphrase, config.ReceivedDir())
	s.poller.Start()

	go s.bridgeMessages()

	s.hub.Broadcast(models.WSEvent{Type: "poll_status", Active: true})
	return nil
}

func (s *Server) stopPolling() {
	if s.poller != nil {
		s.poller.Stop()
	}
	s.hub.Broadcast(models.WSEvent{Type: "poll_status", Active: false})
}

func (s *Server) fetchUpdates() (int, error) {
	if s.poller == nil {
		return 0, fmt.Errorf("poller not initialized — unlock first")
	}
	return s.poller.FetchOnce()
}

func (s *Server) bridgeMessages() {
	if s.poller == nil {
		return
	}

	for msg := range s.poller.Messages() {
		chatIDStr := fmt.Sprintf("%d", msg.ChatID)

		target := ""
		for name, r := range s.cfg.Rooms {
			if r.ChatID == chatIDStr {
				target = name
				break
			}
		}

		if target == "" {
			// Unknown chat — record as a candidate if it's a group chat.
			// Private chats (friend setup) aren't auto-detected; users add them manually.
			s.recordCandidate(chatIDStr, msg.ChatTitle, msg.ChatType, msg.Time)
			log.Printf("[bridge] unknown chatID=%s (type=%s, title=%q) — candidate", chatIDStr, msg.ChatType, msg.ChatTitle)
			continue
		}

		log.Printf("[bridge] matched chatID=%s to %s", chatIDStr, target)
		m := &models.Message{
			Time:     msg.Time,
			Sender:   msg.Sender,
			Content:  msg.Content,
			Type:     msg.Type,
			Filename: msg.Filename,
			FilePath: msg.FilePath,
		}
		s.addMessage(target, m)
		s.forgetCandidate(chatIDStr)
	}
}
