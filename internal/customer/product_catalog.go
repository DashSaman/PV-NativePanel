package customer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type productCatalogStore interface {
	OperationTenantIDTx(context.Context, *sql.Tx) (string, error)
	ListPlansTx(context.Context, *sql.Tx) ([]PlanPreset, error)
	CreatePlanTx(context.Context, *sql.Tx, string, string, string, PlanPreset) (PlanPreset, error)
	ListGroupsTx(context.Context, *sql.Tx) ([]CustomerGroup, error)
	CreateGroupTx(context.Context, *sql.Tx, string, string, string, int) (CustomerGroup, error)
	ListTagsTx(context.Context, *sql.Tx) ([]CustomerTag, error)
	CreateTagTx(context.Context, *sql.Tx, string, string, string, int) (CustomerTag, error)
}

func (s *Service) ListPlans(ctx context.Context, tx *sql.Tx) ([]PlanPreset, error) {
	store, ok := s.store.(productCatalogStore)
	if !ok || tx == nil {
		return nil, errors.New("customer: plan capability is unavailable")
	}
	return store.ListPlansTx(ctx, tx)
}

func (s *Service) CreatePlan(ctx context.Context, tx *sql.Tx, actorID string, plan PlanPreset) (PlanPreset, error) {
	store, ok := s.store.(productCatalogStore)
	if !ok || tx == nil || strings.TrimSpace(actorID) == "" {
		return PlanPreset{}, errors.New("customer: plan capability is unavailable")
	}
	if err := plan.Validate(); err != nil {
		return PlanPreset{}, err
	}
	tenantID, err := store.OperationTenantIDTx(ctx, tx)
	if err != nil {
		return PlanPreset{}, err
	}
	code := fmt.Sprintf("p%x", s.now().UTC().UnixNano())
	return store.CreatePlanTx(ctx, tx, tenantID, actorID, code, plan)
}

func (s *Service) ListGroups(ctx context.Context, tx *sql.Tx) ([]CustomerGroup, error) {
	store, ok := s.store.(productCatalogStore)
	if !ok || tx == nil {
		return nil, errors.New("customer: group capability is unavailable")
	}
	return store.ListGroupsTx(ctx, tx)
}

func (s *Service) CreateGroup(ctx context.Context, tx *sql.Tx, actorID, name string, sortOrder int) (CustomerGroup, error) {
	store, ok := s.store.(productCatalogStore)
	if !ok || tx == nil || strings.TrimSpace(actorID) == "" {
		return CustomerGroup{}, errors.New("customer: group capability is unavailable")
	}
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 120 {
		return CustomerGroup{}, errors.New("customer: group name is invalid")
	}
	tenantID, err := store.OperationTenantIDTx(ctx, tx)
	if err != nil {
		return CustomerGroup{}, err
	}
	return store.CreateGroupTx(ctx, tx, tenantID, actorID, name, sortOrder)
}

func (s *Service) ListTags(ctx context.Context, tx *sql.Tx) ([]CustomerTag, error) {
	store, ok := s.store.(productCatalogStore)
	if !ok || tx == nil {
		return nil, errors.New("customer: tag capability is unavailable")
	}
	return store.ListTagsTx(ctx, tx)
}

func (s *Service) CreateTag(ctx context.Context, tx *sql.Tx, actorID, name string, sortOrder int) (CustomerTag, error) {
	store, ok := s.store.(productCatalogStore)
	if !ok || tx == nil || strings.TrimSpace(actorID) == "" {
		return CustomerTag{}, errors.New("customer: tag capability is unavailable")
	}
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 80 {
		return CustomerTag{}, errors.New("customer: tag name is invalid")
	}
	tenantID, err := store.OperationTenantIDTx(ctx, tx)
	if err != nil {
		return CustomerTag{}, err
	}
	return store.CreateTagTx(ctx, tx, tenantID, actorID, name, sortOrder)
}

func planDurationDays(seconds int64) int {
	if seconds <= 0 {
		return 0
	}
	return int(time.Duration(seconds*int64(time.Second)) / (24 * time.Hour))
}
