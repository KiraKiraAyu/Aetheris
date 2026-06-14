import { useSettingsStore } from '@/stores/settings'
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

interface QueryValue {
  [key: string]: string | number | boolean | undefined
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const settings = useSettingsStore()
  const headers = new Headers(init.headers)
  if (!headers.has('Content-Type') && init.body) {
    headers.set('Content-Type', 'application/json')
  }
  if (settings.apiKey) {
    headers.set('Authorization', `Bearer ${settings.apiKey}`)
  }

  let url = `${settings.apiBaseUrl}${path}`
  if (settings.tenantId) {
    const delimiter = url.includes('?') ? '&' : '?'
    if (!url.includes('tenant_id=')) {
      url = `${url}${delimiter}tenant_id=${encodeURIComponent(settings.tenantId)}`
    }
  }

  try {
    const response = await fetch(url, {
      ...init,
      headers,
    })
    if (!response.ok) {
      let message = `HTTP ${response.status}`
      try {
        const payload = await response.json()
        if (payload?.error) {
          message = payload.error
        }
      } catch {
        // Keep status message.
      }
      throw new Error(message)
    }

    settings.connectionStatus = 'connected'
    settings.connectionError = ''

    if (response.status === 204) {
      return undefined as T
    }
    return (await response.json()) as T
  } catch (err) {
    settings.connectionStatus = 'disconnected'
    const message = err instanceof Error ? err.message : String(err)
    settings.connectionError = message
    settings.triggerToast('error', 'API Request Failed', message, 5000)
    throw err
  }
}

function query(params: QueryValue) {
  const search = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== '') {
      search.set(key, String(value))
    }
  })
  const text = search.toString()
  return text ? `?${text}` : ''
}

export const api = {
  listNotifications(params: {
    recipient?: string
    channel?: Channel | ''
    status?: NotificationStatus | ''
    limit?: number
  }) {
    return request<NotificationRecord[]>(`/notifications${query(params)}`)
  },
  getNotification(id: string) {
    return request<NotificationRecord>(`/notifications/${encodeURIComponent(id)}`)
  },
  listAttempts(notificationId: string) {
    return request<DeliveryAttempt[]>(
      `/notifications/${encodeURIComponent(notificationId)}/attempts`,
    )
  },
  createNotification(payload: CreateNotificationPayload) {
    const settings = useSettingsStore()
    return request<NotificationRecord>('/send', {
      method: 'POST',
      body: JSON.stringify({
        tenant_id: settings.tenantId || undefined,
        ...payload,
      }),
    })
  },
  listInApp(params: { user_id?: string; unread?: boolean; limit?: number }) {
    return request<InAppMessage[]>(`/in-app/messages${query(params)}`)
  },
  markInAppRead(id: string, userId: string) {
    return request<void>(
      `/in-app/messages/${encodeURIComponent(id)}/read${query({ user_id: userId })}`,
      {
        method: 'POST',
      },
    )
  },
  listTemplates(params: { channel?: Channel | ''; key?: string; limit?: number }) {
    return request<NotificationTemplate[]>(`/templates${query(params)}`)
  },
  saveTemplate(template: NotificationTemplate) {
    const settings = useSettingsStore()
    return request<NotificationTemplate>('/templates', {
      method: 'POST',
      body: JSON.stringify({
        tenant_id: settings.tenantId || undefined,
        ...template,
      }),
    })
  },
  deleteTemplate(id: string) {
    return request<void>(`/templates/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    })
  },
  listChannelConfigs() {
    return request<ChannelConfig[]>('/channels')
  },
  saveChannelConfig(config: ChannelConfig) {
    const settings = useSettingsStore()
    return request<ChannelConfig>('/channels', {
      method: 'POST',
      body: JSON.stringify({
        tenant_id: settings.tenantId || undefined,
        ...config,
      }),
    })
  },
}

export function formatDate(value?: string) {
  if (!value) return '-'
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}
