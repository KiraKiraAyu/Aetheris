<script setup lang="ts">
import type { NotificationRecord } from "@/lib/types";
import { useI18nStore } from "@/stores/i18n";

defineProps<{
  selected: NotificationRecord;
}>();

const i18n = useI18nStore();
</script>

<template>
  <div
    class="w-full max-w-135 bg-white dark:bg-surface-900 border border-surface-200 dark:border-surface-800 rounded-xl overflow-hidden text-surface-700 dark:text-surface-350"
  >
    <div
      class="p-3 border-b border-surface-100 dark:border-surface-800/80 bg-surface-50/50 dark:bg-surface-950/20 text-xs font-mono space-y-1 shrink-0"
    >
      <div class="flex gap-2">
        <span class="text-surface-400 dark:text-surface-500 w-12 text-right">{{ i18n.t("notifications.mockup_email_to") }}</span>
        <span class="font-semibold text-surface-700 dark:text-surface-300">{{
          selected.recipient
        }}</span>
      </div>
      <div class="flex gap-2">
        <span class="text-surface-400 dark:text-surface-500 w-12 text-right">{{ i18n.t("notifications.mockup_email_from") }}</span>
        <span class="text-surface-650">aetheris-gateway@tenant.internal</span>
      </div>
      <div class="flex gap-2">
        <span class="text-surface-400 dark:text-surface-500 w-12 text-right">{{ i18n.t("notifications.mockup_email_subject") }}</span>
        <span class="font-bold text-surface-900 dark:text-white">{{
          selected.title || i18n.t("notifications.mockup_email_no_subject")
        }}</span>
      </div>
    </div>
    <div class="p-5 text-sm font-sans min-h-30 leading-relaxed whitespace-pre-wrap">
      {{ selected.body }}
    </div>
  </div>
</template>
