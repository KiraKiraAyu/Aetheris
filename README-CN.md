# Aetheris

<p align="center">
  <img src="web/public/icon.svg" alt="Aetheris Logo" width="120" height="120" />
</p>

<p align="center">
  简体中文 | <a href="./README.md">English</a>
</p>

**Aetheris** 是一个轻量、可扩展且易用的开源聚合通知推送引擎。作为一个统一的通知网关，它默认支持**零外部依赖（SQLite+DBQueue）**的极简部署（极适合个人与中小项目）；同时支持通过 **Redis 与 PostgreSQL** 进行水平扩展以支撑大规模、高并发的企业级业务。一处接入，即可快速打通邮件、短信、站内信及各大办公软件群机器人投递通道。

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 特性

### 支持的推送渠道

- **邮件 (Email)**：支持 SMTP 服务，具备 STARTTLS / SSL / 无加密等连接模式
- **短信 (SMS)**：通用 HTTP 短信，支持在请求体中动态渲染变量与模板
- **通用 Webhook**：隔离私有 IP 沙箱并支持自定义 HMAC-SHA256 签名校验
- **站内信 (In-App)**：基于数据库持久化的收件箱，支持已读/未读状态管理
- **Telegram**：通过 Bot API 发送实时电报消息
- **Slack**：支持发送结构化消息至 Slack 频道 Webhook
- **Discord**：支持推送富文本卡片消息至 Discord Webhook 通路
- **飞书 (Feishu/Lark)**：支持向飞书群助手机器人 Webhook 推送富文本内容
- **钉钉 (DingTalk)**：支持钉钉群助手机器人 Webhook 并具备内置签名验证
- **企业微信 (WeCom)**：支持企业微信群机器人 Webhook 投递

### 核心功能

- **无依赖开箱即用**：默认采用 SQLite 数据库与内置数据库队列，无需安装 Redis 或 PostgreSQL 即可快速体验完整流程。
- **可扩展性架构**：支持无缝启用 Redis（基于 Asynq 队列）实现 Worker 水平扩展，并支持 PostgreSQL 存储以支撑生产级负载。
- **自动重试与退避**：发送失败的任务支持自动进行指数退避重试（最高 5 次）。
- **崩溃恢复机制**：自动检测并重新投递因 Worker 进程意外崩溃而中断的发送任务。
- **双语控制台**：管理后台支持中英文（EN / ZH）切换。
- **动态模板与变量助手**：支持配置变量占位符，并在控制台发送面板自动推导及注入变量模板。
- **多租户隔离与安全**：支持多 API Key 认证，具备安全的租户域边界校验。

## 快速开始

### 一键启动

如果你已经安装 Docker，执行这条命令来启动：

```bash
docker compose up -d
```

默认情况下，这将启动一个无外部依赖的轻量应用，使用 SQLite 数据库和内置数据库队列，不会运行 Redis 或 PostgreSQL 容器，很适合在个人服务器上使用。

启动完成后：

- **管理控制台**：`http://localhost:3000`
- **API 服务端**：`http://localhost:8080` (或通过 Nginx 代理请求：`http://localhost:3000`)

### 附加服务 (Redis / PostgreSQL)

Aetheris 在 Docker Compose 中提供了内置的 Redis 与 PostgreSQL 可选服务，可以通过 Docker Compose **profiles (服务分组)** 来按需启动：

- **启用 PostgreSQL**：
  ```bash
  COMPOSE_PROFILES=postgres DATABASE_URL=postgres://postgres:postgres@postgres:5432/aetheris?sslmode=disable docker compose up -d
  ```
- **启用 Redis**：
  ```bash
  COMPOSE_PROFILES=redis QUEUE_TYPE=redis REDIS_ADDR=redis:6379 docker compose up -d
  ```
