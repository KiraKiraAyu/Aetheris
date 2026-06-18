<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import Button from "primevue/button";
import { useSettingsStore } from "@/stores/settings";
import { useI18nStore } from "@/stores/i18n";
import { getAssetUrl } from "@/lib/api";

defineProps<{ isDark: boolean }>();
const emit = defineEmits<{ (e: "toggle-theme"): void }>();

const route = useRoute();
const settings = useSettingsStore();
const i18n = useI18nStore();

const nav = computed(() => [
  { label: i18n.t("nav.overview"), to: "/", icon: "pi pi-home" },
  { label: i18n.t("nav.notifications"), to: "/notifications", icon: "pi pi-send" },
  { label: i18n.t("nav.inbox"), to: "/inbox", icon: "pi pi-inbox" },
  { label: i18n.t("nav.templates"), to: "/templates", icon: "pi pi-th-large" },
  { label: i18n.t("nav.settings"), to: "/settings", icon: "pi pi-cog" },
]);

const isDemoMode = import.meta.env.MODE === 'demo';
</script>

<template>
  <aside
    class="fixed inset-y-5 left-5 hidden w-24 flex-col items-center rounded-2xl bg-surface-50 dark:bg-surface-900 px-3 py-6 lg:flex"
  >
    <div
      class="grid h-12 w-12 place-items-center rounded-xl"
      aria-label="Aetheris"
    >
      <img :src="getAssetUrl('icon.svg')" class="h-10 w-10 object-contain" alt="Aetheris Logo" />
    </div>

    <!-- Demo Mode Badge -->
    <div
      v-if="isDemoMode"
      class="mt-2 text-[9px] font-black uppercase bg-amber-500/10 dark:bg-amber-500/20 text-amber-500 border border-amber-500/20 px-1.5 py-0.5 rounded-md text-center scale-90 whitespace-nowrap select-none animate-pulse"
      title="Demo Sandbox Active"
    >
      DEMO
    </div>

    <nav class="mt-12 grid gap-4">
      <RouterLink
        v-for="item in nav"
        :key="item.to"
        :to="item.to"
        class="nav-icon-link"
        :class="{ 'is-active': route.path === item.to }"
        :aria-label="item.label"
        :title="item.label"
      >
        <span :class="item.icon" class="text-lg"></span>
      </RouterLink>
    </nav>

    <div class="mt-auto flex flex-col gap-4 items-center">
      <!-- Connection Failure Icon -->
      <div
        v-if="settings.connectionStatus === 'disconnected'"
        class="h-12 w-12 flex items-center justify-center rounded-xl text-red-500 bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-900/30 transition-all cursor-pointer animate-pulse"
        :title="i18n.t('settings.conn_failed') + ': ' + settings.connectionError"
      >
        <span class="pi pi-exclamation-triangle text-lg"></span>
      </div>

      <!-- Dark Mode Switcher -->
      <Button
        text
        rounded
        :icon="isDark ? 'pi pi-sun' : 'pi pi-moon'"
        class="h-12 w-12 text-surface-500 hover:bg-surface-100 dark:hover:bg-surface-800 transition-colors"
        @click="emit('toggle-theme')"
        :title="i18n.t('nav.toggle_theme')"
      />
    </div>
  </aside>
</template>
