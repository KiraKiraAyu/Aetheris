<script setup lang="ts">
import Button from "primevue/button";
import InputText from "primevue/inputtext";
import Select from "primevue/select";
import { ref } from "vue";
import { useSettingsStore } from "@/stores/settings";
import { useToast } from "primevue/usetoast";
import { useI18nStore } from "@/stores/i18n";

const emit = defineEmits<{
  (e: "saved"): void;
}>();

const settings = useSettingsStore();
const toast = useToast();
const i18n = useI18nStore();

const baseUrl = ref(settings.apiBaseUrl);
const apiKey = ref(settings.apiKey);
const tenantId = ref(settings.tenantId);

const testing = ref(false);

const languageOptions = [
  { label: "English", value: "en" },
  { label: "简体中文", value: "zh" },
];

function save() {
  settings.save(baseUrl.value, apiKey.value, tenantId.value);
  toast.add({
    severity: "success",
    summary: i18n.t("settings.api_saved"),
    detail: i18n.t("settings.api_saved_desc"),
    life: 3000,
  });
  emit("saved");
}

async function testConnection() {
  testing.value = true;
  const originalBase = settings.apiBaseUrl;
  const originalKey = settings.apiKey;
  const originalTenant = settings.tenantId;

  settings.apiBaseUrl = baseUrl.value.trim().replace(/\/$/, "") || "/api";
  settings.apiKey = apiKey.value.trim();
  settings.tenantId = tenantId.value.trim() || "default";

  await settings.checkConnection();

  if (settings.connectionStatus === "connected") {
    toast.add({
      severity: "success",
      summary: i18n.t("settings.conn_success"),
      detail: i18n.t("settings.conn_success_desc"),
      life: 3000,
    });
    emit("saved");
  } else {
    toast.add({
      severity: "error",
      summary: i18n.t("settings.conn_failed"),
      detail: settings.connectionError || "Unknown connection error",
      life: 5000,
    });
  }

  settings.apiBaseUrl = originalBase;
  settings.apiKey = originalKey;
  settings.tenantId = originalTenant;

  testing.value = false;
}
</script>

<template>
  <div
    class="p-6 bg-white dark:bg-surface-900 rounded-2xl border border-surface-200 dark:border-surface-800 h-fit"
  >
    <div class="mb-5 flex items-start gap-3">
      <div
        class="grid h-11 w-11 place-items-center rounded-xl bg-surface-950 dark:bg-white text-white dark:text-surface-950 shadow"
      >
        <span class="pi pi-key text-lg"></span>
      </div>
      <div>
        <h2 class="text-lg font-bold text-surface-900 dark:text-white">
          {{ i18n.t("settings.api_params_title") }}
        </h2>
        <p class="text-xs text-surface-500 dark:text-surface-400">
          {{ i18n.t("settings.api_params_desc") }}
        </p>
      </div>
    </div>

    <div class="space-y-4">
      <div class="space-y-1">
        <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
          {{ i18n.t("settings.field_api_url") }}
        </label>
        <div class="relative w-full">
          <span
            class="pi pi-link absolute right-3 top-1/2 -translate-y-1/2 text-surface-400"
          ></span>
          <InputText
            v-model="baseUrl"
            placeholder="/api"
            class="pr-9 h-11 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl w-full text-sm"
          />
        </div>
      </div>

      <div class="space-y-1">
        <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
          {{ i18n.t("settings.field_api_key") }}
        </label>
        <div class="relative w-full">
          <span
            class="pi pi-lock absolute right-3 top-1/2 -translate-y-1/2 text-surface-400"
          ></span>
          <InputText
            v-model="apiKey"
            type="password"
            :placeholder="i18n.t('settings.placeholder_api_key')"
            class="pr-9 h-11 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl w-full text-sm"
          />
        </div>
      </div>

      <div class="space-y-1">
        <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
          {{ i18n.t("settings.field_tenant_id") }}
        </label>
        <div class="relative w-full">
          <span
            class="pi pi-building absolute right-3 top-1/2 -translate-y-1/2 text-surface-400"
          ></span>
          <InputText
            v-model="tenantId"
            placeholder="default"
            class="pr-9 h-11 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl w-full text-sm"
          />
        </div>
      </div>

      <div
        v-if="settings.connectionStatus === 'connected'"
        class="p-3 bg-emerald-50/40 dark:bg-emerald-950/10 border border-emerald-200/50 dark:border-emerald-900/30 rounded-xl text-xs text-emerald-600 dark:text-emerald-400 flex items-center gap-2"
      >
        <span class="pi pi-check-circle text-emerald-500"></span>
        <span>{{ i18n.t("settings.status_verified") }}</span>
      </div>
      <div
        v-else-if="settings.connectionStatus === 'disconnected'"
        class="p-3 bg-rose-50/40 dark:bg-rose-950/10 border border-rose-200/50 dark:border-rose-900/30 rounded-xl text-xs text-rose-600 dark:text-rose-400 flex items-start gap-2"
      >
        <span class="pi pi-exclamation-circle mt-0.5 text-rose-500"></span>
        <span class="break-all">{{ settings.connectionError }}</span>
      </div>

      <div class="grid grid-cols-2 gap-3 pt-3">
        <Button
          severity="secondary"
          outlined
          icon="pi pi-wifi"
          :label="i18n.t('settings.btn_test_ping')"
          :loading="testing"
          class="h-11 rounded-xl"
          @click="testConnection"
        />
        <Button icon="pi pi-save" :label="i18n.t('settings.btn_save')" class="h-11 rounded-xl" @click="save" />
      </div>

      <!-- Language Selector -->
      <div class="border-t border-surface-200 dark:border-surface-800 pt-4 mt-4 space-y-3">
        <div class="flex flex-col">
          <label class="text-xs font-bold text-surface-900 dark:text-white">
            {{ i18n.t("settings.language_section") }}
          </label>
          <span class="text-[11px] text-surface-500 dark:text-surface-400">
            {{ i18n.t("settings.language_desc") }}
          </span>
        </div>
        <Select
          :model-value="i18n.locale"
          @update:model-value="i18n.setLanguage"
          :options="languageOptions"
          option-label="label"
          option-value="value"
          class="h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl text-sm w-full"
        />
      </div>
    </div>
  </div>
</template>
