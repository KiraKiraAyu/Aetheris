<script setup lang="ts">
import Tag from 'primevue/tag'
import { useI18nStore } from '@/stores/i18n'

const props = defineProps<{
  value?: string
}>()

const i18n = useI18nStore()

function severity(value?: string) {
  switch (value) {
    case 'delivered':
      return 'success'
    case 'failed':
      return 'danger'
    case 'running':
      return 'info'
    case 'queued':
      return 'warning'
    default:
      return 'secondary'
  }
}
</script>

<template>
  <Tag :value="props.value ? i18n.t(`status.${props.value}` as any) : '-'" :severity="severity(props.value)" />
</template>
