# svr-admin API 参考

Base URL: `http://localhost:8080/api/v1/admin`

统一响应格式：
```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```
- `code=0` 表示成功，非 0 表示业务错误。
- 分页接口的 `data` 为 `{ "list": [], "total": N, "page": N, "page_size": N }`。

认证方式：`Authorization: Bearer <access_token>`

---

## 认证

### POST /auth/login

管理员登录。

**Request Body:**
```json
{
  "username": "admin",
  "password": "your-password"
}
```

**Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJ...",
    "refresh_token": "eyJ...",
    "expires_in": 900,
    "token_type": "Bearer",
    "admin": {
      "id": 1,
      "username": "admin",
      "name": "超级管理员",
      "email": "admin@nostack.local",
      "status": "active",
      "role_id": 1,
      "role": { "id": 1, "code": "super_admin", "name": "超级管理员" },
      "last_login_at": "2026-06-25T10:00:00Z",
      "last_login_ip": "127.0.0.1"
    }
  }
}
```

### POST /auth/refresh

刷新 Access Token。

**Request Body:**
```json
{
  "refresh_token": "eyJ..."
}
```

### POST /auth/logout

登出（无状态，客户端丢弃 Token 即可）。需认证。

---

## 自身操作

### GET /profile

获取当前管理员信息。需认证。

### PUT /profile/password

修改自身密码。需认证。

**Request Body:**
```json
{
  "old_password": "current-password",
  "new_password": "new-password-8chars-min"
}
```

---

## 管理员管理

需权限：`admin:admin:read` / `admin:admin:write`

### GET /admins

分页查询管理员列表。

**Query Params:** `page`, `page_size`, `keyword`（用户名/姓名/邮箱模糊匹配）

### GET /admins/:id

获取管理员详情。

### POST /admins

创建管理员。

**Request Body:**
```json
{
  "username": "operator01",
  "password": "min8chars!",
  "name": "运营人员",
  "email": "op01@company.com",
  "role_id": 2
}
```

### PUT /admins/:id

更新管理员信息。

**Request Body:**
```json
{
  "name": "新名称",
  "email": "new@email.com",
  "role_id": 3
}
```

### POST /admins/:id/status

启用/禁用管理员。

**Request Body:**
```json
{
  "status": "disabled"
}
```
可选值：`active` / `disabled`

### POST /admins/:id/reset-password

重置管理员密码。

**Request Body:**
```json
{
  "new_password": "new-secure-password"
}
```

---

## 角色与权限

需权限：`admin:role:read` / `admin:role:write` / `admin:permission:read`

### GET /permissions

返回系统全部权限点列表（按 group 分组）。

**Response Data:**
```json
[
  { "id": 1, "code": "admin:admin:read", "name": "查看管理员", "group": "管理员管理" },
  ...
]
```

### GET /roles

角色列表（含关联的权限点）。

### GET /roles/:id

角色详情。

### POST /roles

创建角色。

**Request Body:**
```json
{
  "code": "operator",
  "name": "运营人员",
  "description": "负责日常运营管理"
}
```

### PUT /roles/:id

更新角色基础信息。

**Request Body:**
```json
{
  "name": "新名称",
  "description": "新描述"
}
```

### DELETE /roles/:id

删除角色。系统内置角色（`is_system=true`）和仍有管理员使用的角色不可删除。

### PUT /roles/:id/permissions

全量设置角色权限（按权限码列表替换）。

**Request Body:**
```json
{
  "permission_codes": [
    "admin:admin:read",
    "profile:user:read",
    "profile:user:ban"
  ]
}
```

---

## C 端用户管控

需权限：`profile:user:read` / `profile:user:ban` / `profile:user:reset_password`

> 所有用户数据来源于 svr-profile，通过 internal API 编排。

### GET /users

分页查询用户。

**Query Params:** `page`, `page_size`, `keyword`（邮箱/手机/昵称模糊）, `status`（active/banned）

### GET /users/:id

用户详情。

### POST /users/:id/ban

封禁用户。

### POST /users/:id/unban

解封用户。

### POST /users/:id/reset-password

重置用户密码。

**Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "new_password": "随机生成的新密码"
  }
}
```

---

## 操作日志

需权限：`admin:log:read`

### GET /logs

分页查询操作审计日志。

**Query Params:** `page`, `page_size`, `admin_id`, `action`, `result`（success/fail）

**Response Data Item:**
```json
{
  "id": 1,
  "admin_id": 1,
  "admin_name": "admin",
  "action": "user.ban",
  "resource": "123",
  "detail": "",
  "ip": "192.168.1.1",
  "result": "success",
  "error_msg": "",
  "created_at": "2026-06-25T10:05:00Z"
}
```

---

## 错误码

| 范围 | 模块 |
|---|---|
| 10001-10010 | 认证（登录/刷新/改密） |
| 10101-10199 | 管理员管理 |
| 10201-10299 | 角色权限 |
| 10301-10399 | 用户管控 |
| 40000 | 参数错误 |
| 40100 | 未认证 |
| 40300 | 无权限 |
| 50000 | 内部错误 |
