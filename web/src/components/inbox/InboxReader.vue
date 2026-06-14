<script setup lang="ts">
import Button from "primevue/button";
import { formatDate } from "@/lib/api";
import type { InAppMessage } from "@/lib/types";
import { useI18nStore } from "@/stores/i18n";

defineProps<{
  selectedMessage: InAppMessage | null;
}>();

const emit = defineEmits<{
  (e: "deselect"): void;
  (e: "mark-read", message: InAppMessage): void;
}>();

const i18n = useI18nStore();
</script>

<template>
  <div
    :class="[
      ' flex flex-col border border-surface-200 dark:border-surface-800 bg-surface-0 dark:bg-surface-900 rounded-2xl overflow-hidden',
      !selectedMessage ? 'hidden md:flex' : 'flex',
    ]"
  >
    <!-- Selected Message Reader -->
    <div v-if="selectedMessage" class="flex-1 flex flex-col overflow-hidden">
      <!-- Reader Topbar -->
      <div
        class="p-4 border-b border-surface-100 dark:border-surface-800/80 flex items-center justify-between bg-surface-50/50 dark:bg-surface-950/20"
      >
        <div class="flex items-center gap-2">
          <!-- Back button on Mobile -->
          <Button
            icon="pi pi-arrow-left"
            class="md:hidden h-9 w-9 rounded-xl border border-surface-200 dark:border-surface-800"
            outlined
            severity="secondary"
            @click="emit('deselect')"
          />
          <div>
            <div class="text-xs text-surface-400 dark:text-surface-500">
              {{ i18n.t("in_app.reader_msg_id") }} <span class="font-mono">{{ selectedMessage.id }}</span>
            </div>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <!-- Mark read/unread badge -->
          <span
            class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider border"
            :class="
              selectedMessage.read_at
                ? 'bg-surface-100 text-surface-600 dark:bg-surface-800 dark:text-surface-400 border-surface-200/50 dark:border-surface-700/50'
                : 'bg-indigo-50 text-indigo-700 dark:bg-indigo-950/30 dark:text-indigo-400 border-indigo-200/40 dark:border-indigo-800/40'
            "
          >
            {{ selectedMessage.read_at ? i18n.t("notifications.details_in_app_read_yes") : i18n.t("notifications.details_in_app_read_no") }}
          </span>

          <Button
            v-if="!selectedMessage.read_at"
            size="small"
            icon="pi pi-check"
            :label="i18n.t('in_app.btn_mark_read')"
            class="rounded-xl px-3 h-9 text-xs"
            @click="emit('mark-read', selectedMessage)"
          />
        </div>
      </div>

      <!-- Reader Scrollable Content -->
      <div class="flex-1 overflow-y-auto p-6 space-y-6">
        <!-- Subject and Details -->
        <div class="space-y-3">
          <h2 class="text-xl font-bold text-surface-900 dark:text-white leading-snug">
            {{ selectedMessage.title }}
          </h2>

          <div
            class="flex flex-wrap items-center gap-4 text-xs text-surface-500 dark:text-surface-400 border-t border-b border-surface-50 dark:border-surface-800/50 py-3 mt-4"
          >
            <div class="flex items-center gap-1.5">
              <span class="pi pi-user text-surface-400"></span>
              <span
                >{{ i18n.t("in_app.reader_user") }}
                <strong class="font-mono text-surface-700 dark:text-surface-300">{{
                  selectedMessage.user_id
                }}</strong></span
              >
            </div>
            <div class="flex items-center gap-1.5">
              <span class="pi pi-clock text-surface-400"></span>
              <span
                >{{ i18n.t("in_app.reader_received") }} <strong>{{ formatDate(selectedMessage.created_at) }}</strong></span
              >
            </div>
            <div v-if="selectedMessage.read_at" class="flex items-center gap-1.5">
              <span class="pi pi-eye text-surface-400"></span>
              <span
                >{{ i18n.t("in_app.reader_read_at") }} <strong>{{ formatDate(selectedMessage.read_at) }}</strong></span
              >
            </div>
          </div>
        </div>

        <!-- Message Body Render -->
        <div class="space-y-2">
          <div
            class="text-xs font-bold uppercase tracking-wider text-surface-400 dark:text-surface-500"
          >
            {{ i18n.t("in_app.reader_content") }}
          </div>
          <div
            class="p-5 rounded-xl border border-surface-100 dark:border-surface-800 bg-surface-50/30 dark:bg-surface-950/10 text-sm text-surface-800 dark:text-surface-200 leading-relaxed whitespace-pre-wrap font-sans"
          >
            {{ selectedMessage.body }}
          </div>
        </div>

        <!-- Metadata Key-Value Badges -->
        <div
          v-if="selectedMessage.metadata && Object.keys(selectedMessage.metadata).length"
          class="space-y-2"
        >
          <div
            class="text-xs font-bold uppercase tracking-wider text-surface-400 dark:text-surface-500"
          >
            {{ i18n.t("in_app.reader_metadata") }}
          </div>
          <div class="flex flex-wrap gap-2">
            <div
              v-for="(val, key) in selectedMessage.metadata"
              :key="key"
              class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs bg-surface-50 dark:bg-surface-950 border border-surface-200/40 dark:border-surface-800 font-mono text-surface-700 dark:text-surface-300"
            >
              <span class="text-surface-400 dark:text-surface-500 font-semibold">{{ key }}:</span>
              <span>{{ val }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty State Reader -->
    <div
      v-else
      class="flex-1 flex flex-col items-center justify-center p-6 text-center text-surface-400 dark:text-surface-500 select-none"
    >
      <div
        class="grid h-16 w-16 place-items-center rounded-full bg-surface-50 dark:bg-surface-800 text-surface-300 dark:text-surface-700 mb-4 animate-bounce"
      >
        <span class="pi pi-envelope-open text-2xl"></span>
      </div>
      <h3 class="text-base font-bold text-surface-700 dark:text-surface-300">
        {{ i18n.t("in_app.reader_empty_title") }}
      </h3>
      <p class="text-xs text-surface-400 dark:text-surface-500 mt-1 max-w-70">
        {{ i18n.t("in_app.reader_empty_desc") }}
      </p>
    </div>
  </div>
</template>
