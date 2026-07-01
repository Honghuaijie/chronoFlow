package server

import (
	"bytes"
	"context"
	"io"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"chronoFlow-admin/internal/biz"
	"chronoFlow-admin/internal/conf"
	"chronoFlow-admin/internal/service"

	"github.com/go-kratos/kratos/v2/log"
)

type fakeUserRepo struct {
	createFn func(context.Context, *biz.User) (*biz.User, error)
}

func (r fakeUserRepo) Create(ctx context.Context, user *biz.User) (*biz.User, error) {
	if r.createFn != nil {
		return r.createFn(ctx, user)
	}
	return nil, nil
}

func (fakeUserRepo) GetByID(context.Context, int32) (*biz.User, error) {
	return nil, nil
}

func (fakeUserRepo) List(context.Context) ([]*biz.User, error) {
	return nil, nil
}

func (fakeUserRepo) Update(context.Context, *biz.User) (*biz.User, error) {
	return nil, nil
}

func (fakeUserRepo) Delete(context.Context, int32) error {
	return nil
}

type fakeTx struct{}

func (fakeTx) ExecTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type fakeExecutorRepo struct{}

func (fakeExecutorRepo) Create(context.Context, *biz.Executor) (*biz.Executor, error) {
	return nil, nil
}

func (fakeExecutorRepo) GetByID(context.Context, int64) (*biz.Executor, error) {
	return nil, nil
}

func (fakeExecutorRepo) GetByAddress(context.Context, string) (*biz.Executor, error) {
	return nil, nil
}

func (fakeExecutorRepo) List(context.Context) ([]*biz.Executor, error) {
	return nil, nil
}

func (fakeExecutorRepo) Update(context.Context, *biz.Executor) (*biz.Executor, error) {
	return nil, nil
}

func (fakeExecutorRepo) Delete(context.Context, int64) error {
	return nil
}

type fakeTokenCipher struct{}

func (fakeTokenCipher) Encrypt(plaintext string) (string, error) {
	return plaintext, nil
}

func (fakeTokenCipher) Decrypt(ciphertext string) (string, error) {
	return ciphertext, nil
}

type fakeJobRepo struct{}

func (fakeJobRepo) Create(context.Context, *biz.Job) (*biz.Job, error) {
	return nil, nil
}

func (fakeJobRepo) GetByID(context.Context, int64) (*biz.Job, error) {
	return nil, nil
}

func (fakeJobRepo) List(context.Context, int64) ([]*biz.Job, error) {
	return nil, nil
}

func (fakeJobRepo) Update(context.Context, *biz.Job) (*biz.Job, error) {
	return nil, nil
}

func (fakeJobRepo) Delete(context.Context, int64) error {
	return nil
}

type fakeGlueRepo struct{}

func (fakeGlueRepo) GetByJobID(context.Context, int64) (*biz.Glue, error) {
	return nil, nil
}

func (fakeGlueRepo) Save(context.Context, *biz.Glue) (*biz.Glue, error) {
	return nil, nil
}

type fakeJobLogRepo struct{}

func (fakeJobLogRepo) List(context.Context, biz.JobLogFilter) ([]*biz.JobLog, int64, error) {
	return nil, 0, nil
}

func (fakeJobLogRepo) GetByID(context.Context, int64) (*biz.JobLog, error) {
	return nil, nil
}

func (fakeJobLogRepo) GetRunningByJobID(context.Context, int64) (*biz.JobLog, error) {
	return nil, nil
}

func (fakeJobLogRepo) Create(context.Context, *biz.JobLog) (*biz.JobLog, error) {
	return nil, nil
}

func (r fakeJobLogRepo) CreateRunningIfNoActive(ctx context.Context, jobLog *biz.JobLog) (*biz.JobLog, error) {
	return r.Create(ctx, jobLog)
}

func (fakeJobLogRepo) Update(context.Context, *biz.JobLog) (*biz.JobLog, error) {
	return nil, nil
}

type fakeLogReader struct{}

func (fakeLogReader) Read(context.Context, string) (string, error) {
	return "", nil
}

type fakeLogWriter struct{}

func (fakeLogWriter) Write(context.Context, int64, int64, string) (string, int64, error) {
	return "", 0, nil
}

