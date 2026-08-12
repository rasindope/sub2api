import type { GroupPlatform, ReasoningEffortMapping, ReasoningEffortModelPolicy } from "@/types";

const openAIReasoningEffortValues = [
  "minimal",
  "low",
  "medium",
  "high",
  "xhigh",
  "max",
] as const;

const reasoningEffortValuesForPlatform = (
  platform: GroupPlatform,
): readonly string[] =>
  supportsReasoningEffortPolicyPlatform(platform)
    ? openAIReasoningEffortValues
    : [];

export function supportsReasoningEffortPolicyPlatform(
  platform: GroupPlatform,
): boolean {
  return platform === "openai" || platform === "composite";
}

export function reasoningEffortOptionsForPlatform(platform: GroupPlatform) {
  return reasoningEffortValuesForPlatform(platform).map((value) => ({
    value,
    label: value,
  }));
}

export function normalizeReasoningEffortForPlatform(
  platform: GroupPlatform,
  value: string | null | undefined,
): string {
  const normalized = value?.trim().toLowerCase() ?? "";
  return reasoningEffortValuesForPlatform(platform).some(
    (allowed) => allowed === normalized,
  )
    ? normalized
    : "";
}

export interface ReasoningEffortMappingRow extends ReasoningEffortMapping {
  id: string;
}

export interface ReasoningEffortModelPolicyRow {
  id: string;
  model: string;
  max_effort: string;
  mappings: ReasoningEffortMappingRow[];
  active_days: number[];
  start_time: string;
  end_time: string;
}

export type ReasoningEffortMappingErrorCode =
  | "fromRequired"
  | "toRequired"
  | "duplicateFrom"
  | "unsupportedFrom"
  | "unsupportedTo";

export type ReasoningEffortMappingErrors = Record<
  string,
  Partial<Record<"from" | "to", ReasoningEffortMappingErrorCode>>
>;

let nextMappingRowID = 0;
let nextModelPolicyRowID = 0;

function normalizeReasoningEffortPolicyActiveDays(
  activeDays?: number[] | null,
): number[] {
  return Array.from(
    new Set(
      (activeDays ?? []).filter(
        (day) => Number.isInteger(day) && day >= 1 && day <= 7,
      ),
    ),
  ).sort((left, right) => left - right);
}

export function createReasoningEffortMappingRow(
  mapping: Partial<ReasoningEffortMapping> = {},
): ReasoningEffortMappingRow {
  nextMappingRowID += 1;
  return {
    id: `reasoning-effort-mapping-${nextMappingRowID}`,
    from: mapping.from ?? "",
    to: mapping.to ?? "",
  };
}

export function createReasoningEffortModelPolicyRow(
  policy: Partial<ReasoningEffortModelPolicy> = {},
  platform: GroupPlatform = "openai",
): ReasoningEffortModelPolicyRow {
  nextModelPolicyRowID += 1;
  return {
    id: `reasoning-effort-model-policy-${nextModelPolicyRowID}`,
    model: policy.model?.trim() ?? "",
    max_effort: normalizeReasoningEffortForPlatform(platform, policy.max_effort),
    mappings: reasoningEffortMappingsToRows(policy.mappings, platform),
    active_days: normalizeReasoningEffortPolicyActiveDays(policy.active_days),
    start_time: policy.start_time?.trim() ?? "",
    end_time: policy.end_time?.trim() ?? "",
  };
}

export function reasoningEffortModelPoliciesToRows(
  policies?: ReasoningEffortModelPolicy[] | null,
  platform: GroupPlatform = "openai",
): ReasoningEffortModelPolicyRow[] {
  return (policies ?? []).flatMap((policy) =>
    policy.model?.trim()
      ? [createReasoningEffortModelPolicyRow(policy, platform)]
      : [],
  );
}

export function reasoningEffortModelPoliciesToAPI(
  rows: ReasoningEffortModelPolicyRow[],
): ReasoningEffortModelPolicy[] {
  return rows.map((row) => {
    const activeDays = normalizeReasoningEffortPolicyActiveDays(row.active_days);
    const startTime = row.start_time.trim();
    const endTime = row.end_time.trim();
    return {
      model: row.model.trim(),
      max_effort: row.max_effort.trim(),
      mappings: reasoningEffortMappingsToAPI(row.mappings),
      ...(activeDays.length > 0 ? { active_days: activeDays } : {}),
      ...(startTime ? { start_time: startTime } : {}),
      ...(endTime ? { end_time: endTime } : {}),
    };
  });
}

export function reasoningEffortMappingsToRows(
  mappings?: ReasoningEffortMapping[] | null,
  platform: GroupPlatform = "openai",
): ReasoningEffortMappingRow[] {
  return (mappings ?? []).flatMap((mapping) => {
    const from = normalizeReasoningEffortForPlatform(platform, mapping.from);
    const to = normalizeReasoningEffortForPlatform(platform, mapping.to);
    return from && to
      ? [createReasoningEffortMappingRow({ from, to })]
      : [];
  });
}

export function reasoningEffortMappingsToAPI(
  rows: ReasoningEffortMappingRow[],
): ReasoningEffortMapping[] {
  return rows.map((row) => ({
    from: row.from.trim(),
    to: row.to.trim(),
  }));
}

export function validateReasoningEffortMappings(
  rows: ReasoningEffortMappingRow[],
  platform: GroupPlatform = "openai",
): ReasoningEffortMappingErrors {
  const errors: ReasoningEffortMappingErrors = {};
  const sourceRows = new Map<string, ReasoningEffortMappingRow[]>();

  rows.forEach((row) => {
    const from = row.from.trim();
    const to = row.to.trim();
    if (!from) {
      errors[row.id] = { ...errors[row.id], from: "fromRequired" };
    } else if (!normalizeReasoningEffortForPlatform(platform, from)) {
      errors[row.id] = { ...errors[row.id], from: "unsupportedFrom" };
    } else {
      const key = from.toLowerCase();
      sourceRows.set(key, [...(sourceRows.get(key) ?? []), row]);
    }
    if (!to) {
      errors[row.id] = { ...errors[row.id], to: "toRequired" };
    } else if (!normalizeReasoningEffortForPlatform(platform, to)) {
      errors[row.id] = { ...errors[row.id], to: "unsupportedTo" };
    }
  });

  sourceRows.forEach((duplicateRows) => {
    if (duplicateRows.length < 2) return;
    duplicateRows.forEach((row) => {
      errors[row.id] = { ...errors[row.id], from: "duplicateFrom" };
    });
  });

  return errors;
}
