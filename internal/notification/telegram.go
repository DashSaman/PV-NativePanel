package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type TelegramChannel struct {
	BotToken string
	ChatID   string
	Client   *http.Client
	BaseURL  string
}

func (c *TelegramChannel) Name() string { return "telegram" }

func (c *TelegramChannel) Send(ctx context.Context, message Message) error {
	if c == nil || strings.TrimSpace(c.BotToken) == "" || strings.TrimSpace(c.ChatID) == "" {
		return errors.New("Telegram channel is not configured")
	}
	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	baseURL := strings.TrimRight(c.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.telegram.org"
	}
	safe := Sanitize(message)
	body, err := json.Marshal(map[string]any{
		"chat_id":                  c.ChatID,
		"text":                     strings.TrimSpace(safe.Title + "\n" + safe.Body),
		"disable_web_page_preview": true,
	})
	if err != nil {
		return errors.New("encode Telegram notification")
	}
	endpoint := baseURL + "/bot" + c.BotToken + "/sendMessage"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return errors.New("build Telegram notification request")
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return errors.New("Telegram delivery request failed")
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 4096))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("Telegram delivery returned HTTP %d", res.StatusCode)
	}
	return nil
}

type InAppStore interface {
	CreateNotification(context.Context, Message) error
}

type InAppChannel struct {
	Store InAppStore
}

func (c *InAppChannel) Name() string { return "in_app" }

func (c *InAppChannel) Send(ctx context.Context, message Message) error {
	if c == nil || c.Store == nil {
		return errors.New("in-app notification store is unavailable")
	}
	return c.Store.CreateNotification(ctx, Sanitize(message))
}
