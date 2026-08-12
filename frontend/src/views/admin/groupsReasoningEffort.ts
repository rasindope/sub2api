import type {
  GroupPlatform,
  ReasoningEffortMapping,
  ReasoningEffortMatchType,
  ReasoningEffortModelPolicy,
} from "@/types";

const openAIReasoningEffortValues = [
  "minimal",
  "low",
  "medium",
  "high",
  "xhigh",
  "max",
] as const;
const openAIReasoningEffortSourceValues = [
  "none",
  ...openAIReasoningEffortValues,
] as const;

const anthropicReasoningEffortValues = [
  "low",
  "medium",
  "high",
  "xhigh",
  "max",
] as const;

const reasoningEffortMatchTypes: readonly ReasoningEffortMatchType[] = [
  "exact",
  "prefix",
  "suffix",
];

export const reasoningEffortOverLimitDowngrade = "downgrade";
export const reasoningEffortOverLimitDeny = "deny";
export const reasoningEffortMappingDeny = "deny";

const reasoningEffortValuesForPlatform = (
  platform: GroupPlatform,
): readonly string[] =>
  platform === "anthropic"
    ? anthropicReasoningEffortValues
    : supportsReasoningEffortPolicyPlatform(platform)
      ? openAIReasoningEffortValues
      : [];

export function supportsReasoningEffortPolicyPlatform(
  platform: GroupPlatform,
): boolean {
  return platform === "anthropic" || platform === "openai" || platform === "composite";
}

export function reasoningEffortOptionsForPlatform(platform: GroupPlatform) {
  return reasoningEffortValuesForPlatform(platform).map((value) => ({
    value,
    label: value,
  }));
}

export function reasoningEffortSourceOptionsForPlatform(
  platform: GroupPlatform,
) {
  return (supportsReasoningEffortPolicyPlatform(platform)
    ? openAIReasoningEffortSourceValues
    : []
  ).map((value) => ({ value, label: value }));
}

