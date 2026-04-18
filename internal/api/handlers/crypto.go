package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"antiochus/internal/crypto"
	"antiochus/internal/models"
)

func (h *Handler) Decrypt(w http.ResponseWriter, r *http.Request) {
	var req models.DecryptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if req.Ciphertext == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "ciphertext required"})
		return
	}

	clean := strings.ReplaceAll(req.Ciphertext, "```", "")
	clean = strings.ReplaceAll(clean, "\U0001f510 Antiochus", "")
	clean = strings.TrimSpace(clean)
	clean = strings.Join(strings.Fields(clean), "")

	packet, err := base64.StdEncoding.DecodeString(clean)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid base64"})
		return
	}

	ptype, meta, plaintext, err := crypto.ParsePacket(packet, h.Session.Passphrase())
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "decryption failed: wrong passphrase or corrupted data"})
		return
	}

	result := map[string]string{
		"content": string(plaintext),
	}

	if ptype == crypto.TypeFile {
		result["type"] = "file"
		result["filename"] = meta["filename"]
	} else {
		result["type"] = "text"
	}

	WriteJSON(w, http.StatusOK, result)
}
