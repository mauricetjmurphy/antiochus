package telegram

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"antiochus/internal/crypto"
)

type IncomingMessage struct {
	Sender       string
	SenderUserID int64 // Telegram user ID (stable across bots)
	ChatID       int64
	ChatType     string // "private", "group", "supergroup", "channel"
	ChatTitle    string // group/channel title if applicable
	Content      string
	Type         string // "text" or "file"
	Filename     string
	FilePath     string // local save path for files
	Time         string
}

type Poller struct {
	client     *Client
	passphrase string
	mu         sync.RWMutex
	msgChan    chan IncomingMessage
	stopChan   chan struct{}
	running    bool
	offset     int64
	saveDir    string
}

func NewPoller(client *Client, passphrase, saveDir string) *Poller {
	return &Poller{
		client:     client,
		passphrase: passphrase,
		msgChan:    make(chan IncomingMessage, 100),
		stopChan:   make(chan struct{}),
		saveDir:    saveDir,
	}
}

func (p *Poller) Messages() <-chan IncomingMessage {
	return p.msgChan
}

func (p *Poller) Start() {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return
	}
	p.running = true
	p.mu.Unlock()

	go p.pollLoop()
}

// FetchOnce performs a single getUpdates call with a short timeout and processes
// any updates it finds. Returns the number of updates processed.
func (p *Poller) FetchOnce() (int, error) {
	log.Printf("[fetch] calling getUpdates offset=%d", p.offset)
	updates, err := p.client.GetUpdates(p.offset, 0)
	if err != nil {
		log.Printf("[fetch] error: %v", err)
		return 0, err
	}
	log.Printf("[fetch] getUpdates returned %d update(s)", len(updates))

	count := 0
	for _, update := range updates {
		p.offset = update.UpdateID + 1
		if update.Message == nil {
			continue
		}
		p.processMessage(update.Message)
		count++
	}
	return count, nil
}

func (p *Poller) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.running {
		return
	}
	p.running = false
	close(p.stopChan)
}

func (p *Poller) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.running
}

func (p *Poller) pollLoop() {
	for {
		select {
		case <-p.stopChan:
			return
		default:
		}

		log.Printf("[poll] calling getUpdates offset=%d", p.offset)
		updates, err := p.client.GetUpdates(p.offset, 30)
		if err != nil {
			log.Printf("[poll] error: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		log.Printf("[poll] getUpdates returned %d update(s)", len(updates))

		for _, update := range updates {
			p.offset = update.UpdateID + 1
			if update.Message == nil {
				log.Printf("[poll] update %d has no message, skipping", update.UpdateID)
				continue
			}
			p.processMessage(update.Message)
		}
	}
}

func (p *Poller) processMessage(msg *TgMessage) {
	sender := "unknown"
	if msg.From != nil {
		sender = msg.From.Username
		if sender == "" {
			sender = fmt.Sprintf("user_%d", msg.From.ID)
		}
	}
	chatID := int64(0)
	chatType := ""
	chatTitle := ""
	if msg.Chat != nil {
		chatID = msg.Chat.ID
		chatType = msg.Chat.Type
		chatTitle = msg.Chat.Title
	}
	timestamp := time.Now().Format("15:04")

	log.Printf("[poll] incoming message: sender=%s chatID=%d hasDoc=%v textLen=%d",
		sender, chatID, msg.Document != nil, len(msg.Text))

	p.mu.RLock()
	passphrase := p.passphrase
	p.mu.RUnlock()

	// Check for document messages (.enc files)
	if msg.Document != nil && strings.HasSuffix(msg.Document.FileName, ".enc") {
		fileBytes, err := p.client.GetFile(msg.Document.FileID)
		if err != nil {
			log.Printf("download file error: %v", err)
			return
		}

		ptype, meta, plaintext, err := crypto.ParsePacket(fileBytes, passphrase)
		if err != nil {
			log.Printf("decrypt file error: %v", err)
			return
		}

		if ptype == crypto.TypeFile {
			filename := meta["filename"]
			if filename == "" {
				filename = "unknown_file"
			}
			savePath, err := saveReceivedFile(p.saveDir, filename, plaintext)
			if err != nil {
				log.Printf("save file error: %v", err)
				return
			}
			p.msgChan <- IncomingMessage{
				Sender:    sender,
				ChatID:    chatID,
				ChatType:  chatType,
				ChatTitle: chatTitle,
				Content:   fmt.Sprintf("📎 %s", filename),
				Type:      "file",
				Filename:  filename,
				FilePath:  savePath,
				Time:      timestamp,
			}
		} else {
			p.msgChan <- IncomingMessage{
				Sender:    sender,
				ChatID:    chatID,
				ChatType:  chatType,
				ChatTitle: chatTitle,
				Content:   string(plaintext),
				Type:      "text",
				Time:      timestamp,
			}
		}
		return
	}

	text := msg.Text
	if text == "" {
		return
	}

	// Plain text (no Antiochus marker) — pass through as-is
	if !strings.Contains(text, "ANTIO") && !strings.Contains(text, "\U0001f510") {
		p.msgChan <- IncomingMessage{
			Sender:    sender,
			ChatID:    chatID,
			ChatType:  chatType,
			ChatTitle: chatTitle,
			Content:   text,
			Type:      "text",
			Time:      timestamp,
		}
		return
	}
	log.Printf("[poll] found Antiochus payload in text from %s", sender)

	clean := strings.ReplaceAll(text, "```", "")
	clean = strings.ReplaceAll(clean, "\U0001f510 Antiochus", "")
	clean = strings.TrimSpace(clean)
	clean = strings.Join(strings.Fields(clean), "")

	packet, err := base64.StdEncoding.DecodeString(clean)
	if err != nil {
		log.Printf("base64 decode error: %v", err)
		return
	}

	ptype, meta, plaintext, err := crypto.ParsePacket(packet, passphrase)
	if err != nil {
		log.Printf("decrypt error: %v", err)
		return
	}

	if ptype == crypto.TypeFile {
		filename := meta["filename"]
		if filename == "" {
			filename = "unknown_file"
		}
		savePath, err := saveReceivedFile(p.saveDir, filename, plaintext)
		if err != nil {
			log.Printf("save file error: %v", err)
			return
		}
		p.msgChan <- IncomingMessage{
			Sender:   sender,
			ChatID:   chatID,
			Content:  fmt.Sprintf("📎 %s", filename),
			Type:     "file",
			Filename: filename,
			FilePath: savePath,
			Time:     timestamp,
		}
	} else {
		p.msgChan <- IncomingMessage{
			Sender:  sender,
			ChatID:  chatID,
			Content: string(plaintext),
			Type:    "text",
			Time:    timestamp,
		}
	}
}
