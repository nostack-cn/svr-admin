# svr-admin 认证与鉴权设计

## 认证流程

```
客户端                         svr-admin
  |--- POST /auth/login ------->|
  |    {username, password}     |
  |                             | 1. 查 admins 表
  |                             | 2. bcrypt 校验密码
  |                             | 3. 检查 status=active
  |                             | 4. 加载角色权限码
  |                             | 5. 签发 JWT TokenPair
  |<--- {access_token, ...} ----|
  |                             |
  |--- GET /profile ----------->|
  |    Authorization: Bearer AT |
  |                             | JWTAuth 中间件校验签名/过期
  |                             | 注入 admin_id/role/permissions
  |<--- {admin} ----------------|
```

## JWT 结构

### Access Token (TTL = 15 min)

```json
{
  "admin_id": 1,
  "username": "admin",
  "role": "super_admin",
  "permissions": ["admin:admin:read", "admin:admin:write", ...],
  "token_type": "access",
  "iss": "svr-admin",
  "aud": ["nostack-admin"],
  "exp": 1750849200,
  "iat": 1750848300,
  "jti": "random-hex-32"
}
```

### Refresh Token (TTL = 12 h)

```json
{
  "admin_id": 1,
  "username": "admin",
  "role": "super_admin",
  "token_type": "refresh",
  "iss": "svr-admin",
  "aud": ["nostack-admin"],
  "exp": 1750891800,
  "jti": "random-hex-32"
}
```

## 鉴权中间件

### JWTAuth

- 解析 `Authorization: Bearer <token>`。
- 校验签名、过期时间、`token_type=access`。
- 注入 `admin_id` / `admin_username` / `admin_role` / `admin_permissions` 到 gin.Context。

### RequirePermission(perm)

- 若 `admin_role == "super_admin"`：直接放行（超管拥有全部权限）。
- 否则检查 `admin_permissions` 切片是否包含 `perm`。
- 不通过返回 HTTP 403。

### RequireSuperAdmin

- 仅 `admin_role == "super_admin"` 放行。

### Audit(action)

- 在 handler 执行之后（`c.Next()` 后）自动写入 `admin_operation_logs`。
- handler 可通过 `c.Set("audit_resource", ...)` / `c.Set("audit_detail", ...)` / `c.Set("audit_error", ...)` 注入细节。
- HTTP status >= 400 或 handler 主动 `c.Set("audit_result", "fail")` 时记录为失败。

## 登出

Access Token 仅 15 分钟有效期，采用无状态方案：
- 客户端丢弃 Token 即视为登出。
- 无需 Redis 黑名单。
- 管理员被禁用后，最长 15 分钟内失效；下次 Refresh 时检查 status 立即拒绝。

## 安全防护

- 密码存储：`bcrypt`（DefaultCost=10）。
- JWT Secret：仅通过环境变量注入，不落配置文件。
- 登录失败：可扩展登录失败次数限制（Redis 计数器，预留位置）。
- 高危操作（封禁/重置密码）：已落审计日志；可扩展二次验证（短信/MFA）。
