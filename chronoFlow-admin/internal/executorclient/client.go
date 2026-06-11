package executorclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"chronoFlow-admin/internal/biz"
	"chronoFlow-admin/internal/conf"

	"github.com/google/wire"
)

const executorTokenHeader = "X-Executor-Token"

var ProviderSet = wire.NewSet(
	NewClientFromConf,
	wire.Bind(new(biz.ExecutorRunner), new(*Client)),
	wire.Bind(new(biz.ExecutorHealthClient), new(*Client)),
)

type Client struct {
	httpClient *http.Client
}

func NewClientFromConf(c *conf.Executor) *Client {
	timeout := 10 * time.Second
	if c != nil && c.RequestTimeoutSeconds > 0 {
		timeout = time.Duration(c.RequestTimeoutSeconds) * time.Second
	}
	return NewClient(timeout)
}

func NewClient(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
	}
}

type RunRequest struct {
	JobID          int64  `json:"job_id"`
	LogID          int64  `json:"log_id"`
	Script         string `json:"script"`
	TimeoutSeconds int32  `json:"timeout_seconds"`
	CallbackURL    string `json:"callback_url"`
	CallbackToken  string `json:"callback_token"`
}

type KillRequest struct {
	JobID int64 `json:"job_id"`
	LogID int64 `json:"log_id"`
}

type response[T any] struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type runData struct {
	Status string `json:"status"`
	LogID  int64  `json:"log_id"`
}

type killData struct {
	Status string `json:"status"`
	LogID  int64  `json:"log_id"`
}

func (c *Client) Run(ctx context.Context, address string, token string, req biz.ExecutorRunRequest) error {
	var resp response[runData]
	if err := c.post(ctx, address, "/run", token, RunRequest{
		JobID:          req.JobID,
		LogID:          req.LogID,
		Script:         req.Script,
		TimeoutSeconds: req.TimeoutSeconds,
		CallbackURL:    req.CallbackURL,
		CallbackToken:  req.CallbackToken,
	}, &resp); err != nil {
		return err
	}
	if resp.Code != 0 || resp.Data.Status != "accepted" {
		return fmt.Errorf("executor run not accepted: code=%d status=%s msg=%s", resp.Code, resp.Data.Status, resp.Msg)
	}
	return nil
}

func (c *Client) Kill(ctx context.Context, address string, token string, req biz.ExecutorKillRequest) error {
	var resp response[killData]
	if err := c.post(ctx, address, "/kill", token, KillRequest{
		JobID: req.JobID,
		LogID: req.LogID,
	}, &resp); err != nil {
		return err
	}
	if resp.Code != 0 || resp.Data.Status != "killing" {
		return fmt.Errorf("executor kill not accepted: code=%d status=%s msg=%s", resp.Code, resp.Data.Status, resp.Msg)
	}
	return nil
}

func (c *Client) Health(ctx context.Context, address string, token string) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, joinURL(address, "/health"), nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set(executorTokenHeader, token)
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return fmt.Errorf("executor health returned http %d", httpResp.StatusCode)
	}
	var resp response[struct {
		Status string `json:"status"`
	}]
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("executor health returned code %d", resp.Code)
	}
	return nil
}

func (c *Client) post(ctx context.Context, address string, path string, token string, payload any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, joinURL(address, path), bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(executorTokenHeader, token)
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return fmt.Errorf("executor returned http %d", httpResp.StatusCode)
	}
	return json.NewDecoder(httpResp.Body).Decode(out)
}

func joinURL(address string, path string) string {
	return strings.TrimRight(address, "/") + path
}
