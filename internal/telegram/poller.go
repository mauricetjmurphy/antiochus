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
	Sender   string
	ChatID   int64
	Content  string
	Type     string // "text" or "file"
	Filename string
	FilePath string // local save path for files
	Time     string
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

		updates, err := p.client.GetUpdates(p.offset, 30)
		if err != nil {
			log.Printf("poll error: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for _, update := range updates {
			p.offset = update.UpdateID + 1
			if update.Message == nil {
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
	if msg.Chat != nil {
		chatID = msg.Chat.ID
	}
	timestamp := time.Now().Format("15:04")

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
		return
	}

	// Check for text messages with Antiochus payload
	text := msg.Text
	if !strings.Contains(text, "ANTIO") && !strings.Contains(text, "\U0001f510") {
		return
	}

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
