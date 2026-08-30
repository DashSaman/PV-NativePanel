package subscription

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

const tokenBytes = 32

var (
	ErrInvalidToken = errors.New("subscription: invalid token")
	ErrUnavailable  = errors.New("subscription: unavailable")
)

type AccountingBaseline struct {
	State         string
	Source        string
	CutoffAt      time.Time
	UploadBytes   *int64
	DownloadBytes *int64
}

type Record struct {
	ServiceTermID       string
	RuntimeCredentialID string
	Username            string
	SecretCiphertext    []byte
	SecretNonce         []byte
	EncryptionKeyID     string
	UserState           string
	TermState           string
	QuotaBytes          *int64
	DurationSeconds     int64
	StartPolicy         string
	StartsAt            *time.Time
	FirstConnectedAt    *time.Time
	ExpiresAt           *time.Time
	AccountingBaseline  AccountingBaseline
}

type Profile struct {
	ServiceTermID       string
	RuntimeCredentialID string
	Username            string
	UserState           string
	TermState           string
	QuotaBytes          *int64
	DurationSeconds     int64
	StartPolicy         string
	StartsAt            *time.Time
	FirstConnectedAt    *time.Time
	ExpiresAt           *time.Time
	DirectURI           string
	Available           bool
	UsageAvailable      bool
	UsedBytes           *int64
	RemainingBytes      *int64
	AccountingBaseline  AccountingBaseline
}

type Store interface {
	ResolveToken(context.Context, [32]byte) (Record, error)
}

type Service struct {
	store Store
	key   []byte
	keyID string
	now   func() time.Time
}

func GenerateToken() (string, [32]byte, error) {
	var raw [tokenBytes]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", [32]byte{}, fmt.Errorf("subscription: generate token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw[:])
	hash := sha256.Sum256(raw[:])
	for i := range raw {
		raw[i] = 0
	}
	return token, hash, nil
}

func HashToken(token string) ([32]byte, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(decoded) != tokenBytes || base64.RawURLEncoding.EncodeToString(decoded) != token {
		return [32]byte{}, ErrInvalidToken
	}
	hash := sha256.Sum256(decoded)
	for i := range decoded {
		decoded[i] = 0
	}
	return hash, nil
}

func NewService(store Store, key []byte, keyID string) (*Service, error) {
	if store == nil {
		return nil, errors.New("subscription: store is required")
	}
	if len(key) != 32 {
		return nil, errors.New("subscription: runtime encryption key must be 32 bytes")
	}
	if strings.TrimSpace(keyID) == "" {
		return nil, errors.New("subscription: runtime encryption key id is required")
	}
	keyCopy := append([]byte(nil), key...)
	return &Service{store: store, key: keyCopy, keyID: keyID, now: time.Now}, nil
}

func (s *Service) Resolve(ctx context.Context, rawToken, host string) (string, error) {
	profile, err := s.ResolveProfile(ctx, rawToken, host)
	if err != nil || !profile.Available || profile.DirectURI == "" {
		return "", ErrUnavailable
	}
	return profile.DirectURI, nil
}

func (s *Service) ResolveProfile(ctx context.Context, rawToken, host string) (Profile, error) {
	hash, err := HashToken(rawToken)
	if err != nil {
		return Profile{}, ErrUnavailable
	}
	record, err := s.store.ResolveToken(ctx, hash)
	if err != nil {
		return Profile{}, ErrUnavailable
	}
	if record.EncryptionKeyID != s.keyID || record.Username == "" {
		return Profile{}, ErrUnavailable
	}

	profile := Profile{
		ServiceTermID:       record.ServiceTermID,
		RuntimeCredentialID: record.RuntimeCredentialID,
		Username:            record.Username,
		UserState:           record.UserState,
		TermState:           record.TermState,
		QuotaBytes:          record.QuotaBytes,
		DurationSeconds:     record.DurationSeconds,
		StartPolicy:         record.StartPolicy,
		StartsAt:            record.StartsAt,
		FirstConnectedAt:    record.FirstConnectedAt,
		ExpiresAt:           record.ExpiresAt,
		AccountingBaseline:  cloneAccountingBaseline(record.AccountingBaseline),
		UsageAvailable:      false,
	}
	profile.Available = record.UserState == "active" &&
		(record.TermState == "active" || record.TermState == "pending") &&
		(record.ExpiresAt == nil || record.ExpiresAt.After(s.now().UTC()))
	if !profile.Available {
		return profile, nil
	}

	plaintext, err := runtimecred.DecryptSecret(s.key, record.SecretNonce, record.SecretCiphertext)
	if err != nil {
		return Profile{}, ErrUnavailable
	}
	password := string(plaintext)
	for i := range plaintext {
		plaintext[i] = 0
	}
	uri, err := BuildNaiveURI(record.Username, password, host)
	password = ""
	if err != nil {
		return Profile{}, ErrUnavailable
	}
	profile.DirectURI = uri
	return profile, nil
}

func cloneAccountingBaseline(value AccountingBaseline) AccountingBaseline {
	value.CutoffAt = value.CutoffAt.UTC()
	if value.UploadBytes != nil {
		v := *value.UploadBytes
		value.UploadBytes = &v
	}
	if value.DownloadBytes != nil {
		v := *value.DownloadBytes
		value.DownloadBytes = &v
	}
	return value
}

func BuildNaiveURI(username, password, host string) (string, error) {
	username = strings.TrimSpace(username)
	host = strings.TrimSpace(host)
	if username == "" || password == "" || host == "" {
		return "", errors.New("subscription: username, password and host are required")
	}
	probe, err := url.Parse("https://" + host)
	if err != nil || probe.Host == "" || probe.User != nil || probe.Path != "" || probe.RawQuery != "" || probe.Fragment != "" {
		return "", errors.New("subscription: invalid proxy host")
	}
	u := &url.URL{
		Scheme: "naive+https",
		Host:   probe.Host,
		User:   url.UserPassword(username, password),
	}
	return u.String(), nil
}
