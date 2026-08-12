<template>
  <div class="space-y-4">
    <div>
      <label :for="`${idPrefix}-max-effort`" class="input-label">
        {{ t("admin.groups.form.defaultReasoningEffort") }}
      </label>
      <Select
        :id="`${idPrefix}-max-effort`"
        :model-value="maxEffort"
        :options="reasoningEffortOptions"
        :placeholder="t('admin.groups.form.maxReasoningEffortUnlimited')"
        :aria-label="t('admin.groups.form.maxReasoningEffort')"
        :searchable="false"
        clearable
        @update:model-value="updateMaxEffort"
      />
      <p class="input-hint">
        {{ t("admin.groups.form.defaultReasoningEffortHint") }}
      </p>
    </div>

    <div class="border-t border-gray-200 pt-4 dark:border-dark-600">
      <div class="mb-3 flex items-center justify-between gap-3">
        <div>
          <label class="input-label mb-0">
            {{ t("admin.groups.form.modelReasoningEffortPolicies") }}
          </label>
          <p class="input-hint">
            {{ t("admin.groups.form.modelReasoningEffortPoliciesHint") }}
          </p>
        </div>
        <button
          type="button"
          class="inline-flex min-h-11 items-center gap-1.5 rounded-lg px-2.5 text-sm font-medium text-primary-600 transition-colors hover:bg-primary-50 hover:text-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500/30 dark:text-primary-400 dark:hover:bg-primary-900/20 dark:hover:text-primary-300"
          @click="addModelPolicy"
        >
          <Icon name="plus" size="sm" />
          {{ t("admin.groups.form.addModelReasoningEffortPolicy") }}
        </button>
      </div>
      <div v-if="modelPolicies.length > 0" class="space-y-3">
        <div
          v-for="policy in modelPolicies"
          :key="policy.id"
          class="rounded-lg border border-gray-200 bg-gray-50/40 p-3 dark:border-dark-600 dark:bg-dark-800/40"
        >
          <div
            class="grid gap-3 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] md:items-start"
          >
            <div>
              <label :for="`${idPrefix}-${policy.id}-model`" class="input-label">
                {{ t("admin.groups.form.reasoningEffortPolicyModel") }}
              </label>
              <Select
                :id="`${idPrefix}-${policy.id}-model`"
                :model-value="policy.model"
                :options="modelOptions"
                :placeholder="t('admin.groups.form.reasoningEffortPolicyModelPlaceholder')"
                :searchable="true"
                creatable
                @update:model-value="updateModelPolicy(policy.id, 'model', asString($event))"
              />
            </div>
            <div>
              <label :for="`${idPrefix}-${policy.id}-max-effort`" class="input-label">
                {{ t("admin.groups.form.maxReasoningEffort") }}
              </label>
              <Select
                :id="`${idPrefix}-${policy.id}-max-effort`"
                :model-value="policy.max_effort"
                :options="reasoningEffortOptions"
                :placeholder="t('admin.groups.form.maxReasoningEffortUnlimited')"
                :searchable="false"
                clearable
                @update:model-value="updateModelPolicy(policy.id, 'max_effort', asString($event))"
              />
            </div>
            <button
              type="button"
              class="flex h-11 w-11 items-center justify-center rounded-lg text-gray-400 transition-colors hover:bg-red-50 hover:text-red-500 focus:outline-none focus:ring-2 focus:ring-red-500/30 md:mt-6 dark:hover:bg-red-900/20 dark:hover:text-red-400"
              :title="t('admin.groups.form.removeModelReasoningEffortPolicy')"
              :aria-label="t('admin.groups.form.removeModelReasoningEffortPolicy')"
              @click="removeModelPolicy(policy.id)"
            >
              <Icon name="trash" size="sm" />
            </button>
          </div>
          <div class="mt-3 border-t border-gray-200 pt-3 dark:border-dark-600">
            <div class="grid gap-3 lg:grid-cols-[minmax(0,1fr)_minmax(0,280px)]">
              <div>
                <label class="input-label">
                  {{ t("admin.groups.form.reasoningEffortPolicySchedule") }}
                </label>
                <div class="flex flex-wrap gap-2" role="group" :aria-label="t('admin.groups.form.reasoningEffortPolicyWeekdays')">
                  <label v-for="day in weekdayOptions" :key="day.value" class="cursor-pointer">
                    <input
                      type="checkbox"
                      class="peer sr-only"
                      :checked="isModelPolicyDayActive(policy, day.value)"
                      @change="toggleModelPolicyDay(policy.id, day.value)"
                    />
                    <span class="inline-flex h-9 min-w-10 items-center justify-center rounded-lg border border-gray-300 px-2 text-xs font-medium text-gray-600 transition-colors peer-checked:border-primary-500 peer-checked:bg-primary-50 peer-checked:text-primary-700 dark:border-dark-500 dark:text-gray-300 dark:peer-checked:border-primary-400 dark:peer-checked:bg-primary-900/30 dark:peer-checked:text-primary-200">
                      {{ day.label }}
                    </span>
                  </label>
                </div>
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label :for="`${idPrefix}-${policy.id}-start-time`" class="input-label">
                    {{ t("admin.groups.form.reasoningEffortPolicyStartTime") }}
                  </label>
                  <input
                    :id="`${idPrefix}-${policy.id}-start-time`"
                    :value="policy.start_time"
                    type="time"
                    class="input"
                    @input="updateModelPolicySchedule(policy.id, 'start_time', inputValue($event))"
                  />
                </div>
                <div>
                  <label :for="`${idPrefix}-${policy.id}-end-time`" class="input-label">
                    {{ t("admin.groups.form.reasoningEffortPolicyEndTime") }}
                  </label>
                  <input
                    :id="`${idPrefix}-${policy.id}-end-time`"
                    :value="policy.end_time"
                    type="time"
                    class="input"
                    @input="updateModelPolicySchedule(policy.id, 'end_time', inputValue($event))"
                  />
                </div>
              </div>
            </div>
            <p class="input-hint mt-2">
              {{ t("admin.groups.form.reasoningEffortPolicyScheduleHint") }}
            </p>
            <p
              v-if="showValidation && !isModelPolicyScheduleValid(policy)"
              class="mt-1 text-xs text-red-600 dark:text-red-400"
              role="alert"
            >
              {{ t("admin.groups.form.reasoningEffortPolicyScheduleInvalid") }}
            </p>
          </div>
          <div class="mt-3 border-t border-gray-200 pt-3 dark:border-dark-600">
            <div class="mb-2 flex items-center justify-between gap-3">
              <label class="input-label mb-0">
                {{ t("admin.groups.form.reasoningEffortMappings") }}
              </label>
              <button
                type="button"
                class="inline-flex min-h-9 items-center gap-1 text-sm font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400"
                @click="addModelPolicyMapping(policy.id)"
              >
                <Icon name="plus" size="sm" />
                {{ t("admin.groups.form.addReasoningEffortMapping") }}
              </button>
            </div>
            <div
              v-for="mapping in policy.mappings"
              :key="mapping.id"
              class="mb-2 grid gap-2 md:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)_auto] md:items-center"
            >
              <Select
                :model-value="mapping.from"
                :options="reasoningEffortOptions"
                :placeholder="t('admin.groups.form.reasoningEffortFromPlaceholder')"
                :searchable="false"
                clearable
                @update:model-value="updateModelPolicyMapping(policy.id, mapping.id, 'from', asString($event))"
              />
              <Icon name="arrowRight" size="sm" class="hidden text-gray-400 md:block" />
              <Select
                :model-value="mapping.to"
                :options="reasoningEffortOptions"
                :placeholder="t('admin.groups.form.reasoningEffortToPlaceholder')"
                :searchable="false"
                clearable
                @update:model-value="updateModelPolicyMapping(policy.id, mapping.id, 'to', asString($event))"
              />
              <button
                type="button"
                class="flex h-9 w-9 items-center justify-center rounded-lg text-gray-400 hover:bg-red-50 hover:text-red-500 dark:hover:bg-red-900/20"
                :title="t('admin.groups.form.removeReasoningEffortMapping')"
                @click="removeModelPolicyMapping(policy.id, mapping.id)"
              >
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="border-t border-gray-200 pt-4 dark:border-dark-600">
      <div class="mb-3 flex items-center justify-between gap-3">
        <label class="input-label mb-0">
          {{ t("admin.groups.form.reasoningEffortMappings") }}
        </label>
        <button
          type="button"
          class="inline-flex min-h-11 items-center gap-1.5 rounded-lg px-2.5 text-sm font-medium text-primary-600 transition-colors hover:bg-primary-50 hover:text-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500/30 dark:text-primary-400 dark:hover:bg-primary-900/20 dark:hover:text-primary-300"
          @click="addMapping"
        >
          <Icon name="plus" size="sm" />
          {{ t("admin.groups.form.addReasoningEffortMapping") }}
        </button>
      </div>

      <div v-if="mappings.length > 0" class="space-y-2">
        <div
          v-for="row in mappings"
          :key="row.id"
          class="rounded-lg border border-gray-200 bg-gray-50/40 p-3 dark:border-dark-600 dark:bg-dark-800/40"
        >
          <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)_auto] md:items-start">
            <div>
              <label :for="`${idPrefix}-${row.id}-from`" class="input-label">
                {{ t("admin.groups.form.reasoningEffortFrom") }}
              </label>
              <Select
                :id="`${idPrefix}-${row.id}-from`"
                :model-value="row.from"
                :options="reasoningEffortOptions"
                :placeholder="t('admin.groups.form.reasoningEffortFromPlaceholder')"
                :error="showValidation && !!validationErrors[row.id]?.from"
                :aria-label="t('admin.groups.form.reasoningEffortFrom')"
                :aria-describedby="showValidation && validationErrors[row.id]?.from ? `${idPrefix}-${row.id}-from-error` : undefined"
                :searchable="false"
                clearable
                @update:model-value="updateMapping(row.id, 'from', $event)"
              />
              <p
                v-if="showValidation && validationErrors[row.id]?.from"
                :id="`${idPrefix}-${row.id}-from-error`"
                class="mt-1 text-xs text-red-600 dark:text-red-400"
                role="alert"
              >
                {{ mappingErrorText(validationErrors[row.id]?.from) }}
              </p>
            </div>

            <div class="hidden pt-8 text-gray-400 md:block dark:text-dark-400">
              <Icon name="arrowRight" size="sm" />
            </div>

            <div>
              <label :for="`${idPrefix}-${row.id}-to`" class="input-label">
                {{ t("admin.groups.form.reasoningEffortTo") }}
              </label>
              <Select
                :id="`${idPrefix}-${row.id}-to`"
                :model-value="row.to"
                :options="reasoningEffortOptions"
                :placeholder="t('admin.groups.form.reasoningEffortToPlaceholder')"
                :error="showValidation && !!validationErrors[row.id]?.to"
                :aria-label="t('admin.groups.form.reasoningEffortTo')"
                :aria-describedby="showValidation && validationErrors[row.id]?.to ? `${idPrefix}-${row.id}-to-error` : undefined"
                :searchable="false"
                clearable
                @update:model-value="updateMapping(row.id, 'to', $event)"
              />
              <p
                v-if="showValidation && validationErrors[row.id]?.to"
                :id="`${idPrefix}-${row.id}-to-error`"
                class="mt-1 text-xs text-red-600 dark:text-red-400"
                role="alert"
              >
                {{ mappingErrorText(validationErrors[row.id]?.to) }}
              </p>
            </div>

            <button
              type="button"
              class="flex h-11 w-11 items-center justify-center rounded-lg text-gray-400 transition-colors hover:bg-red-50 hover:text-red-500 focus:outline-none focus:ring-2 focus:ring-red-500/30 md:mt-6 dark:hover:bg-red-900/20 dark:hover:text-red-400"
              :title="t('admin.groups.form.removeReasoningEffortMapping')"
              :aria-label="t('admin.groups.form.removeReasoningEffortMapping')"
              @click="removeMapping(row.id)"
            >
              <Icon name="trash" size="sm" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";
