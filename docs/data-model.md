# svr-admin 数据模型

数据库：MySQL，库名 `svr_admin`（通过 GORM AutoMigrate 自动建表）。

---

## admins - 管理员

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | uint | PK, auto_increment | |
| username | varchar(64) | UNIQUE, NOT NULL | 登录用户名 |
| password | varchar(255) | NOT NULL | bcrypt 哈希 |
| name | varchar(64) | | 姓名 |
| email | varchar(255) | UNIQUE | 邮箱 |
| status | varchar(20) | DEFAULT 'active', INDEX | active / disabled |
| role_id | uint | INDEX, NOT NULL, FK→roles.id | 关联角色 |
| last_login_at | datetime | | 最后登录时间 |
| last_login_ip | varchar(64) | | 最后登录 IP |
| created_at | datetime | | |
| updated_at | datetime | | |
| deleted_at | datetime | INDEX (soft delete) | |

---

## roles - 角色

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | uint | PK | |
| code | varchar(64) | UNIQUE, NOT NULL | 角色码，如 super_admin |
| name | varchar(64) | NOT NULL | 显示名 |
| description | varchar(255) | | 描述 |
| is_system | bool | DEFAULT false | 系统内置角色不可删除 |
| created_at | datetime | | |
| updated_at | datetime | | |
| deleted_at | datetime | INDEX | |

---

## permissions - 权限点

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | uint | PK | |
| code | varchar(128) | UNIQUE, NOT NULL | 权限码，如 profile:user:ban |
| name | varchar(64) | NOT NULL | 显示名 |
| group | varchar(64) | INDEX | 分组（前端展示用） |
| created_at | datetime | | |
| updated_at | datetime | | |
| deleted_at | datetime | INDEX | |

---

## role_permissions - 角色权限关联（多对多）

GORM `many2many:role_permissions` 自动生成。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| role_id | uint | UNIQUE(role_id, permission_id) | |
| permission_id | uint | | |

---

## admin_operation_logs - 操作审计日志

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | uint | PK | |
| admin_id | uint | INDEX | 操作人 |
| admin_name | varchar(64) | | 冗余便于查询 |
| action | varchar(128) | INDEX | 操作类型，如 user.ban |
| resource | varchar(128) | | 操作目标，如 user:1001 |
| detail | text | | JSON 摘要（请求/前后值） |
| ip | varchar(64) | | 操作者 IP |
| result | varchar(20) | DEFAULT 'success', INDEX | success / fail |
| error_msg | text | | 失败原因 |
| created_at | datetime | | |
| updated_at | datetime | | |
| deleted_at | datetime | INDEX | |

---

## ER 关系

```
Admin *--1 Role
Role *--* Permission (via role_permissions)
Admin 1--* AdminOperationLog
```