func newTestHTTPServer(repo fakeUserRepo) *httptest.Server {
	logger := log.NewStdLogger(io.Discard)
	uc := biz.NewUserUsecase(repo, fakeTx{}, logger)
	userSvc := service.NewUserService(uc)
	executorUC := biz.NewExecutorUsecase(fakeExecutorRepo{}, fakeTokenCipher{}, logger)
	executorSvc := service.NewExecutorService(executorUC)
	glueUC := biz.NewGlueUsecase(fakeGlueRepo{}, logger)
	jobUC := biz.NewJobUsecase(fakeJobRepo{}, fakeGlueRepo{}, logger)
	jobLogUC := biz.NewJobLogUsecase(fakeJobLogRepo{}, fakeLogReader{}, logger)
	callbackUC := biz.NewCallbackUsecase(fakeJobLogRepo{}, fakeLogWriter{}, biz.CallbackConfig{MaxLogBytes: 1024}, logger)
	systemSettingsUC := biz.NewSystemSettingUsecase(&serverSystemSettingRepo{}, fakeTokenCipher{}, logger)
	srv := NewHTTPServer(
		nil,
		&conf.Security{JwtSecret: "secret", AdminUsername: "admin", AdminPassword: "admin123", CallbackToken: "callback"},
		service.NewAuthService(&conf.Security{JwtSecret: "secret", AdminUsername: "admin", AdminPassword: "admin123", CallbackToken: "callback"}),
		userSvc,
		executorSvc,
		service.NewJobService(jobUC, nil, nil),
		service.NewGlueService(glueUC),
		service.NewJobLogService(jobLogUC),
		service.NewCallbackService(callbackUC, &conf.Security{CallbackToken: "callback"}),
		service.NewSystemSettingsService(systemSettingsUC, nil),
		logger,
	)
	return httptest.NewServer(srv)
}

type serverSystemSettingRepo struct{}

func (*serverSystemSettingRepo) GetByKey(context.Context, string) (*biz.SystemSetting, error) {
	return nil, nil
}

func (*serverSystemSettingRepo) Upsert(_ context.Context, setting *biz.SystemSetting) (*biz.SystemSetting, error) {
	cp := *setting
	cp.ID = 1
	cp.UpdatedAt = time.Now()
	return &cp, nil
}

func TestHTTPCreateUser_MissingNameOrEmail(t *testing.T) {
	ts := newTestHTTPServer(fakeUserRepo{})
	defer ts.Close()

	body := []byte(`{"name":"","email":"","phone":"123"}`)
	resp, err := stdhttp.Post(ts.URL+"/v1/users/create", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if resp.StatusCode != stdhttp.StatusBadRequest {
		t.Fatalf("unexpected status code: got %d want %d", resp.StatusCode, stdhttp.StatusBadRequest)
	}
	if !strings.Contains(string(respBody), `"code":40001`) {
		t.Fatalf("expected error code 40001, got body %s", string(respBody))
	}
	if !strings.Contains(string(respBody), "name 和 email 不能为空") {
		t.Fatalf("expected message in body, got %s", string(respBody))
	}
}

func TestHTTPCreateUser_Success(t *testing.T) {
	now := time.Now()
	ts := newTestHTTPServer(fakeUserRepo{
		createFn: func(ctx context.Context, user *biz.User) (*biz.User, error) {
			return &biz.User{
				ID:        1,
				Name:      user.Name,
				Email:     user.Email,
				Phone:     user.Phone,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	})
	defer ts.Close()

	body := []byte(`{"name":"Alice","email":"alice@example.com","phone":"123"}`)
	resp, err := stdhttp.Post(ts.URL+"/v1/users/create", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if resp.StatusCode != stdhttp.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", resp.StatusCode, stdhttp.StatusOK)
	}
	if !strings.Contains(string(respBody), `"code":0`) {
		t.Fatalf("expected success code in body, got %s", string(respBody))
	}
	if !strings.Contains(string(respBody), `"message":"CreateUser success"`) {
		t.Fatalf("expected success message in body, got %s", string(respBody))
	}
	if !strings.Contains(string(respBody), `"name":"Alice"`) {
		t.Fatalf("expected response body to contain user name, got %s", string(respBody))
	}
}

func TestHTTPHealth_Success(t *testing.T) {
	ts := newTestHTTPServer(fakeUserRepo{})
	defer ts.Close()

	resp, err := stdhttp.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if resp.StatusCode != stdhttp.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", resp.StatusCode, stdhttp.StatusOK)
	}
	if !strings.Contains(string(respBody), `"status":"ok"`) {
		t.Fatalf("expected health response body, got %s", string(respBody))
	}
}
