<script setup lang="ts">
import { computed } from "vue";
import Button from "primevue/button";
import InputText from "primevue/inputtext";
import { formatDate } from "@/lib/api";
import type { InAppMessage } from "@/lib/types";
import { useI18nStore } from "@/stores/i18n";

const props = defineProps<{
  messages: InAppMessage[];
  loading: boolean;
  filters: { user_id: string; unread: boolean };
  selectedMessage: InAppMessage | null;
}>();

const i18n = useI18nStore();

const emit = defineEmits<{
  (e: "update:filters", value: { user_id: string; unread: boolean }): void;
  (e: "load"): void;
  (e: "select", message: InAppMessage): void;
}>();

const localFilters = computed({
  get: () => props.filters,
  set: (val) => emit("update:filters", val),
});

function toggleUnread() {
  localFilters.value.unread = !localFilters.value.unread;
}
</script>

<template>
  <div
    :class="[
      ' p-4 flex flex-col border border-surface-200 dark:border-surface-800 bg-surface-0 dark:bg-surface-900 rounded-2xl overflow-hidden',
      selectedMessage ? 'hidden md:flex' : 'flex',
    ]"
  >
    <!-- Inbox Filters -->
    <div class="space-y-3 pb-4 border-b border-surface-100 dark:border-surface-800/80">
      <div class="relative w-full">
        <span
          class="pi pi-user absolute right-3 top-1/2 -translate-y-1/2 text-surface-400 dark:text-surface-500 text-sm"
        ></span>
        <InputText
          v-model="localFilters.user_id"
          :placeholder="i18n.t('in_app.filter_user')"
          class="pr-9 h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl text-sm"
          @keyup.enter="emit('load')"
        />
      </div>
      <Button
        :icon="localFilters.unread ? 'pi pi-eye-slash' : 'pi pi-eye'"
        :label="localFilters.unread ? i18n.t('in_app.filter_unread') : i18n.t('in_app.filter_all')"
        class="w-full text-xs rounded-xl h-10 border-surface-200 dark:border-surface-800 text-surface-600 dark:text-surface-300"
        outlined
        @click="toggleUnread"
      />
    </div>

    <!-- Scrollable Feed -->
    <div class="flex-1 overflow-y-auto mt-3 space-y-2 pr-1 scrollbar-thin">
      <div
        v-for="msg in messages"
        :key="msg.id"
        @click="emit('select', msg)"
        class="p-3.5 rounded-xl border border-surface-100 dark:border-surface-800/60 bg-surface-50/20 dark:bg-surface-950/10 cursor-pointer transition-all hover:bg-surface-50 dark:hover:bg-surface-800/40 hover:border-surface-200 dark:hover:border-surface-700 flex items-start gap-3 select-none"
        :class="{
          'border-indigo-200 dark:border-indigo-800 bg-indigo-50/10 dark:bg-indigo-950/15':
            selectedMessage?.id === msg.id,
          'font-semibold': !msg.read_at,
        }"
      >
        <!-- Unread Status Dot -->
        <div
          class="h-2 w-2 rounded-full mt-1.5 shrink-0"
          :class="
            msg.read_at ? 'bg-surface-200 dark:bg-surface-700' : 'bg-indigo-500 animate-pulse'
          "
        ></div>
        <div class="flex-1 min-w-0">
          <div class="flex justify-between items-start gap-1">
            <span
              class="text-xs text-surface-400 dark:text-surface-500 font-mono truncate max-w-30"
              :title="msg.user_id"
              >{{ msg.user_id }}</span
            >
            <span class="text-[10px] text-surface-400 dark:text-surface-500 whitespace-nowrap">{{
              formatDate(msg.created_at)
            }}</span>
          </div>
          <h3
            class="text-sm text-surface-800 dark:text-surface-200 mt-1 truncate"
            :class="{ 'text-surface-950 dark:text-white font-bold': !msg.read_at }"
          >
            {{ msg.title }}
          </h3>
          <p
            class="text-xs text-surface-500 dark:text-surface-400 mt-0.5 line-clamp-1 leading-relaxed"
          >
            {{ msg.body }}
          </p>
        </div>
      </div>

      <!-- Empty State Feed -->
      <div
        v-if="!messages.length && !loading"
        class="py-12 text-center text-surface-400 dark:text-surface-500"
      >
        {{
          !filters.user_id.trim()
            ? i18n.t("in_app.enter_user_id")
            : i18n.t("in_app.no_records")
        }}
      </div>

      <!-- Loading Skeletons -->
      <div v-if="loading" class="space-y-2 mt-1">
        <div
          v-for="n in 3"
          :key="n"
          class="p-3.5 border border-surface-100 dark:border-surface-800/60 rounded-xl animate-pulse space-y-2"
        >
          <div class="h-3 bg-surface-200 dark:bg-surface-800 rounded w-1/4"></div>
          <div class="h-4 bg-surface-200 dark:bg-surface-800 rounded w-3/4"></div>
          <div class="h-3 bg-surface-200 dark:bg-surface-800 rounded w-5/6"></div>
        </div>
      </div>
    </div>
  </div>
</template>
