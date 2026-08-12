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
              v-for="group in policy.mappings"
              :key="group.id"
              class="mb-2 grid gap-2 md:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)_auto] md:items-center"
            >
              <template v-for="pair in group.pairs" :key="pair.id">
                <Select
                  :model-value="pair.from"
                  :options="reasoningEffortOptions"
                  :placeholder="t('admin.groups.form.reasoningEffortFromPlaceholder')"
                  :searchable="false"
                  clearable
                  @update:model-value="updateModelPolicyPair(policy.id, group.id, pair.id, 'from', asString($event))"
                />
                <Icon name="arrowRight" size="sm" class="hidden text-gray-400 md:block" />
                <Select
                  :model-value="pair.to"
                  :options="reasoningEffortOptions"
                  :placeholder="t('admin.groups.form.reasoningEffortToPlaceholder')"
                  :searchable="false"
                  clearable
                  @update:model-value="updateModelPolicyPair(policy.id, group.id, pair.id, 'to', asString($event))"
                />
              </template>
              <button
                type="button"
                class="flex h-9 w-9 items-center justify-center rounded-lg text-gray-400 hover:bg-red-50 hover:text-red-500 dark:hover:bg-red-900/20"
                :title="t('admin.groups.form.removeReasoningEffortMapping')"
                @click="removeModelPolicyMapping(policy.id, group.id)"
              >
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div>
      <label :for="`${idPrefix}-over-limit`" class="input-label">
        {{ t("admin.groups.form.maxReasoningEffortOverLimit") }}
      </label>
      <Select
        :id="`${idPrefix}-over-limit`"
        :model-value="overLimit"
        :options="overLimitOptions"
        :aria-label="t('admin.groups.form.maxReasoningEffortOverLimit')"
        :searchable="false"
        :disabled="!maxEffort"
        @update:model-value="updateOverLimit"
      />
      <p class="input-hint">{{ t("admin.groups.form.maxReasoningEffortOverLimitHint") }}</p>
    </div>

    <div class="border-t border-gray-200 pt-4 dark:border-dark-600">
      <div class="mb-3 flex items-center justify-between gap-3">
        <div>
          <label class="input-label mb-0">
            {{ t("admin.groups.form.reasoningEffortMappings") }}
          </label>
          <p class="input-hint mb-0">
            {{ t("admin.groups.form.reasoningEffortMappingsHint") }}
          </p>
        </div>
        <button
          type="button"
          class="inline-flex min-h-11 shrink-0 items-center gap-1.5 rounded-lg px-2.5 text-sm font-medium text-primary-600 transition-colors hover:bg-primary-50 hover:text-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500/30 dark:text-primary-400 dark:hover:bg-primary-900/20 dark:hover:text-primary-300"
          @click="addGroup"
        >
          <Icon name="plus" size="sm" />
          {{ t("admin.groups.form.addReasoningEffortMapping") }}
        </button>
      </div>

      <div v-if="mappings.length > 0" class="space-y-2">
        <div
          v-for="group in mappings"
          :key="group.id"
          class="space-y-3 rounded-lg border border-gray-200 bg-gray-50/40 p-3 dark:border-dark-600 dark:bg-dark-800/40"
        >
          <div
            class="grid grid-cols-1 items-start gap-3 md:grid-cols-[minmax(0,1fr)_1.25rem_minmax(0,1fr)_2.75rem]"
          >
            <div>
              <label :for="`${idPrefix}-${group.id}-match-type`" class="input-label">
                {{ t("admin.groups.form.reasoningEffortMatchType") }}
              </label>
              <Select
                :id="`${idPrefix}-${group.id}-match-type`"
                :model-value="group.match_type"
                :options="matchTypeOptions"
                :placeholder="t('admin.groups.form.reasoningEffortMatchTypePlaceholder')"
                :error="showValidation && !!groupErrors(group.id).match_type"
                :aria-label="t('admin.groups.form.reasoningEffortMatchType')"
                :searchable="false"
                clearable
                @update:model-value="updateGroup(group.id, 'match_type', $event)"
              />
              <p
                v-if="showValidation && groupErrors(group.id).match_type"
                class="mt-1 text-xs text-red-600 dark:text-red-400"
                role="alert"
              >
                {{ mappingErrorText(groupErrors(group.id).match_type) }}
              </p>
            </div>

            <div class="hidden md:block" aria-hidden="true" />

            <div>
              <label :for="`${idPrefix}-${group.id}-model`" class="input-label">
                {{ t("admin.groups.form.reasoningEffortModel") }}
              </label>
              <input
                :id="`${idPrefix}-${group.id}-model`"
                :value="group.model"
                type="text"
                maxlength="200"
                autocomplete="off"
                class="input"
                :placeholder="t('admin.groups.form.reasoningEffortModelPlaceholder')"
                :aria-label="t('admin.groups.form.reasoningEffortModel')"
                @input="onModelInput(group.id, $event)"
              />
            </div>

            <button
              type="button"
              class="flex h-11 w-11 items-center justify-center self-end rounded-lg text-gray-400 transition-colors hover:bg-red-50 hover:text-red-500 focus:outline-none focus:ring-2 focus:ring-red-500/30 dark:hover:bg-red-900/20 dark:hover:text-red-400"
              :title="t('admin.groups.form.removeReasoningEffortMapping')"
              :aria-label="t('admin.groups.form.removeReasoningEffortMapping')"
              @click="removeGroup(group.id)"
            >
              <Icon name="trash" size="sm" />
            </button>
          </div>

          <p
            v-if="showValidation && groupErrors(group.id).duplicateScope"
            class="text-xs text-red-600 dark:text-red-400"
            role="alert"
          >
            {{ mappingErrorText(groupErrors(group.id).duplicateScope) }}
          </p>

          <div
            v-for="pair in group.pairs"
            :key="pair.id"
            class="grid grid-cols-1 items-start gap-3 md:grid-cols-[minmax(0,1fr)_1.25rem_minmax(0,1fr)_2.75rem]"
          >
            <div>
              <label :for="`${idPrefix}-${pair.id}-from`" class="input-label">
                {{ t("admin.groups.form.reasoningEffortFrom") }}
              </label>
              <Select
                :id="`${idPrefix}-${pair.id}-from`"
                :model-value="pair.from"
                :options="reasoningEffortSourceOptions"
                :placeholder="t('admin.groups.form.reasoningEffortFromPlaceholder')"
                :error="showValidation && !!pairErrors(pair.id).from"
                :aria-label="t('admin.groups.form.reasoningEffortFrom')"
                :searchable="false"
                clearable
                @update:model-value="updatePair(group.id, pair.id, 'from', $event)"
              />
              <p
                v-if="showValidation && pairErrors(pair.id).from"
                class="mt-1 text-xs text-red-600 dark:text-red-400"
                role="alert"
              >
                {{ mappingErrorText(pairErrors(pair.id).from) }}
              </p>
            </div>

            <div class="hidden h-11 items-center justify-center self-end text-gray-400 md:flex dark:text-dark-400">
              <Icon name="arrowRight" size="sm" />
            </div>

            <div>
              <label :for="`${idPrefix}-${pair.id}-to`" class="input-label">
                {{ t("admin.groups.form.reasoningEffortTo") }}
              </label>
              <Select
                :id="`${idPrefix}-${pair.id}-to`"
                :model-value="pair.to"
                :options="reasoningEffortTargetOptions"
                :placeholder="t('admin.groups.form.reasoningEffortToPlaceholder')"
                :error="showValidation && !!pairErrors(pair.id).to"
                :aria-label="t('admin.groups.form.reasoningEffortTo')"
                :searchable="false"
                clearable
                @update:model-value="updatePair(group.id, pair.id, 'to', $event)"
              />
              <p
                v-if="showValidation && pairErrors(pair.id).to"
                class="mt-1 text-xs text-red-600 dark:text-red-400"
                role="alert"
              >
                {{ mappingErrorText(pairErrors(pair.id).to) }}
              </p>
            </div>

            <button
              type="button"
              class="flex h-11 w-11 items-center justify-center self-end rounded-lg text-gray-400 transition-colors hover:bg-red-50 hover:text-red-500 focus:outline-none focus:ring-2 focus:ring-red-500/30 dark:hover:bg-red-900/20 dark:hover:text-red-400"
              :title="t('admin.groups.form.removeReasoningEffortPair')"
              :aria-label="t('admin.groups.form.removeReasoningEffortPair')"
              @click="removePair(group.id, pair.id)"
            >
              <Icon name="trash" size="sm" />
            </button>
          </div>

          <button
            type="button"
            class="inline-flex min-h-9 items-center gap-1.5 text-sm font-medium text-primary-600 transition-colors hover:text-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500/30 dark:text-primary-400 dark:hover:text-primary-300"
            @click="addPair(group.id)"
          >
            <Icon name="plus" size="sm" />
            {{ t("admin.groups.form.addReasoningEffortPair") }}
          </button>
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
  createReasoningEffortMappingPair,
  createReasoningEffortMappingRow,
  normalizeReasoningEffortMatchType,
  createReasoningEffortModelPolicyRow,
  reasoningEffortOptionsForPlatform,
  reasoningEffortSourceOptionsForPlatform,
  reasoningEffortTargetOptionsForPlatform,
  reasoningEffortMappingDeny,
  reasoningEffortOverLimitDeny,
  reasoningEffortOverLimitDowngrade,
  validateReasoningEffortMappings,
  type ReasoningEffortMappingErrorCode,
  type ReasoningEffortMappingRow,
  type ReasoningEffortModelPolicyRow,
} from "@/views/admin/groupsReasoningEffort";

