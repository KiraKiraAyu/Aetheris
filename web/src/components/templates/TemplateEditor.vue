<script setup lang="ts">
import { computed, reactive, watch } from "vue";
import Button from "primevue/button";
import InputText from "primevue/inputtext";
import Select from "primevue/select";
import Textarea from "primevue/textarea";
import type { NotificationTemplate, ChannelOption } from "@/lib/types";
import { useI18nStore } from "@/stores/i18n";

const props = defineProps<{
  form: NotificationTemplate;
  channelOptions: ChannelOption[];
  saving: boolean;
  mockVariablesJson: string;
}>();

const i18n = useI18nStore();

const emit = defineEmits<{
  (e: "update:form", value: NotificationTemplate): void;
  (e: "update:mockVariablesJson", value: string): void;
  (e: "reset"): void;
  (e: "save"): void;
}>();

const localForm = reactive<NotificationTemplate>({ ...props.form });

watch(
  () => props.form,
  (value) => {
    Object.assign(localForm, value);
  },
  { deep: true },
);

watch(
  localForm,
  (value) => {
    emit("update:form", { ...value });
  },
  { deep: true },
);

const localMockJson = computed({
  get: () => props.mockVariablesJson,
  set: (val) => emit("update:mockVariablesJson", val),
});

const isMockJsonValid = computed(() => {
  try {
    JSON.parse(localMockJson.value);
    return true;
  } catch {
    return false;
  }
});

function renderMockTemplate(templateStr: string) {
  if (!templateStr) return "";
  let metadata: Record<string, unknown> = {};
  try {
    metadata = localMockJson.value.trim() ? JSON.parse(localMockJson.value) : {};
  } catch {
    // Wait for valid JSON
  }

  // Parse {{ .Metadata.key }}
  let rendered = templateStr.replace(
    /\{\{\s*\.Metadata\.([a-zA-Z0-9_-]+)\s*\}\}/g,
    (match, key) => {
      return metadata[key] !== undefined ? String(metadata[key]) : match;
    },
  );

  // Parse system properties {{ .TenantID }} etc.
  rendered = rendered
    .replace(/\{\{\s*\.TenantID\s*\}\}/g, "tenant_demo_12")
    .replace(/\{\{\s*\.Recipient\s*\}\}/g, "recipient@example.com")
    .replace(/\{\{\s*\.Channel\s*\}\}/g, localForm.channel);

  return rendered;
}

const previewTitle = computed(() => renderMockTemplate(localForm.title_template || ""));
const previewBody = computed(() => renderMockTemplate(localForm.body_template || ""));
</script>

