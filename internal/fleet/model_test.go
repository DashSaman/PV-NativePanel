package fleet

import "testing"

func TestFleetFoundationDoesNotRequireControllerForStandalone(t *testing.T) {
	if StandaloneRequiresController() {
		t.Fatal("standalone PVNaive must not require a controller")
	}
}

func TestRegistryBlocksDeletingAssignedNodeAndTracksDrift(t *testing.T) {
	registry := NewRegistry()
	node := Node{ID: "node-uuid-1", Name: "primary", CapacityWeight: 100, DesiredRevision: 3, AppliedRevision: 2, Maintenance: MaintenanceActive}
	if err := registry.Upsert(node); err != nil {
		t.Fatal(err)
	}
	if Drift(node) != Pending {
		t.Fatalf("drift=%q", Drift(node))
	}
	if err := registry.Assign("customer-1", node.ID); err != nil {
		t.Fatal(err)
	}
	if registry.CanDelete(node.ID) {
		t.Fatal("assigned node must not be deletable")
	}
	if err := registry.Delete(node.ID); err == nil {
		t.Fatal("expected assignment-aware delete failure")
	}
}

func TestControllerAuthDesignHasMutualIdentityRotationAndReplayBoundary(t *testing.T) {
	design := SecureControllerAuthDesign()
	if design.Transport == "" || design.Identity == "" || design.Rotation == "" || design.Replay == "" {
		t.Fatalf("incomplete auth design: %#v", design)
	}
}
