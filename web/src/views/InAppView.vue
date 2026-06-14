<script setup lang="ts">
import Button from "primevue/button";
import { onMounted, reactive, ref, watch } from "vue";
import { api } from "@/lib/api";
import type { InAppMessage } from "@/lib/types";
import { useToast } from "primevue/usetoast";
import InboxSidebar from "@/components/inbox/InboxSidebar.vue";
import InboxReader from "@/components/inbox/InboxReader.vue";
import { useI18nStore } from "@/stores/i18n";

const toast = useToast();
const i18n = useI18nStore();
const messages = ref<InAppMessage[]>([]);
const loading = ref(false);
const error = ref("");
const filters = reactive({
  user_id: "",
  unread: true,
});

const selectedMessage = ref<InAppMessage | null>(null);

let debounceTimer: ReturnType<typeof setTimeout> | null = null;

// Auto-trigger search with a 350ms debounce when user ID is being typed
watch(
  () => filters.user_id,
  () => {
    if (debounceTimer) clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      load();
    }, 350);
  }
);

// Immediately reload when unread filter toggles
watch(
  () => filters.unread,
  () => {
    if (debounceTimer) clearTimeout(debounceTimer);
    load();
  }
);

async function load() {
  if (debounceTimer) {
    clearTimeout(debounceTimer);
    debounceTimer = null;
  }
  if (!filters.user_id.trim()) {
    messages.value = [];
    selectedMessage.value = null;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    messages.value = await api.listInApp({ ...filters, limit: 100 });
    if (selectedMessage.value) {
      const found = messages.value.find((m) => m.id === selectedMessage.value?.id);
      if (found) {
        selectedMessage.value = found;
      } else {
        selectedMessage.value = null;
      }
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

async function markRead(message: InAppMessage) {
  if (!filters.user_id && !message.user_id) return;
  try {
    await api.markInAppRead(message.id, filters.user_id || message.user_id);
    toast.add({
      severity: "success",
      summary: i18n.t("in_app.toast_read_title"),
      detail: i18n.t("in_app.toast_read_desc"),
      life: 2000,
    });

    message.read_at = new Date().toISOString();
    if (selectedMessage.value && selectedMessage.value.id === message.id) {
      selectedMessage.value.read_at = message.read_at;
    }

    if (filters.unread) {
      await load();
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  }
}

function selectMessage(msg: InAppMessage) {
  selectedMessage.value = msg;
}

function deselectMessage() {
  selectedMessage.value = null;
}

onMounted(load);
</script>

<template>
  <section class="space-y-5">
    <!-- Header -->
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h1 class="mt-1 text-3xl font-bold text-surface-900 dark:text-white">{{ i18n.t("in_app.title") }}</h1>
        <p class="mt-1 text-sm text-surface-500 dark:text-surface-400">
          {{ i18n.t("in_app.desc") }}
        </p>
      </div>
      <Button
        icon="pi pi-refresh"
        :label="i18n.t('in_app.btn_refresh')"
        :loading="loading"
        @click="load"
        class="rounded-xl px-4 h-11"
      />
    </div>

    <!-- Split Pane Inbox -->
    <div
      class="grid gap-5 md:grid-cols-[360px_1fr] lg:grid-cols-[400px_1fr] h-[calc(100vh-230px)] min-h-145"
    >
      <InboxSidebar
        :messages="messages"
        :loading="loading"
        v-model:filters="filters"
        :selected-message="selectedMessage"
        @load="load"
        @select="selectMessage"
      />

      <InboxReader
        :selected-message="selectedMessage"
        @deselect="deselectMessage"
        @mark-read="markRead"
      />
    </div>
  </section>
</template>
