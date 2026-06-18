import type {
  Channel,
  CreateNotificationPayload,
  DeliveryAttempt,
  InAppMessage,
  NotificationRecord,
  NotificationStatus,
  NotificationTemplate,
  ChannelConfig,
} from './types'

// Helper to generate IDs
function genId(prefix: string): string {
  return `${prefix}_${Math.random().toString(36).substring(2, 11)}`
}

// LocalStorage Keys
const KEYS = {
  notifications: 'aetheris.mock.notifications',
  attempts: 'aetheris.mock.attempts',
  inbox: 'aetheris.mock.inbox',
  templates: 'aetheris.mock.templates',
  configs: 'aetheris.mock.configs',
}

// Seed Data
const DEFAULT_CONFIGS: ChannelConfig[] = [
  { channel: 'email', enabled: true, config: '{"host":"smtp.mailtrap.io","port":587,"username":"aetheris_mock","from":"no-reply@aetheris.io"}' },
  { channel: 'sms', enabled: true, config: '{"url":"https://api.sms-gateway.internal/send","method":"POST"}' },
  { channel: 'webhook', enabled: true, config: '{"url":"https://api.mycompany.com/webhook","method":"POST","signing_secret":"whsec_mock_key"}' },
  { channel: 'in_app', enabled: true, config: '{}' },
  { channel: 'telegram', enabled: true, config: '{"bot_token":"682910392:AAH9301823901-XyZ","parse_mode":"markdown"}' },
  { channel: 'slack', enabled: false, config: '{"url":""}' },
  { channel: 'discord', enabled: true, config: '{"url":"https://discord.com/api/webhooks/12345/abcde"}' },
  { channel: 'feishu', enabled: true, config: '{"url":"https://open.feishu.cn/open-apis/bot/v2/hook/xyz"}' },
  { channel: 'dingtalk', enabled: false, config: '{"url":""}' },
  { channel: 'wecom', enabled: false, config: '{"url":""}' },
]

const DEFAULT_TEMPLATES: NotificationTemplate[] = [
  {
    id: 'tpl_welcome',
    key: 'user_welcome',
    channel: 'email',
    title_template: 'Welcome to Aetheris, {{.recipient}}!',
    body_template: 'Hello {{.recipient}},\n\nThank you for signing up to Aetheris! Your tenant workspace is ready. You can now access your notification control center and set up your delivery channels.\n\nCheers,\nTeam Aetheris',
    created_at: new Date(Date.now() - 86400000 * 3).toISOString(),
    updated_at: new Date(Date.now() - 86400000 * 3).toISOString(),
  },
  {
    id: 'tpl_otp',
    key: 'verify_otp',
    channel: 'sms',
    title_template: 'Security Code',
    body_template: 'Your Aetheris verification code is: {{.code}} (valid for 5 minutes). Please do not share this code.',
    created_at: new Date(Date.now() - 86400000 * 2).toISOString(),
    updated_at: new Date(Date.now() - 86400000 * 2).toISOString(),
  },
  {
    id: 'tpl_server_outage',
    key: 'server_outage',
    channel: 'telegram',
    title_template: '🔴 CRITICAL ALERT: Outage Detected',
    body_template: 'Service: {{.service}}\nStatus: Down (HTTP {{.status_code}})\nTime: {{.timestamp}}\nError: Connection reset by peer. Auto-recovering worker instances.',
    created_at: new Date(Date.now() - 3600000 * 4).toISOString(),
    updated_at: new Date(Date.now() - 3600000 * 4).toISOString(),
  },
  {
    id: 'tpl_payment_webhook',
    key: 'payment_success',
    channel: 'webhook',
    title_template: 'Payment Succeeded',
    body_template: '{"event":"payment.succeeded","data":{"invoice":"{{.invoice}}","amount":{{.amount}},"currency":"USD","customer":"{{.customer}}"}}',
    created_at: new Date(Date.now() - 3600000 * 2).toISOString(),
    updated_at: new Date(Date.now() - 3600000 * 2).toISOString(),
  }
]

