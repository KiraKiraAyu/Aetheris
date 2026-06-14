<script setup lang="ts">
import Timeline from "primevue/timeline";
import { formatDate } from "@/lib/api";
import type { DeliveryAttempt } from "@/lib/types";
import { useI18nStore, type TranslationKey } from "@/stores/i18n";

defineProps<{
  attempts: DeliveryAttempt[];
}>();

const i18n = useI18nStore();
</script>

<template>
  <div>
    <h3 class="mb-3 text-sm font-bold text-surface-900 dark:text-white">
      {{ i18n.t("notifications.details_attempts") }}
    </h3>

    <Timeline :value="attempts" class="customized-timeline">
      <template #marker="{ item }">
        <span
          class="flex w-6 h-6 items-center justify-center rounded-full text-[10px] font-bold border"
          :class="
            item.status === 'delivered'
              ? 'bg-reisa-lilac-50 border-reisa-lilac-200 text-reisa-lilac-600 dark:bg-reisa-lilac-950/20 dark:border-reisa-lilac-800/40 dark:text-reisa-lilac-400'
              : item.status === 'failed'
                ? 'bg-rose-50 border-rose-200 text-rose-600 dark:bg-rose-950/20 dark:border-rose-800/40 dark:text-rose-400'
                : 'bg-blue-50 border-blue-200 text-blue-600 dark:bg-blue-950/20 dark:border-blue-800/40 dark:text-blue-400'
          "
        >
          {{ item.attempt }}
        </span>
      </template>

      <template #content="{ item }">
        <div
          class="p-3.5 bg-surface-50 dark:bg-surface-950/40 border border-surface-100 dark:border-surface-800 rounded-xl space-y-1.5 text-xs leading-relaxed"
        >
          <div class="flex flex-wrap justify-between items-center gap-2">
            <div class="flex items-center gap-2">
              <span class="font-bold text-surface-850 dark:text-surface-200"
                >{{ i18n.t("notifications.details_attempt_num", { num: item.attempt }) }}</span
              >
              <span
                class="px-1.5 py-0.5 rounded text-[9px] font-bold uppercase tracking-wider"
                :class="
                  item.status === 'delivered'
                    ? 'bg-reisa-lilac-100 text-reisa-lilac-700 dark:bg-reisa-lilac-900/30 dark:text-reisa-lilac-400'
                    : item.status === 'failed'
                      ? 'bg-rose-100 text-rose-700 dark:bg-rose-900/30 dark:text-rose-400'
                      : 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
                "
              >
                {{ i18n.t(`status.${item.status}` as TranslationKey) }}
              </span>
            </div>
            <div class="text-[10px] text-surface-400 dark:text-surface-500 font-mono">
              {{ formatDate(item.created_at) }}
              <span v-if="item.finished_at"> &rarr; {{ formatDate(item.finished_at) }} </span>
            </div>
          </div>

          <div
            v-if="item.provider_log"
            class="mt-2 bg-white dark:bg-surface-900 border border-surface-200 dark:border-surface-800 p-2.5 rounded-lg overflow-x-auto text-surface-600 dark:text-surface-400 font-mono text-[10px]"
          >
            {{ item.provider_log }}
          </div>
        </div>
      </template>
    </Timeline>
  </div>
</template>

<style scoped>
.customized-timeline :deep(.p-timeline-event-opposite) {
  display: none;
}
.customized-timeline :deep(.p-timeline-event-content) {
  padding-bottom: 1.5rem;
}
</style>
