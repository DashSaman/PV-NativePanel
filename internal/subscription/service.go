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

type Record struct {
	RuntimeCredentialID string
	Username            string
	SecretCiphertext    []byte
	SecretNonce         []byte
	EncryptionKeyID     string
	UserState           string
	TermState           string
	ExpiresAt           *time.Time
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
	hash, err := HashToken(rawToken)
	if err != nil {
		return "", ErrUnavailable
	}
	record, err := s.store.ResolveToken(ctx, hash)
	if err != nil {
		return "", ErrUnavailable
	}
	if record.UserState != "active" || (record.TermState != "active" && record.TermState != "pending") {
		return "", ErrUnavailable
	}
	if record.ExpiresAt != nil && !record.ExpiresAt.After(s.now().UTC()) {
		return "", ErrUnavailable
	}
	if record.EncryptionKeyID != s.keyID || record.Username == "" {
		return "", ErrUnavailable
	}
	plaintext, err := runtimecred.DecryptSecret(s.key, record.SecretNonce, record.SecretCiphertext)
	if err != nil {
		return "", ErrUnavailable
	}
	password := string(plaintext)
	for i := range plaintext {
		plaintext[i] = 0
	}
	uri, err := BuildNaiveURI(record.Username, password, host)
	password = ""
	if err != nil {
		return "", ErrUnavailable
	}
	return uri, nil
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
