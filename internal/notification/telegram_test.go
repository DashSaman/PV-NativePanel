package notification

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTelegramChannelSendsSanitizedPayload(t *testing.T) {
	var path string
	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		data, _ := io.ReadAll(r.Body)
		body = string(data)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	channel := &TelegramChannel{BotToken: "bot-secret-token", ChatID: "123", BaseURL: server.URL, Client: server.Client()}
	message := Message{Title: "Backup warning", Body: "Authorization: Bearer leaked-value password=hunter2"}
	if err := channel.Send(context.Background(), message); err != nil { t.Fatal(err) }
	if path != "/botbot-secret-token/sendMessage" { t.Fatalf("path=%q", path) }
	for _, secret := range []string{"leaked-value", "hunter2"} {
		if strings.Contains(body, secret) { t.Fatalf("Telegram payload leaked %q: %s", secret, body) }
	}
	if !strings.Contains(body, "[REDACTED]") { t.Fatalf("sanitized marker absent: %s", body) }
}

func TestTelegramFailureDoesNotExposeBotToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "bad", http.StatusBadGateway) }))
	defer server.Close()
	const token = "very-secret-bot-token"
	channel := &TelegramChannel{BotToken: token, ChatID: "123", BaseURL: server.URL, Client: server.Client()}
	err := channel.Send(context.Background(), Message{Title: "x", Body: "y"})
	if err == nil { t.Fatal("non-2xx Telegram response must fail") }
	if strings.Contains(err.Error(), token) { t.Fatalf("error leaked bot token: %v", err) }
}
