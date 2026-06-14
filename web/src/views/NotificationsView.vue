<script setup lang="ts">
import Dialog from "primevue/dialog";
import { computed, onMounted, reactive, ref, watch } from "vue";
import StatusBadge from "@/components/common/StatusBadge.vue";
import { api } from "@/lib/api";
import { useToast } from "primevue/usetoast";
import { useSettingsStore } from "@/stores/settings";
import { useI18nStore, type TranslationKey } from "@/stores/i18n";

const i18n = useI18nStore();
import {
  channels,
  statuses,
  type Channel,
  type DeliveryAttempt,
  type NotificationRecord,
  type ChannelOption,
  type FilterChannelOption,
  type ChannelConfig,
  type NotificationTemplate,
} from "@/lib/types";

// Components
import NotificationList from "@/components/notifications/NotificationList.vue";
import DispatchComposer from "@/components/notifications/DispatchComposer.vue";
import AttemptTimeline from "@/components/notifications/AttemptTimeline.vue";
import EmailMockup from "@/components/notifications/DeviceMockups/EmailMockup.vue";
import SmsMockup from "@/components/notifications/DeviceMockups/SmsMockup.vue";
import ChatMockup from "@/components/notifications/DeviceMockups/ChatMockup.vue";
import WebhookMockup from "@/components/notifications/DeviceMockups/WebhookMockup.vue";
import InAppMockup from "@/components/notifications/DeviceMockups/InAppMockup.vue";

const toast = useToast();
const settings = useSettingsStore();
const channelConfigs = ref<Record<string, ChannelConfig>>({});
const notifications = ref<NotificationRecord[]>([]);
const attempts = ref<DeliveryAttempt[]>([]);
const selected = ref<NotificationRecord | null>(null);
const detailVisible = ref(false);
const loading = ref(false);
const saving = ref(false);
const error = ref("");

const composerRef = ref();

const filters = reactive({
  recipient: "",
  channel: "" as Channel | "",
  status: "" as "" | (typeof statuses)[number],
});

const form = reactive({
  recipient: "",
  channel: "in_app" as Channel,
  template_key: "",
  title: "",
  body: "",
  group_key: "",
  idempotency_key: "",
  metadata: "{}",
});

const channelMap = computed(() => {
  const map: Record<string, { label: string; icon: string; tone: string }> = {
    email: {
      label: i18n.t("channels.email"),
      icon: "pi pi-envelope",
      tone: "text-sky-500 bg-sky-50 dark:bg-sky-950/20",
    },
    sms: {
      label: i18n.t("channels.sms"),
      icon: "pi pi-phone",
      tone: "text-indigo-500 bg-indigo-50 dark:bg-indigo-950/20",
    },
    webhook: {
      label: i18n.t("channels.webhook"),
      icon: "pi pi-globe",
      tone: "text-teal-500 bg-teal-50 dark:bg-teal-950/20",
    },
    in_app: {
      label: i18n.t("channels.in_app"),
      icon: "pi pi-inbox",
      tone: "text-amber-500 bg-amber-50 dark:bg-amber-950/20",
    },
    telegram: {
      label: i18n.t("channels.telegram"),
      icon: "pi pi-telegram",
      tone: "text-blue-500 bg-blue-50 dark:bg-blue-950/20",
    },
    slack: {
      label: i18n.t("channels.slack"),
      icon: "pi pi-slack",
      tone: "text-purple-500 bg-purple-50 dark:bg-purple-950/20",
    },
    discord: {
      label: i18n.t("channels.discord"),
      icon: "pi pi-discord",
      tone: "text-violet-500 bg-violet-50 dark:bg-violet-950/20",
    },
    feishu: {
      label: i18n.t("channels.feishu"),
      icon: "pi pi-comments",
      tone: "text-reisa-lilac-500 bg-reisa-lilac-50 dark:bg-reisa-lilac-950/20",
    },
    dingtalk: {
      label: i18n.t("channels.dingtalk"),
      icon: "pi pi-comments",
      tone: "text-cyan-500 bg-cyan-50 dark:bg-cyan-950/20",
    },
    wecom: {
      label: i18n.t("channels.wecom"),
      icon: "pi pi-briefcase",
      tone: "text-rose-500 bg-rose-50 dark:bg-rose-950/20",
    },
  };
  return map;
});

