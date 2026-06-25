# svr-admin 架构设计

## 一、定位与边界

| 服务 | 职责 |
|---|---|
| svr-profile | C 端用户、订单、支付、通知 |
| svr-console | 用户订阅/控制台 |
| **svr-admin** | **运营/管理后台：管理员账号体系、跨服务数据聚合查询与管控、审计** |

核心原则：**svr-admin 自身不直接读写 svr-profile / svr-console 的业务库**，而是通过它们已有的 `/internal/*`（`X-Internal-Key`）接口调用。svr-admin 只拥有自己的库（管理员、角色、权限、审计日志）。

## 二、关键设计决策

### 1. 管理员账号独立

管理员与 C 端用户生命周期、登录入口、风控完全不同。独立后可做更细的 RBAC、登录 IP 白名单、强制 MFA、操作审计，且不污染用户表。

沿用相同的 JWT 结构（`Claims.Role` + `Permissions`），但：
- `Issuer` = `svr-admin`（区别于 svr-profile 的 `svr-profile`）
- JWT 密钥独立（`JWT_SECRET` 环境变量）
- Access Token TTL = 15 min（无状态，无黑名单/Redis 依赖）

### 2. RBAC 落库 + Token 内嵌权限

管理后台需要可视化分配角色/权限，因此 `Role` / `Permission` 落表；登录时把当前角色的权限码快照写进 JWT，鉴权中间件直接读 Token。

- 超级管理员（`super_admin`）跳过权限点校验，拥有全部权限。
- 其他角色按 `RequirePermission("service:resource:action")` 精确控制。

### 3. 跨服务管控走 internal API

对用户封禁、密码重置等操作，svr-admin 调 svr-profile 的 `/internal` 接口（`X-Internal-Key` 鉴权）。svr-admin 是编排层，不直接操作下游数据库。

`profileClient` 为 nil（未配置 `PROFILE_BASE_URL`）时降级返回错误，不影响后台其他功能正常运行。

## 三、数据模型

```
admins              -- 管理员账号
roles               -- 角色（如 super_admin / operator / finance）
permissions         -- 权限点（service:resource:action）
role_permissions    -- 角色-权限关联（多对多，GORM many2many）
admin_operation_logs -- 操作审计日志（所有写操作必记）
```

权限码命名规范：`service:resource:action`
- `admin:admin:read` / `admin:admin:write`
- `admin:role:read` / `admin:role:write`
- `admin:permission:read`
- `admin:log:read`
- `profile:user:read` / `profile:user:ban` / `profile:user:reset_password`

## 四、目录结构

```
svr-admin/
├── config/config.go          # viper 配置：server/database/jwt/profile/seed
├── config.yaml               # 配置模板（敏感项仅 env 注入）
├── model/model.go            # Admin / Role / Permission / AdminOperationLog
├── pkg/
│   ├── auth/jwt.go           # JWT 管理器（Issuer=svr-admin）
│   ├── auth/perm.go          # 权限点常量 + AllPermissions()
│   ├── response/             # 统一响应（复用 svr-profile 约定）
│   ├── pagination/           # 分页参数解析
│   └── profile/client.go     # svr-profile /internal 客户端
├── middleware/middleware.go  # JWTAuth / RequirePermission / RequireSuperAdmin / Audit / CORS / Logger
├── service/
│   ├── admin_service.go      # 管理员 CRUD、登录、改密、种子
│   ├── rbac_service.go       # 角色/权限 CRUD、授权、种子
│   ├── audit_service.go      # 写/查审计日志
│   └── user_admin_service.go # 编排 svr-profile 用户管控
├── handler/
│   ├── auth.go               # 登录/刷新/登出/自身信息/改密
│   ├── admin.go              # 管理员列表/详情/创建/更新/状态/重置密码
│   ├── rbac.go               # 角色/权限 CRUD + 授权
│   ├── user.go               # C 端用户列表/详情/封禁/解封/重置密码
│   ├── log.go                # 操作日志查询
│   └── health.go             # 健康检查
├── router/router.go          # 路由注册
├── cron/cron.go              # 定时任务（预留）
├── main.go                   # DI 装配 + 迁移 + 种子 + 启动
├── Dockerfile
└── Makefile
```

## 五、API 设计

所有接口前缀：`/api/v1/admin`