export function reasoningEffortTargetOptionsForPlatform(
  platform: GroupPlatform,
) {
  const options = reasoningEffortOptionsForPlatform(platform);
  if (options.length === 0) return options;
  return [
    ...options,
    { value: reasoningEffortMappingDeny, label: reasoningEffortMappingDeny },
  ];
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

export function normalizeReasoningEffortSourceForPlatform(
  platform: GroupPlatform,
  value: string | null | undefined,
): string {
  const normalized = value?.trim().toLowerCase() ?? "";
  return supportsReasoningEffortPolicyPlatform(platform) &&
    normalized === "none"
    ? "none"
    : normalizeReasoningEffortForPlatform(platform, value);
}

export function normalizeReasoningEffortTargetForPlatform(
  platform: GroupPlatform,
  value: string | null | undefined,
): string {
  const normalized = value?.trim().toLowerCase() ?? "";
  return supportsReasoningEffortPolicyPlatform(platform) &&
    normalized === reasoningEffortMappingDeny
    ? reasoningEffortMappingDeny
    : normalizeReasoningEffortForPlatform(platform, value);
}

export function normalizeReasoningEffortMatchType(
  value: string | null | undefined,
): ReasoningEffortMatchType | "" {
  const normalized = value?.trim().toLowerCase() ?? "";
  return reasoningEffortMatchTypes.some((allowed) => allowed === normalized)
    ? (normalized as ReasoningEffortMatchType)
    : "";
}

export function normalizeReasoningEffortOverLimit(
  value: string | null | undefined,
): string {
  return value?.trim().toLowerCase() === reasoningEffortOverLimitDeny
    ? reasoningEffortOverLimitDeny
    : reasoningEffortOverLimitDowngrade;
}

export interface ReasoningEffortMappingPair {
  id: string;
  from: string;
  to: string;
}

export interface ReasoningEffortMappingRow {
  id: string;
  match_type: ReasoningEffortMatchType | "";
  model: string;
  pairs: ReasoningEffortMappingPair[];
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
  | "duplicateScope"
  | "unsupportedFrom"
  | "unsupportedTo"
  | "unsupportedMatchType";

export type ReasoningEffortMappingErrors = Record<
  string,
  Partial<
    Record<
      "from" | "to" | "match_type" | "duplicateScope",
      ReasoningEffortMappingErrorCode
    >
  >
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

export function createReasoningEffortMappingPair(
  pair: Partial<Pick<ReasoningEffortMapping, "from" | "to">> = {},
): ReasoningEffortMappingPair {
  nextMappingRowID += 1;
  return {
    id: `reasoning-effort-pair-${nextMappingRowID}`,
    from: pair.from ?? "",
    to: pair.to ?? "",
  };
}

export function createReasoningEffortMappingRow(
  mapping: Partial<Omit<ReasoningEffortMapping, "match_type">> & {
    match_type?: ReasoningEffortMatchType | "";
  } = {},
): ReasoningEffortMappingRow {
  nextMappingRowID += 1;
  return {
    id: `reasoning-effort-group-${nextMappingRowID}`,
    match_type: normalizeReasoningEffortMatchType(mapping.match_type),
    model: mapping.model?.trim() ?? "",
    pairs: [
      createReasoningEffortMappingPair({
        from: mapping.from,
        to: mapping.to,
      }),
    ],
  };
}

function mappingScopeKey(
  matchType: string | null | undefined,
  model: string | null | undefined,
): string {
  const trimmedModel = model?.trim().toLowerCase() ?? "";
  const normalizedType = trimmedModel
    ? normalizeReasoningEffortMatchType(matchType) || "exact"
    : "";
  return `${normalizedType}\0${trimmedModel}`;
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
  const groups: ReasoningEffortMappingRow[] = [];
  const indexByScope = new Map<string, number>();

  (mappings ?? []).forEach((mapping) => {
    const from = normalizeReasoningEffortSourceForPlatform(
      platform,
      mapping.from,
    );
    const to = normalizeReasoningEffortTargetForPlatform(platform, mapping.to);
    if (!from || !to) return;

    const matchType = normalizeReasoningEffortMatchType(mapping.match_type);
    const model = mapping.model?.trim() ?? "";
    const scope = mappingScopeKey(matchType, model);
    const existingIndex = indexByScope.get(scope);
    if (existingIndex !== undefined) {
      groups[existingIndex].pairs.push(
        createReasoningEffortMappingPair({ from, to }),
      );
      return;
    }

    const group = createReasoningEffortMappingRow({
      from,
      to,
      match_type: matchType,
      model,
    });
    indexByScope.set(scope, groups.length);
    groups.push(group);
  });

  return groups;
}

export function reasoningEffortMappingsToAPI(
  rows: ReasoningEffortMappingRow[],
): ReasoningEffortMapping[] {
  return rows.flatMap((row) => {
    const model = row.model.trim();
    return row.pairs.map((pair) => {
      const mapping: ReasoningEffortMapping = {
        from: pair.from.trim(),
        to: pair.to.trim(),
      };
      if (model) {
        mapping.match_type =
          normalizeReasoningEffortMatchType(row.match_type) || "exact";
        mapping.model = model;
      }
      return mapping;
    });
  });
}

export function validateReasoningEffortMappings(
  rows: ReasoningEffortMappingRow[],
  platform: GroupPlatform = "openai",
): ReasoningEffortMappingErrors {
  const errors: ReasoningEffortMappingErrors = {};
  const scopeGroups = new Map<string, ReasoningEffortMappingRow[]>();
  const sourcePairs = new Map<string, ReasoningEffortMappingPair[]>();

  rows.forEach((row) => {
    const rawMatchType = row.match_type.trim().toLowerCase();
    const matchType = normalizeReasoningEffortMatchType(row.match_type);
    if (rawMatchType && !matchType) {
      errors[row.id] = { ...errors[row.id], match_type: "unsupportedMatchType" };
    }

    const scope = mappingScopeKey(row.match_type, row.model);
    scopeGroups.set(scope, [...(scopeGroups.get(scope) ?? []), row]);

    row.pairs.forEach((pair) => {
      const from = pair.from.trim();
      const to = pair.to.trim();
      if (!from) {
        errors[pair.id] = { ...errors[pair.id], from: "fromRequired" };
      } else if (!normalizeReasoningEffortSourceForPlatform(platform, from)) {
        errors[pair.id] = { ...errors[pair.id], from: "unsupportedFrom" };
      } else {
        const key = `${scope}\0${from.toLowerCase()}`;
        sourcePairs.set(key, [...(sourcePairs.get(key) ?? []), pair]);
      }
      if (!to) {
        errors[pair.id] = { ...errors[pair.id], to: "toRequired" };
      } else if (!normalizeReasoningEffortTargetForPlatform(platform, to)) {
        errors[pair.id] = { ...errors[pair.id], to: "unsupportedTo" };
      }
    });
  });

  scopeGroups.forEach((duplicateGroups) => {
    if (duplicateGroups.length < 2) return;
    duplicateGroups.forEach((row) => {
      errors[row.id] = { ...errors[row.id], duplicateScope: "duplicateScope" };
    });
  });

  sourcePairs.forEach((duplicatePairs) => {
    if (duplicatePairs.length < 2) return;
    duplicatePairs.forEach((pair) => {
      errors[pair.id] = { ...errors[pair.id], from: "duplicateFrom" };
    });
  });

  return errors;
}
