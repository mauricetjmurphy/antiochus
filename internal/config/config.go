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
	BotToken         string `yaml:"bot_token"`
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

type Config struct {
	Server   ServerConfig      `yaml:"server"`
	Telegram TelegramConfig    `yaml:"telegram"`
	Crypto   CryptoConfig      `yaml:"crypto"`
	Storage  StorageConfig     `yaml:"storage"`
	Friends  map[string]string `yaml:"friends"`
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
		Friends: make(map[string]string),
	}
}

// ── Computed accessors ─────────────────────────────────────────

// Argon2MemoryKB returns the Argon2 memory cost in KB (what the crypto package expects).
func (c *Config) Argon2MemoryKB() uint32 {
	return uint32(c.Crypto.Argon2MemoryMB) * 1024
}

// MaxFileSize returns the max file size in bytes.
func (c *Config) MaxFileSize() int64 {
	return int64(c.Telegram.MaxFileSizeMB) * 1024 * 1024
}

// ── Paths ──────────────────────────────────────────────────────

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ciphergram")
}

func configPath() string {
	// Look for prod.yml next to the binary first, then fall back to ~/.ciphergram
	candidates := []string{
		"internal/config/prod.yml",
		filepath.Join(configDir(), "ciphergram.yml"),
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
		cfg.Friends = make(map[string]string)
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

func (c *Config) AddFriend(name, chatID string) {
	c.Friends[name] = chatID
}

func (c *Config) RemoveFriend(name string) bool {
	if _, ok := c.Friends[name]; !ok {
		return false
	}
	delete(c.Friends, name)
	return true
}

func (c *Config) ResolveChatID(friendName string) (string, error) {
	if chatID, ok := c.Friends[friendName]; ok {
		return chatID, nil
	}
	return "", fmt.Errorf("friend %q not found", friendName)
}

// FriendsWithMeta returns friends in the format the API handlers expect.
func (c *Config) FriendsWithMeta() []map[string]string {
	result := make([]map[string]string, 0, len(c.Friends))
	for name, chatID := range c.Friends {
		result = append(result, map[string]string{
			"name":    name,
			"chat_id": chatID,
			"added":   time.Now().Format("2006-01-02"),
		})
	}
	return result
}