const props = defineProps<{
  idPrefix: string;
  platform: GroupPlatform;
  maxEffort: string;
  overLimit: string;
  mappings: ReasoningEffortMappingRow[];
  modelPolicies: ReasoningEffortModelPolicyRow[];
  models: string[];
}>();

const emit = defineEmits<{
  (event: "update:maxEffort", value: string): void;
  (event: "update:overLimit", value: string): void;
  (event: "update:mappings", value: ReasoningEffortMappingRow[]): void;
  (event: "update:modelPolicies", value: ReasoningEffortModelPolicyRow[]): void;
}>();

const { t } = useI18n();
const showValidation = ref(false);
const reasoningEffortOptions = computed(() =>
  reasoningEffortOptionsForPlatform(props.platform),
);
const reasoningEffortSourceOptions = computed(() =>
  reasoningEffortSourceOptionsForPlatform(props.platform),
);
const reasoningEffortTargetOptions = computed(() =>
  reasoningEffortTargetOptionsForPlatform(props.platform).map((option) =>
    option.value === reasoningEffortMappingDeny
      ? {
          ...option,
          label: t("admin.groups.form.reasoningEffortToDeny"),
        }
      : option,
  ),
);
const matchTypeOptions = computed(() => [
  {
    value: "exact",
    label: t("admin.groups.form.reasoningEffortMatchExact"),
  },
  {
    value: "prefix",
    label: t("admin.groups.form.reasoningEffortMatchPrefix"),
  },
  {
    value: "suffix",
    label: t("admin.groups.form.reasoningEffortMatchSuffix"),
  },
]);
const overLimitOptions = computed(() => [
  {
    value: reasoningEffortOverLimitDowngrade,
    label: t("admin.groups.form.maxReasoningEffortOverLimitDowngrade"),
  },
  {
    value: reasoningEffortOverLimitDeny,
    label: t("admin.groups.form.maxReasoningEffortOverLimitDeny"),
  },
]);
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

