<script setup lang="ts">
import { reactive, ref, watch, computed } from "vue";
import InputText from "primevue/inputtext";
import Select from "primevue/select";
import Textarea from "primevue/textarea";
import Button from "primevue/button";
import { useI18nStore } from "@/stores/i18n";

import type { DispatchForm, ChannelOption, NotificationTemplate } from "@/lib/types";

const props = defineProps<{
  form: DispatchForm;
  channelOptions: ChannelOption[];
  templates: NotificationTemplate[];
  saving: boolean;
  isFormJsonValid: boolean;
}>();

const emit = defineEmits<{
  (e: "update:form", value: DispatchForm): void;
  (e: "create"): void;
}>();

const i18n = useI18nStore();
const showAdvanced = ref(false);

const localForm = reactive<DispatchForm>({ ...props.form });

watch(
  () => props.form,
  (value) => {
    Object.assign(localForm, value);
  },
  { deep: true },
);

watch(
  localForm,
  (value) => {
    emit("update:form", { ...value });
  },
  { deep: true },
);

const recipientConfig = computed(() => {
  const channel = localForm.channel;
  switch (channel) {
    case "email":
      return {
        label: i18n.t("composer.recipient_email_label"),
        placeholder: i18n.t("composer.recipient_email_placeholder"),
        icon: "pi pi-envelope",
        help: i18n.t("composer.recipient_email_help"),
      };
    case "sms":
      return {
        label: i18n.t("composer.recipient_sms_label"),
        placeholder: i18n.t("composer.recipient_sms_placeholder"),
        icon: "pi pi-phone",
        help: i18n.t("composer.recipient_sms_help"),
      };
    case "telegram":
      return {
        label: i18n.t("composer.recipient_tg_label"),
        placeholder: i18n.t("composer.recipient_tg_placeholder"),
        icon: "pi pi-telegram",
        help: i18n.t("composer.recipient_tg_help"),
      };
    case "slack":
      return {
        label: i18n.t("composer.recipient_slack_label"),
        placeholder: i18n.t("composer.recipient_slack_placeholder"),
        icon: "pi pi-slack",
        help: i18n.t("composer.recipient_slack_help"),
      };
    case "discord":
      return {
        label: i18n.t("composer.recipient_discord_label"),
        placeholder: i18n.t("composer.recipient_discord_placeholder"),
        icon: "pi pi-discord",
        help: i18n.t("composer.recipient_discord_help"),
      };
    case "webhook":
      return {
        label: i18n.t("composer.recipient_webhook_label"),
        placeholder: i18n.t("composer.recipient_webhook_placeholder"),
        icon: "pi pi-globe",
        help: i18n.t("composer.recipient_webhook_help"),
      };
    case "in_app":
      return {
        label: i18n.t("composer.recipient_in_app_label"),
        placeholder: i18n.t("composer.recipient_in_app_placeholder"),
        icon: "pi pi-user",
        help: i18n.t("composer.recipient_in_app_help"),
      };
    case "feishu":
      return {
        label: i18n.t("composer.recipient_feishu_label"),
        placeholder: i18n.t("composer.recipient_feishu_placeholder"),
        icon: "pi pi-comments",
        help: i18n.t("composer.recipient_feishu_help"),
      };
    case "dingtalk":
      return {
        label: i18n.t("composer.recipient_dingtalk_label"),
        placeholder: i18n.t("composer.recipient_dingtalk_placeholder"),
        icon: "pi pi-comments",
        help: i18n.t("composer.recipient_dingtalk_help"),
      };
    case "wecom":
      return {
        label: i18n.t("composer.recipient_wecom_label"),
        placeholder: i18n.t("composer.recipient_wecom_placeholder"),
        icon: "pi pi-briefcase",
        help: i18n.t("composer.recipient_wecom_help"),
      };
    default:
      return {
        label: i18n.t("composer.recipient_default_label"),
        placeholder: i18n.t("composer.recipient_default_placeholder"),
        icon: "pi pi-user",
        help: i18n.t("composer.recipient_default_help"),
      };
  }
});

const filteredTemplates = computed(() => {
  const list = (props.templates || []).filter((t) => t.channel === localForm.channel);
  return [
    { label: i18n.t("composer.custom_template"), value: "" },
    ...list.map((t) => ({ label: `${t.key}`, value: t.key })),
  ];
});