import type { GroupPlatform } from "@/types";
import Icon from "@/components/icons/Icon.vue";
import Select from "@/components/common/Select.vue";
import {
  createReasoningEffortMappingRow,
  createReasoningEffortModelPolicyRow,
  reasoningEffortOptionsForPlatform,
  validateReasoningEffortMappings,
  type ReasoningEffortMappingErrorCode,
  type ReasoningEffortMappingRow,
  type ReasoningEffortModelPolicyRow,
} from "@/views/admin/groupsReasoningEffort";

const props = defineProps<{
  idPrefix: string;
  platform: GroupPlatform;
  maxEffort: string;
  mappings: ReasoningEffortMappingRow[];
  modelPolicies: ReasoningEffortModelPolicyRow[];
  models: string[];
}>();

const emit = defineEmits<{
  (event: "update:maxEffort", value: string): void;
  (event: "update:mappings", value: ReasoningEffortMappingRow[]): void;
  (event: "update:modelPolicies", value: ReasoningEffortModelPolicyRow[]): void;
}>();

const { t } = useI18n();
const showValidation = ref(false);
const reasoningEffortOptions = computed(() =>
  reasoningEffortOptionsForPlatform(props.platform),
);
const modelOptions = computed(() =>
  Array.from(
    new Set([
      ...props.models,
      ...props.modelPolicies.map((policy) => policy.model).filter(Boolean),
    ]),
  ).map((model) => ({ value: model, label: model })),
);
const weekdayOptions = computed(() => [
  { value: 1, label: t("admin.groups.form.reasoningEffortPolicyMonday") },
  { value: 2, label: t("admin.groups.form.reasoningEffortPolicyTuesday") },
  { value: 3, label: t("admin.groups.form.reasoningEffortPolicyWednesday") },
  { value: 4, label: t("admin.groups.form.reasoningEffortPolicyThursday") },
  { value: 5, label: t("admin.groups.form.reasoningEffortPolicyFriday") },
  { value: 6, label: t("admin.groups.form.reasoningEffortPolicySaturday") },
  { value: 7, label: t("admin.groups.form.reasoningEffortPolicySunday") },
]);
const validationErrors = computed(() =>
  validateReasoningEffortMappings(props.mappings, props.platform),
);

