<script setup lang="ts">
import Toast from "primevue/toast";
import { ref, onMounted, watch, computed } from "vue";
import { RouterView, RouterLink, useRoute } from "vue-router";
import { useSettingsStore } from "@/stores/settings";
import AppSidebar from "@/components/layout/AppSidebar.vue";
import { useToast } from "primevue/usetoast";
import { useI18nStore } from "@/stores/i18n";

const settings = useSettingsStore();
const route = useRoute();
const toast = useToast();
const i18n = useI18nStore();
const isDark = ref(false);

watch(
  () => settings.toastEvent,
  (evt) => {
    if (evt) {
      toast.add(evt);
      settings.toastEvent = null; // Clear event after showing
    }
  }
);

const nav = computed(() => [
  { label: i18n.t("nav.overview"), to: "/", icon: "pi pi-home" },
  { label: i18n.t("nav.notifications"), to: "/notifications", icon: "pi pi-send" },
  { label: i18n.t("nav.inbox"), to: "/inbox", icon: "pi pi-inbox" },
  { label: i18n.t("nav.templates"), to: "/templates", icon: "pi pi-th-large" },
  { label: i18n.t("nav.settings"), to: "/settings", icon: "pi pi-cog" },
]);

function toggleDarkMode() {
  isDark.value = !isDark.value;
  if (isDark.value) {
    document.documentElement.classList.add("dark");
    localStorage.setItem("aetheris.theme", "dark");
  } else {
    document.documentElement.classList.remove("dark");
    localStorage.setItem("aetheris.theme", "light");
  }
}

onMounted(() => {
  const savedTheme = localStorage.getItem("aetheris.theme");
  const prefersDark =
    window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches;
  if (savedTheme === "dark" || (!savedTheme && prefersDark)) {
    isDark.value = true;
    document.documentElement.classList.add("dark");
  } else {
    isDark.value = false;
    document.documentElement.classList.remove("dark");
  }
  settings.checkConnection();
});
</script>

<template>
  <Toast />
  <!-- Bottom nav is 4.5rem tall on mobile, pb-22 avoids overlaps -->
  <div class="min-h-screen p-3 md:p-5 pb-22 md:pb-5 transition-colors duration-300">
    <AppSidebar
      :is-dark="isDark"
      @toggle-theme="toggleDarkMode"
    />

    <div class="lg:pl-29">
      <main class="py-3 md:py-6">
        <RouterView />
      </main>
    </div>

    <!-- Floating Bottom Navigation Bar for Mobile & Tablet (hidden on lg screens) -->
    <nav
      class="fixed bottom-3 inset-x-3 z-30 lg:hidden bg-surface-0/90 dark:bg-surface-900/90 backdrop-blur-md border border-surface-200 dark:border-surface-800 rounded-2xl flex justify-around py-3 px-4 shadow-xl safe-bottom transition-all duration-300"
    >
      <RouterLink
        v-for="item in nav"
        :key="item.to"
        :to="item.to"
        class="flex flex-col items-center gap-1.5 text-[10px] font-semibold text-surface-500 dark:text-surface-400 py-1 transition-all"
        :class="{ 'text-primary scale-105 font-bold': route.path === item.to }"
      >
        <span :class="item.icon" class="text-lg"></span>
        <span>{{ item.label }}</span>
      </RouterLink>

      <!-- Connection Failure Icon on Mobile -->
      <div
        v-if="settings.connectionStatus === 'disconnected'"
        class="flex flex-col items-center gap-1 text-[10px] font-semibold text-red-500 py-1 cursor-pointer animate-pulse"
        :title="i18n.t('settings.conn_failed') + ': ' + settings.connectionError"
      >
        <span class="pi pi-exclamation-triangle text-base"></span>
        <span>{{ i18n.t("nav.offline") }}</span>
      </div>
    </nav>
  </div>
</template>
