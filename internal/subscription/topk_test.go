package subscription

import "testing"

func TestTopKCapsAndPreservesOrder(t *testing.T) {
	nodes := make([]Node, 0, TopKDefault+5)
	for i := 0; i < TopKDefault+5; i++ {
		nodes = append(nodes, Node{})
	}
	got := TopKNodes(nodes, 0)
	if len(got) != TopKDefault {
		t.Fatalf("k<=0 must clamp to default: got %d", len(got))
	}
	got = TopKNodes(nodes, 100)
	if len(got) != TopKDefault {
		t.Fatalf("k>default must clamp to default: got %d", len(got))
	}
	got = TopKNodes(nodes, 3)
	if len(got) != 3 {
		t.Fatalf("k=3 must render 3 nodes: got %d", len(got))
	}
	// Order preserved: the caller's primary stays primary.
	if &got[0] != &nodes[0] {
		t.Fatal("primary node must stay first")
	}
	if len(TopKNodes(nil, 10)) != 0 {
		t.Fatal("nil input must render nothing")
	}
	short := []Node{{}, {}}
	if len(TopKNodes(short, 10)) != 2 {
		t.Fatal("short lists pass through unchanged")
	}
}