const channelOptions = computed<ChannelOption[]>(() =>
  channels
    .filter((value) => channelConfigs.value[value]?.enabled === true)
    .map((value) => ({
      label: channelMap.value[value]?.label || value,
      value,
      icon: channelMap.value[value]?.icon || "pi pi-send",
    })),
);

watch(channelOptions, (newOptions) => {
  if (newOptions.length > 0) {
    const firstOption = newOptions[0];
    if (firstOption && !newOptions.some((opt) => opt.value === form.channel)) {
      form.channel = firstOption.value;
    }
  }
});

const filterChannelOptions = computed<FilterChannelOption[]>(() => [
  { label: i18n.t("notifications.filter_channel"), value: "" as const },
  ...channelOptions.value,
]);

const statusOptions = computed(() =>
  statuses.map((value) => ({
    label: i18n.t(`status.${value}` as TranslationKey),
    value,
  })),
);

const isFormJsonValid = computed(() => {
  if (!form.metadata.trim()) return true;
  try {
    JSON.parse(form.metadata);
    return true;
  } catch {
    return false;
  }
});

async function fetchChannelConfigs() {
  if (!settings.tenantId || settings.connectionStatus !== "connected") {
    channelConfigs.value = {};
    return;
  }
  try {
    const list = await api.listChannelConfigs();
    const map: Record<string, ChannelConfig> = {};
    list.forEach((cfg) => {
      map[cfg.channel] = cfg;
    });
    channelConfigs.value = map;
  } catch (err) {
    console.error("fetch channel configs failed", err);
  }
}

const templates = ref<NotificationTemplate[]>([]);

async function fetchTemplates() {
  if (!settings.tenantId || settings.connectionStatus !== "connected") {
    templates.value = [];
    return;
  }
  try {
    templates.value = await api.listTemplates({});
  } catch (err) {
    console.error("fetch templates failed", err);
  }
}

