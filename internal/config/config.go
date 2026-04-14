package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ── YAML config structure ──────────────────────────────────────

type ServerConfig struct {
	Addr string `yaml:"addr"`
}

type TelegramConfig struct {
	BotToken         string `yaml:"bot_token"` // YOUR bot — you poll this for incoming messages
	ChatID           string `yaml:"chat_id"`
	PollTimeout      int    `yaml:"poll_timeout"`
	MaxFileSizeMB    int    `yaml:"max_file_size_mb"`
	MessageCharLimit int    `yaml:"message_char_limit"`
}

type CryptoConfig struct {
	Argon2TimeCost    uint32 `yaml:"argon2_time_cost"`
	Argon2MemoryMB    int    `yaml:"argon2_memory_mb"`
	Argon2Parallelism uint8  `yaml:"argon2_parallelism"`
}

type StorageConfig struct {
	MaxMessagesPerFriend int `yaml:"max_messages_per_friend"`
}

// Friend represents a contact. ChatID is the friend's chat ID with their bot.
// SendBotToken is the friend's bot token, which YOU use to send messages to them.
type Friend struct {
	ChatID       string `yaml:"chat_id"`
	SendBotToken string `yaml:"send_bot_token"`
	Added        string `yaml:"added,omitempty"`
}

type Config struct {
	Server   ServerConfig      `yaml:"server"`
	Telegram TelegramConfig    `yaml:"telegram"`
	Crypto   CryptoConfig      `yaml:"crypto"`
	Storage  StorageConfig     `yaml:"storage"`
	Friends  map[string]Friend `yaml:"friends"`
	path     string
}

// ── Defaults ───────────────────────────────────────────────────

func defaults() *Config {
	return &Config{
		Server: ServerConfig{
			Addr: "127.0.0.1:8080",
		},
		Telegram: TelegramConfig{
			PollTimeout:      30,
			MaxFileSizeMB:    45,
			MessageCharLimit: 4000,
		},
		Crypto: CryptoConfig{
			Argon2TimeCost:    4,
			Argon2MemoryMB:    256,
			Argon2Parallelism: 4,
		},
		Storage: StorageConfig{
			MaxMessagesPerFriend: 500,
		},
		Friends: make(map[string]Friend),
	}
}

// ── Computed accessors ─────────────────────────────────────────

func (c *Config) Argon2MemoryKB() uint32 {
	return uint32(c.Crypto.Argon2MemoryMB) * 1024
}

func (c *Config) MaxFileSize() int64 {
	return int64(c.Telegram.MaxFileSizeMB) * 1024 * 1024
}

// ── Paths ──────────────────────────────────────────────────────

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".antiochus")
}

func configPath() string {
	candidates := []string{
		"internal/config/prod.yml",
		filepath.Join(configDir(), "antiochus.yml"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return candidates[len(candidates)-1]
}

func ReceivedDir() string {
	return filepath.Join(configDir(), "received")
}

// ── Load / Save ────────────────────────────────────────────────

func Load() (*Config, error) {
	p := configPath()
	cfg := defaults()
	cfg.path = p

	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Friends == nil {
		cfg.Friends = make(map[string]Friend)
	}
	return cfg, nil
}

func (c *Config) Save() error {
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(c.path, data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// ── Mutations ──────────────────────────────────────────────────

func (c *Config) SetToken(token string) {
	c.Telegram.BotToken = token
}

func (c *Config) AddFriend(name, chatID, sendBotToken string) {
	c.Friends[name] = Friend{
		ChatID:       chatID,
		SendBotToken: sendBotToken,
		Added:        time.Now().Format("2006-01-02"),
	}
}

func (c *Config) RemoveFriend(name string) bool {
	if _, ok := c.Friends[name]; !ok {
		return false
	}
	delete(c.Friends, name)
	return true
}

func (c *Config) ResolveChatID(friendName string) (string, error) {
	if f, ok := c.Friends[friendName]; ok {
		return f.ChatID, nil
	}
	return "", fmt.Errorf("friend %q not found", friendName)
}

// ResolveSendBotToken returns the bot token used to send messages to this friend.
// Falls back to the user's own bot token if the friend doesn't have a dedicated one.
func (c *Config) ResolveSendBotToken(friendName string) (string, error) {
	if f, ok := c.Friends[friendName]; ok {
		if f.SendBotToken != "" {
			return f.SendBotToken, nil
		}
		return c.Telegram.BotToken, nil
	}
	return "", fmt.Errorf("friend %q not found", friendName)
}

// FriendsWithMeta returns friends for the API — does NOT expose send_bot_token.
func (c *Config) FriendsWithMeta() []map[string]string {
	result := make([]map[string]string, 0, len(c.Friends))
	for name, f := range c.Friends {
		hasToken := "false"
		if f.SendBotToken != "" {
			hasToken = "true"
		}
		result = append(result, map[string]string{
			"name":               name,
			"chat_id":            f.ChatID,
			"added":              f.Added,
			"has_send_bot_token": hasToken,
		})
	}
	return result
}
