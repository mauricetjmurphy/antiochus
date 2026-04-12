package models

type Message struct {
	Time     string `json:"time"`
	Sender   string `json:"sender"`
	Content  string `json:"content"`
	Type     string `json:"type"`     // "text" or "file"
	Filename string `json:"filename,omitempty"`
	FilePath string `json:"filePath,omitempty"`
}

type WSEvent struct {
	Type    string   `json:"type"`    // "new_message", "poll_status", "error"
	Friend  string   `json:"friend,omitempty"`
	Message *Message `json:"message,omitempty"`
	Active  bool     `json:"active,omitempty"`
	Error   string   `json:"error,omitempty"`
}

type DecryptRequest struct {
	Ciphertext string `json:"ciphertext"`
}

type FriendRequest struct {
	Name   string `json:"name"`
	ChatID string `json:"chat_id"`
}

type ConfigUpdateRequest struct {
	Token string `json:"token"`
}
