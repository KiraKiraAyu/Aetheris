<script setup lang="ts">
import { reactive, watch, ref, onMounted } from "vue";
import InputText from "primevue/inputtext";
import { useI18nStore } from "@/stores/i18n";

const props = defineProps<{
  modelValue: {
    tg_bot_token: string;
    tg_api_base_url: string;
    tg_parse_mode: string;
    tg_disable_link: boolean;
    tg_timeout_seconds: number;
    tg_headers: string;
    tg_body_template: string;
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
  validateHeaders(localForm.tg_headers);
});
</script>

<template>
  <div class="space-y-3">
    <div class="space-y-1">
      <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
        {{ i18n.t("channel_dialog.tg_bot_token") }}
      </label>
      <InputText
        v-model="localForm.tg_bot_token"
        placeholder="0000000000:AAxxxxxxxxx..."
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
      <div class="space-y-1">
        <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
          {{ i18n.t("channel_dialog.tg_api_url") }}
        </label>
        <InputText
          v-model="localForm.tg_api_base_url"
          placeholder="https://api.telegram.org (Default)"
          class="h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full text-xs rounded-xl"
        />
      </div>

      <div class="grid grid-cols-2 gap-3">
        <div class="space-y-1">
          <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
            {{ i18n.t("channel_dialog.tg_parse_mode") }}
          </label>
          <select
            v-model="localForm.tg_parse_mode"
            class="h-10 px-3 border border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full text-xs rounded-xl dark:text-white"
          >
            <option value="HTML">{{ i18n.t("channel_dialog.tg_parse_html") }}</option>
            <option value="Markdown">{{ i18n.t("channel_dialog.tg_parse_markdown") }}</option>
            <option value="MarkdownV2">{{ i18n.t("channel_dialog.tg_parse_markdown_v2") }}</option>
            <option value="">{{ i18n.t("channel_dialog.tg_parse_plain") }}</option>
          </select>
        </div>
        <div class="space-y-1">
          <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
            {{ i18n.t("channel_dialog.tg_timeout") }}
          </label>
          <input
            v-model.number="localForm.tg_timeout_seconds"
            type="number"
            class="h-10 px-3 border border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full text-xs rounded-xl dark:text-white"
          />
        </div>
      </div>

      <div class="flex items-center gap-2 p-1">
        <input
          type="checkbox"
          v-model="localForm.tg_disable_link"
          id="tg_disable_link"
          class="rounded border-surface-300 text-reisa-lilac-600 focus:ring-reisa-lilac-500"
        />
        <label for="tg_disable_link" class="text-xs font-semibold text-surface-600 dark:text-surface-400 cursor-pointer">
          {{ i18n.t("channel_dialog.tg_link_preview") }}
        </label>
      </div>

      <div class="space-y-1">
        <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
          {{ i18n.t("channel_dialog.tg_body_template") }}
        </label>
        <textarea
          v-model="localForm.tg_body_template"
          placeholder='{"chat_id":"{{.Recipient}}","text":"*{{.Title}}*\n{{.Body}}","parse_mode":"HTML"}'
          rows="3"
          class="font-mono text-[11px] p-3 border border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full rounded-xl dark:text-white focus:outline-none"
        ></textarea>
      </div>

      <div class="space-y-1">
        <div class="flex justify-between items-center">
          <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
            {{ i18n.t("channel_dialog.tg_headers") }}
          </label>
          <span v-if="headersError" class="text-[9px] text-rose-500 font-bold uppercase tracking-wider">{{
            headersError
          }}</span>
        </div>
        <textarea
          v-model="localForm.tg_headers"
          rows="3"
          @input="validateHeaders(localForm.tg_headers)"
          class="font-mono text-[11px] p-3 border border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full rounded-xl dark:text-white focus:outline-none"
          :class="{ 'border-rose-400': headersError }"
        ></textarea>
      </div>
    </div>
  </div>
</template>
