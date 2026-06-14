<script setup lang="ts">
import { computed } from "vue";
import InputText from "primevue/inputtext";
import Select from "primevue/select";
import Button from "primevue/button";
import DataTable from "primevue/datatable";
import Column from "primevue/column";
import StatusBadge from "@/components/common/StatusBadge.vue";
import { formatDate } from "@/lib/api";
import { useI18nStore, type TranslationKey } from "@/stores/i18n";
import type {
  NotificationRecord,
  FilterChannelOption,
  StatusOption,
  ChannelMap,
} from "@/lib/types";

const i18n = useI18nStore();

const props = defineProps<{
  notifications: NotificationRecord[];
  loading: boolean;
  filters: { recipient: string; channel: string; status: string };
  filterChannelOptions: FilterChannelOption[];
  statusOptions: StatusOption[];
  channelMap: ChannelMap;
}>();

const emit = defineEmits<{
  (e: "update:filters", value: { recipient: string; channel: string; status: string }): void;
  (e: "load"): void;
  (e: "inspect", record: NotificationRecord): void;
}>();

// Use a local writable proxy for filters
const localFilters = computed({
  get: () => props.filters,
  set: (val) => emit("update:filters", val),
});
</script>

<template>
  <div class="space-y-5">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h1 class="mt-1 text-3xl font-bold text-surface-900 dark:text-white">{{ i18n.t("notifications.title") }}</h1>
        <p class="mt-1 text-sm text-surface-500 dark:text-surface-400">
          {{ i18n.t("notifications.desc") }}
        </p>
      </div>
      <Button
        icon="pi pi-refresh"
        :label="i18n.t('overview.refresh')"
        :loading="loading"
        @click="emit('load')"
        class="rounded-xl px-4 h-11"
      />
    </div>

    <div
      class="p-6 bg-surface-0 dark:bg-surface-900 rounded-2xl border border-surface-200 dark:border-surface-800"
    >
      <div class="mb-4 grid gap-3 sm:grid-cols-[2fr_1.5fr_1.5fr_auto]">
        <div class="relative w-full">
          <span
            class="pi pi-user absolute right-3 top-1/2 -translate-y-1/2 text-surface-400 dark:text-surface-500 text-sm"
          ></span>
          <InputText
            v-model="localFilters.recipient"
            :placeholder="i18n.t('notifications.search_recipient')"
            class="pr-9 h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl text-sm"
          />
        </div>

        <Select
          v-model="localFilters.channel"
          :options="filterChannelOptions"
          option-label="label"
          option-value="value"
          :placeholder="i18n.t('notifications.filter_channel')"
          class="h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl text-sm"
        >
          <template #option="{ option }">
            <div class="flex items-center gap-2 text-xs">
              <span v-if="option.icon" :class="option.icon" class="text-surface-400"></span>
              <span class="capitalize">{{ option.label }}</span>
            </div>
          </template>
        </Select>

        <Select
          v-model="localFilters.status"
          :options="statusOptions"
          option-label="label"
          option-value="value"
          :placeholder="i18n.t('notifications.filter_status')"
          show-clear
          class="h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl text-sm"
        />

        <Button
          icon="pi pi-filter"
          :label="i18n.t('notifications.btn_apply')"
          class="h-10 rounded-xl px-4 text-xs font-semibold"
          @click="emit('load')"
        />
      </div>

      <DataTable
        :value="notifications"
        :loading="loading"
        data-key="id"
        class="p-datatable-sm"
        responsive-layout="scroll"
      >
        <Column :header="i18n.t('overview.table_status')" class="py-3">
          <template #body="{ data }">
            <StatusBadge :value="data.status" />
          </template>
        </Column>
        <Column
          field="title"
          :header="i18n.t('overview.table_title')"
          class="py-3 font-semibold text-surface-850 dark:text-surface-200 truncate max-w-50"
        />
        <Column
          field="channel"
          :header="i18n.t('overview.table_channel')"
          class="py-3 uppercase text-[10px] tracking-wider font-bold"
        >
          <template #body="{ data }">
            <span class="inline-flex items-center gap-1.5">
              <span
                :class="channelMap[data.channel]?.icon"
                class="text-surface-400 dark:text-surface-500 text-xs"
              ></span>
              <span>{{ i18n.t(`channels.${data.channel}` as TranslationKey) }}</span>
            </span>
          </template>
        </Column>
        <Column
          field="recipient"
          :header="i18n.t('overview.table_recipient')"
          class="py-3 text-surface-500 dark:text-surface-400 truncate max-w-37.5"
        />
        <Column :header="i18n.t('overview.table_time')" class="py-3 text-surface-400 dark:text-surface-500">
          <template #body="{ data }">
            <span class="text-xs font-mono">{{ formatDate(data.created_at) }}</span>
          </template>
        </Column>
        <Column header="" class="py-3">
          <template #body="{ data }">
            <Button
              size="small"
              text
              icon="pi pi-external-link"
              :label="i18n.t('notifications.btn_inspect')"
              class="h-8 rounded-lg"
              @click="emit('inspect', data)"
            />
          </template>
        </Column>
        <template #empty>
          <div class="py-12 text-center text-surface-400 dark:text-surface-500">
            {{ i18n.t("notifications.no_records") }}
          </div>
        </template>
      </DataTable>
    </div>
  </div>
</template>
