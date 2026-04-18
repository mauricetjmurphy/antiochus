package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

	"antiochus/internal/config"

	"github.com/go-chi/chi/v5"
)

// DownloadFile serves a decrypted file from the received-files directory.
// The filename is the base name of the message's FilePath.
func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "filename")
	if name == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "filename required"})
		return
	}

	// Defensive: reject traversal attempts
	clean := filepath.Base(name)
	if clean == "" || clean == "." || clean == ".." ||
		strings.ContainsAny(clean, `/\`) || strings.Contains(clean, "..") {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid filename"})
		return
	}

	dir := config.ReceivedDir()
	full := filepath.Join(dir, clean)

	// Ensure the resolved path is inside the received dir
	absDir, _ := filepath.Abs(dir)
	absPath, _ := filepath.Abs(full)
	if !strings.HasPrefix(absPath, absDir+string(filepath.Separator)) {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid path"})
		return
	}

	w.Header().Set("Content-Disposition", `attachment; filename="`+clean+`"`)
	http.ServeFile(w, r, full)
}
