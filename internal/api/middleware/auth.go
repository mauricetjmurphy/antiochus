package middleware

import (
	"encoding/json"
	"net/http"
)

// SessionChecker is a function that returns true if a valid session exists.
type SessionChecker func() bool

// RequireSession returns middleware that rejects requests if no session is active.
func RequireSession(hasSession SessionChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !hasSession() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "session not established — unlock first",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
