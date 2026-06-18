<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { useSettingsStore } from "@/stores/settings";
import { useI18nStore } from "@/stores/i18n";
import { api, getAssetUrl } from "@/lib/api";
import type { ChannelConfig } from "@/lib/types";
import ApiAccessPanel from "@/components/settings/ApiAccessPanel.vue";
import ChannelConfigDialog from "@/components/settings/ChannelConfigDialog.vue";

const settings = useSettingsStore();
const i18n = useI18nStore();

const loadingConfigs = ref(false);
const channelConfigs = ref<Record<string, ChannelConfig>>({});

const displayConfigDialog = ref(false);
const selectedChannel = ref<string | null>(null);

async function fetchConfigs() {
  if (!settings.tenantId || settings.connectionStatus !== "connected") {
    channelConfigs.value = {};
    return;
  }
  loadingConfigs.value = true;
  try {
    const list = await api.listChannelConfigs();
    const map: Record<string, ChannelConfig> = {};
    list.forEach((cfg) => {
      map[cfg.channel] = cfg;
    });
    channelConfigs.value = map;
  } catch (err) {
    console.error("fetch channel configs failed", err);
  } finally {
    loadingConfigs.value = false;
  }
}

function isChannelEnabled(name: string) {
  return channelConfigs.value[name]?.enabled || false;
}

function openConfig(channelName: string) {
  selectedChannel.value = channelName;
  displayConfigDialog.value = true;
}

function onChannelSaved(saved: ChannelConfig) {
  channelConfigs.value[saved.channel] = saved;
}

onMounted(() => {
  fetchConfigs();
});

watch(
  () => settings.tenantId,
  () => {
    fetchConfigs();
  },
);

watch(
  () => settings.connectionStatus,
  (status) => {
    if (status === "connected") {
      fetchConfigs();
    }
  },
);

const channelsList = [
  {
    name: "email",
    icon: "pi pi-envelope",
    color: "text-sky-500 bg-sky-50 dark:bg-sky-950/20 border-sky-100 dark:border-sky-900/30",
    desc: "SMTP server configurations to send transactional and digest emails.",
  },
  {
    name: "sms",
    icon: "pi pi-phone",
    color:
      "text-indigo-500 bg-indigo-50 dark:bg-indigo-950/20 border-indigo-100 dark:border-indigo-900/30",
    desc: "HTTP webhook templates to route direct SMS/text target messages.",
  },
  {
    name: "webhook",
    icon: "pi pi-globe",
    color: "text-teal-500 bg-teal-50 dark:bg-teal-950/20 border-teal-100 dark:border-teal-900/30",
    desc: "Outgoing HTTP POST callbacks with security validation signature headers.",
  },
  {
    name: "in_app",
    icon: "pi pi-inbox",
    color:
      "text-amber-500 bg-amber-50 dark:bg-amber-950/20 border-amber-100 dark:border-amber-900/30",
    desc: "In-app notifications stored directly in database for user inbox queues.",
  },
  {
    name: "telegram",
    icon: "pi pi-telegram",
    color: "text-blue-500 bg-blue-50 dark:bg-blue-950/20 border-blue-100 dark:border-blue-900/30",
    desc: "Deliver alert target updates to Telegram bot client networks.",
  },
  {
    name: "slack",
    icon: "pi pi-slack",
    color:
      "text-purple-500 bg-purple-50 dark:bg-purple-950/20 border-purple-100 dark:border-purple-900/30",
    desc: "Route message payloads to Slack incoming webhook endpoints.",
  },
  {
    name: "discord",
    icon: "pi pi-discord",
    color:
      "text-violet-500 bg-violet-50 dark:bg-violet-950/20 border-violet-100 dark:border-violet-900/30",
    desc: "Push rich card messages directly to Discord webhook server channels.",
  },
  {
    name: "feishu",
    icon: "pi pi-comments",
    color:
      "text-reisa-lilac-500 bg-reisa-lilac-50 dark:bg-reisa-lilac-950/20 border-reisa-lilac-100 dark:border-reisa-lilac-900/30",
    desc: "Post structured messages to Feishu/Lark chatbot webhook URLs.",
  },
  {
    name: "dingtalk",
    icon: "pi pi-comments",
    color: "text-cyan-500 bg-cyan-50 dark:bg-cyan-950/20 border-cyan-100 dark:border-cyan-900/30",
    desc: "Corporate chat alerts with DingTalk robot signature verification.",
  },
  {
    name: "wecom",
    icon: "pi pi-briefcase",
    color: "text-rose-500 bg-rose-50 dark:bg-rose-950/20 border-rose-100 dark:border-rose-900/30",
    desc: "Push alert target payloads to WeCom robot webhook channels.",
  },
];

