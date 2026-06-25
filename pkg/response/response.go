package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// R 统一响应结构
type R struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageResult 分页结果
type PageResult struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// OK 成功响应
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, R{Code: 0, Message: "success", Data: data})
}

// OKMsg 带自定义消息的成功响应
func OKMsg(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, R{Code: 0, Message: msg, Data: data})
}

// OKPage 分页成功响应
func OKPage(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	OK(c, PageResult{List: list, Total: total, Page: page, PageSize: pageSize})
}

// Fail 失败响应
func Fail(c *gin.Context, httpCode int, code int, msg string) {
	c.JSON(httpCode, R{Code: code, Message: msg})
}

// BadRequest 400 参数错误
func BadRequest(c *gin.Context, msg string) {
	Fail(c, http.StatusBadRequest, 40000, msg)
}

// Unauthorized 401 未认证
func Unauthorized(c *gin.Context, msg string) {
	Fail(c, http.StatusUnauthorized, 40100, msg)
}

// Forbidden 403 无权限
func Forbidden(c *gin.Context, msg string) {
	Fail(c, http.StatusForbidden, 40300, msg)
}

// NotFound 404 资源不存在
func NotFound(c *gin.Context, msg string) {
	Fail(c, http.StatusNotFound, 40400, msg)
}

// InternalError 500 服务器内部错误
func InternalError(c *gin.Context, msg string) {
	Fail(c, http.StatusInternalServerError, 50000, msg)
}

// BusinessError 业务逻辑错误（HTTP 200，但 code 非 0）
func BusinessError(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, R{Code: code, Message: msg})
}
