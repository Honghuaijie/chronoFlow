package callback

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"chronoFlow-exec/internal/conf"
	"chronoFlow-exec/internal/store"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewClientFromConf)

type Client struct {
	httpClient *http.Client
}

func NewClient(timeout time.Duration) *Client {
	return &Client{httpClient: &http.Client{Timeout: timeout}}
}

func NewClientFromConf(c *conf.Callback) *Client {
	timeout := 10 * time.Second
	if c != nil && c.RequestTimeoutSeconds > 0 {
		timeout = time.Duration(c.RequestTimeoutSeconds) * time.Second
	}
	return NewClient(timeout)
}

type response struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

func (c *Client) Send(item *store.CallbackItem) error {
	body := map[string]any{
		"log_id":        item.LogID,
		"job_id":        item.JobID,
		"status":        item.Status,
		"exit_code":     item.ExitCode,
		"log_content":   item.LogContent,
		"log_truncated": item.LogTruncated,
		"start_time":    item.StartTime.Format("2006-01-02 15:04:05"),
		"end_time":      item.EndTime.Format("2006-01-02 15:04:05"),
		"duration_ms":   item.DurationMS,
		"error_message": item.ErrorMessage,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, item.CallbackURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Callback-Token", item.CallbackToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("callback returned http %d", resp.StatusCode)
	}
	var decoded response
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return err
	}
	if decoded.Code != 0 {
		return fmt.Errorf("callback returned code %d msg=%s", decoded.Code, decoded.Msg)
	}
	return nil
}
