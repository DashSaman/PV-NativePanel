package subscription

// R5 GATE STEER-006 (per-user node subset): mobile subscriptions must carry
// at most TopKDefault nodes — never the whole pool (battery + CL. noise).
// The subset is chosen by score upstream (steering aggregates); this cap is
// the render-side guarantee that K is respected regardless of caller error.
const TopKDefault = 10

// TopKNodes caps the rendered node list at k entries (k <= 0 -> TopKDefault).
// Order is preserved: the caller's primary stays primary. Pure and total —
// nil/empty input returns nil, short lists pass through unchanged.
func TopKNodes(nodes []Node, k int) []Node {
	if k <= 0 || k > TopKDefault {
		k = TopKDefault
	}
	if len(nodes) <= k {
		return nodes
	}
	return nodes[:k:k]
}
