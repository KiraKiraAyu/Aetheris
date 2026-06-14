<script setup lang="ts">
import { RouterLink } from "vue-router";
import DataTable from "primevue/datatable";
import Column from "primevue/column";
import StatusBadge from "@/components/common/StatusBadge.vue";
import { formatDate } from "@/lib/api";
import type { NotificationRecord } from "@/lib/types";
import { useI18nStore } from "@/stores/i18n";

defineProps<{
  notifications: NotificationRecord[];
  loading: boolean;
}>();

const i18n = useI18nStore();
</script>

<template>
  <div
    class="p-6 bg-surface-0 dark:bg-surface-900 rounded-2xl border border-surface-200 dark:border-surface-800"
  >
    <div class="mb-4 flex items-center justify-between">
      <div>
        <h2 class="text-lg font-bold text-surface-900 dark:text-white">{{ i18n.t("overview.recent_activity") }}</h2>
        <p class="text-xs text-surface-500 dark:text-surface-400">
          {{ i18n.t("overview.recent_activity_desc") }}
        </p>
      </div>
      <RouterLink
        class="text-xs font-bold text-indigo-600 dark:text-indigo-400 hover:underline flex items-center gap-1"
        to="/notifications"
      >
        <span>{{ i18n.t("overview.view_history") }}</span>
        <span class="pi pi-arrow-right text-[10px]"></span>
      </RouterLink>
    </div>

    <DataTable
      :value="notifications"
      :loading="loading"
      data-key="id"
      class="p-datatable-sm"
      responsive-layout="scroll"
    >
      <Column :header="i18n.t('overview.table_status')" class="py-2.5">
        <template #body="{ data }">
          <StatusBadge :value="data.status" />
        </template>
      </Column>
      <Column
        field="title"
        :header="i18n.t('overview.table_title')"
        class="py-2.5 font-medium text-surface-900 dark:text-surface-100 max-w-50 truncate"
      />
      <Column field="channel" :header="i18n.t('overview.table_channel')" class="py-2.5 uppercase text-xs tracking-wider" />
      <Column
        field="recipient"
        :header="i18n.t('overview.table_recipient')"
        class="py-2.5 text-surface-500 dark:text-surface-400 truncate max-w-37.5"
      />
      <Column :header="i18n.t('overview.table_time')" class="py-2.5 text-surface-400 dark:text-surface-500">
        <template #body="{ data }">
          <span class="text-xs font-mono">{{ formatDate(data.created_at) }}</span>
        </template>
      </Column>
      <template #empty>
        <div class="py-8 text-center text-surface-400 dark:text-surface-500">
          {{ i18n.t("overview.table_empty") }}
        </div>
      </template>
    </DataTable>
  </div>
</template>
