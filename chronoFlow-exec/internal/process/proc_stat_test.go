package process

import "testing"

func TestParseProcStatProcessGroupID(t *testing.T) {
	stat := "12345 (worker name) S 1 6789 6789 0 -1 4194560 1 2 3 4 5"

	info, err := parseProcStat(stat)
	if err != nil {
		t.Fatalf("parseProcStat returned error: %v", err)
	}
	if info.ProcessGroupID != 6789 {
		t.Fatalf("expected pgid 6789, got %d", info.ProcessGroupID)
	}
	if info.State != "S" {
		t.Fatalf("expected state S, got %q", info.State)
	}
}

func TestParseProcStatDetectsZombieState(t *testing.T) {
	stat := "42 (sleep) Z 1 40 1 0 -1 4228364 102 0 0 0"

	info, err := parseProcStat(stat)
	if err != nil {
		t.Fatalf("parseProcStat returned error: %v", err)
	}
	if !info.Zombie {
		t.Fatal("expected zombie state")
	}
}
