<script setup lang="ts">
import Button from "primevue/button";
import Dialog from "primevue/dialog";
import { ref, watch } from "vue";
import { useSettingsStore } from "@/stores/settings";
import { useI18nStore } from "@/stores/i18n";
import { useToast } from "primevue/usetoast";
import { api } from "@/lib/api";
import type { ChannelConfig, Channel } from "@/lib/types";
import EmailConfigForm from "./EmailConfigForm.vue";
import TelegramConfigForm from "./TelegramConfigForm.vue";
import WebhookConfigForm from "./WebhookConfigForm.vue";

const i18n = useI18nStore();

interface ParsedChannelConfig {
  default_recipient?: string;
  host?: string;
  port?: number;
  username?: string;
  password?: string;
  from?: string;
  tls_mode?: string;
  timeout_seconds?: number;
  headers?: Record<string, string>;
  bot_token?: string;
  api_base_url?: string;
  parse_mode?: string;
  disable_link?: boolean;
  url_template?: string;
  method?: string;
  body_template?: string;
  success_status_min?: number;
  success_status_max?: number;
  response_id_header?: string;
  response_id_json_field?: string;
  allowed_hosts?: string[];
  allow_private_ips?: boolean;
  signing_secret?: string;
}

const props = defineProps<{
  visible: boolean;
  channel: string | null;
  channelConfig: ChannelConfig | undefined;
}>();

const emit = defineEmits<{
  (e: "update:visible", value: boolean): void;
  (e: "saved", savedConfig: ChannelConfig): void;
}>();

const settings = useSettingsStore();
const toast = useToast();

const savingConfig = ref(false);
const isFormValid = ref(true);

const form = ref({
  enabled: false,
  default_recipient: "",
  // Email settings
  email_host: "",
  email_port: 587,
  email_username: "",
  email_password: "",
  email_from: "",
  email_tls_mode: "starttls",
  email_timeout_seconds: 10,
  email_headers: "{}",

  // HTTP webhook settings (SMS, webhook, slack, discord, feishu, dingtalk, wecom)
  http_url_template: "",
  http_method: "POST",
  http_headers: "{}",
  http_body_template: "",
  http_timeout_seconds: 10,
  http_success_status_min: 200,
  http_success_status_max: 299,
  http_response_id_header: "",
  http_response_id_json_field: "",
  http_allowed_hosts: "",
  http_allow_private_ips: false,
  http_signing_secret: "",

  // Telegram settings
  tg_bot_token: "",
  tg_api_base_url: "",
  tg_parse_mode: "HTML",
  tg_disable_link: false,
  tg_timeout_seconds: 10,
  tg_headers: "{}",
  tg_body_template: "",
});

function initForm() {
  if (!props.channel) return;
  isFormValid.value = true;

  const existing = props.channelConfig;
  form.value.enabled = existing?.enabled || false;

  let parsedConfig: ParsedChannelConfig = {};
  if (existing?.config) {
    try {
      parsedConfig = JSON.parse(existing.config);
    } catch (e) {
      console.error("parse existing config json", e);
    }
  }

  form.value.default_recipient = parsedConfig.default_recipient || "";

  const channelName = props.channel;
  if (channelName === "email") {
    form.value.email_host = parsedConfig.host || "";
    form.value.email_port = parsedConfig.port || 587;
    form.value.email_username = parsedConfig.username || "";
    form.value.email_password = parsedConfig.password || "";
    form.value.email_from = parsedConfig.from || "";
    form.value.email_tls_mode = parsedConfig.tls_mode || "starttls";
    form.value.email_timeout_seconds = parsedConfig.timeout_seconds || 10;
    form.value.email_headers = parsedConfig.headers
      ? JSON.stringify(parsedConfig.headers, null, 2)
      : "{}";
  } else if (channelName === "telegram") {
    form.value.tg_bot_token = parsedConfig.bot_token || "";
    form.value.tg_api_base_url = parsedConfig.api_base_url || "";
    form.value.tg_parse_mode = parsedConfig.parse_mode || "HTML";
    form.value.tg_disable_link = parsedConfig.disable_link || false;
    form.value.tg_timeout_seconds = parsedConfig.timeout_seconds || 10;
    form.value.tg_headers = parsedConfig.headers
      ? JSON.stringify(parsedConfig.headers, null, 2)
      : "{}";
    form.value.tg_body_template = parsedConfig.body_template || "";
  } else {
    // For SMS, webhook, slack, discord, feishu, dingtalk, wecom
    form.value.http_url_template = parsedConfig.url_template || "";
    form.value.http_method = parsedConfig.method || "POST";
    form.value.http_headers = parsedConfig.headers
      ? JSON.stringify(parsedConfig.headers, null, 2)
      : "{}";
    form.value.http_body_template = parsedConfig.body_template || "";
    form.value.http_timeout_seconds = parsedConfig.timeout_seconds || 10;
    form.value.http_success_status_min = parsedConfig.success_status_min || 200;
    form.value.http_success_status_max = parsedConfig.success_status_max || 299;
    form.value.http_response_id_header = parsedConfig.response_id_header || "";
    form.value.http_response_id_json_field = parsedConfig.response_id_json_field || "";
    form.value.http_allowed_hosts = parsedConfig.allowed_hosts
      ? parsedConfig.allowed_hosts.join(", ")
      : "";
    form.value.http_allow_private_ips = parsedConfig.allow_private_ips || false;
    form.value.http_signing_secret = parsedConfig.signing_secret || "";
  }
}

