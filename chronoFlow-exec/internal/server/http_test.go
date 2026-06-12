package server

import (
	"io"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"chronoFlow-exec/internal/conf"
	"chronoFlow-exec/internal/service"

	"github.com/go-kratos/kratos/v2/log"
)

func newTestHTTPServer() *httptest.Server {
	executorConf := &conf.Executor{Name: "exec-test", Token: "executor-token"}
	executorSvc := service.NewExecutorService(executorConf, nil, nil, nil)
	srv := NewHTTPServer(nil, executorConf, executorSvc, log.NewStdLogger(io.Discard))
	return httptest.NewServer(srv)
}

func TestHTTPHealthRequiresToken(t *testing.T) {
	ts := newTestHTTPServer()
	defer ts.Close()

	resp, err := stdhttp.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != stdhttp.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestHTTPHealthSuccess(t *testing.T) {
	ts := newTestHTTPServer()
	defer ts.Close()

	req, err := stdhttp.NewRequest(stdhttp.MethodGet, ts.URL+"/health", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}
	req.Header.Set("X-Executor-Token", "executor-token")
	resp, err := stdhttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /health returned error: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}

	if resp.StatusCode != stdhttp.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
	}
	if !strings.Contains(string(body), `"status":"online"`) {
		t.Fatalf("expected online body, got %s", string(body))
	}
}
