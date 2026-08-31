package customer

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

type task14ProductStore struct{ plan PlanPreset }

func (s task14ProductStore) OperationTenantIDTx(context.Context, *sql.Tx) (string, error) {
	return "tenant", nil
}
func (s task14ProductStore) PlanByIDTx(context.Context, *sql.Tx, string) (PlanPreset, error) {
	return s.plan, nil
}
func (s task14ProductStore) CreateProductServiceTermTx(context.Context, *sql.Tx, CreateServiceTermRecord) (ServiceTerm, error) {
	return ServiceTerm{}, errors.New("not used")
}
func (s task14ProductStore) ApplyCustomerMetadataTx(context.Context, *sql.Tx, string, string, string, string, []string, bool) error {
	return errors.New("not used")
}

func TestProductCreateFromPlanCarriesConcurrencyLimit(t *testing.T) {
	limit := 2
	plan := PlanPreset{
		ID: "plan-limited", Name: "Limited", ValiditySeconds: 30 * 86400,
		StartPolicy: StartOnCreation, ResetStrategy: ResetNone, Enabled: true,
		ConcurrencyLimit: &limit,
	}
	service := &Service{}
	record, _, _, err := service.productCreateTerm(context.Background(), &sql.Tx{}, task14ProductStore{plan: plan}, "tenant", "user", ProductCreateCustomerInput{PlanID: plan.ID}, time.Unix(1_700_000_000, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	if record.ConcurrencyLimit == nil || *record.ConcurrencyLimit != 2 {
		t.Fatalf("plan concurrency limit not copied to new term: %#v", record.ConcurrencyLimit)
	}
}

func TestRenewalsPreserveConcurrencyLimit(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	limitCurrent := 1
	store := &renewalStoreStub{
		ctx: RenewalContext{TenantID: "tenant", UserID: "user", RuntimeCredentialID: "runtime", Current: ServiceTerm{ID: "old", UserID: "user", State: TermActive, DurationSeconds: 86400, StartPolicy: StartOnCreation, ConcurrencyLimit: &limitCurrent}},
	}
	service := &Service{store: store, now: func() time.Time { return now }}
	if _, err := service.RenewCustomer(context.Background(), &sql.Tx{}, "actor", "user", RenewalInput{Mode: RenewalCurrent}); err != nil {
		t.Fatal(err)
	}
	if store.created.ConcurrencyLimit == nil || *store.created.ConcurrencyLimit != 1 {
		t.Fatalf("renew-current lost concurrency limit: %#v", store.created.ConcurrencyLimit)
	}

	limitPlan := 4
	store.created = CreateRenewalTermRecord{}
	store.plan = PlanPreset{ID: "plan-4", Name: "Four", ValiditySeconds: 86400, StartPolicy: StartOnCreation, ResetStrategy: ResetNone, Enabled: true, ConcurrencyLimit: &limitPlan}
	if _, err := service.RenewCustomer(context.Background(), &sql.Tx{}, "actor", "user", RenewalInput{Mode: RenewalUsingPlan, PlanID: "plan-4"}); err != nil {
		t.Fatal(err)
	}
	if store.created.ConcurrencyLimit == nil || *store.created.ConcurrencyLimit != 4 {
		t.Fatalf("renew-plan lost concurrency limit: %#v", store.created.ConcurrencyLimit)
	}
}

func TestPostgresStoresPersistConcurrencyLimit(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve source")
	}
	dir := filepath.Dir(filename)
	for _, name := range []string{"product_catalog_store.go", "renewal_store.go", "product_create_store.go", "store.go"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "concurrency_limit") {
			t.Fatalf("%s does not persist/read concurrency_limit", name)
		}
	}
}
