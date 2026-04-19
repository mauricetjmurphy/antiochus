package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed prod.yml
var defaultConfigYAML []byte

// ── YAML config structure ──────────────────────────────────────

type ServerConfig struct {
	Addr string `yaml:"addr"`
}

type TelegramConfig struct {
	BotToken         string `yaml:"bot_token"`
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
	MaxMessagesPerRoom int `yaml:"max_messages_per_room"`
}

// Room represents a Telegram group chat used as an Antiochus room.
// ChatID is the group's chat_id. Title is the group title from Telegram.
type Room struct {
	ChatID string `yaml:"chat_id"`
	Title  string `yaml:"title,omitempty"`
	Added  string `yaml:"added,omitempty"`
}

type Config struct {
	Server   ServerConfig    `yaml:"server"`
	Telegram TelegramConfig  `yaml:"telegram"`
	Crypto   CryptoConfig    `yaml:"crypto"`
	Storage  StorageConfig   `yaml:"storage"`
	Rooms    map[string]Room `yaml:"rooms"`
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
			MaxMessagesPerRoom: 500,
		},
		Rooms: make(map[string]Room),
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

// configPath returns the runtime config path in the user's home directory so
// that Antiochus works the same whether it's run from a dev tree, an installed
// binary, or a read-only AppImage mount. The embedded prod.yml supplies the
// defaults on first launch.
func configPath() string {
	return filepath.Join(configDir(), "prod.yml")
}

func (c *Config) Path() string {
	return c.path
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
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			return nil, fmt.Errorf("create config dir: %w", err)
		}
		if err := os.WriteFile(p, defaultConfigYAML, 0600); err != nil {
			return nil, fmt.Errorf("write default config: %w", err)
		}
		data = defaultConfigYAML
	} else if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Rooms == nil {
		cfg.Rooms = make(map[string]Room)
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

func (c *Config) AddRoom(name, chatID, title string) {
	c.Rooms[name] = Room{
		ChatID: chatID,
		Title:  title,
		Added:  time.Now().Format("2006-01-02"),
	}
}

func (c *Config) RemoveRoom(name string) bool {
	if _, ok := c.Rooms[name]; !ok {
		return false
	}
	delete(c.Rooms, name)
	return true
}

func (c *Config) ResolveChatID(name string) (string, error) {
	if r, ok := c.Rooms[name]; ok {
		return r.ChatID, nil
	}
	return "", fmt.Errorf("%q not found", name)
}

// RoomsWithMeta returns rooms for the API.
func (c *Config) RoomsWithMeta() []map[string]string {
	result := make([]map[string]string, 0, len(c.Rooms))
	for name, r := range c.Rooms {
		result = append(result, map[string]string{
			"name":    name,
			"chat_id": r.ChatID,
			"title":   r.Title,
			"added":   r.Added,
		})
	}
	return result
}