watch(
  () => props.visible,
  (newVal) => {
    if (newVal) {
      initForm();
    }
  },
);

async function saveConfig() {
  if (!props.channel) return;

  if (!isFormValid.value) {
    toast.add({
      severity: "error",
      summary: i18n.t("channel_dialog.headers_error"),
      detail: i18n.t("channel_dialog.headers_error_desc"),
      life: 4000,
    });
    return;
  }

  savingConfig.value = true;
  try {
    let configObj: ParsedChannelConfig = {};

    if (props.channel === "email") {
      configObj = {
        host: form.value.email_host.trim(),
        port: Number(form.value.email_port),
        username: form.value.email_username.trim(),
        password: form.value.email_password.trim(),
        from: form.value.email_from.trim(),
        tls_mode: form.value.email_tls_mode,
        timeout_seconds: Number(form.value.email_timeout_seconds),
        headers: JSON.parse(form.value.email_headers || "{}"),
      };
    } else if (props.channel === "telegram") {
      configObj = {
        bot_token: form.value.tg_bot_token.trim(),
        api_base_url: form.value.tg_api_base_url.trim(),
        parse_mode: form.value.tg_parse_mode,
        disable_link: form.value.tg_disable_link,
        timeout_seconds: Number(form.value.tg_timeout_seconds),
        headers: JSON.parse(form.value.tg_headers || "{}"),
        body_template: form.value.tg_body_template.trim(),
      };
    } else if (props.channel === "in_app") {
      configObj = {};
    } else {
      let allowedHostsList: string[] = [];
      if (form.value.http_allowed_hosts.trim()) {
        allowedHostsList = form.value.http_allowed_hosts
          .split(",")
          .map((s) => s.trim())
          .filter(Boolean);
      }
      configObj = {
        url_template: form.value.http_url_template.trim(),
        method: form.value.http_method,
        headers: JSON.parse(form.value.http_headers || "{}"),
        body_template: form.value.http_body_template.trim(),
        timeout_seconds: Number(form.value.http_timeout_seconds),
        success_status_min: Number(form.value.http_success_status_min),
        success_status_max: Number(form.value.http_success_status_max),
        response_id_header: form.value.http_response_id_header.trim(),
        response_id_json_field: form.value.http_response_id_json_field.trim(),
        signing_secret: form.value.http_signing_secret.trim(),
      };

      if (props.channel === "webhook") {
        configObj.allowed_hosts = allowedHostsList;
        configObj.allow_private_ips = form.value.http_allow_private_ips;
      }
    }

    if (form.value.default_recipient.trim()) {
      configObj.default_recipient = form.value.default_recipient.trim();
    }

    const payload: ChannelConfig = {
      id: props.channelConfig?.id || undefined,
      tenant_id: settings.tenantId || undefined,
      channel: props.channel as Channel,
      enabled: form.value.enabled,
      config: JSON.stringify(configObj),
    };

    const saved = await api.saveChannelConfig(payload);

    toast.add({
      severity: "success",
      summary: i18n.t("channel_dialog.save_success"),
      detail: i18n.t("channel_dialog.save_success_desc", { channel: props.channel }),
      life: 3000,
    });
    emit("saved", saved);
    emit("update:visible", false);
  } catch (err: unknown) {
    const errMsg = err instanceof Error ? err.message : String(err);
    toast.add({
      severity: "error",
      summary: i18n.t("channel_dialog.save_failed"),
      detail: errMsg || "Failed to save configuration.",
      life: 5000,
    });
  } finally {
    savingConfig.value = false;
  }
}
</script>

