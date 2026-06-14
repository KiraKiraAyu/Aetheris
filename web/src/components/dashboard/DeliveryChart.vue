<script setup lang="ts">
import Chart from "primevue/chart";

defineProps<{
  chartData: {
    labels: string[];
    datasets: Array<{
      label: string;
      backgroundColor: string;
      borderRadius?: number;
      data: number[];
    }>;
  };
  chartOptions: Record<string, unknown>;
}>();
</script>

<template>
  <div
    class="p-6 bg-surface-0 dark:bg-surface-900 rounded-2xl border border-surface-200 dark:border-surface-800"
  >
    <div class="mb-5 flex items-start justify-between">
      <div>
        <h2 class="text-lg font-bold text-surface-900 dark:text-white">
          Delivery Performance Trend
        </h2>
        <p class="text-xs text-surface-500 dark:text-surface-400">
          Distribution of notifications status across the monthly timeline.
        </p>
      </div>
      <div
        class="h-9 w-9 grid place-items-center rounded-lg bg-surface-50 dark:bg-surface-800 text-surface-400 dark:text-surface-500"
      >
        <span class="pi pi-chart-bar"></span>
      </div>
    </div>
    <div class="h-70">
      <Chart type="bar" :data="chartData" :options="chartOptions" class="h-full" />
    </div>
  </div>
</template>