watch(
  () => [localForm.template_key, localForm.channel],
  ([newKey, newChannel]) => {
    if (!newKey) return;
    const template = (props.templates || []).find(
      (t) => t.key === newKey && t.channel === newChannel,
    );
    if (template) {
      const combined = `${template.title_template || ""} ${template.body_template || ""}`;
      const regex = /\{\{([^}]+)\}\}/g;
      const variables = new Set<string>();
      let match;
      while ((match = regex.exec(combined)) !== null) {
        if (match[1]) {
          variables.add(match[1].trim());
        }
      }
      const metaObj: Record<string, string> = {};
      variables.forEach((v) => {
        metaObj[v] = "";
      });
      localForm.metadata = JSON.stringify(metaObj, null, 2);
    }
  },
);

function resetTab() {
  showAdvanced.value = false;
}
defineExpose({ resetTab });
</script>

<template>
  <div class="space-y-5">
    <div
      class="p-6 bg-surface-0 dark:bg-surface-900 rounded-2xl border border-surface-200 dark:border-surface-800"
    >
      <div class="mb-4 flex items-start gap-3">
        <div
          class="grid h-11 w-11 place-items-center rounded-xl bg-blue-50 dark:bg-blue-950/30 text-blue-600 dark:text-blue-400"
        >
          <span class="pi pi-send text-lg"></span>
        </div>
        <div>
          <h2 class="text-lg font-bold text-surface-900 dark:text-white">
            {{ i18n.t("composer.title") }}
          </h2>
          <p class="text-xs text-surface-500 dark:text-surface-400">
            {{ i18n.t("composer.desc") }}
          </p>
        </div>
      </div>

      <div class="mt-5 space-y-6">
        <!-- Section 1: Route & Destination -->
        <div class="space-y-4">
          <div
            class="flex items-center justify-between border-b border-surface-200 dark:border-surface-800/80 pb-1.5"
          >
            <h3
              class="text-xs font-bold uppercase tracking-wider text-surface-400 dark:text-surface-500"
            >
              {{ i18n.t("composer.sec_route") }}
            </h3>
          </div>

          <div class="space-y-4">
            <div class="space-y-1">
              <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">{{
                i18n.t("composer.field_channel")
              }}</label>
              <Select
                v-model="localForm.channel"
                :options="channelOptions"
                option-label="label"
                option-value="value"
                class="h-11 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl text-sm"
              >
                <template #option="{ option }">
                  <div class="flex items-center gap-2.5 text-xs font-medium">
                    <span :class="[option.icon, 'text-surface-400']"></span>
                    <span class="capitalize">{{ option.label }}</span>
                  </div>
                </template>
              </Select>
            </div>

            <div class="space-y-1">
              <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">{{
                recipientConfig.label
              }}</label>
              <div class="relative w-full">
                <span
                  :class="[
                    recipientConfig.icon,
                    'absolute right-3 top-1/2 -translate-y-1/2 text-surface-400 dark:text-surface-500',
                  ]"
                ></span>
                <InputText
                  v-model="localForm.recipient"
                  :placeholder="recipientConfig.placeholder"
                  class="pr-9 h-11 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl text-sm"
                />
              </div>
            </div>

            <div
              class="p-3 bg-surface-50 dark:bg-surface-950/30 rounded-xl border border-surface-100 dark:border-surface-800/80 text-[11px] text-surface-500 dark:text-surface-400 leading-normal flex items-start gap-2.5"
            >
              <span class="pi pi-info-circle text-surface-400 shrink-0 mt-0.5"></span>
              <span>{{ recipientConfig.help }}</span>
            </div>
          </div>
        </div>

        <!-- Section 2: Message Content -->
        <div class="space-y-4">
          <div
            class="flex items-center justify-between border-b border-surface-200 dark:border-surface-800/80 pb-1.5"
          >
            <h3
              class="text-xs font-bold uppercase tracking-wider text-surface-400 dark:text-surface-500"
            >
              {{ i18n.t("composer.sec_content") }}
            </h3>
          </div>

          <div class="space-y-4">
            <div class="space-y-1">
              <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">{{
                i18n.t("composer.field_template")
              }}</label>
              <Select
                v-model="localForm.template_key"
                :options="filteredTemplates"
                option-label="label"
                option-value="value"
                class="h-11 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl text-sm"
                :placeholder="i18n.t('composer.field_template_placeholder')"
              >
                <template #option="{ option }">
                  <span class="font-mono text-xs">{{ option.label }}</span>
                </template>
              </Select>
              <span class="text-[9px] text-surface-400 dark:text-surface-500">
                {{ i18n.t("composer.field_template_help") }}
              </span>
            </div>

            <!-- Template variables info banner -->
            <div
              v-if="localForm.template_key"
              class="p-3 bg-amber-50/80 dark:bg-amber-950/20 rounded-xl border border-amber-200/50 dark:border-amber-900/30 text-[11px] text-amber-700 dark:text-amber-400 leading-normal flex items-start gap-2"
            >
              <span class="pi pi-exclamation-triangle shrink-0 mt-0.5 text-amber-500"></span>
              <span>
                {{ i18n.t("composer.template_banner") }}
              </span>
            </div>

            <div class="space-y-1">
              <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">{{
                i18n.t("composer.field_title")
              }}</label>
              <InputText
                v-model="localForm.title"
                :placeholder="i18n.t('composer.field_title_placeholder')"
                class="h-11 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl text-sm"
                :disabled="Boolean(localForm.template_key)"
              />
            </div>

            <div class="space-y-1">
              <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">{{
                i18n.t("composer.field_body")
              }}</label>
              <Textarea
                v-model="localForm.body"
                rows="4"
                :placeholder="i18n.t('composer.field_body_placeholder')"
                class="border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl text-xs"
                :disabled="Boolean(localForm.template_key)"
              />
            </div>
          </div>
        </div>

        <!-- Section 3: Advanced Options Toggle Header -->
        <div class="pt-2 border-t border-surface-100 dark:border-surface-800">
          <button
            type="button"
            @click="showAdvanced = !showAdvanced"
            class="flex items-center justify-between w-full py-2 text-left text-xs font-bold text-surface-650 dark:text-surface-400 hover:text-surface-900 dark:hover:text-white transition-colors"
          >
            <div class="flex items-center gap-2">
              <span
                :class="[showAdvanced ? 'pi pi-chevron-up' : 'pi pi-chevron-down', 'text-[10px]']"
              ></span>
              <span>{{ i18n.t("composer.sec_advanced") }}</span>
              <span
                v-if="!isFormJsonValid"
                class="h-2 w-2 rounded-full bg-rose-500 animate-pulse shrink-0"
                title="Invalid JSON syntax"
              ></span>
            </div>
            <span class="text-[9px] font-normal text-surface-400 dark:text-surface-500">
              {{ showAdvanced ? i18n.t("composer.collapse") : i18n.t("composer.expand") }}
            </span>
          </button>

          <!-- Collapsible Content -->
          <div v-show="showAdvanced" class="mt-3 space-y-4 pt-1">
            <div class="space-y-1">
              <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">{{
                i18n.t("composer.field_group")
              }}</label>
              <InputText
                v-model="localForm.group_key"
                :placeholder="i18n.t('composer.field_group_placeholder')"
                class="h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl text-sm font-mono"
              />
            </div>

            <div class="space-y-1">
              <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">{{
                i18n.t("composer.field_idempotency")
              }}</label>
              <InputText
                v-model="localForm.idempotency_key"
                :placeholder="i18n.t('composer.field_idempotency_placeholder')"
                class="h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl text-sm font-mono"
              />
            </div>

            <div class="space-y-1">
              <div class="flex justify-between items-center">
                <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">{{
                  i18n.t("composer.field_metadata")
                }}</label>
                <span
                  v-if="localForm.metadata.trim()"
                  class="text-[9px] font-bold uppercase tracking-wider"
                  :class="isFormJsonValid ? 'text-reisa-lilac-500' : 'text-rose-500'"
                >
                  {{
                    isFormJsonValid ? i18n.t("composer.json_valid") : i18n.t("composer.json_error")
                  }}
                </span>
              </div>
              <Textarea
                v-model="localForm.metadata"
                rows="5"
                class="font-mono text-xs border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl p-3"
                :class="{
                  'border-rose-300 focus:border-rose-500 focus:ring-rose-500': !isFormJsonValid,
                }"
              />
            </div>
          </div>
        </div>
      </div>

      <div class="mt-4 pt-3 border-t border-surface-100 dark:border-surface-800 flex justify-end">
        <Button
          icon="pi pi-plus"
          :label="i18n.t('composer.btn_dispatch')"
          :loading="saving"
          class="w-full h-11 rounded-xl font-semibold"
          @click="emit('create')"
        />
      </div>
    </div>
  </div>
</template>
