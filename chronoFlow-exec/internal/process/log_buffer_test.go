package process

import "testing"

func TestLogBufferKeepsWithinLimitAndMarksTruncated(t *testing.T) {
	buf := NewLogBuffer(10)

	buf.Write([]byte("1234567890"))
	buf.Write([]byte("abcdef"))

	if !buf.Truncated() {
		t.Fatal("expected buffer to be truncated")
	}
	if len(buf.String()) > 10 {
		t.Fatalf("expected content <= 10 bytes, got %d", len(buf.String()))
	}
	if buf.String() != "12345bcdef" {
		t.Fatalf("unexpected content %q", buf.String())
	}
}
