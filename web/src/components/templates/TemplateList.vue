<script setup lang="ts">
import { computed } from "vue";
import Button from "primevue/button";
import Column from "primevue/column";
import DataTable from "primevue/datatable";
import InputText from "primevue/inputtext";
import Select from "primevue/select";
import type { NotificationTemplate, FilterChannelOption } from "@/lib/types";
import { useI18nStore, type TranslationKey } from "@/stores/i18n";

const props = defineProps<{
  templates: NotificationTemplate[];
  loading: boolean;
  filters: { key: string; channel: string };
  filterChannelOptions: FilterChannelOption[];
}>();

const i18n = useI18nStore();

const emit = defineEmits<{
  (e: "update:filters", value: { key: string; channel: string }): void;
  (e: "load"): void;
  (e: "edit", template: NotificationTemplate): void;
  (e: "remove", template: NotificationTemplate): void;
}>();

const localFilters = computed({
  get: () => props.filters,
  set: (val) => emit("update:filters", val),
});
</script>

<template>
  <div class="space-y-5">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h1 class="mt-1 text-3xl font-bold text-surface-900 dark:text-white">{{ i18n.t("templates.title") }}</h1>
        <p class="mt-1 text-sm text-surface-500 dark:text-surface-400">
          {{ i18n.t("templates.desc") }}
        </p>
      </div>
      <Button
        icon="pi pi-refresh"
        :label="i18n.t('templates.btn_refresh')"
        :loading="loading"
        @click="emit('load')"
        class="rounded-xl px-4 h-11"
      />
    </div>

    <!-- Templates DataTable -->
    <div
      class="p-6 bg-surface-0 dark:bg-surface-900 rounded-2xl border border-surface-200 dark:border-surface-800"
    >
      <!-- Filter Bar -->
      <div class="mb-4 grid gap-3 sm:grid-cols-[2fr_1.5fr_auto]">
        <div class="relative w-full">
          <span
            class="pi pi-search absolute right-3 top-1/2 -translate-y-1/2 text-surface-400 dark:text-surface-500 text-sm"
          ></span>
          <InputText
            v-model="localFilters.key"
            :placeholder="i18n.t('templates.search_key')"
            class="pr-9 h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl text-sm"
          />
        </div>
        <Select
          v-model="localFilters.channel"
          :options="filterChannelOptions"
          option-label="label"
          option-value="value"
          :placeholder="i18n.t('templates.filter_channel_placeholder')"
          class="h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl text-sm"
        >
          <template #option="{ option }">
            <div class="flex items-center gap-2 text-xs">
              <span v-if="option.icon" :class="option.icon" class="text-surface-400"></span>
              <span class="capitalize">{{ option.label }}</span>
            </div>
          </template>
        </Select>
        <Button
          icon="pi pi-filter"
          :label="i18n.t('notifications.btn_apply')"
          class="h-10 rounded-xl px-4 text-xs font-semibold"
          @click="emit('load')"
        />
      </div>

      <DataTable
        :value="templates"
        :loading="loading"
        data-key="id"
        responsive-layout="scroll"
        class="p-datatable-sm"
      >
        <Column
          field="key"
          :header="i18n.t('templates.table_key')"
          class="py-3 font-semibold text-surface-800 dark:text-surface-200 font-mono text-xs"
        />
        <Column
          field="channel"
          :header="i18n.t('templates.table_channel')"
          class="py-3 uppercase text-[10px] tracking-wider font-bold"
        >
          <template #body="{ data }">
            <span class="inline-flex items-center gap-1">
              <span class="h-1.5 w-1.5 rounded-full bg-indigo-500"></span>
              <span>{{ i18n.t(`channels.${data.channel}` as TranslationKey) }}</span>
            </span>
          </template>
        </Column>
        <Column
          field="title_template"
          :header="i18n.t('templates.table_title_template')"
          class="py-3 text-surface-500 dark:text-surface-400 truncate max-w-50"
        />
        <Column :header="i18n.t('notifications.table_actions')" class="py-3">
          <template #body="{ data }">
            <div class="flex gap-1.5 justify-end">
              <Button
                size="small"
                text
                icon="pi pi-pencil"
                class="h-8 w-8 rounded-lg"
                @click="emit('edit', data)"
                :title="i18n.t('templates.action_edit')"
              />
              <Button
                size="small"
                text
                severity="danger"
                icon="pi pi-trash"
                class="h-8 w-8 rounded-lg"
                @click="emit('remove', data)"
                :title="i18n.t('templates.action_delete')"
              />
            </div>
          </template>
        </Column>
        <template #empty>
          <div class="py-12 text-center text-surface-400 dark:text-surface-500">
            {{ i18n.t('templates.empty_state') }}
          </div>
        </template>
      </DataTable>
    </div>
  </div>
</template>
