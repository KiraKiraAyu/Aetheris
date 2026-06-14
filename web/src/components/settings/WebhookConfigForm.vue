<script setup lang="ts">
import { reactive, watch, ref, onMounted } from "vue";
import InputText from "primevue/inputtext";
import { useI18nStore } from "@/stores/i18n";

const props = defineProps<{
  modelValue: {
    http_url_template: string;
    http_method: string;
    http_headers: string;
    http_body_template: string;
    http_timeout_seconds: number;
    http_success_status_min: number;
    http_success_status_max: number;
    http_response_id_header: string;
    http_response_id_json_field: string;
    http_allowed_hosts: string;
    http_allow_private_ips: boolean;
    http_signing_secret: string;
  };
  channel: string;
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
  validateHeaders(localForm.http_headers);
});
</script>

<template>
  <div class="space-y-3">
    <div class="space-y-1">
      <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
        {{ ["sms", "webhook"].includes(props.channel) ? i18n.t("channel_dialog.webhook_sms_url") : i18n.t("channel_dialog.webhook_url") }}
      </label>
      <InputText
        v-model="localForm.http_url_template"
        :placeholder="
          ['slack', 'discord', 'feishu', 'dingtalk', 'wecom'].includes(props.channel)
            ? 'https://oapi.webhook.com/...'
            : 'https://api.sms-gateway.com/send?to={recipient}'
        "
        class="h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full text-xs rounded-xl"
      />
      <span class="text-[10px] text-surface-400 dark:text-surface-500" v-if="['sms', 'webhook'].includes(props.channel)">
        {{ i18n.t("channel_dialog.webhook_help") }}
      </span>
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
      <!-- Advanced webhook properties -->
      <div class="space-y-3" v-if="['sms', 'webhook'].includes(props.channel)">
        <div class="grid grid-cols-3 gap-3">
          <div class="space-y-1">
            <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
              {{ i18n.t("channel_dialog.webhook_method") }}
            </label>
            <select
              v-model="localForm.http_method"
              class="h-10 px-3 border border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full text-xs rounded-xl dark:text-white"
            >
              <option value="POST">POST</option>
              <option value="PUT">PUT</option>
              <option value="GET">GET</option>
            </select>
          </div>
          <div class="space-y-1">
            <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
              {{ i18n.t("channel_dialog.webhook_min_success") }}
            </label>
            <input
              v-model.number="localForm.http_success_status_min"
              type="number"
              class="h-10 px-3 border border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full text-xs rounded-xl dark:text-white"
            />
          </div>
          <div class="space-y-1">
            <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
              {{ i18n.t("channel_dialog.webhook_max_success") }}
            </label>
            <input
              v-model.number="localForm.http_success_status_max"
              type="number"
              class="h-10 px-3 border border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full text-xs rounded-xl dark:text-white"
            />
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div class="space-y-1">
            <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
              {{ i18n.t("channel_dialog.webhook_resp_id_header") }}
            </label>
            <InputText
              v-model="localForm.http_response_id_header"
              placeholder="X-Message-ID"
              class="h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full text-xs rounded-xl"
            />
          </div>
          <div class="space-y-1">
            <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
              {{ i18n.t("channel_dialog.webhook_resp_id_json") }}
            </label>
            <InputText
              v-model="localForm.http_response_id_json_field"
              placeholder="data.id"
              class="h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full text-xs rounded-xl"
            />
          </div>
        </div>

        <div class="space-y-1">
          <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
            {{ i18n.t("channel_dialog.webhook_body") }}
          </label>
          <textarea
            v-model="localForm.http_body_template"
            placeholder='{"to":"{{.Recipient}}","text":{{quote .Body}}}'
            rows="3"
            class="font-mono text-[11px] p-3 border border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full rounded-xl dark:text-white focus:outline-none"
          ></textarea>
        </div>
      </div>

      <!-- Specific Webhook whitelist domain properties -->
      <div class="space-y-3" v-if="props.channel === 'webhook'">
        <div class="space-y-1">
          <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
            {{ i18n.t("channel_dialog.webhook_hosts") }}
          </label>
          <InputText
            v-model="localForm.http_allowed_hosts"
            placeholder="api.company.com, *.partners.org"
            class="h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full text-xs rounded-xl"
          />
        </div>

        <div class="flex items-center gap-2 p-1">
          <input
            type="checkbox"
            v-model="localForm.http_allow_private_ips"
            id="http_allow_private_ips"
            class="rounded border-surface-300 text-reisa-lilac-600 focus:ring-reisa-lilac-500"
          />
          <label
            for="http_allow_private_ips"
            class="text-xs font-semibold text-surface-600 dark:text-surface-400 cursor-pointer"
          >
            {{ i18n.t("channel_dialog.webhook_private_ips") }}
          </label>
        </div>
      </div>

      <div class="grid grid-cols-2 gap-3">
        <div class="space-y-1">
          <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
            {{ i18n.t("channel_dialog.webhook_signing_secret") }}
          </label>
          <InputText
            v-model="localForm.http_signing_secret"
            type="password"
            placeholder="Generate signature key header"
            class="h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full text-xs rounded-xl"
          />
        </div>
        <div class="space-y-1">
          <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
            {{ i18n.t("channel_dialog.tg_timeout") }}
          </label>
          <input
            v-model.number="localForm.http_timeout_seconds"
            type="number"
            class="h-10 px-3 border border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full text-xs rounded-xl dark:text-white"
          />
        </div>
      </div>

      <div class="space-y-1">
        <div class="flex justify-between items-center">
          <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">
            {{ i18n.t("channel_dialog.webhook_headers") }}
          </label>
          <span v-if="headersError" class="text-[9px] text-rose-500 font-bold uppercase tracking-wider">{{
            headersError
          }}</span>
        </div>
        <textarea
          v-model="localForm.http_headers"
          rows="4"
          @input="validateHeaders(localForm.http_headers)"
          class="font-mono text-[11px] p-3 border border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 w-full rounded-xl dark:text-white focus:outline-none"
          :class="{ 'border-rose-400': headersError }"
        ></textarea>
      </div>
    </div>
  </div>
</template>
