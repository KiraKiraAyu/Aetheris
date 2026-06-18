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

import { mockEngine } from './mockEngine'

export const api = {
  async listNotifications(params: {
    recipient?: string
    channel?: Channel | ''
    status?: NotificationStatus | ''
    limit?: number
  }) {
    if (import.meta.env.MODE === 'demo') {
      return mockEngine.listNotifications(params)
    }
    return request<NotificationRecord[]>(`/notifications${query(params)}`)
  },
  async getNotification(id: string) {
    if (import.meta.env.MODE === 'demo') {
      return mockEngine.getNotification(id)
    }
    return request<NotificationRecord>(`/notifications/${encodeURIComponent(id)}`)
  },
  async listAttempts(notificationId: string) {
    if (import.meta.env.MODE === 'demo') {
      return mockEngine.listAttempts(notificationId)
    }
    return request<DeliveryAttempt[]>(
      `/notifications/${encodeURIComponent(notificationId)}/attempts`,
    )
  },
  async createNotification(payload: CreateNotificationPayload) {
    if (import.meta.env.MODE === 'demo') {
      return mockEngine.createNotification(payload)
    }
    const settings = useSettingsStore()
    return request<NotificationRecord>('/send', {
      method: 'POST',
      body: JSON.stringify({
        tenant_id: settings.tenantId || undefined,
        ...payload,
      }),
    })
  },
  async listInApp(params: { user_id?: string; unread?: boolean; limit?: number }) {
    if (import.meta.env.MODE === 'demo') {
      return mockEngine.listInApp(params)
    }
    return request<InAppMessage[]>(`/in-app/messages${query(params)}`)
  },
  async markInAppRead(id: string, userId: string) {
    if (import.meta.env.MODE === 'demo') {
      return mockEngine.markInAppRead(id, userId)
    }
    return request<void>(
      `/in-app/messages/${encodeURIComponent(id)}/read${query({ user_id: userId })}`,
      {
        method: 'POST',
      },
    )
  },
  async listTemplates(params: { channel?: Channel | ''; key?: string; limit?: number }) {
    if (import.meta.env.MODE === 'demo') {
      return mockEngine.listTemplates(params)
    }
    return request<NotificationTemplate[]>(`/templates${query(params)}`)
  },
  async saveTemplate(template: NotificationTemplate) {
    if (import.meta.env.MODE === 'demo') {
      return mockEngine.saveTemplate(template)
    }
    const settings = useSettingsStore()
    return request<NotificationTemplate>('/templates', {
      method: 'POST',
      body: JSON.stringify({
        tenant_id: settings.tenantId || undefined,
        ...template,
      }),
    })
  },
  async deleteTemplate(id: string) {
    if (import.meta.env.MODE === 'demo') {
      return mockEngine.deleteTemplate(id)
    }
    return request<void>(`/templates/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    })
  },
  async listChannelConfigs() {
    if (import.meta.env.MODE === 'demo') {
      return mockEngine.listChannelConfigs()
    }
    return request<ChannelConfig[]>('/channels')
  },
  async saveChannelConfig(config: ChannelConfig) {
    if (import.meta.env.MODE === 'demo') {
      return mockEngine.saveChannelConfig(config)
    }
    const settings = useSettingsStore()
    return request<ChannelConfig>('/channels', {
      method: 'POST',
      body: JSON.stringify({
        tenant_id: settings.tenantId || undefined,
        ...config,
      }),
    })
  },
  async clearLogs() {
    if (import.meta.env.MODE === 'demo') {
      mockEngine.clearLogs()
    }
  },
}

export function formatDate(value?: string) {
  if (!value) return '-'
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}

export function getAssetUrl(path: string): string {
  const cleanPath = path.startsWith('/') ? path.substring(1) : path
  if (import.meta.env.MODE === 'demo') {
    const pathname = window.location.pathname
    let base = pathname
    if (pathname.endsWith('.html')) {
      base = pathname.substring(0, pathname.lastIndexOf('/') + 1)
    } else if (!pathname.endsWith('/')) {
      base = pathname + '/'
    }
    return `${base}${cleanPath}`
  }
  const baseUrl = import.meta.env.BASE_URL
  const normalizedBase = baseUrl.endsWith('/') ? baseUrl : `${baseUrl}/`
  return `${normalizedBase}${cleanPath}`
}

