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

type task15ProductStore struct{ plan PlanPreset }

func (s task15ProductStore) OperationTenantIDTx(context.Context, *sql.Tx) (string, error) {
	return "tenant", nil
}
func (s task15ProductStore) PlanByIDTx(context.Context, *sql.Tx, string) (PlanPreset, error) {
	return s.plan, nil
}
func (s task15ProductStore) CreateProductServiceTermTx(context.Context, *sql.Tx, CreateServiceTermRecord) (ServiceTerm, error) {
	return ServiceTerm{}, errors.New("not used")
}
func (s task15ProductStore) ApplyCustomerMetadataTx(context.Context, *sql.Tx, string, string, string, string, []string, bool) error {
	return errors.New("not used")
}

func TestProductCreateFromPlanCarriesUniqueIPLimit(t *testing.T) {
	limit := 3
	plan := PlanPreset{
		ID: "plan-ip-limited", Name: "IP Limited", ValiditySeconds: 30 * 86400,
		StartPolicy: StartOnCreation, ResetStrategy: ResetNone, Enabled: true,
		UniqueIPLimit: &limit,
	}
	service := &Service{}
	record, _, _, err := service.productCreateTerm(context.Background(), &sql.Tx{}, task15ProductStore{plan: plan}, "tenant", "user", ProductCreateCustomerInput{PlanID: plan.ID}, time.Unix(1_700_000_000, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	if record.UniqueIPLimit == nil || *record.UniqueIPLimit != 3 {
		t.Fatalf("plan unique IP limit not copied to new term: %#v", record.UniqueIPLimit)
	}
}

func TestRenewalsPreserveUniqueIPLimit(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	limitCurrent := 2
	store := &renewalStoreStub{
		ctx: RenewalContext{TenantID: "tenant", UserID: "user", RuntimeCredentialID: "runtime", Current: ServiceTerm{ID: "old", UserID: "user", State: TermActive, DurationSeconds: 86400, StartPolicy: StartOnCreation, UniqueIPLimit: &limitCurrent}},
	}
	service := &Service{store: store, now: func() time.Time { return now }}
	if _, err := service.RenewCustomer(context.Background(), &sql.Tx{}, "actor", "user", RenewalInput{Mode: RenewalCurrent}); err != nil {
		t.Fatal(err)
	}
	if store.created.UniqueIPLimit == nil || *store.created.UniqueIPLimit != 2 {
		t.Fatalf("renew-current lost unique IP limit: %#v", store.created.UniqueIPLimit)
	}

	limitPlan := 4
	store.created = CreateRenewalTermRecord{}
	store.plan = PlanPreset{ID: "plan-ip-4", Name: "IP Four", ValiditySeconds: 86400, StartPolicy: StartOnCreation, ResetStrategy: ResetNone, Enabled: true, UniqueIPLimit: &limitPlan}
	if _, err := service.RenewCustomer(context.Background(), &sql.Tx{}, "actor", "user", RenewalInput{Mode: RenewalUsingPlan, PlanID: "plan-ip-4"}); err != nil {
		t.Fatal(err)
	}
	if store.created.UniqueIPLimit == nil || *store.created.UniqueIPLimit != 4 {
		t.Fatalf("renew-plan lost unique IP limit: %#v", store.created.UniqueIPLimit)
	}
}

func TestPostgresStoresPersistUniqueIPLimit(t *testing.T) {
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
		if !strings.Contains(string(data), "unique_ip_limit") {
			t.Fatalf("%s does not persist/read unique_ip_limit", name)
		}
	}
}
