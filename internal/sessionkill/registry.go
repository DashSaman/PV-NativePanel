// Package sessionkill provides the isolated, exact-session disconnect primitive for PVNaive.
package sessionkill

import "sync"

type Key struct {
	RuntimeCredentialID string
	NodeID string
	BootID string
	SessionID string
}

type KillResult struct { Found bool; Killed bool }

type session struct {
	key Key
	cancel func(Key)
	killed bool
	cancelOnce sync.Once
}

type Registry struct {
	mu sync.Mutex
	live map[Key]*session
}

func New() *Registry { return &Registry{live: make(map[Key]*session)} }

func (r *Registry) Register(key Key, cancel func(Key)) (unregister func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.live[key] = &session{key: key, cancel: cancel}
	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.live[key] != nil && r.live[key].key == key { delete(r.live, key) }
	}
}

func (r *Registry) Unregister(key Key) {
	r.mu.Lock(); defer r.mu.Unlock(); delete(r.live, key)
}

func (r *Registry) Kill(key Key) (KillResult, error) {
	r.mu.Lock()
	sess := r.live[key]
	if sess == nil { r.mu.Unlock(); return KillResult{}, nil }
	if sess.killed { r.mu.Unlock(); return KillResult{Found: true}, nil }
	sess.killed = true
	target := sess
	r.mu.Unlock()
	target.cancelOnce.Do(func() { if target.cancel != nil { target.cancel(key) } })
	return KillResult{Found: true, Killed: true}, nil
}

func (r *Registry) IsLive(key Key) bool {
	r.mu.Lock(); defer r.mu.Unlock(); sess := r.live[key]; return sess != nil && !sess.killed
}

func (r *Registry) Count() int {
	r.mu.Lock(); defer r.mu.Unlock(); n := 0; for _, sess := range r.live { if !sess.killed { n++ } }; return n
}
