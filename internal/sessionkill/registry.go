// Package sessionkill provides the isolated, exact-session disconnect
// primitive for PVNaive.
//
// A live Naive CONNECT is identified by the full opaque 4-tuple that the
// forwardproxy overlay already uses for exact accounting:
//
//	(RuntimeCredentialID, NodeID, BootID, SessionID)
//
// The SessionID is a random UUID generated in-process by forwardproxy at
// CONNECT time and is never guessable or re-usable. Because each live CONNECT
// therefore owns a unique tuple, a kill addressed with the exact tuple targets
// exactly one live stream, and a forged or partial tuple can never match a
// session it does not own.
//
// The registry is the in-process map the data plane uses to record live
// streams and to translate an exact-tuple kill request into terminating that
// one stream. The stream termination is delivered through a per-session
// cancel function; the forwardproxy overlay arranges for that cancellation to
// make the stream copy loop return so the normal exact final-accounting close
// path runs. Killing never revokes a whole credential and never reloads or
// restarts Caddy.
//
// This control contract is intentionally narrow and local-only: it accepts
// only an exact full tuple, it is idempotent, and it is authenticated by
// construction (only callers able to observe a session ID may target it). It
// is never on the public data-plane path.
package sessionkill

import "sync"

// Key is the full opaque session identity tuple. Two different live CONNECTs
// never share a Key because SessionID is unique per CONNECT.
type Key struct {
	RuntimeCredentialID string
	NodeID              string
	BootID              string
	SessionID           string
}

// KillResult reports the outcome of an exact-tuple kill request.
type KillResult struct {
	// Found reports whether a matching live session was registered.
	Found bool
	// Killed reports whether this request actually cancelled the stream
	// handle. It is false for an already-cancelled (idempotent repeat) or
	// already-closed session.
	Killed bool
}

type session struct {
	key    Key
	cancel func(Key)
	killed bool

	cancelOnce sync.Once
}

// Registry maps exact session tuples to live stream handles. It is safe for
// concurrent use.
type Registry struct {
	mu   sync.Mutex
	live map[Key]*session
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{live: make(map[Key]*session)}
}

// Register records a live stream under its exact tuple and returns an
// unregister closure to call when the stream ends normally. Registering the
// same tuple again atomically replaces the prior handle so that exactly one
// cancel can ever fire for the current handle.
func (r *Registry) Register(key Key, cancel func(Key)) (unregister func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.live[key] = &session{key: key, cancel: cancel}
	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.live[key] != nil && r.live[key].key == key {
			delete(r.live, key)
		}
	}
}

// Unregister removes a live session by its exact tuple. It is a no-op if the
// tuple is not registered. The returned closure from Register is preferred to
// calling Unregister directly because it removes only the matching handle.
func (r *Registry) Unregister(key Key) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.live, key)
}

// Kill terminates the single live stream whose exact tuple equals key.
//
// It is idempotent: killing an already-cancelled session (one still tearing
// down after a prior kill) is a harmless no-op with Killed=false but Found
// still true; killing a session that was never registered, or already
// unregistered after a normal close, reports Found=false. It is never a
// credential revocation and never touches any other stream, even one sharing
// the same user/runtime-credential/node/boot.
func (r *Registry) Kill(key Key) (KillResult, error) {
	r.mu.Lock()
	sess := r.live[key]
	if sess == nil {
		r.mu.Unlock()
		return KillResult{}, nil
	}
	if sess.killed {
		r.mu.Unlock()
		return KillResult{Found: true}, nil
	}
	sess.killed = true
	target := sess
	r.mu.Unlock()

	target.cancelOnce.Do(func() {
		if target.cancel != nil {
			target.cancel(key)
		}
	})
	return KillResult{Found: true, Killed: true}, nil
}

// IsLive reports whether an exact tuple currently has a registered, not-yet-
// killed stream.
func (r *Registry) IsLive(key Key) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	sess := r.live[key]
	return sess != nil && !sess.killed
}

// Count returns the number of currently registered, still-live (not killed)
// streams.
func (r *Registry) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, sess := range r.live {
		if !sess.killed {
			n++
		}
	}
	return n
}
