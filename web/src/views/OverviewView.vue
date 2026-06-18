<script setup lang="ts">
import Button from "primevue/button";
import { computed, onMounted, ref } from "vue";
import { api } from "@/lib/api";
import type { NotificationRecord } from "@/lib/types";
import StatCard from "@/components/dashboard/StatCard.vue";
import DeliveryChart from "@/components/dashboard/DeliveryChart.vue";
import RecentActivityTable from "@/components/dashboard/RecentActivityTable.vue";
import { useI18nStore } from "@/stores/i18n";

const i18n = useI18nStore();

const loading = ref(false);
const error = ref("");
const notifications = ref<NotificationRecord[]>([]);

const totals = computed(() => {
  const result = { queued: 0, delivered: 0, failed: 0, total: notifications.value.length };
  notifications.value.forEach((item) => {
    if (item.status === "queued" || item.status === "delivered" || item.status === "failed") {
      result[item.status] += 1;
    }
  });
  return result;
});

const statusCards = computed(() => {
  const total = totals.value.total;
  return [
    {
      label: i18n.t("overview.kpi_total"),
      value: total,
      subtext: i18n.t("overview.kpi_total_desc"),
      icon: "pi pi-send",
      tone: "bg-blue-50 text-blue-600 dark:bg-blue-950/30 dark:text-blue-400",
      progress: 100,
      accent: "from-blue-500 to-cyan-400",
    },
    {
      label: i18n.t("overview.kpi_delivered"),
      value: totals.value.delivered,
      subtext: i18n.t("overview.kpi_delivered_desc", {
        pct: total ? Math.round((totals.value.delivered / total) * 100) : 0,
      }),
      icon: "pi pi-check-circle",
      tone: "bg-emerald-50 text-emerald-600 dark:bg-emerald-950/30 dark:text-emerald-400",
      progress: total ? (totals.value.delivered / total) * 100 : 0,
      accent: "from-emerald-500 to-teal-400",
    },
    {
      label: i18n.t("overview.kpi_queued"),
      value: totals.value.queued,
      subtext: i18n.t("overview.kpi_queued_desc"),
      icon: "pi pi-clock",
      tone: "bg-amber-50 text-amber-600 dark:bg-amber-950/30 dark:text-amber-400",
      progress: total ? (totals.value.queued / total) * 100 : 0,
      accent: "from-amber-500 to-orange-400",
    },
    {
      label: i18n.t("overview.kpi_failed"),
      value: totals.value.failed,
      subtext: i18n.t("overview.kpi_failed_desc"),
      icon: "pi pi-exclamation-triangle",
      tone: "bg-rose-50 text-rose-600 dark:bg-rose-950/30 dark:text-rose-400",
      progress: total ? (totals.value.failed / total) * 100 : 0,
      accent: "from-rose-500 to-red-400",
    },
  ];
});

// Chart.js Configuration
const chartData = computed(() => {
  const months = Array.from({ length: 12 }, (_, i) => {
    const d = new Date();
    d.setMonth(d.getMonth() - 11 + i);
    return d.toLocaleDateString(i18n.locale, { month: "short" });
  });
  const deliveredData = Array.from({ length: 12 }, () => 0);
  const queuedData = Array.from({ length: 12 }, () => 0);
  const failedData = Array.from({ length: 12 }, () => 0);

  const isDark = document.documentElement.classList.contains("dark");
  const deliveredColor = isDark ? "oklch(0.72 0.16 150)" : "oklch(0.65 0.16 150)";

  const now = new Date();
  const startMonth = new Date(now.getFullYear(), now.getMonth() - 11, 1);

  notifications.value.forEach((item) => {
    const date = new Date(item.created_at);
    const diffMonths = (date.getFullYear() - startMonth.getFullYear()) * 12 + (date.getMonth() - startMonth.getMonth());
    if (diffMonths >= 0 && diffMonths < 12) {
      if (item.status === "delivered")
        deliveredData[diffMonths] = (deliveredData[diffMonths] ?? 0) + 1;
      else if (item.status === "queued")
        queuedData[diffMonths] = (queuedData[diffMonths] ?? 0) + 1;
      else if (item.status === "failed")
        failedData[diffMonths] = (failedData[diffMonths] ?? 0) + 1;
    }
  });

  return {
    labels: months,
    datasets: [
      {
        label: i18n.t("overview.chart_delivered"),
        backgroundColor: deliveredColor,
        borderRadius: 6,
        data: deliveredData,
      },
      {
        label: i18n.t("overview.chart_queued"),
        backgroundColor: "#f59e0b",
        borderRadius: 6,
        data: queuedData,
      },
      {
        label: i18n.t("overview.chart_failed"),
        backgroundColor: "#f43f5e",
        borderRadius: 6,
        data: failedData,
      },
    ],
  };
});

const chartOptions = computed(() => {
  const isDark = document.documentElement.classList.contains("dark");
  const gridColor = isDark ? "rgba(255, 255, 255, 0.06)" : "rgba(15, 23, 42, 0.05)";
  const textColor = isDark ? "#94a3b8" : "#64748b";

  return {
    maintainAspectRatio: false,
    plugins: {
      legend: {
        labels: {
          color: textColor,
          font: { family: "Inter", weight: "500" },
          boxWidth: 12,
          usePointStyle: true,
          pointStyle: "circle",
        },
      },
      tooltip: {
        padding: 12,
        backgroundColor: isDark ? "#0f172a" : "#ffffff",
        titleColor: isDark ? "#ffffff" : "#0f172a",
        bodyColor: isDark ? "#cbd5e1" : "#475569",
        borderColor: isDark ? "#334155" : "#e2e8f0",
        borderWidth: 1,
        usePointStyle: true,
      },
    },
    scales: {
      x: {
        stacked: true,
        grid: { display: false },
        ticks: { color: textColor, font: { family: "Inter" } },
      },
      y: {
        stacked: true,
        grid: { color: gridColor },
        ticks: { color: textColor, font: { family: "Inter" } },
      },
    },
  };
});

async function load() {
  loading.value = true;
  error.value = "";
  try {
    notifications.value = await api.listNotifications({ limit: 50 });
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <section class="space-y-6">
    <!-- Page Header -->
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h1 class="mt-1 text-3xl font-bold text-surface-900 dark:text-white">{{ i18n.t("overview.title") }}</h1>
        <p class="mt-1 text-sm text-surface-500 dark:text-surface-400">
          {{ i18n.t("overview.desc") }}
        </p>
      </div>
      <Button
        icon="pi pi-refresh"
        :label="i18n.t('overview.refresh')"
        :loading="loading"
        @click="load"
        class="rounded-xl px-4 h-11"
      />
    </div>

    <!-- KPI Metrics -->
    <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <StatCard v-for="card in statusCards" :key="card.label" v-bind="card" />
    </div>

    <!-- Trend Chart -->
    <DeliveryChart :chart-data="chartData" :chart-options="chartOptions" />

    <!-- Secondary Grid: Recent Notifications -->
    <div class="w-full">
      <RecentActivityTable :notifications="notifications.slice(0, 6)" :loading="loading" />
    </div>
  </section>
</template>
