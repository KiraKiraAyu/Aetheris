<script setup lang="ts">
import Card from "primevue/card";

defineProps<{
  label: string;
  value: number | string;
  subtext: string;
  icon: string;
  tone: string;
  progress: number;
  accent: string;
}>();
</script>

<template>
  <Card
    class="overflow-hidden border border-surface-200 dark:border-surface-800 bg-surface-0 dark:bg-surface-900 shadow-none!"
  >
    <template #content>
      <div class="flex items-start justify-between gap-3">
        <div>
          <div
            class="text-xs font-bold uppercase tracking-wider text-surface-400 dark:text-surface-500"
          >
            {{ label }}
          </div>
          <div class="mt-2 text-3xl font-bold text-surface-900 dark:text-white">
            {{ value }}
          </div>
          <div class="mt-1 text-[11px] text-surface-500 dark:text-surface-400">
            {{ subtext }}
          </div>
        </div>
        <div
          :class="['grid h-12 w-12 place-items-center rounded-2xl dark:border-surface-800', tone]"
        >
          <span :class="icon" class="text-lg"></span>
        </div>
      </div>
      <!-- Progress bar -->
      <div class="mt-5 h-1.5 overflow-hidden rounded-full bg-surface-100 dark:bg-surface-800">
        <div
          class="h-full rounded-full bg-linear-to-r transition-all duration-1000"
          :class="accent"
          :style="{ width: `${progress}%` }"
        ></div>
      </div>
    </template>
  </Card>
</template>
