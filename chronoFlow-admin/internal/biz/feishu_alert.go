package biz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const maxAlertErrorLength = 500

type FeishuAlertSender struct {
	client *http.Client
}

func NewFeishuAlertSender() *FeishuAlertSender {
	return &FeishuAlertSender{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *FeishuAlertSender) SendCard(ctx context.Context, webhook string, payload any) error {
	body, err := json.Marshal(map[string]any{
		"msg_type": "interactive",
		"card":     payload,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("feishu webhook http status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	if len(respBody) == 0 {
		return nil
	}
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil
	}
	if code, ok := feishuResponseCode(result); ok && code != 0 {
		return fmt.Errorf("feishu webhook response code %.0f: %s", code, feishuResponseMessage(result))
	}
	return nil
}

func feishuResponseCode(result map[string]any) (float64, bool) {
	if code, ok := result["StatusCode"].(float64); ok {
		return code, true
	}
	if code, ok := result["code"].(float64); ok {
		return code, true
	}
	return 0, false
}

func feishuResponseMessage(result map[string]any) string {
	for _, key := range []string{"StatusMessage", "msg", "message"} {
		if msg, ok := result[key].(string); ok && msg != "" {
			return msg
		}
	}
	return "unknown error"
}

type FeishuAlertCardInput struct {
	JobLogID     int64
	JobID        int64
	JobName      string
	ExecutorName string
	Status       string
	TriggerType  string
	StartTime    time.Time
	EndTime      *time.Time
	DurationMS   int64
	ExitCode     *int32
	ErrorMessage string
}

func buildFailureAlertTitle(status string) string {
	if status == JobLogStatusTimeout {
		return "ChronoFlow 任务执行超时"
	}
	return "ChronoFlow 任务执行失败"
}

func buildFeishuAlertCard(input FeishuAlertCardInput) map[string]any {
	fields := []map[string]any{
		markdownField("任务", input.JobName),
		markdownField("日志 ID", fmt.Sprintf("%d", input.JobLogID)),
		markdownField("任务 ID", fmt.Sprintf("%d", input.JobID)),
		markdownField("执行器", input.ExecutorName),
		markdownField("状态", input.Status),
		markdownField("触发方式", input.TriggerType),
		markdownField("开始时间", formatAlertTime(input.StartTime)),
		markdownField("结束时间", formatAlertPtrTime(input.EndTime)),
	}
	if input.DurationMS > 0 {
		fields = append(fields, markdownField("耗时", fmt.Sprintf("%d ms", input.DurationMS)))
	}
	if input.ExitCode != nil {
		fields = append(fields, markdownField("Exit Code", fmt.Sprintf("%d", *input.ExitCode)))
	}
	errorMessage := truncateAlertError(input.ErrorMessage)
	if errorMessage == "" {
		errorMessage = "-"
	}
	return map[string]any{
		"config": map[string]any{
			"wide_screen_mode": true,
		},
		"header": map[string]any{
			"template": "red",
			"title": map[string]any{
				"tag":     "plain_text",
				"content": buildFailureAlertTitle(input.Status),
			},
		},
		"elements": []any{
			map[string]any{
				"tag": "div",
				"fields": func() []any {
					items := make([]any, 0, len(fields))
					for _, field := range fields {
						items = append(items, field)
					}
					return items
				}(),
			},
			map[string]any{
				"tag": "hr",
			},
			map[string]any{
				"tag": "div",
				"text": map[string]any{
					"tag":     "lark_md",
					"content": "**错误信息**\n" + errorMessage,
				},
			},
		},
	}
}

func markdownField(label, value string) map[string]any {
	if value == "" {
		value = "-"
	}
	return map[string]any{
		"is_short": true,
		"text": map[string]any{
			"tag":     "lark_md",
			"content": fmt.Sprintf("**%s**\n%s", label, value),
		},
	}
}

func truncateAlertError(message string) string {
	message = strings.TrimSpace(message)
	if len([]rune(message)) <= maxAlertErrorLength {
		return message
	}
	runes := []rune(message)
	return string(runes[:maxAlertErrorLength]) + "..."
}

func formatAlertTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04:05")
}

func formatAlertPtrTime(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return formatAlertTime(*t)
}
