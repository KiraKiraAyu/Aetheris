<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { api } from '@/lib/api'
import { channels, type Channel, type NotificationTemplate, type ChannelOption, type FilterChannelOption } from '@/lib/types'
import { useToast } from 'primevue/usetoast'
import TemplateList from '@/components/templates/TemplateList.vue'
import TemplateEditor from '@/components/templates/TemplateEditor.vue'
import { useI18nStore, type TranslationKey } from '@/stores/i18n'

const toast = useToast()
const i18n = useI18nStore()
const templates = ref<NotificationTemplate[]>([])
const loading = ref(false)
const saving = ref(false)
const error = ref('')

const filters = reactive({
  channel: '' as Channel | '',
  key: '',
})

const form = reactive<NotificationTemplate>({
  id: undefined,
  key: '',
  channel: 'in_app',
  title_template: '',
  body_template: '',
})

const mockVariablesJson = ref('{\n "name": "Jane Doe",\n "code": "4820",\n "amount": "$149.99"\n}')

const channelOptions = computed<ChannelOption[]>(() =>
  channels.map((value) => {
    let icon = 'pi pi-envelope'
    if (value === 'sms') icon = 'pi pi-phone'
    else if (value === 'webhook') icon = 'pi pi-globe'
    else if (value === 'in_app') icon = 'pi pi-inbox'
    else if (value === 'telegram') icon = 'pi pi-telegram'
    else if (value === 'slack') icon = 'pi pi-slack'
    else if (value === 'discord') icon = 'pi pi-discord'
    else if (value === 'feishu' || value === 'dingtalk') icon = 'pi pi-comments'
    else if (value === 'wecom') icon = 'pi pi-briefcase'

    return { label: i18n.t(`channels.${value}` as TranslationKey), value, icon }
  }),
)

const filterChannelOptions = computed<FilterChannelOption[]>(() => [
  { label: i18n.t("notifications.filter_channel"), value: '' as const },
  ...channelOptions.value,
])

async function load() {
  loading.value = true
  error.value = ''
  try {
    templates.value = await api.listTemplates({ ...filters, limit: 100 })
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err)
  } finally {
    loading.value = false
  }
}

function edit(template: NotificationTemplate) {
  form.id = template.id
  form.key = template.key
  form.channel = template.channel
  form.title_template = template.title_template
  form.body_template = template.body_template

  toast.add({
    severity: 'info',
    summary: i18n.t("templates.toast_edit_title"),
    detail: i18n.t("templates.toast_edit_desc", { key: template.key }),
    life: 2000,
  })
}

function reset() {
  form.id = undefined
  form.key = ''
  form.channel = 'in_app'
  form.title_template = ''
  form.body_template = ''
}

async function save() {
  saving.value = true
  error.value = ''
  try {
    await api.saveTemplate({ ...form })
    toast.add({
      severity: 'success',
      summary: i18n.t("templates.toast_save_title"),
      detail: i18n.t("templates.toast_save_desc", { key: form.key }),
      life: 3000,
    })
    reset()
    await load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err)
  } finally {
    saving.value = false
  }
}

async function remove(template: NotificationTemplate) {
  if (!template.id) return
  try {
    await api.deleteTemplate(template.id)
    toast.add({
      severity: 'warn',
      summary: i18n.t("templates.toast_delete_title"),
      detail: i18n.t("templates.toast_delete_desc"),
      life: 3000,
    })
    await load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err)
  }
}

function updateTemplateForm(nextForm: NotificationTemplate) {
  Object.assign(form, nextForm)
}

onMounted(load)
</script>

<template>
  <section class="grid gap-6 xl:grid-cols-[1fr_450px]">
    
    <div class="space-y-5">

      <TemplateList
        :templates="templates"
        :loading="loading"
        v-model:filters="filters"
        :filter-channel-options="filterChannelOptions"
        @load="load"
        @edit="edit"
        @remove="remove"
      />
    </div>

    <!-- Right Column: Edit / New Template Form -->
    <TemplateEditor
      :form="form"
      :channel-options="channelOptions"
      :saving="saving"
      v-model:mockVariablesJson="mockVariablesJson"
      @update:form="updateTemplateForm"
      @reset="reset"
      @save="save"
    />
  </section>
</template>
