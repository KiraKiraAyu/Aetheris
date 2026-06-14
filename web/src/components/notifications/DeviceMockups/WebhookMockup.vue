<script setup lang="ts">
import type { NotificationRecord } from "@/lib/types";

defineProps<{
  selected: NotificationRecord;
}>();
</script>

<template>
  <div
    class="w-full max-w-155 bg-surface-950 border border-surface-850 rounded-xl overflow-hidden font-mono text-[11px] text-indigo-300 leading-normal"
  >
    <div
      class="p-2 bg-surface-900 border-b border-surface-850 text-surface-400 flex items-center justify-between"
    >
      <span>POST {{ selected.recipient }}</span>
      <span
        class="bg-reisa-lilac-900/40 text-reisa-lilac-400 px-1.5 py-0.5 rounded text-[9px] font-bold border border-reisa-lilac-800/50"
        >200 OK</span
      >
    </div>
    <pre class="p-4 overflow-x-auto text-surface-300">
{
 "event": "notification.delivered",
 "id": "{{ selected.id }}",
 "channel": "{{ selected.channel }}",
 "recipient": "{{ selected.recipient }}",
 "title": "{{ selected.title }}",
 "body": "{{ selected.body }}",
 "aggregate_count": {{ selected.aggregate_count }}
}
    </pre>
  </div>
</template>