const DEFAULT_NOTIFICATIONS: NotificationRecord[] = [
  // --- JUNE 2026 ---
  {
    id: 'ntf_1',
    tenant_id: 'default',
    recipient: 'test-user@company.com',
    channel: 'email',
    template_key: 'user_welcome',
    title: 'Welcome to Aetheris, test-user@company.com!',
    body: 'Hello test-user@company.com,\n\nThank you for signing up to Aetheris! Your tenant workspace is ready. You can now access your notification control center and set up your delivery channels.\n\nCheers,\nTeam Aetheris',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-06-17T18:00:00Z',
    created_at: '2026-06-17T18:00:00Z',
    updated_at: '2026-06-17T18:00:00Z',
  },
  {
    id: 'ntf_2',
    tenant_id: 'default',
    recipient: '+1234567890',
    channel: 'sms',
    template_key: 'verify_otp',
    title: 'Security Code',
    body: 'Your Aetheris verification code is: 582190 (valid for 5 minutes). Please do not share this code.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-06-17T14:30:00Z',
    created_at: '2026-06-17T14:30:00Z',
    updated_at: '2026-06-17T14:30:00Z',
  },
  {
    id: 'ntf_3',
    tenant_id: 'default',
    recipient: '@ops_channel',
    channel: 'telegram',
    template_key: 'server_outage',
    title: '🔴 CRITICAL ALERT: Outage Detected',
    body: 'Service: main-db-cluster\nStatus: Down (HTTP 503)\nTime: 2026-06-17 12:00:00\nError: Connection reset by peer. Auto-recovering worker instances.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-06-17T12:05:00Z',
    created_at: '2026-06-17T12:00:00Z',
    updated_at: '2026-06-17T12:05:00Z',
  },
  {
    id: 'ntf_4',
    tenant_id: 'default',
    recipient: 'https://api.mycompany.com/webhook',
    channel: 'webhook',
    template_key: 'payment_success',
    title: 'Payment Succeeded',
    body: '{"event":"payment.succeeded","data":{"invoice":"inv_10284","amount":49,"currency":"USD","customer":"cus_89102"}}',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-06-16T10:15:00Z',
    created_at: '2026-06-16T10:15:00Z',
    updated_at: '2026-06-16T10:15:00Z',
  },
  {
    id: 'ntf_5',
    tenant_id: 'default',
    recipient: 'user_998',
    channel: 'in_app',
    title: 'System Maintenance Complete',
    body: 'We have finished upgrading the database cluster to PostgreSQL 16. Services are fully operational and query latencies have improved by 15%.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-06-15T09:00:00Z',
    created_at: '2026-06-15T09:00:00Z',
    updated_at: '2026-06-15T09:00:00Z',
  },
  {
    id: 'ntf_6',
    tenant_id: 'default',
    recipient: '#general-alerts',
    channel: 'slack',
    title: 'Weekly Performance Report',
    body: 'Aetheris Weekly Summary:\n- Total Dispatched: 1,482\n- Delivery Rate: 99.4%\n- API Response Time: 48ms',
    status: 'failed',
    last_error: 'Post "https://hooks.slack.com/services/...": dial tcp: lookup hooks.slack.com on 127.0.0.11:53: no such host',
    aggregate_count: 1,
    created_at: '2026-06-14T22:00:00Z',
    updated_at: '2026-06-14T22:00:00Z',
  },
  {
    id: 'ntf_7',
    tenant_id: 'default',
    recipient: 'user_998',
    channel: 'in_app',
    title: 'Welcome to your In-App notifications!',
    body: 'This inbox simulator will display any notification requests targeted to the "in_app" channel for this user.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-06-14T08:00:00Z',
    created_at: '2026-06-14T08:00:00Z',
    updated_at: '2026-06-14T08:00:00Z',
  },
  {
    id: 'ntf_8',
    tenant_id: 'default',
    recipient: 'https://discord.com/api/webhooks/alerts',
    channel: 'discord',
    title: 'New User Registered',
    body: 'Username: alex_dev_99\nEmail: alex.jones@outlook.com\nReferrer: Twitter Ad Campaign',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-06-12T16:40:00Z',
    created_at: '2026-06-12T16:40:00Z',
    updated_at: '2026-06-12T16:40:00Z',
  },
  {
    id: 'ntf_9',
    tenant_id: 'default',
    recipient: 'https://open.feishu.cn/open-apis/bot/xyz',
    channel: 'feishu',
    title: 'Daily Sales Briefing',
    body: '📊 Daily Summary (June 10):\n- Total ARR: $1,204,920 (+2.4%)\n- New signups: 142\n- Churn rate: 0.1%',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-06-10T18:30:00Z',
    created_at: '2026-06-10T18:30:00Z',
    updated_at: '2026-06-10T18:30:00Z',
  },
  {
    id: 'ntf_10',
    tenant_id: 'default',
    recipient: 'https://oapi.dingtalk.com/robot/send',
    channel: 'dingtalk',
    title: 'Attendance Reminder',
    body: 'Team leads, please verify your department attendance records in the admin console before 6:00 PM today.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-06-08T09:10:00Z',
    created_at: '2026-06-08T09:10:00Z',
    updated_at: '2026-06-08T09:10:00Z',
  },
  {
    id: 'ntf_11',
    tenant_id: 'default',
    recipient: 'https://qyapi.weixin.qq.com/webhook',
    channel: 'wecom',
    title: 'Project Kickoff Announcement',
    body: 'Project Aetheris v2 kickoff starts tomorrow at 10:00 AM in Meeting Room 4A. Remote Zoom link is attached.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-06-05T14:00:00Z',
    created_at: '2026-06-05T14:00:00Z',
    updated_at: '2026-06-05T14:00:00Z',
  },

  // --- MAY 2026 ---
  {
    id: 'ntf_12',
    tenant_id: 'default',
    recipient: 'billing-contact@myclient.com',
    channel: 'email',
    title: 'Monthly Subscription Invoice - May 2026',
    body: 'Hello,\n\nYour invoice for May 2026 (inv_09841) is now available. An automatic charge of $99.00 has been applied to your card on file.\n\nThank you,\nAetheris Billing',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-05-28T08:15:00Z',
    created_at: '2026-05-28T08:15:00Z',
    updated_at: '2026-05-28T08:15:00Z',
  },
  {
    id: 'ntf_13',
    tenant_id: 'default',
    recipient: '+1234567890',
    channel: 'sms',
    title: 'Security Alert',
    body: 'Your Aetheris password was changed on May 24, 2026 at 15:42 UTC. If this was not you, contact support immediately.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-05-24T15:43:00Z',
    created_at: '2026-05-24T15:42:00Z',
    updated_at: '2026-05-24T15:43:00Z',
  },
  {
    id: 'ntf_14',
    tenant_id: 'default',
    recipient: 'https://api.mycompany.com/webhook',
    channel: 'webhook',
    title: 'User Created Callback',
    body: '{"event":"user.created","data":{"id":"usr_0891a","email":"new-dev@mycompany.com","role":"editor"}}',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-05-20T11:30:00Z',
    created_at: '2026-05-20T11:30:00Z',
    updated_at: '2026-05-20T11:30:00Z',
  },
  {
    id: 'ntf_15',
    tenant_id: 'default',
    recipient: 'user_998',
    channel: 'in_app',
    title: 'New Comment on Pull Request',
    body: 'kyle_architect commented on PR #412: "Looks solid. We should double check if the SQLite database lock handles high concurrent connection writes under stress."',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-05-18T16:22:00Z',
    created_at: '2026-05-18T16:22:00Z',
    updated_at: '2026-05-18T16:22:00Z',
  },
  {
    id: 'ntf_16',
    tenant_id: 'default',
    recipient: '@ops_channel',
    channel: 'telegram',
    title: 'Server Backup Complete',
    body: 'System backup task bk_20260515 completed. Total files archived: 148,203. Backup stored securely in S3 bucket.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-05-15T03:00:00Z',
    created_at: '2026-05-15T03:00:00Z',
    updated_at: '2026-05-15T03:00:00Z',
  },
  {
    id: 'ntf_17',
    tenant_id: 'default',
    recipient: '#general-alerts',
    channel: 'slack',
    title: 'Deployment Succeeded',
    body: 'Environment: Production\nCommit: [a9f821b] Merge branch features/redis-postgres\nDuration: 2m 45s\nStatus: Healthy',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-05-12T10:05:00Z',
    created_at: '2026-05-12T10:03:00Z',
    updated_at: '2026-05-12T10:05:00Z',
  },
  {
    id: 'ntf_18',
    tenant_id: 'default',
    recipient: 'https://discord.com/api/webhooks/alerts',
    channel: 'discord',
    title: 'Community Announcement',
    body: '📢 Version 1.8.0-RC is now live in beta testing! Head over to the beta-testing channel to get early release binaries.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-05-08T18:00:00Z',
    created_at: '2026-05-08T18:00:00Z',
    updated_at: '2026-05-08T18:00:00Z',
  },
  {
    id: 'ntf_19',
    tenant_id: 'default',
    recipient: 'https://open.feishu.cn/open-apis/bot/xyz',
    channel: 'feishu',
    title: 'Approval Request Received',
    body: 'Leave Request: Vacation (Sarah Connor)\nDates: May 20 - May 24 (5 days)\nWaiting for your approval.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-05-04T09:15:00Z',
    created_at: '2026-05-04T09:15:00Z',
    updated_at: '2026-05-04T09:15:00Z',
  },

  // --- APRIL 2026 ---
  {
    id: 'ntf_20',
    tenant_id: 'default',
    recipient: 'admin@mycompany.com',
    channel: 'email',
    title: 'Security Alert: Login from New IP',
    body: 'A login request was detected from IP address 192.168.1.185 (Chicago, US) using Chrome on Windows. If this wasn\'t you, please reset your credentials.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-04-28T22:12:00Z',
    created_at: '2026-04-28T22:12:00Z',
    updated_at: '2026-04-28T22:12:00Z',
  },
  {
    id: 'ntf_21',
    tenant_id: 'default',
    recipient: '+1234567890',
    channel: 'sms',
    title: 'Low Balance Reminder',
    body: 'Your SMS gateway wallet balance is below $10.00. Recharge now to avoid auto-routing failures on transactional codes.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-04-24T10:00:00Z',
    created_at: '2026-04-24T10:00:00Z',
    updated_at: '2026-04-24T10:00:00Z',
  },
  {
    id: 'ntf_22',
    tenant_id: 'default',
    recipient: 'https://api.mycompany.com/webhook',
    channel: 'webhook',
    title: 'Order Dispatched Callback',
    body: '{"event":"order.dispatched","data":{"order_id":"ord_00984","tracking_number":"TRK92810392","carrier":"UPS"}}',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-04-20T15:30:00Z',
    created_at: '2026-04-20T15:30:00Z',
    updated_at: '2026-04-20T15:30:00Z',
  },
  {
    id: 'ntf_23',
    tenant_id: 'default',
    recipient: 'user_998',
    channel: 'in_app',
    title: 'New Task Assigned',
    body: 'You have been assigned to task: "Integrate Discord Webhook templates inside templates designer". Due Date: April 25.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-04-18T09:12:00Z',
    created_at: '2026-04-18T09:12:00Z',
    updated_at: '2026-04-18T09:12:00Z',
  },
  {
    id: 'ntf_24',
    tenant_id: 'default',
    recipient: '@ops_channel',
    channel: 'telegram',
    title: '⚠️ System Performance Alert',
    body: 'Service: frontend-server\nMetric: Memory usage exceeded 90%\nAction: Retrying auto-restarts.',
    status: 'failed',
    last_error: 'Telegram API request failed: HTTP 400 Bad Request: Chat not found',
    aggregate_count: 1,
    created_at: '2026-04-15T04:10:00Z',
    updated_at: '2026-04-15T04:10:00Z',
  },
  {
    id: 'ntf_25',
    tenant_id: 'default',
    recipient: '#general-alerts',
    channel: 'slack',
    title: 'GitHub Action Run Failed',
    body: 'Workflow: Continuous Integration\nBranch: master\nRun ID: 9182390123\nError: Lint checks failed on web module.',
    status: 'failed',
    last_error: 'Slack Webhook returned HTTP 404 Not Found: Channel archived',
    aggregate_count: 1,
    created_at: '2026-04-10T11:45:00Z',
    updated_at: '2026-04-10T11:45:00Z',
  },

  // --- MARCH 2026 ---
  {
    id: 'ntf_26',
    tenant_id: 'default',
    recipient: 'security-admin@mycompany.com',
    channel: 'email',
    title: 'New API Key Generated',
    body: 'Hello,\n\nA new API Key (key_****************abcd) has been generated for tenant "default". If you did not execute this action, please revoke the key immediately.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-03-28T09:15:00Z',
    created_at: '2026-03-28T09:15:00Z',
    updated_at: '2026-03-28T09:15:00Z',
  },
  {
    id: 'ntf_27',
    tenant_id: 'default',
    recipient: '+1234567890',
    channel: 'sms',
    title: 'OTP Verification Code',
    body: 'Your Aetheris verification code is: 489201. Do not share this code.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-03-24T18:02:00Z',
    created_at: '2026-03-24T18:02:00Z',
    updated_at: '2026-03-24T18:02:00Z',
  },
  {
    id: 'ntf_28',
    tenant_id: 'default',
    recipient: 'https://api.mycompany.com/webhook',
    channel: 'webhook',
    title: 'Subscription Cancelled Callback',
    body: '{"event":"subscription.cancelled","data":{"id":"sub_0019283","customer":"cus_89102","reason":"unsubscribed"}}',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-03-20T16:30:00Z',
    created_at: '2026-03-20T16:30:00Z',
    updated_at: '2026-03-20T16:30:00Z',
  },
  {
    id: 'ntf_29',
    tenant_id: 'default',
    recipient: 'user_998',
    channel: 'in_app',
    title: 'Team Member Joined Workspace',
    body: 'User "kyle_architect" (kyle@company.com) has accepted your invitation and joined workspace: "default".',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-03-15T10:04:00Z',
    created_at: '2026-03-15T10:04:00Z',
    updated_at: '2026-03-15T10:04:00Z',
  },
  {
    id: 'ntf_30',
    tenant_id: 'default',
    recipient: 'https://discord.com/api/webhooks/alerts',
    channel: 'discord',
    title: 'Workspace Boosted',
    body: '🚀 tenant "default" has successfully enabled Postgres integration! Metrics database is now running on a dedicated AWS Aurora instance.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-03-10T14:20:00Z',
    created_at: '2026-03-10T14:20:00Z',
    updated_at: '2026-03-10T14:20:00Z',
  },

  // --- FEBRUARY 2026 ---
  {
    id: 'ntf_31',
    tenant_id: 'default',
    recipient: 'newsletter@mycompany.com',
    channel: 'email',
    title: 'Newsletter Update: Q1 Features',
    body: 'Hello,\n\nWe are excited to launch Aetheris v1.5! In this release: SMTP custom headers, Telegram Markdown parsing support, and automatic exponential backoff retries.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-02-28T09:00:00Z',
    created_at: '2026-02-28T09:00:00Z',
    updated_at: '2026-02-28T09:00:00Z',
  },
  {
    id: 'ntf_32',
    tenant_id: 'default',
    recipient: '+1234567890',
    channel: 'sms',
    title: 'Delivery Notice',
    body: 'Your package with tracking TRK91823901 has arrived at the reception desk. Pick up before 8:00 PM.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-02-24T17:15:00Z',
    created_at: '2026-02-24T17:15:00Z',
    updated_at: '2026-02-24T17:15:00Z',
  },
  {
    id: 'ntf_33',
    tenant_id: 'default',
    recipient: 'https://api.mycompany.com/webhook',
    channel: 'webhook',
    title: 'Charge Refunded Callback',
    body: '{"event":"charge.refunded","data":{"charge_id":"chg_8910283","amount":49,"currency":"USD"}}',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-02-20T12:00:00Z',
    created_at: '2026-02-20T12:00:00Z',
    updated_at: '2026-02-20T12:00:00Z',
  },
  {
    id: 'ntf_34',
    tenant_id: 'default',
    recipient: 'user_998',
    channel: 'in_app',
    title: 'System Upgrade Notice',
    body: 'Aetheris core engine is upgrading to v1.5.0-rc2 tonight at 02:00 UTC. Minor send latency fluctuations (50ms) are expected during worker container re-scheduling.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-02-15T09:30:00Z',
    created_at: '2026-02-15T09:30:00Z',
    updated_at: '2026-02-15T09:30:00Z',
  },

  // --- JANUARY 2026 ---
  {
    id: 'ntf_35',
    tenant_id: 'default',
    recipient: 'welcome-verify@newcompany.com',
    channel: 'email',
    title: 'Account Verification',
    body: 'Hello,\n\nPlease verify your account by clicking the following link: https://aetheris.io/verify?token=verify_mock_token_9823901\n\nThanks,\nTeam Aetheris',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-01-28T10:12:00Z',
    created_at: '2026-01-28T10:12:00Z',
    updated_at: '2026-01-28T10:12:00Z',
  },
  {
    id: 'ntf_36',
    tenant_id: 'default',
    recipient: '+1234567890',
    channel: 'sms',
    title: 'Bank Verification Code',
    body: 'Your bank wire transfer confirmation code is: 104829 (valid for 3 minutes). Do not share.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-01-24T11:05:00Z',
    created_at: '2026-01-24T11:05:00Z',
    updated_at: '2026-01-24T11:05:00Z',
  },
  {
    id: 'ntf_37',
    tenant_id: 'default',
    recipient: 'https://api.mycompany.com/webhook',
    channel: 'webhook',
    title: 'Customer Updated Callback',
    body: '{"event":"customer.updated","data":{"customer_id":"cus_89102","name":"John Doe","email":"john@doe.com"}}',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-01-20T14:40:00Z',
    created_at: '2026-01-20T14:40:00Z',
    updated_at: '2026-01-20T14:40:00Z',
  },
  {
    id: 'ntf_38',
    tenant_id: 'default',
    recipient: 'user_998',
    channel: 'in_app',
    title: 'Storage Quota Warning',
    body: 'Your workspace log retention database has reached 80% capacity (7.9 GB / 10 GB). Older logs will be auto-archived after 30 days.',
    status: 'delivered',
    aggregate_count: 1,
    delivered_at: '2026-01-15T09:00:00Z',
    created_at: '2026-01-15T09:00:00Z',
    updated_at: '2026-01-15T09:00:00Z',
  }
]