<template>
  <Dialog
    :visible="props.visible"
    @update:visible="(val) => emit('update:visible', val)"
    modal
    :header="i18n.t('channel_dialog.title', { name: (props.channel || '').toUpperCase() })"
    class="w-[92vw] max-w-160 rounded-2xl"
  >
    <div v-if="props.channel" class="flex flex-col">
      <div class="py-2 space-y-4 overflow-y-auto max-h-[60vh] scrollbar-none pr-1">
        <!-- Enable toggle -->
        <div
          class="flex items-center justify-between p-4 bg-surface-50 dark:bg-surface-950 border border-surface-100 dark:border-surface-800 rounded-xl"
        >
          <div>
            <div class="text-sm font-bold text-surface-900 dark:text-white">
              {{ i18n.t("channel_dialog.enable_label") }}
            </div>
            <div class="text-xs text-surface-500 dark:text-surface-400 mt-0.5">
              {{ i18n.t("channel_dialog.enable_desc") }}
            </div>
          </div>
          <label class="relative inline-flex items-center cursor-pointer select-none">
            <input type="checkbox" v-model="form.enabled" class="sr-only peer" />
            <div
              class="w-11 h-6 bg-surface-200 dark:bg-surface-800 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:left-0.5 after:bg-white after:border-surface-300 dark:after:border-surface-600 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-reisa-lilac-500"
            ></div>
          </label>
        </div>

        <div v-show="form.enabled" class="space-y-4 pt-1">
          <!-- Default Recipient fallback configuration -->
          <div
            class="flex flex-col gap-2 p-4 bg-surface-50 dark:bg-surface-950 border border-surface-100 dark:border-surface-800 rounded-xl"
          >
            <div class="flex items-center justify-between">
              <label class="text-sm font-bold text-surface-900 dark:text-white">
                {{ i18n.t("channel_dialog.default_recipient") }}
              </label>
              <span class="text-[10px] text-surface-400 dark:text-surface-500 font-bold bg-surface-200 dark:bg-surface-850 px-1.5 py-0.5 rounded">
                {{ i18n.t("channel_dialog.optional_label") }}
              </span>
            </div>
            <input
              type="text"
              v-model="form.default_recipient"
              class="w-full h-10 px-3 text-xs bg-white dark:bg-surface-900 border border-surface-200 dark:border-surface-800 focus:border-reisa-lilac-500 rounded-xl outline-none"
              :placeholder="i18n.t('channel_dialog.default_recipient_placeholder')"
            />
            <div class="text-[10px] text-surface-500 dark:text-surface-400">
              {{ i18n.t("channel_dialog.default_recipient_desc") }}
            </div>
          </div>

          <!-- 1. EMAIL FORM -->
          <EmailConfigForm
            v-if="props.channel === 'email'"
            v-model="form"
            @valid="(val) => isFormValid = val"
          />

          <!-- 2. TELEGRAM FORM -->
          <TelegramConfigForm
            v-else-if="props.channel === 'telegram'"
            v-model="form"
            @valid="(val) => isFormValid = val"
          />

          <!-- 3. IN_APP FORM -->
          <div
            v-else-if="props.channel === 'in_app'"
            class="p-4 bg-surface-50 dark:bg-surface-950 border border-surface-100 dark:border-surface-800 rounded-xl text-xs text-surface-500 dark:text-surface-400 flex items-start gap-2.5"
          >
            <span class="pi pi-info-circle text-reisa-lilac-500 mt-0.5"></span>
            <span>{{ i18n.t("channel_dialog.in_app_desc") }}</span>
          </div>

          <!-- 4. HTTP / WEBHOOK / CHAT PROVIDERS FORM (sms, webhook, slack, discord, feishu, dingtalk, wecom) -->
          <WebhookConfigForm
            v-else
            v-model="form"
            :channel="props.channel"
            @valid="(val) => isFormValid = val"
          />
        </div>
      </div>

      <!-- Dialog Footer Actions -->
      <div
        class="mt-6 pt-4 border-t border-surface-100 dark:border-surface-800 flex justify-end gap-3"
      >
        <Button
          severity="secondary"
          outlined
          :label="i18n.t('channel_dialog.btn_cancel')"
          class="h-10 px-4 rounded-xl text-xs font-bold"
          @click="emit('update:visible', false)"
        />
        <Button
          icon="pi pi-check"
          :label="i18n.t('channel_dialog.btn_save_config')"
          :loading="savingConfig"
          class="h-10 px-5 rounded-xl text-xs font-bold"
          @click="saveConfig"
        />
      </div>
    </div>
  </Dialog>
</template>
