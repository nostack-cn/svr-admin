# svr-admin 配置说明

## 配置优先级

命令行参数 > 环境变量 > 配置文件 (`config.yaml`) > 默认值

## 启动参数

```bash
./svr-admin -c /path/to/config.yaml
```

如不指定 `-c`，默认读取当前目录下 `config.yaml`；支持 `config.local.yaml` 覆盖。

## 环境变量

| 环境变量 | 配置键 | 默认值 | 说明 |
|---|---|---|---|
| `SERVER_PORT` | `server.port` | 8080 | 服务监听端口 |
| `SERVER_MODE` | `server.mode` | debug | Gin 模式 (debug/release/test) |
| `DB_HOST` | `database.host` | 127.0.0.1 | MySQL 主机 |
| `DB_PORT` | `database.port` | 3306 | MySQL 端口 |
| `DB_USER` | `database.user` | root | MySQL 用户名 |
| `DB_PASSWORD` | `database.password` | (空) | MySQL 密码 |
| `DB_NAME` | `database.dbname` | svr_admin | MySQL 库名 |
| `JWT_SECRET` | `jwt.secret` | (必须配置) | JWT 签名密钥 |
| `PROFILE_BASE_URL` | `profile.base_url` | (空=禁用用户管控) | svr-profile 基础地址 |
| `PROFILE_INTERNAL_KEY` | `profile.internal_key` | (空) | 调用 svr-profile internal 鉴权密钥 |
| `SEED_SUPER_ADMIN_USERNAME` | `seed.super_admin_username` | admin | 初始超管用户名 |
| `SEED_SUPER_ADMIN_PASSWORD` | `seed.super_admin_password` | (空=不创建) | 初始超管密码 |
| `SEED_SUPER_ADMIN_EMAIL` | `seed.super_admin_email` | admin@nostack.local | 初始超管邮箱 |

## 种子逻辑

启动时自动执行（幂等）：

1. **同步权限点**：将 `pkg/auth/perm.go` 中定义的全部权限点写入 `permissions` 表（仅插入缺失项）。
2. **确保超管角色**：创建/更新 `super_admin` 角色并关联全部权限点。
3. **创建初始超管**：仅当 `admins` 表为空 **且** `SEED_SUPER_ADMIN_PASSWORD` 非空时创建。

## 示例 docker-compose 环境变量

```yaml
services:
  svr-admin:
    image: svr-admin:latest
    ports:
      - "8080:8080"
    environment:
      - SERVER_MODE=release
      - DB_HOST=mysql
      - DB_PASSWORD=secret
      - DB_NAME=svr_admin
      - JWT_SECRET=my-secure-jwt-secret-at-least-32-chars
      - PROFILE_BASE_URL=http://svr-profile:8082
      - PROFILE_INTERNAL_KEY=shared-internal-key
      - SEED_SUPER_ADMIN_PASSWORD=Admin@2026!
```

## 降级行为

| 未配置项 | 影响范围 | 行为 |
|---|---|---|
| `PROFILE_BASE_URL` | 用户管控功能 | 返回错误 "svr-profile 下游未配置" |
| `SEED_SUPER_ADMIN_PASSWORD` | 初始超管 | 不自动创建，需手动通过 API 创建第一个管理员 |
| `JWT_SECRET` 使用默认值 | 安全性 | **仅开发环境**可用，生产必须配置 |
