package customer

import "time"

// AccountingSnapshot is the trusted Direct Naive read model for one immutable
// ServiceTerm. Known ledger bytes remain visible even when completeness is
// false, but incomplete accounting never claims an exact remaining balance or
// exact presence/quota status.
type AccountingSnapshot struct {
	Present            bool
	AccountingComplete bool
	UploadBytes        int64
	DownloadBytes      int64
	UsedBytes          int64
	RemainingBytes     *int64
	Online             bool
	OnlineSessions     int64
	LastOnline         *time.Time
}

func ApplyAccountingSnapshot(view *CustomerView, snapshot AccountingSnapshot) {
	if view == nil {
		return
	}
	view.UploadBytes = snapshot.UploadBytes
	view.DownloadBytes = snapshot.DownloadBytes
	view.UsedBytes = snapshot.UsedBytes
	view.RemainingBytes = snapshot.RemainingBytes
	view.AccountingComplete = snapshot.Present && snapshot.AccountingComplete
	view.Online = snapshot.Online
	view.OnlineSessions = snapshot.OnlineSessions
	view.LastOnline = snapshot.LastOnline

	if !snapshot.Present {
		view.UsageCapability = UsageCapability{Available: false, Reason: "accounting_unavailable"}
	} else if !snapshot.AccountingComplete {
		view.UsageCapability = UsageCapability{Available: false, Reason: "accounting_incomplete"}
	} else {
		view.UsageCapability = UsageCapability{Available: true}
	}

	var used *int64
	var presence *PresenceStatus
	if view.UsageCapability.Available {
		usedValue := snapshot.UsedBytes
		used = &usedValue
		presenceValue := PresenceOffline
		if snapshot.Online {
			presenceValue = PresenceOnline
		}
		presence = &presenceValue
	}
	view.StatusDimensions = DeriveStatusDimensions(StatusInput{
		UserState:           view.Status,
		TermState:           view.ServiceState,
		StartPolicy:         view.StartPolicy,
		QuotaBytes:          view.QuotaBytes,
		ExpiresAt:           view.ExpiresAt,
		OnHold:              view.OnHold,
		AccountingAvailable: view.UsageCapability.Available,
		UsedBytes:           used,
		Presence:            presence,
	})
}
