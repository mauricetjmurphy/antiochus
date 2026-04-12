package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type Client struct {
	token      string
	baseURL    string
	httpClient *http.Client
}

type BotUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	IsBot    bool   `json:"is_bot"`
}

type TgResponse struct {
	OK          bool            `json:"ok"`
	Description string          `json:"description"`
	Result      json.RawMessage `json:"result"`
}

type Update struct {
	UpdateID int64   `json:"update_id"`
	Message  *TgMessage `json:"message"`
}

type TgMessage struct {
	MessageID int64       `json:"message_id"`
	From      *TgUser     `json:"from"`
	Chat      *TgChat     `json:"chat"`
	Text      string      `json:"text"`
	Document  *TgDocument `json:"document"`
}

type TgUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type TgChat struct {
	ID int64 `json:"id"`
}

type TgDocument struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
}

type TgFile struct {
	FileID   string `json:"file_id"`
	FilePath string `json:"file_path"`
}

func NewClient(token string) *Client {
	return &Client{
		token:   token,
		baseURL: fmt.Sprintf("https://api.telegram.org/bot%s", token),
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *Client) GetMe() (*BotUser, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/getMe")
	if err != nil {
		return nil, fmt.Errorf("getMe request: %w", err)
	}
	defer resp.Body.Close()

	var tgResp TgResponse
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if !tgResp.OK {
		return nil, fmt.Errorf("getMe failed: %s", tgResp.Description)
	}

	var user BotUser
	if err := json.Unmarshal(tgResp.Result, &user); err != nil {
		return nil, fmt.Errorf("decode user: %w", err)
	}
	return &user, nil
}

func (c *Client) SendMessage(chatID string, text string) (int64, error) {
	body, _ := json.Marshal(map[string]string{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	})

	resp, err := c.httpClient.Post(c.baseURL+"/sendMessage", "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("sendMessage request: %w", err)
	}
	defer resp.Body.Close()

	var tgResp TgResponse
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return 0, fmt.Errorf("decode response: %w", err)
	}
	if !tgResp.OK {
		return 0, fmt.Errorf("sendMessage failed: %s", tgResp.Description)
	}

	var msg TgMessage
	if err := json.Unmarshal(tgResp.Result, &msg); err != nil {
		return 0, fmt.Errorf("decode message: %w", err)
	}
	return msg.MessageID, nil
}

func (c *Client) SendDocument(chatID string, fileBytes []byte, filename, caption string) error {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	_ = w.WriteField("chat_id", chatID)
	if caption != "" {
		_ = w.WriteField("caption", caption)
	}

	fw, err := w.CreateFormFile("document", filename)
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	if _, err := fw.Write(fileBytes); err != nil {
		return fmt.Errorf("write file data: %w", err)
	}
	w.Close()

	resp, err := c.httpClient.Post(c.baseURL+"/sendDocument", w.FormDataContentType(), &buf)
	if err != nil {
		return fmt.Errorf("sendDocument request: %w", err)
	}
	defer resp.Body.Close()

	var tgResp TgResponse
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if !tgResp.OK {
		return fmt.Errorf("sendDocument failed: %s", tgResp.Description)
	}
	return nil
}

func (c *Client) GetFile(fileID string) ([]byte, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/getFile?file_id=%s", c.baseURL, fileID))
	if err != nil {
		return nil, fmt.Errorf("getFile request: %w", err)
	}
	defer resp.Body.Close()

	var tgResp TgResponse
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if !tgResp.OK {
		return nil, fmt.Errorf("getFile failed: %s", tgResp.Description)
	}

	var file TgFile
	if err := json.Unmarshal(tgResp.Result, &file); err != nil {
		return nil, fmt.Errorf("decode file: %w", err)
	}

	dlURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", c.token, file.FilePath)
	dlResp, err := c.httpClient.Get(dlURL)
	if err != nil {
		return nil, fmt.Errorf("download file: %w", err)
	}
	defer dlResp.Body.Close()

	return io.ReadAll(dlResp.Body)
}

func (c *Client) GetUpdates(offset int64, timeout int) ([]Update, error) {
	url := fmt.Sprintf("%s/getUpdates?offset=%d&timeout=%d", c.baseURL, offset, timeout)

	client := &http.Client{Timeout: time.Duration(timeout+5) * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("getUpdates request: %w", err)
	}
	defer resp.Body.Close()

	var tgResp TgResponse
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if !tgResp.OK {
		return nil, fmt.Errorf("getUpdates failed: %s", tgResp.Description)
	}

	var updates []Update
	if err := json.Unmarshal(tgResp.Result, &updates); err != nil {
		return nil, fmt.Errorf("decode updates: %w", err)
	}
	return updates, nil
}
