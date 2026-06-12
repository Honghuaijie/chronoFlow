package callback

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"chronoFlow-exec/internal/store"
)

func TestClientSendUsesCallbackToken(t *testing.T) {
	var gotToken string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.Header.Get("X-Callback-Token")
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"log_id":1,"status":"success"}}`))
	}))
	defer server.Close()

	client := NewClient(time.Second)
	err := client.Send(&store.CallbackItem{
		LogID:         1,
		JobID:         2,
		CallbackURL:   server.URL,
		CallbackToken: "callback-token",
		Status:        "success",
	})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if gotToken != "callback-token" {
		t.Fatalf("expected callback token, got %q", gotToken)
	}
}
