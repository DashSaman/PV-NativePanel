package protocol

import "context"

type AccountingLevel string

const (
	AccountingExact     AccountingLevel = "exact"
	AccountingEstimated AccountingLevel = "estimated"
	AccountingNone      AccountingLevel = "none"
)

type Capabilities struct {
	Accounting           AccountingLevel
	SessionVisibility    bool
	SpeedLimit            bool
	ConcurrencyLimit      bool
	DeviceLimit           bool
	IPObservation         bool
	AtomicApply           bool
	ZeroDowntimeReload    bool
	Rollback              bool
	PaddingControl        bool
	DestinationMetadata  bool
	MultiEndpoint         bool
}

type Descriptor struct {
	ID                  string
	Version             string
	RuntimeFamily       string
	ConfigSchemaVersion string
	Capabilities        Capabilities
	SubscriptionFormats []string
}

type DesiredState struct {
	Revision string
	Config   []byte
}

type ValidationResult struct {
	Valid    bool
	Warnings []string
}

type Health struct {
	Status  string
	Message string
}

type Adapter interface {
	Descriptor() Descriptor
	Validate(context.Context, DesiredState) (ValidationResult, error)
	Stage(context.Context, DesiredState) error
	Apply(context.Context, string) error
	Rollback(context.Context, string) error
	Health(context.Context) (Health, error)
}