const groupErrors = (id: string) => validationErrors.value[id] ?? {};
const pairErrors = (id: string) => validationErrors.value[id] ?? {};

const updateMaxEffort = (value: string | number | boolean | null) => {
  emit("update:maxEffort", asString(value));
};

const updateOverLimit = (value: string | number | boolean | null) => {
  emit(
    "update:overLimit",
    asString(value) || reasoningEffortOverLimitDowngrade,
  );
};

const updateGroup = (
  id: string,
  field: "match_type" | "model",
  value: string | number | boolean | null,
) => {
  const nextValue =
    field === "match_type"
      ? normalizeReasoningEffortMatchType(asString(value))
      : asString(value);
  emit(
    "update:mappings",
    props.mappings.map((group) =>
      group.id === id ? { ...group, [field]: nextValue } : group,
    ),
  );
};

const onModelInput = (id: string, event: Event) => {
  const target = event.target as HTMLInputElement | null;
  updateGroup(id, "model", target?.value ?? "");
};

const updatePair = (
  groupId: string,
  pairId: string,
  field: "from" | "to",
  value: string | number | boolean | null,
) => {
  emit(
    "update:mappings",
    props.mappings.map((group) =>
      group.id === groupId
        ? {
            ...group,
            pairs: group.pairs.map((pair) =>
              pair.id === pairId ? { ...pair, [field]: asString(value) } : pair,
            ),
          }
        : group,
    ),
  );
};

const addGroup = () => {
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
const updateModelPolicyPair = (
  policyID: string,
  groupID: string,
  pairID: string,
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
            mappings: policy.mappings.map((group) =>
              group.id !== groupID
                ? group
                : {
                    ...group,
                    pairs: group.pairs.map((pair) =>
                      pair.id === pairID ? { ...pair, [field]: value } : pair,
                    ),
                  },
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
const removeModelPolicyMapping = (policyID: string, groupID: string) => {
  emit(
    "update:modelPolicies",
    props.modelPolicies.map((policy) =>
      policy.id === policyID
        ? {
            ...policy,
            mappings: policy.mappings.filter((group) => group.id !== groupID),
          }
        : policy,
    ),
  );
};

const removeGroup = (id: string) => {
  emit(
    "update:mappings",
    props.mappings.filter((group) => group.id !== id),
  );
};

const addPair = (groupId: string) => {
  emit(
    "update:mappings",
    props.mappings.map((group) =>
      group.id === groupId
        ? { ...group, pairs: [...group.pairs, createReasoningEffortMappingPair()] }
        : group,
    ),
  );
};

const removePair = (groupId: string, pairId: string) => {
  emit(
    "update:mappings",
    props.mappings.flatMap((group) => {
      if (group.id !== groupId) return [group];
      const pairs = group.pairs.filter((pair) => pair.id !== pairId);
      return pairs.length > 0 ? [{ ...group, pairs }] : [];
    }),
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
