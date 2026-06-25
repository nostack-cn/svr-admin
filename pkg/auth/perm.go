package auth

// 权限点常量（格式: service:resource:action）
const (
	// 管理员与 RBAC
	PermAdminRead  = "admin:admin:read"
	PermAdminWrite = "admin:admin:write"
	PermRoleRead   = "admin:role:read"
	PermRoleWrite  = "admin:role:write"
	PermPermRead   = "admin:permission:read"
	PermLogRead    = "admin:log:read"

	// 用户管控（编排 svr-profile）
	PermUserRead          = "profile:user:read"
	PermUserBan           = "profile:user:ban"
	PermUserResetPassword = "profile:user:reset_password"

	// 订单管控（编排 svr-profile）
	PermOrderRead   = "profile:order:read"
	PermOrderRefund = "profile:order:refund"

	// 订阅管控（编排 svr-console）
	PermSubscriptionCancel = "console:subscription:cancel"
)

// AllPermissions 返回系统全部权限点定义（用于种子数据与权限点列表展示）
// Group 用于前端分组展示。
func AllPermissions() []PermissionDef {
	return []PermissionDef{
		{Code: PermAdminRead, Name: "查看管理员", Group: "管理员管理"},
		{Code: PermAdminWrite, Name: "管理管理员", Group: "管理员管理"},
		{Code: PermRoleRead, Name: "查看角色", Group: "角色权限"},
		{Code: PermRoleWrite, Name: "管理角色", Group: "角色权限"},
		{Code: PermPermRead, Name: "查看权限点", Group: "角色权限"},
		{Code: PermLogRead, Name: "查看操作日志", Group: "审计"},
		{Code: PermUserRead, Name: "查看用户", Group: "用户管理"},
		{Code: PermUserBan, Name: "封禁/解封用户", Group: "用户管理"},
		{Code: PermUserResetPassword, Name: "重置用户密码", Group: "用户管理"},
		{Code: PermOrderRead, Name: "查看订单", Group: "订单管理"},
		{Code: PermOrderRefund, Name: "订单退款", Group: "订单管理"},
		{Code: PermSubscriptionCancel, Name: "取消用户订阅", Group: "订阅管理"},
	}
}

// PermissionDef 权限点定义
type PermissionDef struct {
	Code  string
	Name  string
	Group string
}
