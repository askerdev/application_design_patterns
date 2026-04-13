package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client — adaptee for the Adapter pattern — wraps the Telegram Bot HTTP API
type Client struct {
	token  string
	chatID int64
	http   *http.Client
}

func NewClient(token string, chatID int64) *Client {
	return &Client{
		token:  token,
		chatID: chatID,
		http:   &http.Client{Timeout: 10 * time.Second},
	}
}

// IsConfigured returns false if token or chatID is missing
func (c *Client) IsConfigured() bool {
	return c.token != "" && c.chatID != 0
}

type sendMessageRequest struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

// SendMessage sends a text message via Telegram Bot API
func (c *Client) SendMessage(text string) error {
	if !c.IsConfigured() {
		return nil
	}
	body, err := json.Marshal(sendMessageRequest{ChatID: c.chatID, Text: text})
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.token)
	resp, err := c.http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned %d", resp.StatusCode)
	}
	return nil
}
