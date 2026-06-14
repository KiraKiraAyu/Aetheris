import './assets/main.css'
import 'primeicons/primeicons.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import PrimeVue from 'primevue/config'
import Aura from '@primeuix/themes/aura'
import { definePreset } from '@primeuix/themes'

import App from './App.vue'
import router from './router'
import ToastService from 'primevue/toastservice'

import { useI18nStore } from '@/stores/i18n'

const app = createApp(App)

const MyPreset = definePreset(Aura, {
  semantic: {
    primary: {
      50: 'oklch(0.96 0.018 301)',
      100: 'oklch(0.92 0.030 301)',
      200: 'oklch(0.88 0.042 301)',
      300: 'oklch(0.84 0.052 301)',
      400: 'oklch(0.76 0.060 301)',
      500: 'oklch(0.66 0.058 301)',
      600: 'oklch(0.56 0.052 301)',
      700: 'oklch(0.46 0.044 301)',
      800: 'oklch(0.36 0.036 301)',
      900: 'oklch(0.26 0.028 301)',
      950: 'oklch(0.19 0.100 301)',
    }
  }
})

app.use(createPinia())
app.use(router)
app.use(ToastService)
app.use(PrimeVue, {
  theme: {
    preset: MyPreset,
    options: {
      darkModeSelector: '.dark',
    },
  },
})

const i18n = useI18nStore()
i18n.detectLanguage()

app.mount('#app')