async function load() {
  loading.value = true;
  error.value = "";
  try {
    notifications.value = await api.listNotifications({ ...filters, limit: 100 });
    await fetchChannelConfigs();
    await fetchTemplates();
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

watch(
  () => settings.tenantId,
  () => {
    load();
  },
);

watch(
  () => settings.connectionStatus,
  (status) => {
    if (status === "connected") {
      load();
    }
  },
);

async function openDetail(record: NotificationRecord) {
  detailVisible.value = true;
  attempts.value = [];
  try {
    selected.value = await api.getNotification(record.id);
    attempts.value = await api.listAttempts(record.id);
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  }
}

async function createNotification() {
  if (!form.recipient) {
    toast.add({
      severity: "error",
      summary: i18n.t("notifications.toast_missing_recipient_title"),
      detail: i18n.t("notifications.toast_missing_recipient_desc"),
      life: 3000,
    });
    return;
  }
  if (!isFormJsonValid.value) {
    toast.add({
      severity: "error",
      summary: i18n.t("notifications.toast_invalid_json_title"),
      detail: i18n.t("notifications.toast_invalid_json_desc"),
      life: 3000,
    });
    return;
  }

  saving.value = true;
  error.value = "";
  try {
    const metadata = form.metadata.trim() ? JSON.parse(form.metadata) : {};
    await api.createNotification({
      recipient: form.recipient,
      channel: form.channel,
      template_key: form.template_key || undefined,
      title: form.title || undefined,
      body: form.body || undefined,
      group_key: form.group_key || undefined,
      idempotency_key: form.idempotency_key || undefined,
      metadata,
    });

    toast.add({
      severity: "success",
      summary: i18n.t("notifications.toast_dispatched_title"),
      detail: i18n.t("notifications.toast_dispatched_desc", { recipient: form.recipient }),
      life: 3000,
    });

    // Reset some fields
    form.title = "";
    form.body = "";
    form.idempotency_key = "";
    form.metadata = "{}";
    if (composerRef.value) {
      composerRef.value.resetTab();
    }

    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
    toast.add({
      severity: "error",
      summary: i18n.t("notifications.toast_failed_title"),
      detail: error.value,
      life: 5000,
    });
  } finally {
    saving.value = false;
  }
}

function updateDispatchForm(nextForm: typeof form) {
  Object.assign(form, nextForm);
}

onMounted(load);
</script>

<template>
  <section class="grid gap-6 xl:grid-cols-[1fr_420px]">
    <div class="space-y-5">
      <NotificationList
        :notifications="notifications"
        :loading="loading"
        v-model:filters="filters"
        :filter-channel-options="filterChannelOptions"
        :status-options="statusOptions"
        :channel-map="channelMap"
        @load="load"
        @inspect="openDetail"
      />
    </div>

    <!-- Right Column: Create Notification Form -->
    <DispatchComposer
      ref="composerRef"
      :form="form"
      :channel-options="channelOptions"
      :templates="templates"
      :saving="saving"
      :is-form-json-valid="isFormJsonValid"
      @update:form="updateDispatchForm"
      @create="createNotification"
    />

    <!-- Notification Inspector Detail Dialog -->
    <Dialog
      v-model:visible="detailVisible"
      modal
      :header="i18n.t('notifications.details_inspector_title')"
      class="w-[92vw] max-w-235 rounded-2xl"
    >
      <div v-if="selected" class="grid gap-6 py-2 overflow-y-auto max-h-[80vh] scrollbar-thin">
        <!-- Info Cards Header -->
        <div class="grid gap-3 sm:grid-cols-3">
          <div
            class="p-3 bg-surface-50 dark:bg-surface-950 border border-surface-100 dark:border-surface-800/80 rounded-xl"
          >
            <div class="text-[10px] uppercase font-bold text-surface-400 dark:text-surface-500">
              {{ i18n.t("notifications.details_status") }}
            </div>
            <StatusBadge class="mt-1" :value="selected.status" />
          </div>
          <div
            class="p-3 bg-surface-50 dark:bg-surface-950 border border-surface-100 dark:border-surface-800/80 rounded-xl"
          >
            <div class="text-[10px] uppercase font-bold text-surface-400 dark:text-surface-500">
              {{ i18n.t("notifications.details_channel_scope") }}
            </div>
            <div
              class="mt-1 flex items-center gap-1.5 font-bold capitalize text-sm text-surface-800 dark:text-surface-200"
            >
              <span
                :class="channelMap[selected.channel]?.icon"
                class="text-surface-450 text-xs"
              ></span>
              {{ i18n.t(`channels.${selected.channel}` as any) }}
            </div>
          </div>
          <div
            class="p-3 bg-surface-50 dark:bg-surface-950 border border-surface-100 dark:border-surface-800/80 rounded-xl"
          >
            <div class="text-[10px] uppercase font-bold text-surface-400 dark:text-surface-500">
              {{ i18n.t("notifications.details_recipient") }}
            </div>
            <div
              class="mt-1 font-mono font-bold text-xs text-surface-800 dark:text-surface-200 truncate"
              :title="selected.recipient"
            >
              {{ selected.recipient }}
            </div>
          </div>
        </div>

        <!-- Visual Device Message Preview -->
        <div class="space-y-2">
          <div
            class="text-xs font-bold uppercase tracking-wider text-surface-400 dark:text-surface-500"
          >
            {{ i18n.t("notifications.details_preview_title") }}
          </div>

          <div
            class="w-full flex justify-center bg-surface-100 dark:bg-surface-950/60 p-4 sm:p-6 rounded-2xl border border-surface-200 dark:border-surface-800/80"
          >
            <EmailMockup v-if="selected.channel === 'email'" :selected="selected" />
            <SmsMockup v-else-if="selected.channel === 'sms'" :selected="selected" />
            <ChatMockup
              v-else-if="
                ['telegram', 'slack', 'discord', 'feishu', 'dingtalk', 'wecom'].includes(
                  selected.channel,
                )
              "
              :selected="selected"
              :channel-map="channelMap"
            />
            <WebhookMockup v-else-if="selected.channel === 'webhook'" :selected="selected" />
            <InAppMockup v-else :selected="selected" />
          </div>
        </div>

        <!-- Attempts list as a timeline -->
        <AttemptTimeline :attempts="attempts" />
      </div>
    </Dialog>
  </section>
</template>
