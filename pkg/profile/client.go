// Package profile 封装对 svr-profile 内部 API 的调用。
// 管理后台通过 svr-profile 的 /internal 接口（X-Internal-Key 鉴权）查询与管控 C 端用户，
// svr-admin 自身不直接读写 svr-profile 业务库。
package profile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client svr-profile 内部 API 客户端
type Client struct {
	BaseURL     string
	InternalKey string
	httpClient  *http.Client
}

// NewClient 创建 svr-profile 客户端
func NewClient(baseURL, internalKey string) *Client {
	return &Client{
		BaseURL:     strings.TrimRight(baseURL, "/"),
		InternalKey: internalKey,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

// User svr-profile 返回的用户视图（按需裁剪）
type User struct {
	ID        uint       `json:"id"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone"`
	Nickname  string     `json:"nickname"`
	Role      string     `json:"role"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	LastLogin *time.Time `json:"last_login_at,omitempty"`
}

// ListUsersParams 用户列表查询参数
type ListUsersParams struct {
	Page     int
	PageSize int
	Keyword  string // 邮箱/手机号/昵称模糊匹配
	Status   string
}

// ListUsersResult 用户列表结果
type ListUsersResult struct {
	List  []User `json:"list"`
	Total int64  `json:"total"`
}

// internalEnvelope svr-profile 统一响应解包
type internalEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// do 执行请求并解包统一响应；非 0 code 视为业务错误。
func (c *Client) do(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("序列化请求失败: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reader)
	if err != nil {
		return fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Key", c.InternalKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求 svr-profile 失败: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("svr-profile 返回 HTTP %d: %s", resp.StatusCode, string(raw))
	}

	var env internalEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("解析 svr-profile 响应失败: %w", err)
	}
	if env.Code != 0 {
		return fmt.Errorf("svr-profile 业务错误(%d): %s", env.Code, env.Message)
	}
	if out != nil && len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return fmt.Errorf("解析 svr-profile 数据失败: %w", err)
		}
	}
	return nil
}

// ListUsers GET /internal/users
func (c *Client) ListUsers(ctx context.Context, p ListUsersParams) (*ListUsersResult, error) {
	q := url.Values{}
	q.Set("page", strconv.Itoa(p.Page))
	q.Set("page_size", strconv.Itoa(p.PageSize))
	if p.Keyword != "" {
		q.Set("keyword", p.Keyword)
	}
	if p.Status != "" {
		q.Set("status", p.Status)
	}
	var out ListUsersResult
	if err := c.do(ctx, http.MethodGet, "/internal/users?"+q.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetUser GET /internal/users/:id
func (c *Client) GetUser(ctx context.Context, id uint) (*User, error) {
	var out User
	if err := c.do(ctx, http.MethodGet, "/internal/users/"+strconv.FormatUint(uint64(id), 10), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SetUserStatusRequest 封禁/解封请求
type SetUserStatusRequest struct {
	Status string `json:"status"` // active / banned
}

// SetUserStatus POST /internal/users/:id/status
func (c *Client) SetUserStatus(ctx context.Context, id uint, status string) error {
	return c.do(ctx, http.MethodPost,
		"/internal/users/"+strconv.FormatUint(uint64(id), 10)+"/status",
		SetUserStatusRequest{Status: status}, nil)
}

// ResetPasswordResult 重置密码结果
type ResetPasswordResult struct {
	NewPassword string `json:"new_password"`
}

// ResetUserPassword POST /internal/users/:id/reset-password
func (c *Client) ResetUserPassword(ctx context.Context, id uint) (*ResetPasswordResult, error) {
	var out ResetPasswordResult
	if err := c.do(ctx, http.MethodPost,
		"/internal/users/"+strconv.FormatUint(uint64(id), 10)+"/reset-password", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---- 订单相关 ----

// Order svr-profile 返回的订单视图
type Order struct {
	ID          uint   `json:"id"`
	UserID      uint   `json:"user_id"`
	OrderNo     string `json:"order_no"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	Status      string `json:"status"`
	ProductType string `json:"product_type"`
	ProductID   string `json:"product_id"`
	Description string `json:"description"`
	PayChannel  string `json:"pay_channel,omitempty"`
	PaidAt      string `json:"paid_at,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// OrderListResult 订单列表结果
type OrderListResult struct {
	List     []Order `json:"list"`
	Total    int64   `json:"total"`
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
}

// ListOrdersParams 订单列表查询参数
type ListOrdersParams struct {
	UserID   uint
	Status   string
	Page     int
	PageSize int
}

// ListOrders GET /internal/orders
func (c *Client) ListOrders(ctx context.Context, p ListOrdersParams) (*OrderListResult, error) {
	q := url.Values{}
	if p.UserID > 0 {
		q.Set("user_id", strconv.FormatUint(uint64(p.UserID), 10))
	}
	if p.Status != "" {
		q.Set("status", p.Status)
	}
	q.Set("page", strconv.Itoa(p.Page))
	q.Set("page_size", strconv.Itoa(p.PageSize))
	var out OrderListResult
	if err := c.do(ctx, http.MethodGet, "/internal/orders?"+q.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetOrder GET /internal/orders/:id
func (c *Client) GetOrder(ctx context.Context, orderID uint) (*Order, error) {
	var out Order
	if err := c.do(ctx, http.MethodGet,
		"/internal/orders/"+strconv.FormatUint(uint64(orderID), 10), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RefundRequest 退款请求
type RefundRequest struct {
	Amount     int64  `json:"amount"`      // 退款金额（分）
	RefundDesc string `json:"refund_desc"` // 退款原因
}

// Refund 退款结果
type Refund struct {
	ID               uint   `json:"id"`
	OrderID          uint   `json:"order_id"`
	UserID           uint   `json:"user_id"`
	OutTradeRefundNo string `json:"out_trade_refund_no"`
	Amount           int64  `json:"amount"`
	Status           string `json:"status"`
	RefundDesc       string `json:"refund_desc"`
	CreatedAt        string `json:"created_at"`
}

// RefundOrder POST /internal/orders/:id/refund
func (c *Client) RefundOrder(ctx context.Context, orderID uint, req *RefundRequest) (*Refund, error) {
	var out Refund
	if err := c.do(ctx, http.MethodPost,
		"/internal/orders/"+strconv.FormatUint(uint64(orderID), 10)+"/refund", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
