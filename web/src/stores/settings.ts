import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

const baseKey = 'aetheris.apiBaseUrl'
const apiKeyKey = 'aetheris.apiKey'
const tenantIdKey = 'aetheris.tenantId'

const getInitialBaseUrl = () => {
  const saved = localStorage.getItem(baseKey)
  if (!saved) return '/api'
  if (saved === 'http://localhost:8080' || saved === '/api/v1') {
    localStorage.setItem(baseKey, '/api')
    return '/api'
  }
  return saved
}

export const useSettingsStore = defineStore('settings', () => {
  const apiBaseUrl = ref(getInitialBaseUrl())
  const apiKey = ref(localStorage.getItem(apiKeyKey) || '')
  const tenantId = ref(localStorage.getItem(tenantIdKey) || 'default')

  const connectionStatus = ref<'connected' | 'disconnected' | 'checking' | 'unconfigured'>('checking')
  const connectionError = ref('')
  
  // Reactive toast event queue/channel
  const toastEvent = ref<{ severity: string; summary: string; detail: string; life?: number } | null>(null)

  const isAuthenticated = computed(() => apiKey.value.trim().length > 0)

  function triggerToast(severity: string, summary: string, detail: string, life = 5000) {
    toastEvent.value = { severity, summary, detail, life }
  }

  function save(nextBaseUrl: string, nextApiKey: string, nextTenantId: string) {
    apiBaseUrl.value = nextBaseUrl.trim().replace(/\/$/, '') || '/api'
    apiKey.value = nextApiKey.trim()
    tenantId.value = nextTenantId.trim() || 'default'
    localStorage.setItem(baseKey, apiBaseUrl.value)
    localStorage.setItem(apiKeyKey, apiKey.value)
    localStorage.setItem(tenantIdKey, tenantId.value)
    checkConnection()
  }

  async function checkConnection() {
    connectionStatus.value = 'checking'
    try {
      // 1. Check if the server is reachable and alive via public /healthz
      const healthUrl = `${apiBaseUrl.value}/healthz`
      const healthRes = await fetch(healthUrl)
      if (!healthRes.ok) {
        throw new Error(`Server health check failed: HTTP ${healthRes.status}`)
      }

      // 2. If API Key is configured, verify it on a protected endpoint
      if (apiKey.value.trim()) {
        const { api } = await import('@/lib/api')
        try {
          await api.listNotifications({ limit: 1 })
        } catch (err: unknown) {
          const errMsg = err instanceof Error ? err.message : String(err)
          if (errMsg.includes('401') || errMsg.toLowerCase().includes('unauthorized')) {
            connectionStatus.value = 'disconnected'
            connectionError.value = 'Unauthorized: Invalid API secret key'
            return
          }
          // Other validation errors (e.g. 400 Bad Request) mean we are reachable and authenticated!
        }
      }

      connectionStatus.value = 'connected'
      connectionError.value = ''
    } catch (err) {
      connectionStatus.value = 'disconnected'
      connectionError.value = err instanceof Error ? err.message : String(err)
    }
  }

  return {
    apiBaseUrl,
    apiKey,
    tenantId,
    connectionStatus,
    connectionError,
    toastEvent,
    isAuthenticated,
    save,
    checkConnection,
    triggerToast,
  }
})