- **同时启用两者**：
  ```bash
  COMPOSE_PROFILES=redis,postgres QUEUE_TYPE=redis REDIS_ADDR=redis:6379 DATABASE_URL=postgres://postgres:postgres@postgres:5432/aetheris?sslmode=disable docker compose up -d
  ```

如果你已经有独立的 PostgreSQL 实例，并希望在 Docker 中运行 Aetheris 应用服务：

```bash
DATABASE_URL="postgres://username:password@your-host:5432/dbname" docker compose up -d
```

### 本地运行

请确保系统安装了 [Go 1.25+](https://golang.org) 以及 [Node.js 20+](https://nodejs.org)。

本地启动：

```bash
pnpm install
pnpm dev
```

默认情况下，本地开发使用 SQLite 和内置数据库队列（DBQueue），不需要在本地运行 Redis 或 PostgreSQL 服务。

若要在本地使用 Redis 或 PostgreSQL，请复制 `.env.example` 为 `.env` 并进行配置：

- **Redis**：设置 `QUEUE_TYPE=redis` 并确保 `REDIS_ADDR` 指向你本地运行的的 Redis 实例。
- **PostgreSQL**：设置 `DATABASE_URL` 为你的 PostgreSQL 连接字符串。

启动后，访问前端管理台：`http://localhost:5178`

### 配置送信渠道

在设置页面启用并配置送信渠道。具体渠道配置可参考 [Aetheris 使用与配置指引](docs/GUIDE-CN.md)。

### API 接入

配置送信渠道后，发送通知只需向接口发送一个 HTTP POST 请求：

```bash
curl -X POST http://localhost:8080/send \
  -H 'Content-Type: application/json' \
  -d '{
    "recipient": "user@example.com",
    "channel": "email",
    "body": "邮件正文内容"
  }'
```

如果配置了默认收件人，则无需携带 `recipient` 字段：

```bash
curl -X POST http://localhost:8080/send \
  -H 'Content-Type: application/json' \
  -d '{
    "channel": "email",
    "body": "邮件正文内容"
  }'
```

如果配置了 API Key，只需添加 `X-API-Key` 请求头：

```bash
curl -X POST http://localhost:8080/send \
  -H 'Content-Type: application/json' \
  -H 'X-API-Key: your-tenant-api-key' \
  -d '{
    "channel": "email",
    "body": "邮件正文内容"
  }'
```

## 多租户管理

Aetheris 原生支持多租户。通过在后端进行静态的 API Key 与租户 ID（Tenant ID）映射，单个 Aetheris 部署即可同时向多个独立的团队或组织安全地提供推送服务。

### 1. 注册与授权租户（管理员）

要添加或修改租户，编辑项目根目录下的 `.env` 配置文件。在 `API_KEYS` 环境变量中，以逗号分隔指定 API Key 与租户 ID 的映射对：

```env
API_KEYS=key_of_shop:tenant_shop,key_of_ops:tenant_ops
```

保存并重启后端服务，授权即时生效。

### 2. 接入租户控制台（租户用户）

1. 在浏览器中打开 Aetheris 控制台（`http://localhost:3000`）。
2. 进入 **设置（Settings）** 页面。
3. 在 **API 访问配置（API Access）** 面板中，填写管理员分配的凭证：
   - **API Key**: 租户对应的密钥（例如 `key_of_shop`）。
   - **Tenant ID**: 租户 ID（例如 `tenant_shop`）。
4. 点击保存保存配置。

保存后，控制台会自动建立与后端的连接，并将所有页面（数据概览、历史通知、消息模板、通道设置）自动过滤并锁定在该租户的安全域内，实现与其他租户的完全隔离。

## 进阶配置指南

如果你需要了解所有 API 接口细节、各通道配置规范（JSON 字段）以及动态模板的渲染变量，请参阅文档：[Aetheris 使用与配置指引](docs/GUIDE-CN.md)。

## 开源协议

基于 MIT 协议开源。详情参见 `LICENSE` 文件。