<template>
  <div class="space-y-5">
    <!-- Form Panel -->
    <div
      class="p-6 bg-surface-0 dark:bg-surface-900 rounded-2xl border border-surface-200 dark:border-surface-800"
    >
      <div class="mb-5 flex items-start gap-3">
        <div
          class="grid h-11 w-11 place-items-center rounded-xl bg-reisa-lilac-50 dark:bg-reisa-lilac-950/30 text-reisa-lilac-600 dark:text-reisa-lilac-400"
        >
          <span class="pi pi-th-large text-lg"></span>
        </div>
        <div>
          <h2 class="text-lg font-bold text-surface-900 dark:text-white">
            {{
              localForm.id
                ? i18n.t("templates.edit_dialog_title")
                : i18n.t("templates.create_dialog_title")
            }}
          </h2>
          <p class="text-xs text-surface-500 dark:text-surface-400">
            {{ i18n.t("templates.editor_desc") }}
          </p>
        </div>
      </div>

      <div class="mt-4 space-y-4">
        <div class="space-y-1">
          <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">{{
            i18n.t("templates.field_key")
          }}</label>
          <InputText
            v-model="localForm.key"
            placeholder="e.g. auth.otp-verification"
            class="h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl text-sm font-mono"
          />
        </div>

        <div class="space-y-1">
          <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">{{
            i18n.t("templates.field_channel")
          }}</label>
          <Select
            v-model="localForm.channel"
            :options="channelOptions"
            option-label="label"
            option-value="value"
            class="h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl text-sm"
          >
            <template #option="{ option }">
              <div class="flex items-center gap-2 text-xs">
                <span :class="option.icon" class="text-surface-400"></span>
                <span class="capitalize">{{ option.label }}</span>
              </div>
            </template>
          </Select>
        </div>

        <div class="space-y-1">
          <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">{{
            i18n.t("templates.field_title")
          }}</label>
          <InputText
            v-model="localForm.title_template"
            placeholder="Verify code: {{ .Metadata.code }}"
            class="h-10 border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl text-sm"
          />
        </div>

        <div class="space-y-1">
          <label class="text-xs font-semibold text-surface-500 dark:text-surface-400">{{
            i18n.t("templates.field_body")
          }}</label>
          <Textarea
            v-model="localForm.body_template"
            rows="6"
            placeholder="Dear {{ .Metadata.name }},\nYour OTP authentication code is {{ .Metadata.code }}."
            class="border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl text-xs leading-relaxed"
          />
        </div>

        <div class="flex gap-2 pt-2 border-t border-surface-100 dark:border-surface-800">
          <Button
            icon="pi pi-undo"
            :label="i18n.t('templates.btn_reset')"
            class="flex-1 rounded-xl h-10 text-xs text-surface-500 dark:text-surface-400"
            outlined
            severity="secondary"
            @click="emit('reset')"
          />
          <Button
            icon="pi pi-save"
            :label="i18n.t('templates.btn_save')"
            class="flex-1 rounded-xl h-10 text-xs font-semibold"
            :loading="saving"
            @click="emit('save')"
          />
        </div>
      </div>
    </div>

    <!-- Preview Mock Configurations -->
    <div
      class="p-6 bg-surface-0 dark:bg-surface-900 rounded-2xl border border-surface-200 dark:border-surface-800"
    >
      <div class="mb-4 flex items-center justify-between">
        <div>
          <h3 class="text-sm font-bold text-surface-900 dark:text-white">
            {{ i18n.t("templates.sandbox_title") }}
          </h3>
          <p class="text-[10px] text-surface-400 dark:text-surface-500">
            {{ i18n.t("templates.sandbox_desc") }}
          </p>
        </div>
        <span
          class="h-6 px-2 rounded-lg text-[9px] uppercase tracking-wider font-bold border flex items-center gap-1 select-none"
          :class="
            isMockJsonValid
              ? 'bg-reisa-lilac-50 border-reisa-lilac-200 text-reisa-lilac-600 dark:bg-reisa-lilac-950/20 dark:border-reisa-lilac-800/40 dark:text-reisa-lilac-400'
              : 'bg-rose-50 border-rose-200 text-rose-600 dark:bg-rose-950/20 dark:border-rose-800/40 dark:text-rose-400'
          "
        >
          <span
            class="h-1.5 w-1.5 rounded-full"
            :class="isMockJsonValid ? 'bg-reisa-lilac-500' : 'bg-rose-500 animate-pulse'"
          ></span>
          {{ isMockJsonValid ? i18n.t("composer.json_valid") : i18n.t("composer.json_error") }}
        </span>
      </div>

      <div class="space-y-4">
        <Textarea
          v-model="localMockJson"
          rows="5"
          class="font-mono text-xs border-surface-200 dark:border-surface-800 bg-surface-50/50 dark:bg-surface-950/50 rounded-xl leading-normal p-3 w-full"
          :class="{
            'border-rose-300 focus:border-rose-500 focus:ring-rose-500': !isMockJsonValid,
          }"
        />

        <!-- Live Output Preview -->
        <div
          class="p-4 rounded-xl border border-surface-100 dark:border-surface-800 bg-surface-50/40 dark:bg-surface-950/20 space-y-3"
        >
          <div>
            <div
              class="text-[9px] font-bold uppercase tracking-wider text-surface-400 dark:text-surface-500"
            >
              {{ i18n.t("templates.sandbox_rendered_title") }}
            </div>
            <div class="mt-1 text-xs font-bold text-surface-800 dark:text-surface-200">
              {{ previewTitle || i18n.t("templates.sandbox_empty_title") }}
            </div>
          </div>
          <div class="border-t border-surface-100 dark:border-surface-800/60 pt-2">
            <div
              class="text-[9px] font-bold uppercase tracking-wider text-surface-400 dark:text-surface-500"
            >
              {{ i18n.t("templates.sandbox_rendered_body") }}
            </div>
            <div
              class="mt-1 text-xs text-surface-600 dark:text-surface-300 whitespace-pre-wrap leading-relaxed"
            >
              {{ previewBody || i18n.t("templates.sandbox_empty_body") }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
