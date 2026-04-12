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

type Server struct {
	cfg            *config.Config
	hub            *WSHub
	router         chi.Router
	handler        *handlers.Handler
	poller         *telegram.Poller
	messages       map[string][]*models.Message
	msgMu          sync.RWMutex
	frontendFS     fs.FS
	setupGuideHTML string
}

func New(cfg *config.Config, frontendFS fs.FS, setupGuideMD string) *Server {
	s := &Server{
		cfg:            cfg,
		hub:            NewWSHub(),
		router:         chi.NewRouter(),
		messages:       make(map[string][]*models.Message),
		frontendFS:     frontendFS,
		setupGuideHTML: renderMarkdownToHTML(setupGuideMD),
	}

	session := handlers.NewSession()

	s.handler = &handlers.Handler{
		Cfg:        cfg,
		Session:    session,
		AddMessage: func(friend string, msg interface{}) { s.addMessage(friend, msg.(*models.Message)) },
		GetMsgs:    func(friend string) interface{} { return s.getMessages(friend) },
		OnStartPoll: func(passphrase string) error {
			// Once session is established, wire up WS encryption
			s.hub.EncryptFunc = session.EncryptJSON
			return s.startPolling(passphrase)
		},
		OnStopPoll: func() {
			s.stopPolling()
			s.hub.EncryptFunc = nil
		},
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

// handleWSUpgrade validates the session before allowing a WebSocket connection.
func (s *Server) handleWSUpgrade(w http.ResponseWriter, r *http.Request) {
	if !s.handler.Session.HasPassphrase() {
		http.Error(w, `{"error":"session not established"}`, http.StatusUnauthorized)
		return
	}
	s.hub.HandleUpgrade(w, r)
}

func (s *Server) addMessage(friend string, msg *models.Message) {
	s.msgMu.Lock()
	s.messages[friend] = append(s.messages[friend], msg)
	if len(s.messages[friend]) > s.cfg.Storage.MaxMessagesPerFriend {
		s.messages[friend] = s.messages[friend][len(s.messages[friend])-s.cfg.Storage.MaxMessagesPerFriend:]
	}
	s.msgMu.Unlock()

	s.hub.Broadcast(models.WSEvent{
		Type:    "new_message",
		Friend:  friend,
		Message: msg,
	})
}

func (s *Server) getMessages(friend string) []*models.Message {
	s.msgMu.RLock()
	msgs := s.messages[friend]
	s.msgMu.RUnlock()
	if msgs == nil {
		return []*models.Message{}
	}
	return msgs
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

func (s *Server) bridgeMessages() {
	if s.poller == nil {
		return
	}

	for msg := range s.poller.Messages() {
		friendName := ""
		chatIDStr := fmt.Sprintf("%d", msg.ChatID)
		for name, chatID := range s.cfg.Friends {
			if chatID == chatIDStr {
				friendName = name
				break
			}
		}
		if friendName == "" {
			friendName = msg.Sender
		}

		m := &models.Message{
			Time:     msg.Time,
			Sender:   msg.Sender,
			Content:  msg.Content,
			Type:     msg.Type,
			Filename: msg.Filename,
			FilePath: msg.FilePath,
		}
		s.addMessage(friendName, m)
	}
}
