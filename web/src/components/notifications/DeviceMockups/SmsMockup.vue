<script setup lang="ts">
import type { NotificationRecord } from '@/lib/types'

defineProps<{
  selected: NotificationRecord
}>()
</script>

<template>
  <div
    class="w-[280px] bg-surface-200 dark:bg-surface-850 border-4 border-surface-400 dark:border-surface-800 rounded-[28px] overflow-hidden flex flex-col font-sans"
  >
    <div
      class="p-2 border-b border-surface-300 dark:border-surface-700 bg-surface-100/90 dark:bg-surface-800/90 text-[10px] text-center font-bold text-surface-700 dark:text-surface-300"
    >
      {{ selected.recipient }}
    </div>
    <div
      class="p-3 flex-1 flex flex-col justify-end min-h-[160px] bg-surface-100 dark:bg-surface-950"
    >
      <div
        class="p-3 bg-indigo-500 text-white rounded-2xl rounded-br-none text-[11px] leading-normal max-w-[90%] self-end"
      >
        <div v-if="selected.title" class="font-bold mb-0.5">{{ selected.title }}</div>
        <div class="whitespace-pre-wrap">{{ selected.body }}</div>
      </div>
    </div>
  </div>
</template>
