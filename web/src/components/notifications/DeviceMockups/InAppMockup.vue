<script setup lang="ts">
import type { NotificationRecord } from "@/lib/types";
import { formatDate } from "@/lib/api";
import { useI18nStore } from "@/stores/i18n";

defineProps<{
  selected: NotificationRecord;
}>();

const i18n = useI18nStore();
</script>

<template>
  <div
    class="w-full max-w-125 p-4 bg-white dark:bg-surface-900 border border-surface-200 dark:border-surface-800 rounded-xl flex gap-3.5 font-sans"
  >
    <div
      class="h-9 w-9 rounded-full bg-amber-50 text-amber-500 dark:bg-amber-950/30 dark:text-amber-400 flex items-center justify-center border border-amber-100 dark:border-amber-800/30 shrink-0"
    >
      <span class="pi pi-inbox text-sm"></span>
    </div>
    <div class="flex-1 min-w-0">
      <div class="flex justify-between items-baseline">
        <h4 class="text-xs font-bold text-surface-900 dark:text-white leading-normal truncate">
          {{ selected.title || i18n.t("notifications.mockup_in_app_no_title") }}
        </h4>
        <span class="text-[9px] text-surface-400 dark:text-surface-500 whitespace-nowrap ml-2">{{
          formatDate(selected.created_at)
        }}</span>
      </div>
      <p
        class="mt-1.5 text-xs text-surface-600 dark:text-surface-300 whitespace-pre-wrap leading-normal bg-surface-50 dark:bg-surface-950 p-2.5 rounded-lg border border-surface-100 dark:border-surface-850"
      >
        {{ selected.body }}
      </p>
    </div>
  </div>
</template>
