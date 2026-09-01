package sessionkill

import (
	"sync/atomic"
	"testing"
)

func key(session string) Key {
	return Key{RuntimeCredentialID: "11111111-1111-4111-8111-111111111111", NodeID: "node-a", BootID: "22222222-2222-4222-8222-222222222222", SessionID: session}
}

func TestRegistryExactKillPreservesSiblingAndIsIdempotent(t *testing.T) {
	r := New()
	a, b := key("33333333-3333-4333-8333-333333333333"), key("44444444-4444-4444-8444-444444444444")
	var killedA, killedB atomic.Int32
	unregA := r.Register(a, func(got Key) {
		if got != a {
			t.Errorf("cancel key=%+v", got)
		}
		killedA.Add(1)
	})
	defer unregA()
	unregB := r.Register(b, func(got Key) {
		if got != b {
			t.Errorf("cancel key=%+v", got)
		}
		killedB.Add(1)
	})
	defer unregB()
	first, err := r.Kill(a)
	if err != nil || !first.Found || !first.Killed {
		t.Fatalf("first kill=%+v err=%v", first, err)
	}
	if killedA.Load() != 1 || killedB.Load() != 0 {
		t.Fatalf("cancel counts a=%d b=%d", killedA.Load(), killedB.Load())
	}
	if r.IsLive(a) || !r.IsLive(b) {
		t.Fatalf("live state a=%v b=%v", r.IsLive(a), r.IsLive(b))
	}
	second, err := r.Kill(a)
	if err != nil || !second.Found || second.Killed {
		t.Fatalf("second kill=%+v err=%v", second, err)
	}
	if killedA.Load() != 1 || killedB.Load() != 0 {
		t.Fatalf("idempotence counts a=%d b=%d", killedA.Load(), killedB.Load())
	}
}

func TestRegistryForgedTupleDoesNotMatch(t *testing.T) {
	r := New()
	real := key("33333333-3333-4333-8333-333333333333")
	var cancelled atomic.Int32
	r.Register(real, func(Key) { cancelled.Add(1) })
	forged := real
	forged.BootID = "99999999-9999-4999-8999-999999999999"
	got, err := r.Kill(forged)
	if err != nil || got.Found || got.Killed {
		t.Fatalf("forged kill=%+v err=%v", got, err)
	}
	if cancelled.Load() != 0 || !r.IsLive(real) {
		t.Fatalf("forged kill touched real session")
	}
}

func TestRegistryUnregisterMakesSessionUnkillable(t *testing.T) {
	r := New()
	k := key("33333333-3333-4333-8333-333333333333")
	var cancelled atomic.Int32
	unregister := r.Register(k, func(Key) { cancelled.Add(1) })
	unregister()
	got, err := r.Kill(k)
	if err != nil || got.Found || got.Killed {
		t.Fatalf("kill after unregister=%+v err=%v", got, err)
	}
	if cancelled.Load() != 0 || r.Count() != 0 {
		t.Fatalf("unexpected state cancelled=%d count=%d", cancelled.Load(), r.Count())
	}
}
