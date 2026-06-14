import { describe, it, expect } from 'vitest'

import { mount } from '@vue/test-utils'
import App from '../App.vue'
import { createTestingRouter } from './helpers'
import { createPinia } from 'pinia'
import PrimeVue from 'primevue/config'
import Aura from '@primeuix/themes/aura'
import ToastService from 'primevue/toastservice'

describe('App', () => {
  it('renders the management shell', async () => {
    const router = createTestingRouter()
    router.push('/')
    await router.isReady()
    const wrapper = mount(App, {
      global: {
        plugins: [
          createPinia(),
          router,
          ToastService,
          [
            PrimeVue,
            {
              theme: {
                preset: Aura,
                options: { darkModeSelector: false },
              },
            },
          ],
        ],
      },
    })
    expect(wrapper.text()).toContain('Overview')
    expect(wrapper.text()).toContain('Notifications')
    expect(wrapper.text()).toContain('Templates')
  })
})
