// Package consoleclient 封装对 svr-console 内部 API 的调用。
// 管理后台通过 svr-console 的 /internal 接口（X-Internal-Key 鉴权）管理用户订阅。
package consoleclient

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

// Client svr-console 内部 API 客户端
type Client struct {
	BaseURL     string
	InternalKey string
	httpClient  *http.Client
}

// NewClient 创建 svr-console 客户端
func NewClient(baseURL, internalKey string) *Client {
	return &Client{
		BaseURL:     strings.TrimRight(baseURL, "/"),
		InternalKey: internalKey,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Key", c.InternalKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求 svr-console 失败: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("svr-console 返回 HTTP %d: %s", resp.StatusCode, string(raw))
	}
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("解析 svr-console 响应失败: %w", err)
	}
	if env.Code != 0 {
		return fmt.Errorf("svr-console 业务错误(%d): %s", env.Code, env.Message)
	}
	if out != nil && len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return fmt.Errorf("解析 data 失败: %w", err)
		}
	}
	return nil
}

// CancelSubscriptionRequest 取消订阅请求
type CancelSubscriptionRequest struct {
	UserID uint   `json:"user_id"`
	Reason string `json:"reason"`
}

// CancelSubscription POST /internal/subscriptions/cancel
func (c *Client) CancelSubscription(ctx context.Context, userID uint, reason string) error {
	return c.do(ctx, http.MethodPost, "/internal/subscriptions/cancel",
		CancelSubscriptionRequest{UserID: userID, Reason: reason}, nil)
}