const DEFAULT_ATTEMPTS: DeliveryAttempt[] = [
  // --- JUNE ---
  { id: 'att_1', notification_id: 'ntf_1', tenant_id: 'default', channel: 'email', attempt: 1, status: 'delivered', provider_message_id: '<msg-1@smtp.mailtrap.io>', started_at: '2026-06-17T18:00:00Z', finished_at: '2026-06-17T18:00:01Z', duration_ms: 780 },
  { id: 'att_2', notification_id: 'ntf_2', tenant_id: 'default', channel: 'sms', attempt: 1, status: 'delivered', provider_message_id: 'sms_msg_002', started_at: '2026-06-17T14:30:00Z', finished_at: '2026-06-17T14:30:01Z', duration_ms: 310 },
  { id: 'att_3', notification_id: 'ntf_3', tenant_id: 'default', channel: 'telegram', attempt: 1, status: 'delivered', provider_message_id: 'tg_up_003', started_at: '2026-06-17T12:00:00Z', finished_at: '2026-06-17T12:00:01Z', duration_ms: 220 },
  { id: 'att_4', notification_id: 'ntf_4', tenant_id: 'default', channel: 'webhook', attempt: 1, status: 'delivered', provider_message_id: 'wh_att_004', started_at: '2026-06-16T10:15:00Z', finished_at: '2026-06-16T10:15:01Z', duration_ms: 480 },
  { id: 'att_5', notification_id: 'ntf_5', tenant_id: 'default', channel: 'in_app', attempt: 1, status: 'delivered', provider_message_id: 'inapp_msg_005', started_at: '2026-06-15T09:00:00Z', finished_at: '2026-06-15T09:00:00Z', duration_ms: 45 },
  { id: 'att_6_1', notification_id: 'ntf_6', tenant_id: 'default', channel: 'slack', attempt: 1, status: 'failed', last_error: 'dial tcp: lookup hooks.slack.com: no such host', started_at: '2026-06-14T22:00:00Z', finished_at: '2026-06-14T22:00:05Z', duration_ms: 5000 },
  { id: 'att_6_2', notification_id: 'ntf_6', tenant_id: 'default', channel: 'slack', attempt: 2, status: 'failed', last_error: 'dial tcp: lookup hooks.slack.com: no such host', started_at: '2026-06-14T22:01:00Z', finished_at: '2026-06-14T22:01:05Z', duration_ms: 5000 },
  { id: 'att_7', notification_id: 'ntf_7', tenant_id: 'default', channel: 'in_app', attempt: 1, status: 'delivered', provider_message_id: 'inapp_msg_007', started_at: '2026-06-14T08:00:00Z', finished_at: '2026-06-14T08:00:00Z', duration_ms: 30 },
  { id: 'att_8', notification_id: 'ntf_8', tenant_id: 'default', channel: 'discord', attempt: 1, status: 'delivered', provider_message_id: 'disc_008', started_at: '2026-06-12T16:40:00Z', finished_at: '2026-06-12T16:40:01Z', duration_ms: 310 },
  { id: 'att_9', notification_id: 'ntf_9', tenant_id: 'default', channel: 'feishu', attempt: 1, status: 'delivered', provider_message_id: 'feishu_009', started_at: '2026-06-10T18:30:00Z', finished_at: '2026-06-10T18:30:01Z', duration_ms: 290 },
  { id: 'att_10', notification_id: 'ntf_10', tenant_id: 'default', channel: 'dingtalk', attempt: 1, status: 'delivered', provider_message_id: 'ding_010', started_at: '2026-06-08T09:10:00Z', finished_at: '2026-06-08T09:10:01Z', duration_ms: 180 },
  { id: 'att_11', notification_id: 'ntf_11', tenant_id: 'default', channel: 'wecom', attempt: 1, status: 'delivered', provider_message_id: 'wecom_011', started_at: '2026-06-05T14:00:00Z', finished_at: '2026-06-05T14:00:01Z', duration_ms: 220 },

  // --- MAY ---
  { id: 'att_12', notification_id: 'ntf_12', tenant_id: 'default', channel: 'email', attempt: 1, status: 'delivered', provider_message_id: '<msg-12@smtp.mailtrap.io>', started_at: '2026-05-28T08:15:00Z', finished_at: '2026-05-28T08:15:01Z', duration_ms: 650 },
  { id: 'att_13', notification_id: 'ntf_13', tenant_id: 'default', channel: 'sms', attempt: 1, status: 'delivered', provider_message_id: 'sms_msg_013', started_at: '2026-05-24T15:42:00Z', finished_at: '2026-05-24T15:43:00Z', duration_ms: 280 },
  { id: 'att_14', notification_id: 'ntf_14', tenant_id: 'default', channel: 'webhook', attempt: 1, status: 'delivered', provider_message_id: 'wh_att_014', started_at: '2026-05-20T11:30:00Z', finished_at: '2026-05-20T11:30:01Z', duration_ms: 380 },
  { id: 'att_15', notification_id: 'ntf_15', tenant_id: 'default', channel: 'in_app', attempt: 1, status: 'delivered', provider_message_id: 'inapp_msg_015', started_at: '2026-05-18T16:22:00Z', finished_at: '2026-05-18T16:22:00Z', duration_ms: 40 },
  { id: 'att_16', notification_id: 'ntf_16', tenant_id: 'default', channel: 'telegram', attempt: 1, status: 'delivered', provider_message_id: 'tg_up_016', started_at: '2026-05-15T03:00:00Z', finished_at: '2026-05-15T03:00:01Z', duration_ms: 190 },
  { id: 'att_17', notification_id: 'ntf_17', tenant_id: 'default', channel: 'slack', attempt: 1, status: 'delivered', provider_message_id: 'slack_017', started_at: '2026-05-12T10:03:00Z', finished_at: '2026-05-12T10:05:00Z', duration_ms: 490 },
  { id: 'att_18', notification_id: 'ntf_18', tenant_id: 'default', channel: 'discord', attempt: 1, status: 'delivered', provider_message_id: 'disc_018', started_at: '2026-05-08T18:00:00Z', finished_at: '2026-05-08T18:00:01Z', duration_ms: 310 },
  { id: 'att_19', notification_id: 'ntf_19', tenant_id: 'default', channel: 'feishu', attempt: 1, status: 'delivered', provider_message_id: 'feishu_019', started_at: '2026-05-04T09:15:00Z', finished_at: '2026-05-04T09:15:01Z', duration_ms: 220 },

  // --- APRIL ---
  { id: 'att_20', notification_id: 'ntf_20', tenant_id: 'default', channel: 'email', attempt: 1, status: 'delivered', provider_message_id: '<msg-20@smtp.mailtrap.io>', started_at: '2026-04-28T22:12:00Z', finished_at: '2026-04-28T22:12:01Z', duration_ms: 810 },
  { id: 'att_21', notification_id: 'ntf_21', tenant_id: 'default', channel: 'sms', attempt: 1, status: 'delivered', provider_message_id: 'sms_msg_021', started_at: '2026-04-24T10:00:00Z', finished_at: '2026-04-24T10:00:01Z', duration_ms: 240 },
  { id: 'att_22', notification_id: 'ntf_22', tenant_id: 'default', channel: 'webhook', attempt: 1, status: 'delivered', provider_message_id: 'wh_att_022', started_at: '2026-04-20T15:30:00Z', finished_at: '2026-04-20T15:30:01Z', duration_ms: 410 },
  { id: 'att_23', notification_id: 'ntf_23', tenant_id: 'default', channel: 'in_app', attempt: 1, status: 'delivered', provider_message_id: 'inapp_msg_023', started_at: '2026-04-18T09:12:00Z', finished_at: '2026-04-18T09:12:00Z', duration_ms: 50 },
  { id: 'att_24', notification_id: 'ntf_24', tenant_id: 'default', channel: 'telegram', attempt: 1, status: 'failed', last_error: 'Telegram API request failed: HTTP 400 Bad Request', started_at: '2026-04-15T04:10:00Z', finished_at: '2026-04-15T04:10:02Z', duration_ms: 2100 },
  { id: 'att_25', notification_id: 'ntf_25', tenant_id: 'default', channel: 'slack', attempt: 1, status: 'failed', last_error: 'Slack Webhook returned HTTP 404 Not Found', started_at: '2026-04-10T11:45:00Z', finished_at: '2026-04-10T11:45:01Z', duration_ms: 1200 },

  // --- MARCH ---
  { id: 'att_26', notification_id: 'ntf_26', tenant_id: 'default', channel: 'email', attempt: 1, status: 'delivered', provider_message_id: '<msg-26@smtp.mailtrap.io>', started_at: '2026-03-28T09:15:00Z', finished_at: '2026-03-28T09:15:01Z', duration_ms: 910 },
  { id: 'att_27', notification_id: 'ntf_27', tenant_id: 'default', channel: 'sms', attempt: 1, status: 'delivered', provider_message_id: 'sms_msg_027', started_at: '2026-03-24T18:02:00Z', finished_at: '2026-03-24T18:02:01Z', duration_ms: 280 },
  { id: 'att_28', notification_id: 'ntf_28', tenant_id: 'default', channel: 'webhook', attempt: 1, status: 'delivered', provider_message_id: 'wh_att_028', started_at: '2026-03-20T16:30:00Z', finished_at: '2026-03-20T16:30:01Z', duration_ms: 320 },
  { id: 'att_29', notification_id: 'ntf_29', tenant_id: 'default', channel: 'in_app', attempt: 1, status: 'delivered', provider_message_id: 'inapp_msg_029', started_at: '2026-03-15T10:04:00Z', finished_at: '2026-03-15T10:04:00Z', duration_ms: 30 },
  { id: 'att_30', notification_id: 'ntf_30', tenant_id: 'default', channel: 'discord', attempt: 1, status: 'delivered', provider_message_id: 'disc_030', started_at: '2026-03-10T14:20:00Z', finished_at: '2026-03-10T14:20:01Z', duration_ms: 290 },

  // --- FEBRUARY ---
  { id: 'att_31', notification_id: 'ntf_31', tenant_id: 'default', channel: 'email', attempt: 1, status: 'delivered', provider_message_id: '<msg-31@smtp.mailtrap.io>', started_at: '2026-02-28T09:00:00Z', finished_at: '2026-02-28T09:00:01Z', duration_ms: 710 },
  { id: 'att_32', notification_id: 'ntf_32', tenant_id: 'default', channel: 'sms', attempt: 1, status: 'delivered', provider_message_id: 'sms_msg_032', started_at: '2026-02-24T17:15:00Z', finished_at: '2026-02-24T17:15:01Z', duration_ms: 290 },
  { id: 'att_33', notification_id: 'ntf_33', tenant_id: 'default', channel: 'webhook', attempt: 1, status: 'delivered', provider_message_id: 'wh_att_033', started_at: '2026-02-20T12:00:00Z', finished_at: '2026-02-20T12:00:01Z', duration_ms: 390 },
  { id: 'att_34', notification_id: 'ntf_34', tenant_id: 'default', channel: 'in_app', attempt: 1, status: 'delivered', provider_message_id: 'inapp_msg_034', started_at: '2026-02-15T09:30:00Z', finished_at: '2026-02-15T09:30:00Z', duration_ms: 35 },

  // --- JANUARY ---
  { id: 'att_35', notification_id: 'ntf_35', tenant_id: 'default', channel: 'email', attempt: 1, status: 'delivered', provider_message_id: '<msg-35@smtp.mailtrap.io>', started_at: '2026-01-28T10:12:00Z', finished_at: '2026-01-28T10:12:01Z', duration_ms: 690 },
  { id: 'att_36', notification_id: 'ntf_36', tenant_id: 'default', channel: 'sms', attempt: 1, status: 'delivered', provider_message_id: 'sms_msg_036', started_at: '2026-01-24T11:05:00Z', finished_at: '2026-01-24T11:05:01Z', duration_ms: 310 },
  { id: 'att_37', notification_id: 'ntf_37', tenant_id: 'default', channel: 'webhook', attempt: 1, status: 'delivered', provider_message_id: 'wh_att_037', started_at: '2026-01-20T14:40:00Z', finished_at: '2026-01-20T14:40:01Z', duration_ms: 340 },
  { id: 'att_38', notification_id: 'ntf_38', tenant_id: 'default', channel: 'in_app', attempt: 1, status: 'delivered', provider_message_id: 'inapp_msg_038', started_at: '2026-01-15T09:00:00Z', finished_at: '2026-01-15T09:00:00Z', duration_ms: 40 }
]

