package notification

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/observability"
)

type EventType string

const (
	ExpiryWarning    EventType = "expiry_warning"
	Expired          EventType = "expired"
	QuotaWarning     EventType = "quota_warning"
	QuotaDepleted    EventType = "quota_depleted"
	RuntimeDown      EventType = "runtime_down"
	RuntimeRecovered EventType = "runtime_recovered"
	DBIssue          EventType = "db_issue"
	BackupFailure    EventType = "backup_failure"
	BackupSuccess    EventType = "backup_success"
	UpdateAvailable  EventType = "update_available"
	UpdateFailure    EventType = "update_failure"
)

type Capabilities struct {
	ExactAccounting bool
}

func Enabled(eventType EventType, capabilities Capabilities) bool {
	switch eventType {
	case QuotaWarning, QuotaDepleted:
		return capabilities.ExactAccounting
	default:
		return true
	}
}

type Event struct {
	Type  EventType
	Key   string
	Title string
	Body  string
}

type Message struct {
	Title string
	Body  string
}

type Channel interface {
	Name() string
	Send(context.Context, Message) error
}

type Config struct {
	MaxAttempts int
	DedupeTTL   time.Duration
	BaseBackoff time.Duration
	Now         func() time.Time
	Sleep       func(context.Context, time.Duration) error
}

type Engine struct {
	config Config
	mu     sync.Mutex
	sent   map[string]time.Time
}

func NewEngine(config Config) *Engine {
	if config.MaxAttempts < 1 {
		config.MaxAttempts = 3
	}
	if config.DedupeTTL <= 0 {
		config.DedupeTTL = 5 * time.Minute
	}
	if config.BaseBackoff <= 0 {
		config.BaseBackoff = 500 * time.Millisecond
	}
	if config.Now == nil {
		config.Now = time.Now
	}
	if config.Sleep == nil {
		config.Sleep = sleepContext
	}
	return &Engine{config: config, sent: make(map[string]time.Time)}
}

func (e *Engine) Deliver(ctx context.Context, channel Channel, event Event) error {
	if e == nil || channel == nil {
		return errors.New("notification channel is unavailable")
	}
	if event.Type == "" || event.Key == "" {
		return errors.New("notification event type and key are required")
	}
	key := channel.Name() + ":" + string(event.Type) + ":" + event.Key
	now := e.config.Now().UTC()
	if e.isDuplicate(key, now) {
		return nil
	}
	message := Sanitize(Message{Title: event.Title, Body: event.Body})
	var lastErr error
	for attempt := 1; attempt <= e.config.MaxAttempts; attempt++ {
		if err := channel.Send(ctx, message); err == nil {
			e.markDelivered(key, e.config.Now().UTC())
			return nil
		} else {
			lastErr = err
		}
		if attempt == e.config.MaxAttempts {
			break
		}
		delay := e.config.BaseBackoff * time.Duration(1<<(attempt-1))
		if err := e.config.Sleep(ctx, delay); err != nil {
			return err
		}
	}
	return fmt.Errorf("notification delivery failed after %d attempts: %w", e.config.MaxAttempts, lastErr)
}

func (e *Engine) isDuplicate(key string, now time.Time) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	last, ok := e.sent[key]
	return ok && now.Sub(last) < e.config.DedupeTTL
}

func (e *Engine) markDelivered(key string, now time.Time) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.sent[key] = now
	for storedKey, last := range e.sent {
		if now.Sub(last) >= e.config.DedupeTTL*2 {
			delete(e.sent, storedKey)
		}
	}
}

func Sanitize(message Message) Message {
	message.Title = observability.RedactText(message.Title)
	message.Body = observability.RedactText(message.Body)
	return message
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
