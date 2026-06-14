# Aetheris

<p align="center">
  <img src="web/public/icon.svg" alt="Aetheris Logo" width="120" height="120" />
</p>

<p align="center">
  <a href="./README-CN.md">简体中文</a> | English
</p>

**Aetheris** is a lightweight, scalable, and easy-to-use open-source aggregated notification delivery engine. Acting as a unified notification gateway, it supports a zero-dependency default startup (using SQLite & DBQueue) for lightweight deployments while seamlessly scaling to high-concurrency enterprise workloads via Redis & PostgreSQL. Connect once, and instantly reach email, SMS, in-app inboxes, and all major workplace chat bots.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

### Supported Delivery Channels

- **Email**: SMTP with STARTTLS / SSL / no encryption modes
- **SMS**: Dynamic template resolution for HTTP API request bodies across SMS providers
- **Generic Webhook**: Private IP sandbox isolation and custom HMAC-SHA256 payload signing
- **In-App**: Database-persisted user inbox with read/unread tracking
- **Telegram**: Native Bot API delivery
- **Slack**: Message routing to incoming webhook endpoints
- **Discord**: Rich card messages pushed to Discord webhook channels
- **Feishu (Lark)**: Structured messages posted to Feishu/Lark chatbot webhook URLs
- **DingTalk**: Group chatbot webhook with built-in signature verification
- **WeCom (WeChat Work)**: Group robot webhook message delivery

### Core Capabilities

- **Zero-Dependency Startup**: Run instantly using a local SQLite database and an in-database queue without needing external services like Redis or PostgreSQL.
- **Optional Scaling (Redis & PostgreSQL)**: Seamlessly opt-in to Redis for horizontal worker scaling (via Asynq) and PostgreSQL for production-ready persistence.
- **Automatic Retries & Backoff**: Failed delivery attempts are automatically retried using exponential backoff (up to 5 attempts).
- **Worker Crash Recovery**: Automatically reclaims and restarts tasks that were left running if a worker process crashes mid-delivery.
- **Bilingual Console**: A fully responsive management dashboard localized in English (EN) and Chinese (ZH).
- **Dynamic Templates**: Dynamic placeholders and variables helper to preview and resolve subject/body payloads automatically.
- **Tenant Scope & Security**: Secure multi-tenant API key authentication with tenant isolation checks.

## Quick Start

### One-Click Launch

If you have Docker installed, run this to get started:

```bash
docker compose up -d
```

By default, this launches a lightweight, zero-dependency application with SQLite and an embedded database queue — no Redis or PostgreSQL containers are spun up, making it ideal for personal servers.

Once started:

- **Management Console**: `http://localhost:3000`
- **API Server**: `http://localhost:8080` (or via Nginx proxy: `http://localhost:3000`)

### Optional Services (Redis / PostgreSQL)

Aetheris includes optional built-in services for Redis and PostgreSQL managed via Docker Compose **profiles**:

- **Enable PostgreSQL**:
  ```bash
  COMPOSE_PROFILES=postgres DATABASE_URL=postgres://postgres:postgres@postgres:5432/aetheris?sslmode=disable docker compose up -d
  ```
- **Enable Redis**:
  ```bash
  COMPOSE_PROFILES=redis QUEUE_TYPE=redis REDIS_ADDR=redis:6379 docker compose up -d
  ```
- **Enable Both**:
  ```bash
  COMPOSE_PROFILES=redis,postgres QUEUE_TYPE=redis REDIS_ADDR=redis:6379 DATABASE_URL=postgres://postgres:postgres@postgres:5432/aetheris?sslmode=disable docker compose up -d
  ```

If you already have a PostgreSQL instance and want to run Aetheris services in Docker:

1. Start the services specifying your custom `DATABASE_URL`:
   ```bash
   DATABASE_URL="postgres://username:password@your-host:5432/dbname?sslmode=require" docker compose up -d
   ```

### Local Development

Make sure your system has [Go 1.25+](https://golang.org) and [Node.js 20+](https://nodejs.org) installed.

To start locally:

```bash
pnpm install
pnpm dev
```

By default, this uses SQLite and the Database Queue (DBQueue), requiring **no running Redis or PostgreSQL server**.

To use Redis or PostgreSQL locally, copy `.env.example` to `.env` and configure:

- **Redis**: Set `QUEUE_TYPE=redis` and ensure `REDIS_ADDR` points to your running Redis instance.
- **PostgreSQL**: Set `DATABASE_URL` to your PostgreSQL connection string.

Once running, access the management console at `http://localhost:5178`.

### Configure Delivery Channels

Enable and configure delivery channels on the Settings page. See the [Aetheris Guide](docs/GUIDE.md) for per-channel configuration details.

### API Access

Once channels are configured, sending a notification is a single HTTP POST request:

```bash
curl -X POST http://localhost:8080/send \
  -H 'Content-Type: application/json' \
  -d '{
    "recipient": "user@example.com",
    "channel": "email",
    "body": "Email body content"
  }'
```

If a default recipient is configured, the `recipient` field can be omitted:

```bash
curl -X POST http://localhost:8080/send \
  -H 'Content-Type: application/json' \
  -d '{
    "channel": "email",
    "body": "Email body content"
  }'
```

If an API Key is configured, add the `X-API-Key` header:

```bash
curl -X POST http://localhost:8080/send \
  -H 'Content-Type: application/json' \
  -H 'X-API-Key: your-tenant-api-key' \
  -d '{
    "channel": "email",
    "body": "Email body content"
  }'
```

## Multi-Tenancy Management

Aetheris supports multi-tenancy natively through static environment-level API Key mappings. This allows a single deployment of Aetheris to serve multiple independent organizations or teams securely.

### 1. Registering Tenants (Admin)

To create or register tenants, edit the `API_KEYS` variable in your `.env` configuration file. Specify API keys mapped to unique Tenant IDs using a comma-separated format:

```env
API_KEYS=key_for_team_a:tenant_a,key_for_team_b:tenant_b
```

Restart the containers or server. The API keys are now securely associated with their respective tenants.

### 2. Accessing the Tenant Space (User)

1. Open the Aetheris console (`http://localhost:3000`).
2. Navigate to the **Settings** page.
3. In the **API Access** panel, fill in:
   - **API Key**: The key assigned to your team (e.g., `key_for_team_a`).
   - **Tenant ID**: The ID assigned to your team (e.g., `tenant_a`).
4. Save the configuration.

Once saved, the management console will connect to the backend and automatically filter and isolate all views, configurations, templates, and history logs to your tenant space.

## Advanced Configuration

For detailed API endpoint documentation, per-channel configuration schemas (JSON fields), and dynamic template variables, see the [Aetheris Guide](docs/GUIDE.md).

## License

Open source under the MIT License. See the `LICENSE` file for details.