const asString = (value: string | number | boolean | null): string =>
  value == null ? "" : String(value);
const inputValue = (event: Event): string =>
  (event.target as HTMLInputElement | null)?.value ?? "";

const updateMaxEffort = (value: string | number | boolean | null) => {
  emit("update:maxEffort", asString(value));
};

const updateMapping = (
  id: string,
  field: "from" | "to",
  value: string | number | boolean | null,
) => {
  emit(
    "update:mappings",
    props.mappings.map((row) =>
      row.id === id ? { ...row, [field]: asString(value) } : row,
    ),
  );
};

const addMapping = () => {
  emit("update:mappings", [
    ...props.mappings,
    createReasoningEffortMappingRow(),
  ]);
};

const updateModelPolicy = (
  id: string,
  field: "model" | "max_effort",
  value: string,
) => {
  emit(
    "update:modelPolicies",
    props.modelPolicies.map((policy) =>
      policy.id === id ? { ...policy, [field]: value } : policy,
    ),
  );
};
const updateModelPolicySchedule = (
  id: string,
  field: "start_time" | "end_time",
  value: string,
) => {
  emit(
    "update:modelPolicies",
    props.modelPolicies.map((policy) =>
      policy.id === id ? { ...policy, [field]: value } : policy,
    ),
  );
};
const isModelPolicyDayActive = (
  policy: ReasoningEffortModelPolicyRow,
  day: number,
): boolean => policy.active_days.length === 0 || policy.active_days.includes(day);
const toggleModelPolicyDay = (id: string, day: number) => {
  emit(
    "update:modelPolicies",
    props.modelPolicies.map((policy) => {
      if (policy.id !== id) return policy;
      const activeDays = new Set(policy.active_days.length > 0 ? policy.active_days : [1, 2, 3, 4, 5, 6, 7]);
      if (activeDays.has(day)) {
        if (activeDays.size === 1) return policy;
        activeDays.delete(day);
      } else {
        activeDays.add(day);
      }
      const nextDays = [...activeDays].sort((left, right) => left - right);
      return { ...policy, active_days: nextDays.length === 7 ? [] : nextDays };
    }),
  );
};
const addModelPolicy = () => {
  emit("update:modelPolicies", [
    ...props.modelPolicies,
    createReasoningEffortModelPolicyRow({}, props.platform),
  ]);
};
const removeModelPolicy = (id: string) => {
  emit("update:modelPolicies", props.modelPolicies.filter((policy) => policy.id !== id));
};
const updateModelPolicyMapping = (
  policyID: string,
  mappingID: string,
  field: "from" | "to",
  value: string,
) => {
  emit(
    "update:modelPolicies",
    props.modelPolicies.map((policy) =>
      policy.id !== policyID
        ? policy
        : {
            ...policy,
            mappings: policy.mappings.map((mapping) =>
              mapping.id === mappingID ? { ...mapping, [field]: value } : mapping,
            ),
          },
    ),
  );
};
const addModelPolicyMapping = (policyID: string) => {
  emit(
    "update:modelPolicies",
    props.modelPolicies.map((policy) =>
      policy.id === policyID
        ? { ...policy, mappings: [...policy.mappings, createReasoningEffortMappingRow()] }
        : policy,
    ),
  );
};
const removeModelPolicyMapping = (policyID: string, mappingID: string) => {
  emit(
    "update:modelPolicies",
    props.modelPolicies.map((policy) =>
      policy.id === policyID
        ? {
            ...policy,
            mappings: policy.mappings.filter((mapping) => mapping.id !== mappingID),
          }
        : policy,
    ),
  );
};