function getTextColorClass(colorStr: string) {
  return colorStr.split(" ").find((c) => c.startsWith("text-")) || "";
}
</script>

<template>
  <section class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <div class="flex items-center gap-3 mt-1">
          <h1 class="text-3xl font-bold text-surface-900 dark:text-white">
            {{ i18n.t("settings.title") }}
          </h1>
          <span
            class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-md text-xs font-semibold cursor-pointer select-none transition-all duration-300 hover:scale-105"
            :class="{
              'bg-emerald-100 text-emerald-700 border border-emerald-200 dark:bg-emerald-900/30 dark:text-emerald-400 dark:border-emerald-800/40':
                settings.connectionStatus === 'connected',
              'bg-red-100 text-red-700 border border-red-200 dark:bg-red-900/30 dark:text-red-400 dark:border-red-800/40':
                settings.connectionStatus === 'disconnected',
              'bg-surface-100 text-surface-600 border border-surface-200 dark:bg-surface-800/30 dark:text-surface-400 dark:border-surface-700/40':
                settings.connectionStatus === 'checking',
              'bg-orange-100 text-orange-700 border border-orange-200 dark:bg-orange-900/30 dark:text-orange-400 dark:border-orange-800/40':
                settings.connectionStatus === 'unconfigured',
            }"
            @click="settings.checkConnection()"
            :title="settings.connectionError || 'Click to check connection'"
          >
            <span
              class="h-1.5 w-1.5 rounded-full"
              :class="{
                'bg-emerald-500 animate-pulse': settings.connectionStatus === 'connected',
                'bg-red-500 animate-ping': settings.connectionStatus === 'disconnected',
                'bg-surface-400 animate-spin': settings.connectionStatus === 'checking',
                'bg-orange-500 animate-pulse': settings.connectionStatus === 'unconfigured',
              }"
            ></span>
            {{
              settings.connectionStatus === "connected"
                ? i18n.t("settings.status_connected")
                : settings.connectionStatus === "disconnected"
                  ? i18n.t("settings.status_offline")
                  : settings.connectionStatus === "checking"
                    ? i18n.t("settings.status_checking")
                    : i18n.t("settings.status_unconfigured")
            }}
          </span>
        </div>
        <p class="mt-1.5 text-sm text-surface-500 dark:text-surface-400">
          {{ i18n.t("settings.desc") }}
        </p>
      </div>
    </div>

    <div class="grid gap-6 xl:grid-cols-[400px_1fr]">
      <!-- API Access Panel -->
      <ApiAccessPanel @saved="fetchConfigs" />

      <!-- Supported Channels Panel -->
      <div
        class="p-6 bg-white dark:bg-surface-900 rounded-2xl border border-surface-200 dark:border-surface-800"
      >
        <div class="mb-5 flex items-center justify-between">
          <div>
            <h2 class="text-lg font-bold text-surface-900 dark:text-white">
              {{ i18n.t("settings.channels_title") }}
            </h2>
            <p class="text-xs text-surface-500 dark:text-surface-400">
              {{ i18n.t("settings.channels_desc") }}
            </p>
          </div>
          <div
            class="h-9 w-9 grid place-items-center rounded-lg bg-surface-50 dark:bg-surface-800 text-surface-500"
          >
            <span class="pi pi-sitemap text-sm"></span>
          </div>
        </div>

        <div
          v-if="settings.connectionStatus !== 'connected'"
          class="flex flex-col items-center justify-center p-8 text-center bg-surface-50/50 dark:bg-surface-950/20 border border-dashed border-surface-200 dark:border-surface-800 rounded-2xl"
        >
          <span
            class="pi pi-exclamation-triangle text-3xl text-surface-400 dark:text-surface-500 mb-3"
          ></span>
          <h3 class="text-sm font-bold text-surface-700 dark:text-surface-300">
            {{ i18n.t("settings.channels_backend_conn") }}
          </h3>
          <p class="text-xs text-surface-500 dark:text-surface-400 mt-1 max-w-xs">
            {{ i18n.t("settings.channels_backend_conn_desc") }}
          </p>
        </div>

        <div v-else class="grid gap-3 sm:grid-cols-2 lg:grid-cols-2 xl:grid-cols-3">
          <div
            v-for="channel in channelsList"
            :key="channel.name"
            class="flex flex-col p-4 rounded-xl border border-surface-100 dark:border-surface-800 bg-surface-50/30 dark:bg-surface-950/20 hover:border-reisa-lilac-200 dark:hover:border-reisa-lilac-700 hover:bg-surface-0 dark:hover:bg-surface-900 transition-all duration-200 group cursor-pointer hover:shadow-sm"
            @click="openConfig(channel.name)"
          >
            <div class="flex items-center gap-3">
              <div class="grid h-9 w-9 place-items-center">
                <img
                  v-if="
                    ['telegram', 'slack', 'discord', 'feishu', 'dingtalk', 'wecom'].includes(
                      channel.name,
                    )
                  "
                  :src="getAssetUrl(`logos/${channel.name}.svg`)"
                  class="h-8 w-8 object-contain shrink-0"
                  :alt="channel.name"
                />
                <span
                  v-else
                  :class="[channel.icon, getTextColorClass(channel.color)]"
                  style="font-size: 2rem"
                ></span>
              </div>
              <div class="flex-1 min-w-0">
                <div class="text-sm font-semibold text-surface-800 dark:text-surface-200 truncate">
                  {{ i18n.t(`channels.${channel.name}` as any) }}
                </div>
                <div class="flex items-center gap-1 mt-0.5">
                  <span
                    class="h-1.5 w-1.5 rounded-full"
                    :class="
                      isChannelEnabled(channel.name)
                        ? 'bg-emerald-500 animate-pulse'
                        : 'bg-surface-300 dark:bg-surface-700'
                    "
                  ></span>
                  <span
                    class="text-[9px] uppercase tracking-wide font-bold"
                    :class="
                      isChannelEnabled(channel.name)
                        ? 'text-emerald-500'
                        : 'text-surface-400 dark:text-surface-500'
                    "
                  >
                    {{
                      isChannelEnabled(channel.name)
                        ? i18n.t("settings.channels_active")
                        : i18n.t("settings.channels_disabled")
                    }}
                  </span>
                </div>
              </div>
            </div>
            <p
              class="mt-3 text-[11px] text-surface-500 dark:text-surface-400 leading-normal flex-1 group-hover:text-surface-700 dark:group-hover:text-surface-300 transition-colors"
            >
              {{ i18n.t(`channels.${channel.name}_desc` as any) }}
            </p>
          </div>
        </div>

        <div
          class="mt-5 p-3.5 rounded-xl bg-surface-50/80 dark:bg-surface-950/40 border border-surface-100 dark:border-surface-800/80 text-xs text-surface-500 dark:text-surface-400 leading-relaxed flex items-start gap-2.5"
        >
          <span class="pi pi-info-circle text-surface-400 mt-0.5 shrink-0"></span>
          <span>
            {{ i18n.t("settings.security_notice") }}
          </span>
        </div>
      </div>
    </div>

    <!-- Configuration Dialog Modal -->
    <ChannelConfigDialog
      v-model:visible="displayConfigDialog"
      :channel="selectedChannel"
      :channel-config="selectedChannel ? channelConfigs[selectedChannel] : undefined"
      @saved="onChannelSaved"
    />
  </section>
</template>
