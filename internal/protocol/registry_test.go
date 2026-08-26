package protocol

import (
	"context"
	"testing"
)

type fakeAdapter struct{ descriptor Descriptor }
func (f fakeAdapter) Descriptor() Descriptor { return f.descriptor }
func (fakeAdapter) Validate(context.Context, DesiredState) (ValidationResult, error) { return ValidationResult{Valid:true}, nil }
func (fakeAdapter) Stage(context.Context, DesiredState) error { return nil }
func (fakeAdapter) Apply(context.Context, string) error { return nil }
func (fakeAdapter) Rollback(context.Context, string) error { return nil }
func (fakeAdapter) Health(context.Context) (Health, error) { return Health{Status:"ok"}, nil }

func TestRegistryRejectsMissingAccountingDeclaration(t *testing.T) {
	r := NewRegistry()
	err := r.Register(fakeAdapter{Descriptor{ID:"naive", Version:"1", RuntimeFamily:"caddy"}})
	if err == nil { t.Fatal("expected capability validation error") }
}

func TestRegistryRejectsDuplicateProtocol(t *testing.T) {
	r := NewRegistry()
	a := fakeAdapter{Descriptor{ID:"naive", Version:"1", RuntimeFamily:"caddy", Capabilities:Capabilities{Accounting:AccountingExact}}}
	if err := r.Register(a); err != nil { t.Fatal(err) }
	if err := r.Register(a); err == nil { t.Fatal("expected duplicate error") }
}