const DEFAULT_INBOX: InAppMessage[] = [
  {
    id: 'inapp_msg_1',
    notification_id: 'ntf_5',
    tenant_id: 'default',
    user_id: 'user_998',
    title: 'System Maintenance Complete',
    body: 'We have finished upgrading the database cluster to PostgreSQL 16. Services are fully operational and query latencies have improved by 15%.',
    created_at: '2026-06-15T09:00:00Z',
  },
  {
    id: 'inapp_msg_2',
    notification_id: 'ntf_7',
    tenant_id: 'default',
    user_id: 'user_998',
    title: 'Welcome to your In-App notifications!',
    body: 'This inbox simulator will display any notification requests targeted to the "in_app" channel for this user.',
    read_at: '2026-06-14T23:00:00Z',
    created_at: '2026-06-14T08:00:00Z',
  },
  {
    id: 'inapp_msg_3',
    notification_id: 'ntf_15',
    tenant_id: 'default',
    user_id: 'user_998',
    title: 'New Comment on Pull Request',
    body: 'kyle_architect commented on PR #412: "Looks solid. We should double check if the SQLite database lock handles high concurrent connection writes under stress."',
    created_at: '2026-05-18T16:22:00Z',
  },
  {
    id: 'inapp_msg_4',
    notification_id: 'ntf_23',
    tenant_id: 'default',
    user_id: 'user_998',
    title: 'New Task Assigned',
    body: 'You have been assigned to task: "Integrate Discord Webhook templates inside templates designer". Due Date: April 25.',
    read_at: '2026-04-19T08:00:00Z',
    created_at: '2026-04-18T09:12:00Z',
  },
  {
    id: 'inapp_msg_5',
    notification_id: 'ntf_29',
    tenant_id: 'default',
    user_id: 'user_998',
    title: 'Team Member Joined Workspace',
    body: 'User "kyle_architect" (kyle@company.com) has accepted your invitation and joined workspace: "default".',
    created_at: '2026-03-15T10:04:00Z',
  },
  {
    id: 'inapp_msg_6',
    notification_id: 'ntf_34',
    tenant_id: 'default',
    user_id: 'user_998',
    title: 'System Upgrade Notice',
    body: 'Aetheris core engine is upgrading to v1.5.0-rc2 tonight at 02:00 UTC. Minor send latency fluctuations (50ms) are expected during worker container re-scheduling.',
    read_at: '2026-02-15T10:00:00Z',
    created_at: '2026-02-15T09:30:00Z',
  },
  {
    id: 'inapp_msg_7',
    notification_id: 'ntf_38',
    tenant_id: 'default',
    user_id: 'user_998',
    title: 'Storage Quota Warning',
    body: 'Your workspace log retention database has reached 80% capacity (7.9 GB / 10 GB). Older logs will be auto-archived after 30 days.',
    created_at: '2026-01-15T09:00:00Z',
  }
]

