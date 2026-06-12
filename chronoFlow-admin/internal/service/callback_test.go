package service

import (
	"context"
	"testing"

	v1 "chronoFlow-admin/api/all-pb-go/v1"
	"chronoFlow-admin/internal/conf"
)

func TestCallbackServiceRejectsContextWithoutHTTPToken(t *testing.T) {
	svc := NewCallbackService(nil, &conf.Security{CallbackToken: "callback"})

	_, err := svc.CallbackJobRun(context.Background(), &v1.CallbackJobRunRequest{
		LogId:  1,
		JobId:  2,
		Status: "success",
	})
	if err == nil {
		t.Fatal("expected invalid token error, got nil")
	}
}
