package process

import "testing"

func TestParseProcStatProcessGroupID(t *testing.T) {
	stat := "12345 (worker name) S 1 6789 6789 0 -1 4194560 1 2 3 4 5"

	pgid, err := parseProcStatProcessGroupID(stat)
	if err != nil {
		t.Fatalf("parseProcStatProcessGroupID returned error: %v", err)
	}
	if pgid != 6789 {
		t.Fatalf("expected pgid 6789, got %d", pgid)
	}
}
