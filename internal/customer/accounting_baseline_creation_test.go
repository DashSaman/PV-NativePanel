package customer

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

func assertKnownZeroBaseline(t *testing.T, got AccountingBaseline, cutoff time.Time) {
	t.Helper()
	if got.State != AccountingBaselineKnown || got.Source != AccountingBaselineFreshManagedTerm {
		t.Fatalf("baseline state/source = %#v", got)
	}
	if !got.CutoffAt.Equal(cutoff) {
		t.Fatalf("baseline cutoff = %v, want %v", got.CutoffAt, cutoff)
	}
	if got.UploadBytes == nil || got.DownloadBytes == nil || *got.UploadBytes != 0 || *got.DownloadBytes != 0 {
		t.Fatalf("fresh baseline bytes = %#v", got)
	}
}

func TestCreateCustomerRecordsProvableKnownZeroBaseline(t *testing.T) {
	now := time.Date(2026, 8, 30, 1, 0, 0, 0, time.UTC)
	store := &fakeCustomerStore{}
	mutation := &fakeCustomerRuntimeMutation{view: runtimecred.CredentialView{ID: "runtime-fresh", Username: "fresh"}}
	service := NewService(store, func(context.Context, *sql.Tx, string, string, runtimecred.CreateInput) (RuntimeMutation, error) {
		return mutation, nil
	}, func() time.Time { return now })

	_, err := service.CreateCustomer(context.Background(), nil, "owner-1", "baseline-create-0001", CreateCustomerInput{
		Username: "fresh", Password: "valid password 123", Validity: ValidityInput{Mode: ValidityOnCreation, DurationDays: 30},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertKnownZeroBaseline(t, store.createdTerm.AccountingBaseline, now)
}

func TestAdoptRuntimeRecordsUnknownHistoricalBaselineAtTrustedBoundary(t *testing.T) {
	now := time.Date(2026, 8, 30, 1, 30, 0, 0, time.UTC)
	store := &adoptUpdateCustomerStore{
		adoptable: runtimecred.CredentialView{ID: "runtime-legacy", Username: "legacy", Status: runtimecred.CredentialActive, Origin: runtimecred.CredentialImported},
	}
	service := NewService(store, func(context.Context, *sql.Tx, string, string, runtimecred.CreateInput) (RuntimeMutation, error) {
		return nil, errors.New("adoption must not create runtime")
	}, func() time.Time { return now })

	_, err := service.AdoptRuntimeCredential(context.Background(), nil, "owner-1", AdoptRuntimeInput{
		RuntimeCredentialID: "runtime-legacy",
		Validity:            ValidityInput{Mode: ValidityOnCreation, DurationDays: 30},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := store.createdTerm.AccountingBaseline
	if got.State != AccountingBaselineUnknown || got.Source != AccountingBaselineLegacyUnavailable || !got.CutoffAt.Equal(now) {
		t.Fatalf("adopted baseline = %#v", got)
	}
	if got.UploadBytes != nil || got.DownloadBytes != nil {
		t.Fatalf("adoption fabricated historical bytes: %#v", got)
	}
}

func TestRenewalStartsNewManagedAccountingEpochAtKnownZero(t *testing.T) {
	now := time.Date(2026, 8, 30, 2, 0, 0, 0, time.UTC)
	quota := int64(10 * BytesPerCustomerGB)
	store := &renewalStoreStub{
		ctx: RenewalContext{
			TenantID: "tenant", UserID: "user", RuntimeCredentialID: "runtime-stable",
			Current: ServiceTerm{ID: "old-term", UserID: "user", State: TermEnded},
		},
		plan: PlanPreset{ID: "plan", QuotaBytes: &quota, ValiditySeconds: 30 * 86400, StartPolicy: StartOnCreation, Enabled: true},
	}
	service := &Service{store: store, now: func() time.Time { return now }}
	if _, err := service.RenewCustomer(context.Background(), &sql.Tx{}, "actor", "user", RenewalInput{Mode: RenewalUsingPlan, PlanID: "plan"}); err != nil {
		t.Fatal(err)
	}
	assertKnownZeroBaseline(t, store.created.AccountingBaseline, now)
}

type baselineProductStore struct{}

func (baselineProductStore) OperationTenantIDTx(context.Context, *sql.Tx) (string, error) {
	return "tenant", nil
}
func (baselineProductStore) PlanByIDTx(context.Context, *sql.Tx, string) (PlanPreset, error) {
	return PlanPreset{}, errors.New("plan lookup not expected")
}
func (baselineProductStore) CreateProductServiceTermTx(context.Context, *sql.Tx, CreateServiceTermRecord) (ServiceTerm, error) {
	return ServiceTerm{}, errors.New("not used")
}
func (baselineProductStore) ApplyCustomerMetadataTx(context.Context, *sql.Tx, string, string, string, string, []string, bool) error {
	return errors.New("not used")
}

func TestProductCreateTermStartsKnownZeroAccountingEpoch(t *testing.T) {
	now := time.Date(2026, 8, 30, 2, 30, 0, 0, time.UTC)
	quota := int64(25)
	service := &Service{}
	record, _, _, err := service.productCreateTerm(context.Background(), &sql.Tx{}, baselineProductStore{}, "tenant", "fresh-product", ProductCreateCustomerInput{
		QuotaGB:  &quota,
		Validity: ValidityInput{Mode: ValidityOnCreation, DurationDays: 30},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	assertKnownZeroBaseline(t, record.AccountingBaseline, now)
}
