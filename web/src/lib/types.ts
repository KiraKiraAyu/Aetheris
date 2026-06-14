export type Channel =
  | 'email'
  | 'sms'
  | 'webhook'
  | 'in_app'
  | 'telegram'
  | 'slack'
  | 'discord'
  | 'feishu'
  | 'dingtalk'
  | 'wecom'

export type NotificationStatus = 'queued' | 'delivered' | 'failed'
export type AttemptStatus = 'running' | 'delivered' | 'failed'

export interface NotificationRecord {
  id: string
  tenant_id: string
  recipient: string
  channel: Channel
  template_key?: string
  title: string
  body: string
  group_key?: string
  status: NotificationStatus
  idempotency_key?: string
  aggregate_count: number
  metadata?: Record<string, string>
  provider_message_id?: string
  last_error?: string
  delivered_at?: string
  created_at: string
  updated_at: string
}

export interface DeliveryAttempt {
  id: string
  notification_id: string
  tenant_id: string
  channel: Channel
  attempt: number
  status: AttemptStatus
  provider_message_id?: string
  last_error?: string
  started_at: string
  finished_at?: string
  duration_ms: number
}

export interface InAppMessage {
  id: string
  notification_id: string
  tenant_id: string
  user_id: string
  title: string
  body: string
  metadata?: Record<string, string>
  read_at?: string
  created_at: string
}

export interface NotificationTemplate {
  id?: string
  tenant_id?: string
  key: string
  channel: Channel
  title_template: string
  body_template: string
  created_at?: string
  updated_at?: string
}

export interface CreateNotificationPayload {
  tenant_id?: string
  recipient: string
  channel: Channel
  template_key?: string
  title?: string
  body?: string
  group_key?: string
  idempotency_key?: string
  metadata?: Record<string, string>
}

export const channels: Channel[] = [
  'email',
  'sms',
  'webhook',
  'in_app',
  'telegram',
  'slack',
  'discord',
  'feishu',
  'dingtalk',
  'wecom',
]

export const statuses: NotificationStatus[] = ['queued', 'delivered', 'failed']

export interface ChannelOption {
  label: string
  value: Channel
  icon: string
}

export interface FilterChannelOption {
  label: string
  value: Channel | ''
  icon?: string
}

export interface ChannelInfo {
  label: string
  icon: string
  tone: string
}

export type ChannelMap = Record<string, ChannelInfo>

export interface StatusOption {
  label: string
  value: string
}

export interface DispatchForm {
  recipient: string
  channel: Channel
  template_key: string
  title: string
  body: string
  group_key: string
  idempotency_key: string
  metadata: string
}

export interface ChannelConfig {
  id?: string
  tenant_id?: string
  channel: Channel
  enabled: boolean
  config: string
  created_at?: string
  updated_at?: string
}