// Storage Accessors
function getStorageItem<T>(key: string, defaultValue: T): T {
  const data = localStorage.getItem(key)
  if (!data) return defaultValue
  try {
    return JSON.parse(data) as T
  } catch {
    return defaultValue
  }
}

function setStorageItem<T>(key: string, value: T): void {
  localStorage.setItem(key, JSON.stringify(value))
}

// Global initialization
export function initMockDb() {
  const versionKey = 'aetheris.mock.version'
  const currentVersion = '3' // Incremented version to force seed update
  const savedVersion = localStorage.getItem(versionKey)

  if (savedVersion !== currentVersion) {
    localStorage.removeItem(KEYS.configs)
    localStorage.removeItem(KEYS.templates)
    localStorage.removeItem(KEYS.notifications)
    localStorage.removeItem(KEYS.attempts)
    localStorage.removeItem(KEYS.inbox)
    localStorage.setItem(versionKey, currentVersion)
  }

  if (!localStorage.getItem(KEYS.configs)) {
    setStorageItem(KEYS.configs, DEFAULT_CONFIGS)
  }
  if (!localStorage.getItem(KEYS.templates)) {
    setStorageItem(KEYS.templates, DEFAULT_TEMPLATES)
  }
  if (!localStorage.getItem(KEYS.notifications)) {
    setStorageItem(KEYS.notifications, DEFAULT_NOTIFICATIONS)
  }
  if (!localStorage.getItem(KEYS.attempts)) {
    setStorageItem(KEYS.attempts, DEFAULT_ATTEMPTS)
  }
  if (!localStorage.getItem(KEYS.inbox)) {
    setStorageItem(KEYS.inbox, DEFAULT_INBOX)
  }
}