const removeMapping = (id: string) => {
  emit(
    "update:mappings",
    props.mappings.filter((row) => row.id !== id),
  );
};

const mappingErrorText = (
  code: ReasoningEffortMappingErrorCode | undefined,
): string => (code ? t(`admin.groups.form.${code}`) : "");

const isModelPolicyScheduleValid = (policy: ReasoningEffortModelPolicyRow): boolean => {
  const startTime = policy.start_time.trim();
  const endTime = policy.end_time.trim();
  return (startTime === "" && endTime === "") ||
    (startTime !== "" && endTime !== "" && startTime !== endTime);
};

const validate = (): boolean => {
  showValidation.value = true;
  const policyModels = props.modelPolicies.map((policy) =>
    policy.model.trim().toLowerCase(),
  );
  return (
    Object.keys(validationErrors.value).length === 0 &&
    props.modelPolicies.every(
      (policy) =>
        policy.model.trim() &&
        (policy.max_effort || policy.mappings.length > 0) &&
        isModelPolicyScheduleValid(policy) &&
        Object.keys(validateReasoningEffortMappings(policy.mappings, props.platform)).length === 0,
    ) &&
    new Set(policyModels).size === policyModels.length
  );
};

const resetValidation = () => {
  showValidation.value = false;
};

defineExpose({ validate, resetValidation });
</script>
