package fleet

import (
	"errors"
	"strings"
	"time"
)

type NodeState string
type MaintenanceState string

const (
	NodeUnknown  NodeState = "unknown"
	NodeHealthy  NodeState = "healthy"
	NodeDegraded NodeState = "degraded"
	NodeOffline  NodeState = "offline"

	MaintenanceActive   MaintenanceState = "active"
	MaintenanceDraining MaintenanceState = "draining"
	MaintenanceDisabled MaintenanceState = "disabled"
)

type Node struct {
	ID              string
	Name            string
	Version         string
	RuntimeState    string
	Health          NodeState
	DesiredRevision uint64
	AppliedRevision uint64
	CapacityWeight  uint16
	Maintenance     MaintenanceState
	LastSeenAt      time.Time
}

func (n Node) Validate() error {
	if strings.TrimSpace(n.ID) == "" {
		return errors.New("fleet: stable node id is required")
	}
	if strings.TrimSpace(n.Name) == "" {
		return errors.New("fleet: node name is required")
	}
	if n.CapacityWeight == 0 {
		return errors.New("fleet: capacity weight must be positive")
	}
	return nil
}

type DriftState string

const (
	InSync  DriftState = "in_sync"
	Pending DriftState = "pending"
	Ahead   DriftState = "ahead"
)

func Drift(n Node) DriftState {
	switch {
	case n.AppliedRevision == n.DesiredRevision:
		return InSync
	case n.AppliedRevision < n.DesiredRevision:
		return Pending
	default:
		return Ahead
	}
}

type Registry struct {
	nodes       map[string]Node
	assignments map[string]map[string]struct{}
}

func NewRegistry() *Registry {
	return &Registry{nodes: make(map[string]Node), assignments: make(map[string]map[string]struct{})}
}

func (r *Registry) Upsert(node Node) error {
	if r == nil {
		return errors.New("fleet: registry unavailable")
	}
	if err := node.Validate(); err != nil {
		return err
	}
	r.nodes[node.ID] = node
	return nil
}

func (r *Registry) Assign(customerID, nodeID string) error {
	if r == nil || r.nodes[nodeID].ID == "" {
		return errors.New("fleet: target node does not exist")
	}
	if strings.TrimSpace(customerID) == "" {
		return errors.New("fleet: customer id is required")
	}
	set := r.assignments[nodeID]
	if set == nil {
		set = make(map[string]struct{})
		r.assignments[nodeID] = set
	}
	set[customerID] = struct{}{}
	return nil
}

func (r *Registry) CanDelete(nodeID string) bool {
	if r == nil {
		return false
	}
	_, exists := r.nodes[nodeID]
	return exists && len(r.assignments[nodeID]) == 0
}

func (r *Registry) Delete(nodeID string) error {
	if !r.CanDelete(nodeID) {
		return errors.New("fleet: node has assignments or does not exist")
	}
	delete(r.nodes, nodeID)
	delete(r.assignments, nodeID)
	return nil
}

func StandaloneRequiresController() bool { return false }

type ControllerAuthDesign struct {
	Transport string
	Identity  string
	Rotation  string
	Replay    string
}

func SecureControllerAuthDesign() ControllerAuthDesign {
	return ControllerAuthDesign{
		Transport: "mutual TLS on a dedicated control endpoint",
		Identity:  "stable node UUID bound to a short-lived client certificate",
		Rotation:  "controller-issued certificate rotation with overlap and explicit revocation",
		Replay:    "monotonic desired revision plus signed request timestamp/idempotency key",
	}
}
