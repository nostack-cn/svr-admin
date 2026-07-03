# SSL 证书详解：从申请到部署

SSL 证书是网站安全的基础设施，它加密浏览器与服务器之间的数据传输，防止中间人攻击。本文从原理到实践，带你全面掌握 SSL 证书。

## 一、为什么需要 SSL

- **数据加密** — 防止数据在传输过程中被窃听和篡改
- **身份验证** — 确认你访问的网站是真实的，而非钓鱼站点
- **SEO 提升** — Google 将 HTTPS 作为排名信号
- **浏览器信任** — 没有 HTTPS 的网站会被标记为"不安全"

## 二、SSL/TLS 工作原理

SSL 握手过程（简化版）：

1. 客户端发起链接，发送支持的加密套件
2. 服务端返回 SSL 证书（包含公钥）
3. 客户端验证证书链的有效性
4. 客户端生成会话密钥，用公钥加密后发送给服务端
5. 服务端用私钥解密，获取会话密钥
6. 双方开始使用对称加密通信

## 三、证书类型对比

| 类型 | 验证级别 | 适用场景 | 价格 |
|------|---------|---------|------|
| DV（域名验证） | 低 | 个人博客、小型网站 | 免费 - 数百元 |
| OV（组织验证） | 中 | 企业官网、电商 | 数千元 |
| EV（扩展验证） | 高 | 银行、金融、大型企业 | 万元以上 |

## 四、免费证书方案

### 4.1 Let's Encrypt（推荐）

使用 Certbot 自动申请和续签：

```bash
# 安装 Certbot
sudo apt install certbot python3-certbot-nginx

# 申请证书（自动配置 Nginx）
sudo certbot --nginx -d example.com -d www.example.com

# 测试自动续签
sudo certbot renew --dry-run
```

> **关键提示**：Let's Encrypt 证书有效期 90 天，但 certbot 会通过 systemd timer 自动续签，配置好后无需手动干预。

### 4.2 云服务商免费证书

- 腾讯云 — 每个主域名 20 个免费 DV 证书（有效期 1 年）
- 阿里云 — 每个账号 20 个免费 DV 证书（DigiCert 品牌）
- Cloudflare — 免费 Universal SSL（反向代理模式）

## 五、Nginx 部署配置

```nginx
server {
    listen 80;
    server_name example.com www.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name example.com www.example.com;

    # 证书文件路径
    ssl_certificate     /etc/letsencrypt/live/example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/example.com/privkey.pem;

    # 安全配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # HSTS（告诉浏览器强制使用 HTTPS）
    add_header Strict-Transport-Security "max-age=31536000" always;

    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 六、常见问题

**Q：证书过期会怎样？**
A：浏览器显示"您的连接不是私密连接"，用户无法正常访问。建议设置到期前监控告警。

**Q：多域名如何配置？**
A：使用 SAN（Subject Alternative Name）证书，一个证书可以包含多个域名。

**Q：HTTP/2 需要 SSL 吗？**
A：是的，主流浏览器要求 HTTP/2 必须通过 HTTPS 启用。
