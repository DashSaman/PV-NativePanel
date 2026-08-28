package runtimeevent

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type FirstUseEvent struct {
	RuntimeCredentialID string
	Method              string
	Authenticated       bool
	ObservedAt          time.Time
}

type FirstUseActivator interface {
	ActivateFirstUse(context.Context, *sql.Tx, string, time.Time) (bool, error)
}

func HandleFirstUse(ctx context.Context, tx *sql.Tx, activator FirstUseActivator, event FirstUseEvent) (bool, error) {
	if activator == nil {
		return false, errors.New("runtimeevent: first-use activator is required")
	}
	if !event.Authenticated || strings.ToUpper(strings.TrimSpace(event.Method)) != "CONNECT" {
		return false, nil
	}
	credentialID := strings.TrimSpace(event.RuntimeCredentialID)
	if credentialID == "" {
		return false, errors.New("runtimeevent: runtime credential id is required")
	}
	if event.ObservedAt.IsZero() {
		return false, errors.New("runtimeevent: observed time is required")
	}
	return activator.ActivateFirstUse(ctx, tx, credentialID, event.ObservedAt.UTC())
}
