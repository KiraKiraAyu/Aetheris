<script setup lang="ts">
import { reactive, watch, ref, onMounted } from "vue";
import InputText from "primevue/inputtext";
import { useI18nStore } from "@/stores/i18n";

const props = defineProps<{
  modelValue: {
    email_host: string;
    email_port: number;
    email_username: string;
    email_password: string;
    email_from: string;
    email_tls_mode: string;
    email_timeout_seconds: number;
    email_headers: string;
  };
}>();

const emit = defineEmits<{
  (e: "update:modelValue", value: typeof props.modelValue): void;
  (e: "valid", isValid: boolean): void;
}>();

const i18n = useI18nStore();
const showAdvanced = ref(false);
const headersError = ref("");
const localForm = reactive({ ...props.modelValue });

watch(
  () => props.modelValue,
  (newVal) => {
    Object.assign(localForm, newVal);
  },
  { deep: true },
);

watch(
  localForm,
  (newVal) => {
    emit("update:modelValue", { ...newVal });
  },
  { deep: true },
);

function validateHeaders(val: string) {
  if (!val.trim()) {
    headersError.value = "";
    emit("valid", true);
    return true;
  }
  try {
    const parsed = JSON.parse(val);
    if (typeof parsed !== "object" || parsed === null || Array.isArray(parsed)) {
      headersError.value = "Headers must be a JSON object (key-value pairs)";
      emit("valid", false);
      return false;
    }
    headersError.value = "";
    emit("valid", true);
    return true;
  } catch (e: unknown) {
    const errMsg = e instanceof Error ? e.message : String(e);
    headersError.value = `Invalid JSON: ${errMsg}`;
    emit("valid", false);
    return false;
  }
}

onMounted(() => {
  validateHeaders(localForm.email_headers);
});
</script>

<template>
  <div class="space-y-3">
    <div class="grid grid-cols-3 gap-3">
      <div class="col-span-2 space-y-1">
        <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
          {{ i18n.t("channel_dialog.email_host") }}
        </label>
        <InputText
          v-model="localForm.email_host"
          placeholder="smtp.gmail.com"
          class="h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full text-xs rounded-xl"
        />
      </div>
      <div class="space-y-1">
        <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
          {{ i18n.t("channel_dialog.email_port") }}
        </label>
        <input
          v-model.number="localForm.email_port"
          type="number"
          placeholder="587"
          class="h-10 px-3 border border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full text-xs rounded-xl dark:text-white"
        />
      </div>
    </div>

    <div class="grid grid-cols-2 gap-3">
      <div class="space-y-1">
        <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
          {{ i18n.t("channel_dialog.email_username") }}
        </label>
        <InputText
          v-model="localForm.email_username"
          placeholder="user@example.com"
          class="h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full text-xs rounded-xl"
        />
      </div>
      <div class="space-y-1">
        <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
          {{ i18n.t("channel_dialog.email_password") }}
        </label>
        <InputText
          v-model="localForm.email_password"
          type="password"
          placeholder="••••••••"
          class="h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full text-xs rounded-xl"
        />
      </div>
    </div>

    <div class="space-y-1">
      <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
        {{ i18n.t("channel_dialog.email_from") }}
      </label>
      <InputText
        v-model="localForm.email_from"
        placeholder="Aetheris <noreply@example.com>"
        class="h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full text-xs rounded-xl"
      />
    </div>

    <!-- Advanced Settings toggle -->
    <div class="pt-2">
      <button
        type="button"
        @click="showAdvanced = !showAdvanced"
        class="flex items-center gap-1.5 text-xs font-semibold text-reisa-lilac-500 hover:text-reisa-lilac-600 dark:text-reisa-lilac-400 dark:hover:text-reisa-lilac-300 transition-colors focus:outline-none"
      >
        <span :class="['pi text-[10px]', showAdvanced ? 'pi-chevron-down' : 'pi-chevron-right']"></span>
        <span>{{ i18n.t("channel_dialog.email_advanced") }}</span>
      </button>
    </div>

    <div v-show="showAdvanced" class="space-y-3 pt-2 border-t border-surface-100 dark:border-surface-800/80">
      <div class="grid grid-cols-2 gap-3">
        <div class="space-y-1">
          <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
            {{ i18n.t("channel_dialog.email_tls") }}
          </label>
          <select
            v-model="localForm.email_tls_mode"
            class="h-10 px-3 border border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full text-xs rounded-xl dark:text-white"
          >
            <option value="none">{{ i18n.t("channel_dialog.email_tls_none") }}</option>
            <option value="ssl">{{ i18n.t("channel_dialog.email_tls_ssl") }}</option>
            <option value="starttls">{{ i18n.t("channel_dialog.email_tls_starttls") }}</option>
          </select>
        </div>
        <div class="space-y-1">
          <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
            {{ i18n.t("channel_dialog.email_timeout") }}
          </label>
          <input
            v-model.number="localForm.email_timeout_seconds"
            type="number"
            placeholder="10"
            class="h-10 px-3 border border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full text-xs rounded-xl dark:text-white"
          />
        </div>
      </div>

      <div class="space-y-1">
        <div class="flex justify-between items-center">
          <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
            {{ i18n.t("channel_dialog.email_headers") }}
          </label>
          <span v-if="headersError" class="text-[9px] text-rose-500 font-bold uppercase tracking-wider">{{
            headersError
          }}</span>
        </div>
        <textarea
          v-model="localForm.email_headers"
          rows="4"
          @input="validateHeaders(localForm.email_headers)"
          class="font-mono text-[11px] p-3 border border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full rounded-xl dark:text-white focus:outline-none"
          :class="{ 'border-rose-400': headersError }"
        ></textarea>
      </div>
    </div>
  </div>
</template>
