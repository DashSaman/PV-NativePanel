package httpapi

import (
	"context"
	"net/http"

	"github.com/DashSaman/PV-NaivePanel/internal/auth"
)

type authContextKey struct{}

type authenticatedRequest struct {
	Bound           *auth.AuthenticatedTx
	RawSessionToken string
}

func withAuthenticatedRequest(r *http.Request, bound *auth.AuthenticatedTx, rawSessionToken string) *http.Request {
	value := &authenticatedRequest{Bound: bound, RawSessionToken: rawSessionToken}
	return r.WithContext(context.WithValue(r.Context(), authContextKey{}, value))
}

func authenticatedFromRequest(r *http.Request) (*authenticatedRequest, bool) {
	value, ok := r.Context().Value(authContextKey{}).(*authenticatedRequest)
	return value, ok && value != nil && value.Bound != nil
}
