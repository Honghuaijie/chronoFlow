package executorclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"chronoFlow-admin/internal/biz"
)

func TestClientRunSendsTokenAndPayload(t *testing.T) {
	var gotToken string
	var gotReq RunRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/run" {
			t.Fatalf("path = %q, want /run", r.URL.Path)
		}
		gotToken = r.Header.Get("X-Executor-Token")
		if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(response[runData]{
			Code: 0,
			Msg:  "accepted",
			Data: runData{Status: "accepted", LogID: 2001},
		})
	}))
	defer server.Close()

	client := NewClient(2 * time.Second)
	err := client.Run(context.Background(), server.URL, "executor-token", biz.ExecutorRunRequest{
		JobID:          1001,
		LogID:          2001,
		Script:         "#!/bin/bash\necho hello",
		TimeoutSeconds: 600,
		CallbackURL:    "http://admin/internal/job-runs/callback",
		CallbackToken:  "callback-token",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if gotToken != "executor-token" {
		t.Fatalf("token = %q, want executor-token", gotToken)
	}
	if gotReq.LogID != 2001 || gotReq.JobID != 1001 || gotReq.CallbackToken != "callback-token" {
		t.Fatalf("unexpected request payload: %+v", gotReq)
	}
}

func TestClientRunRejectsNonAcceptedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(response[runData]{
			Code: 0,
			Msg:  "ok",
			Data: runData{Status: "unexpected", LogID: 2001},
		})
	}))
	defer server.Close()

	client := NewClient(2 * time.Second)
	err := client.Run(context.Background(), server.URL, "executor-token", biz.ExecutorRunRequest{JobID: 1001, LogID: 2001})
	if err == nil {
		t.Fatal("Run() error = nil, want non-accepted error")
	}
}

func TestClientKillSendsToken(t *testing.T) {
	var gotToken string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/kill" {
			t.Fatalf("path = %q, want /kill", r.URL.Path)
		}
		gotToken = r.Header.Get("X-Executor-Token")
		_ = json.NewEncoder(w).Encode(response[killData]{
			Code: 0,
			Msg:  "killing",
			Data: killData{Status: "killing", LogID: 2001},
		})
	}))
	defer server.Close()

	client := NewClient(2 * time.Second)
	err := client.Kill(context.Background(), server.URL, "executor-token", biz.ExecutorKillRequest{JobID: 1001, LogID: 2001})
	if err != nil {
		t.Fatalf("Kill() error = %v", err)
	}
	if gotToken != "executor-token" {
		t.Fatalf("token = %q, want executor-token", gotToken)
	}
}
