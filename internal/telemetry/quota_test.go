package telemetry

import (
	"sync"
	"testing"
)

func TestSharedQuotaBudgetNeverDoubleSpendsAcrossSessions(t *testing.T) {
	budget, err := NewSharedQuotaBudget(100)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan int64, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			granted, err := budget.Claim(100)
			if err != nil {
				t.Errorf("claim: %v", err)
				return
			}
			results <- granted
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	var total int64
	for granted := range results {
		total += granted
	}
	if total != 100 {
		t.Fatalf("shared quota double-spent: total granted=%d, want 100", total)
	}
	if got := budget.Remaining(); got != 0 {
		t.Fatalf("remaining=%d, want 0", got)
	}
}

func TestSharedQuotaBudgetClaimsPartialRemainder(t *testing.T) {
	budget, err := NewSharedQuotaBudget(100)
	if err != nil {
		t.Fatal(err)
	}
	first, err := budget.Claim(70)
	if err != nil || first != 70 {
		t.Fatalf("first claim=(%d,%v), want (70,nil)", first, err)
	}
	second, err := budget.Claim(70)
	if err != nil || second != 30 {
		t.Fatalf("second claim=(%d,%v), want (30,nil)", second, err)
	}
	third, err := budget.Claim(1)
	if err != nil || third != 0 {
		t.Fatalf("depleted claim=(%d,%v), want (0,nil)", third, err)
	}
}

func TestSharedQuotaBudgetRejectsInvalidClaims(t *testing.T) {
	if _, err := NewSharedQuotaBudget(-1); err == nil {
		t.Fatal("negative quota must fail")
	}
	budget, err := NewSharedQuotaBudget(10)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := budget.Claim(0); err == nil {
		t.Fatal("zero claim must fail")
	}
	if _, err := budget.Claim(-1); err == nil {
		t.Fatal("negative claim must fail")
	}
}
