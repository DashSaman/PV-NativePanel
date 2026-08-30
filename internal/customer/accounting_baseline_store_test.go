package customer

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"
)

func baselineStoreFixture() (AccountingBaseline, time.Time) {
	now := time.Date(2026, 8, 30, 3, 0, 0, 0, time.UTC)
	zeroUpload, zeroDownload := int64(0), int64(0)
	return AccountingBaseline{
		State:         AccountingBaselineKnown,
		Source:        AccountingBaselineFreshManagedTerm,
		CutoffAt:      now,
		UploadBytes:   &zeroUpload,
		DownloadBytes: &zeroDownload,
	}, now
}

func TestCreateServiceTermPersistsAndReturnsAccountingBaseline(t *testing.T) {
	baseline, now := baselineStoreFixture()
	quota := int64(1000)
	conn := &customerScriptConn{queries: []customerScriptQuery{{
		contains: "accounting_baseline_state",
		columns:  []string{"id", "tenant_id", "user_id", "quota_bytes", "duration_seconds", "start_policy", "purchased_at", "starts_at", "first_connected_at", "expires_at", "state", "accounting_baseline_state", "accounting_baseline_source", "accounting_baseline_cutoff_at", "accounting_baseline_upload_bytes", "accounting_baseline_download_bytes", "revision"},
		values:   []driver.Value{"term-1", "tenant-1", "user-1", quota, int64(3600), "on_creation", now, now, nil, now.Add(time.Hour), "active", "known", "fresh_managed_term", now, int64(0), int64(0), int64(1)},
	}}}
	tx := newCustomerStoreTx(t, conn)
	term, err := NewPostgresStore().CreateServiceTermTx(context.Background(), tx, CreateServiceTermRecord{
		TenantID: "tenant-1", UserID: "user-1", QuotaBytes: &quota, DurationSeconds: 3600,
		StartPolicy: StartOnCreation, PurchasedAt: now, StartsAt: &now, ExpiresAt: baselinePtrTime(now.Add(time.Hour)),
		State: TermActive, AccountingBaseline: baseline,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertKnownZeroBaseline(t, term.AccountingBaseline, now)
}

func TestCreateProductServiceTermPersistsAndReturnsAccountingBaseline(t *testing.T) {
	baseline, now := baselineStoreFixture()
	quota := int64(2000)
	conn := &customerScriptConn{queries: []customerScriptQuery{{
		contains: "accounting_baseline_state",
		columns:  []string{"id", "tenant_id", "user_id", "plan_id", "quota_bytes", "duration_seconds", "no_expiry", "start_policy", "purchased_at", "starts_at", "first_connected_at", "expires_at", "state", "renewal_kind", "renewed_from_term_id", "accounting_baseline_state", "accounting_baseline_source", "accounting_baseline_cutoff_at", "accounting_baseline_upload_bytes", "accounting_baseline_download_bytes", "revision"},
		values:   []driver.Value{"term-product", "tenant-1", "user-1", nil, quota, int64(3600), false, "on_creation", now, now, nil, now.Add(time.Hour), "active", "initial", nil, "known", "fresh_managed_term", now, int64(0), int64(0), int64(1)},
	}}}
	tx := newCustomerStoreTx(t, conn)
	term, err := NewPostgresStore().CreateProductServiceTermTx(context.Background(), tx, CreateServiceTermRecord{
		TenantID: "tenant-1", UserID: "user-1", QuotaBytes: &quota, DurationSeconds: 3600,
		StartPolicy: StartOnCreation, PurchasedAt: now, StartsAt: &now, ExpiresAt: baselinePtrTime(now.Add(time.Hour)),
		State: TermActive, RenewalKind: "initial", AccountingBaseline: baseline,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertKnownZeroBaseline(t, term.AccountingBaseline, now)
}

func TestCreateRenewalTermPersistsAndReturnsAccountingBaseline(t *testing.T) {
	baseline, now := baselineStoreFixture()
	quota := int64(3000)
	conn := &customerScriptConn{queries: []customerScriptQuery{{
		contains: "accounting_baseline_state",
		columns:  []string{"id", "tenant_id", "user_id", "plan_id", "quota_bytes", "duration_seconds", "no_expiry", "start_policy", "purchased_at", "starts_at", "first_connected_at", "expires_at", "state", "renewal_kind", "renewed_from_term_id", "accounting_baseline_state", "accounting_baseline_source", "accounting_baseline_cutoff_at", "accounting_baseline_upload_bytes", "accounting_baseline_download_bytes", "revision"},
		values:   []driver.Value{"term-renew", "tenant-1", "user-1", nil, quota, int64(3600), false, "on_creation", now, now, nil, now.Add(time.Hour), "active", "renew_current", "old-term", "known", "fresh_managed_term", now, int64(0), int64(0), int64(1)},
	}}}
	tx := newCustomerStoreTx(t, conn)
	term, err := NewPostgresStore().CreateRenewalTermTx(context.Background(), tx, CreateRenewalTermRecord{
		TenantID: "tenant-1", UserID: "user-1", QuotaBytes: &quota, DurationSeconds: 3600,
		StartPolicy: StartOnCreation, PurchasedAt: now, StartsAt: &now, ExpiresAt: baselinePtrTime(now.Add(time.Hour)),
		State: TermActive, RenewalKind: "renew_current", RenewedFromTermID: "old-term", AccountingBaseline: baseline,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertKnownZeroBaseline(t, term.AccountingBaseline, now)
}

func baselinePtrTime(value time.Time) *time.Time { return &value }
