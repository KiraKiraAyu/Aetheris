<script setup lang="ts">
import type { NotificationRecord, ChannelMap } from "@/lib/types";
import { formatDate } from "@/lib/api";
import { useI18nStore } from "@/stores/i18n";

defineProps<{
  selected: NotificationRecord;
  channelMap: ChannelMap;
}>();

const i18n = useI18nStore();
</script>

<template>
  <div
    class="w-full max-w-135 bg-white dark:bg-surface-900 border border-surface-100 dark:border-surface-800 rounded-xl p-4 flex gap-3 font-sans"
  >
    <!-- Branded Avatar -->
    <div
      :class="[
        'grid h-9 w-9 place-items-center rounded-lg border text-white font-bold',
        selected.channel === 'telegram'
          ? 'bg-blue-500 border-blue-400'
          : selected.channel === 'slack'
            ? 'bg-purple-600 border-purple-500'
            : selected.channel === 'discord'
              ? 'bg-indigo-600 border-indigo-500'
              : 'bg-reisa-lilac-600 border-reisa-lilac-500',
      ]"
    >
      <span :class="channelMap[selected.channel]?.icon || 'pi pi-send'" class="text-sm"></span>
    </div>
    <div class="flex-1 min-w-0">
      <div class="flex items-baseline gap-2">
        <span class="text-xs font-bold text-surface-900 dark:text-white capitalize"
          >{{ i18n.t('notifications.mockup_bot', { channel: i18n.t(`channels.${selected.channel}` as any) }) }}</span
        >
        <span
          class="bg-indigo-50 text-indigo-600 dark:bg-indigo-950/20 dark:text-indigo-400 px-1 rounded text-[8px] font-bold uppercase tracking-wide"
          >APP</span
        >
        <span class="text-[9px] text-surface-400 dark:text-surface-500 font-mono">{{
          formatDate(selected.created_at)
        }}</span>
      </div>
      <!-- Box Content -->
      <div
        class="mt-2 p-3 bg-surface-50 dark:bg-surface-950 border border-surface-100 dark:border-surface-850 rounded-xl space-y-1.5 text-xs text-surface-700 dark:text-surface-300"
      >
        <div v-if="selected.title" class="font-bold text-surface-900 dark:text-white">
          {{ selected.title }}
        </div>
        <div class="whitespace-pre-wrap leading-normal">{{ selected.body }}</div>
      </div>
    </div>
  </div>
</template>
