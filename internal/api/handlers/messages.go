package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"antiochus/internal/crypto"
	"antiochus/internal/models"
	"antiochus/internal/telegram"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		To   string `json:"to"`
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if req.To == "" || req.Text == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "to and text required"})
		return
	}

	chatID, err := h.Cfg.ResolveChatID(req.To)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if h.Cfg.Telegram.BotToken == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "bot token not configured"})
		return
	}

	passphrase := h.Session.Passphrase()
	packet, err := crypto.BuildPacket(crypto.TypeText, map[string]string{}, []byte(req.Text), passphrase)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "encrypt: " + err.Error()})
		return
	}

	client := telegram.NewClient(h.Cfg.Telegram.BotToken)

	encoded := base64.StdEncoding.EncodeToString(packet)
	if len(encoded) <= h.Cfg.Telegram.MessageCharLimit {
		msgText := fmt.Sprintf("\U0001f510 CipherGram\n```\n%s\n```", encoded)
		msgID, err := client.SendMessage(chatID, msgText)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "send: " + err.Error()})
			return
		}

		msg := &models.Message{
			Time:    time.Now().Format("15:04"),
			Sender:  "you",
			Content: req.Text,
			Type:    "text",
		}
		h.AddMessage(req.To, msg)

		WriteJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "message_id": msgID})
	} else {
		err := client.SendDocument(chatID, packet, "ciphergram.enc", "\U0001f510 CipherGram")
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "send: " + err.Error()})
			return
		}

		msg := &models.Message{
			Time:    time.Now().Format("15:04"),
			Sender:  "you",
			Content: req.Text,
			Type:    "text",
		}
		h.AddMessage(req.To, msg)

		WriteJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
	}
}

func (h *Handler) SendFile(w http.ResponseWriter, r *http.Request) {
	maxSize := h.Cfg.MaxFileSize()
	if err := r.ParseMultipartForm(maxSize + 1024*1024); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "parse form: " + err.Error()})
		return
	}

	to := r.FormValue("to")
	if to == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "to required"})
		return
	}

	chatID, err := h.Cfg.ResolveChatID(to)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "file required"})
		return
	}
	defer file.Close()

	if header.Size > maxSize {
		WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("file too large (max %d MB)", h.Cfg.Telegram.MaxFileSizeMB),
		})
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "read file: " + err.Error()})
		return
	}

	passphrase := h.Session.Passphrase()
	filename := header.Filename
	meta := map[string]string{"filename": filename}
	packet, err := crypto.BuildPacket(crypto.TypeFile, meta, fileBytes, passphrase)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "encrypt: " + err.Error()})
		return
	}

	client := telegram.NewClient(h.Cfg.Telegram.BotToken)
	caption := fmt.Sprintf("\U0001f510 CipherGram [\U0001f4ce %s]", filename)
	if err := client.SendDocument(chatID, packet, "ciphergram.enc", caption); err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "send: " + err.Error()})
		return
	}

	msg := &models.Message{
		Time:     time.Now().Format("15:04"),
		Sender:   "you",
		Content:  fmt.Sprintf("\U0001f4ce %s (sent)", filename),
		Type:     "file",
		Filename: filename,
	}
	h.AddMessage(to, msg)

	WriteJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "filename": filename})
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	friend := chi.URLParam(r, "friend")
	if friend == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "friend name required"})
		return
	}

	msgs := h.GetMsgs(friend)
	WriteJSON(w, http.StatusOK, msgs)
}