### 认证（免登录）

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/auth/login` | 管理员登录（username + password） |
| POST | `/auth/refresh` | 刷新 Access Token |

### 需要登录（JWT 认证）

**自身操作：**

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/auth/logout` | 登出（无状态，客户端丢弃 Token） |
| GET | `/profile` | 获取当前管理员信息 |
| PUT | `/profile/password` | 修改自身密码 |

**管理员管理（需 `admin:admin:*` 权限）：**

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/admins` | 管理员列表（分页 + keyword） |
| GET | `/admins/:id` | 管理员详情 |
| POST | `/admins` | 创建管理员 |
| PUT | `/admins/:id` | 更新管理员 |
| POST | `/admins/:id/status` | 启用/禁用管理员 |
| POST | `/admins/:id/reset-password` | 重置管理员密码 |

**角色与权限（需 `admin:role:*` / `admin:permission:*`）：**

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/permissions` | 权限点列表 |
| GET | `/roles` | 角色列表（含权限） |
| GET | `/roles/:id` | 角色详情 |
| POST | `/roles` | 创建角色 |
| PUT | `/roles/:id` | 更新角色 |
| DELETE | `/roles/:id` | 删除角色（系统角色/在用角色禁止） |
| PUT | `/roles/:id/permissions` | 全量设置角色权限 |

**C 端用户管控（编排 svr-profile，需 `profile:user:*`）：**

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/users` | 用户列表（分页 + keyword + status） |
| GET | `/users/:id` | 用户详情 |
| POST | `/users/:id/ban` | 封禁用户 |
| POST | `/users/:id/unban` | 解封用户 |
| POST | `/users/:id/reset-password` | 重置用户密码 |

**操作日志（需 `admin:log:read`）：**

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/logs` | 操作日志列表（分页 + admin_id/action/result 过滤） |

## 六、跨服务依赖

svr-admin 调用 svr-profile 的以下 internal 接口（`X-Internal-Key` 鉴权）：

| 方法 | svr-profile 路径 | 说明 |
|---|---|---|
| GET | `/internal/users` | 分页查询用户（keyword/status） |
| GET | `/internal/users/:id` | 获取用户详情 |
| POST | `/internal/users/:id/status` | 封禁/解封 |
| POST | `/internal/users/:id/reset-password` | 重置密码 |

## 七、安全要点

- 管理员密码 `bcrypt`；JWT 密钥与 svr-profile 完全独立（env `JWT_SECRET`）。
- 所有敏感配置（JWT Secret、Profile Internal Key、DB Password、Seed Password）仅通过环境变量注入。
- `super_admin` 走角色跳过兜底；其余按 `RequirePermission` 精确控制；越权返回 403。
- 所有写操作（创建/更新/删除/封禁/重置密码）经 `Audit` 中间件自动落 `admin_operation_logs`。
- Profile 客户端仅访问已知下游服务地址（配置注入），不接受外部传入 URL，避免 SSRF。
- Access Token 15min 短 TTL，登出无需 Redis 黑名单，客户端丢弃即可。

## 八、可靠性

- 下游 internal 调用失败：高危写操作同步返回错误并在审计日志记录 `result=fail`，不静默吞掉。
- `profileClient` 为 nil（未配置下游）时，用户管控接口返回明确错误，后台其他功能不受影响。
- 种子逻辑幂等：权限点仅插入缺失项；超管角色 + 初始超管仅在库中无管理员时创建。

## 九、配置项

| 配置键 | 环境变量 | 默认值 | 说明 |
|---|---|---|---|
| `server.port` | `SERVER_PORT` | 8080 | 服务端口 |
| `server.mode` | `SERVER_MODE` | debug | Gin 模式 |
| `database.*` | `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` | 127.0.0.1:3306/svr_admin | 数据库 |
| `jwt.secret` | `JWT_SECRET` | (必须配置) | JWT 签名密钥 |
| `profile.base_url` | `PROFILE_BASE_URL` | (空=禁用) | svr-profile 基础地址 |
| `profile.internal_key` | `PROFILE_INTERNAL_KEY` | (空) | 调用 svr-profile internal 的鉴权密钥 |
| `seed.super_admin_username` | `SEED_SUPER_ADMIN_USERNAME` | admin | 初始超管用户名 |
| `seed.super_admin_password` | `SEED_SUPER_ADMIN_PASSWORD` | (空=不创建) | 初始超管密码 |
| `seed.super_admin_email` | `SEED_SUPER_ADMIN_EMAIL` | admin@nostack.local | 初始超管邮箱 |