// Engine Methods
export const mockEngine = {
  listNotifications(params: {
    recipient?: string
    channel?: Channel | ''
    status?: NotificationStatus | ''
    limit?: number
  }): NotificationRecord[] {
    initMockDb()
    let list = getStorageItem<NotificationRecord[]>(KEYS.notifications, [])

    // Sort descending by created_at
    list.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())

    if (params.recipient) {
      const query = params.recipient.toLowerCase()
      list = list.filter((n) => n.recipient.toLowerCase().includes(query))
    }
    if (params.channel) {
      list = list.filter((n) => n.channel === params.channel)
    }
    if (params.status) {
      list = list.filter((n) => n.status === params.status)
    }
    if (params.limit) {
      list = list.slice(0, params.limit)
    }
    return list
  },

  getNotification(id: string): NotificationRecord {
    initMockDb()
    const list = getStorageItem<NotificationRecord[]>(KEYS.notifications, [])
    const found = list.find((n) => n.id === id)
    if (!found) throw new Error('Notification not found')
    return found
  },

  listAttempts(notificationId: string): DeliveryAttempt[] {
    initMockDb()
    const list = getStorageItem<DeliveryAttempt[]>(KEYS.attempts, [])
    return list.filter((att) => att.notification_id === notificationId)
  },

  createNotification(payload: CreateNotificationPayload): NotificationRecord {
    initMockDb()
    const notifications = getStorageItem<NotificationRecord[]>(KEYS.notifications, [])
    const attempts = getStorageItem<DeliveryAttempt[]>(KEYS.attempts, [])
    const inbox = getStorageItem<InAppMessage[]>(KEYS.inbox, [])
    const configs = getStorageItem<ChannelConfig[]>(KEYS.configs, [])

    const now = new Date().toISOString()
    const id = genId('ntf')

    // Find channel config to check if disabled
    const cfg = configs.find((c) => c.channel === payload.channel)
    const isChannelEnabled = cfg ? cfg.enabled : true

    // Set status
    const status: NotificationStatus = isChannelEnabled
      ? (Math.random() > 0.05 ? 'delivered' : 'failed')
      : 'failed'
    const errorMsg = isChannelEnabled
      ? (status === 'failed' ? 'Simulated delivery timeout' : undefined)
      : `Channel ${payload.channel} is disabled in settings`

    // Extract title & body, supporting variable injection from metadata
    let title = payload.title || ''
    let body = payload.body || ''

    if (payload.template_key) {
      const templates = getStorageItem<NotificationTemplate[]>(KEYS.templates, [])
      const foundTemplate = templates.find((t) => t.key === payload.template_key)
      if (foundTemplate) {
        title = foundTemplate.title_template
        body = foundTemplate.body_template

        // Replace metadata placeholder variables e.g. {{.recipient}} or {{.code}}
        const metadataMap: Record<string, string> = {
          recipient: payload.recipient,
          ...payload.metadata,
        }

        Object.entries(metadataMap).forEach(([k, v]) => {
          const regex = new RegExp(`\\{\\{\\s*\\.${k}\\s*\\}\\}`, 'g')
          title = title.replace(regex, v)
          body = body.replace(regex, v)
        })
      }
    }

    const newNotification: NotificationRecord = {
      id,
      tenant_id: payload.tenant_id || 'default',
      recipient: payload.recipient,
      channel: payload.channel,
      template_key: payload.template_key || undefined,
      title,
      body,
      group_key: payload.group_key || undefined,
      status,
      idempotency_key: payload.idempotency_key || undefined,
      aggregate_count: 1,
      metadata: payload.metadata,
      provider_message_id: status === 'delivered' ? genId('prov') : undefined,
      last_error: errorMsg,
      delivered_at: status === 'delivered' ? now : undefined,
      created_at: now,
      updated_at: now,
    }

    // Add attempts
    if (status === 'delivered') {
      attempts.push({
        id: genId('att'),
        notification_id: id,
        tenant_id: payload.tenant_id || 'default',
        channel: payload.channel,
        attempt: 1,
        status: 'delivered',
        provider_message_id: newNotification.provider_message_id,
        started_at: now,
        finished_at: now,
        duration_ms: Math.floor(Math.random() * 300) + 50,
      })
    } else {
      attempts.push({
        id: genId('att'),
        notification_id: id,
        tenant_id: payload.tenant_id || 'default',
        channel: payload.channel,
        attempt: 1,
        status: 'failed',
        last_error: errorMsg,
        started_at: now,
        finished_at: now,
        duration_ms: Math.floor(Math.random() * 2000) + 1000,
      })
    }

    // Add to inbox if channel is in_app
    if (payload.channel === 'in_app' && status === 'delivered') {
      inbox.push({
        id: genId('inapp_msg'),
        notification_id: id,
        tenant_id: payload.tenant_id || 'default',
        user_id: payload.recipient,
        title,
        body,
        metadata: payload.metadata,
        created_at: now,
      })
    }

    notifications.push(newNotification)
    setStorageItem(KEYS.notifications, notifications)
    setStorageItem(KEYS.attempts, attempts)
    setStorageItem(KEYS.inbox, inbox)

    return newNotification
  },

  listInApp(params: { user_id?: string; unread?: boolean; limit?: number }): InAppMessage[] {
    initMockDb()
    let list = getStorageItem<InAppMessage[]>(KEYS.inbox, [])

    // Sort descending
    list.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())

    if (params.user_id) {
      list = list.filter((m) => m.user_id === params.user_id)
    }
    if (params.unread) {
      list = list.filter((m) => !m.read_at)
    }
    if (params.limit) {
      list = list.slice(0, params.limit)
    }
    return list
  },

  markInAppRead(id: string, userId: string): void {
    initMockDb()
    const inbox = getStorageItem<InAppMessage[]>(KEYS.inbox, [])
    const found = inbox.find((m) => m.id === id && m.user_id === userId)
    if (found) {
      found.read_at = new Date().toISOString()
      setStorageItem(KEYS.inbox, inbox)
    }
  },

  listTemplates(params: { channel?: Channel | ''; key?: string; limit?: number }): NotificationTemplate[] {
    initMockDb()
    let list = getStorageItem<NotificationTemplate[]>(KEYS.templates, [])

    if (params.channel) {
      list = list.filter((t) => t.channel === params.channel)
    }
    if (params.key) {
      const q = params.key.toLowerCase()
      list = list.filter((t) => t.key.toLowerCase().includes(q))
    }
    if (params.limit) {
      list = list.slice(0, params.limit)
    }
    return list
  },

  saveTemplate(template: NotificationTemplate): NotificationTemplate {
    initMockDb()
    const list = getStorageItem<NotificationTemplate[]>(KEYS.templates, [])
    const now = new Date().toISOString()

    if (template.id) {
      const index = list.findIndex((t) => t.id === template.id)
      if (index !== -1) {
        const existing = list[index]
        const updated: NotificationTemplate = {
          ...existing,
          ...template,
          updated_at: now,
        }
        list[index] = updated
        setStorageItem(KEYS.templates, list)
        return updated
      }
    }

    const newTpl: NotificationTemplate = {
      ...template,
      id: genId('tpl'),
      created_at: now,
      updated_at: now,
    }
    list.push(newTpl)
    setStorageItem(KEYS.templates, list)
    return newTpl
  },

  deleteTemplate(id: string): void {
    initMockDb()
    let list = getStorageItem<NotificationTemplate[]>(KEYS.templates, [])
    list = list.filter((t) => t.id !== id)
    setStorageItem(KEYS.templates, list)
  },

  listChannelConfigs(): ChannelConfig[] {
    initMockDb()
    return getStorageItem<ChannelConfig[]>(KEYS.configs, [])
  },

  saveChannelConfig(config: ChannelConfig): ChannelConfig {
    initMockDb()
    const list = getStorageItem<ChannelConfig[]>(KEYS.configs, [])
    const index = list.findIndex((c) => c.channel === config.channel)
    const now = new Date().toISOString()

    const updated: ChannelConfig = {
      ...config,
      id: config.id || genId('cfg'),
      created_at: config.created_at || now,
      updated_at: now,
    }

    if (index !== -1) {
      list[index] = updated
    } else {
      list.push(updated)
    }

    setStorageItem(KEYS.configs, list)
    return updated
  },

  clearLogs(): void {
    setStorageItem(KEYS.notifications, [])
    setStorageItem(KEYS.attempts, [])
    setStorageItem(KEYS.inbox, [])
  }
}
