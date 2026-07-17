# Group Model Whitelist Enforcement Design

## Goal

Use the existing group `models_list_config` as the model access whitelist for OpenAI/Codex API keys. A key inherits the whitelist from its assigned group, so the operator can manage a small number of model plans without configuring every key separately.

## Current Behavior

`models_list_config` currently filters only `/v1/models`. It is already stored on the group, included in API key authentication snapshots, and invalidated when a group changes, but it does not reject requests for models outside the configured list.

## Rules

- Enforcement applies only to groups whose platform is `openai`.
- If the feature is disabled, all models remain allowed.
- If the configured model list is empty, all models remain allowed.
- When enabled with at least one model, matching is exact, case-sensitive, and performed after trimming whitespace.
- The check uses the original client-requested model before channel or account model mapping. Mapping cannot bypass the whitelist.
- Existing groups and keys remain unrestricted until an administrator enables the list and selects models.

## Backend Design

Add a group-level helper that answers whether a requested OpenAI model is allowed by `models_list_config`. Reuse that helper in every OpenAI handler that accepts a model:

- Responses HTTP, including `/responses/compact` and remote compaction body signals
- Responses WebSocket, for the first request and every later turn
- Chat Completions
- Embeddings
- Images generation and edits, using the effective default model when the request omits one
- Other OpenAI/Codex handlers that accept an explicit model, such as alpha search

Run the check immediately after parsing or resolving the requested model and before content moderation, account scheduling, concurrency acquisition, upstream forwarding, or billing. A rejected request therefore consumes no upstream quota and creates no usage record.

The OpenAI `/v1/messages` compatibility endpoint is out of scope because it uses a different client model namespace and dispatch mapping.

## Model Catalogs

The existing `/v1/models` filtering remains unchanged.

Codex Desktop also loads a separate manifest from `/models` or `/backend-api/codex/models`. When the group whitelist is active, parse that manifest and retain only entries whose `slug` is allowed. Because the upstream ETag describes the unfiltered representation, do not forward `If-None-Match` or return the upstream ETag on filtered responses. Returning `200` for the infrequent model refresh is simpler and prevents stale model lists after a group update.

## Errors And Logging

Rejected HTTP requests return an OpenAI-compatible `403 permission_error` with a message that names the requested model and says it is not allowed for the API key's group.

Rejected WebSocket requests use the existing policy-violation close path with the same clear reason. Later turns are checked again so a client cannot open a connection with one model and switch to another.

Write a structured warning containing `api_key_id`, `group_id`, and `requested_model`. Do not log the API key value or request body.

## Admin UI

For OpenAI groups, rename the existing section from `Custom /v1/models Model List` to `Model Whitelist`. Update its description to state that enabled models are both displayed and callable. Keep the existing toggle, selection list, ordering, select-all, and invert controls; no second switch is needed.

For non-OpenAI groups, retain the existing display-only label and behavior.

## Storage And Cache

No database migration is required. The existing `models_list_config` JSON value remains the source of truth, and the API key auth snapshot already carries it. Existing group update code already invalidates authentication caches for keys in that group.

## Verification

Add focused tests for:

- disabled and empty configurations allowing all models
- exact allowed and denied model matches
- denial before scheduling for Responses, Compact, Chat Completions, Embeddings, and Images
- WebSocket denial on both the first and a later turn
- `/v1/models` and Codex manifest filtering
- OpenAI error shape and denial log fields
- non-OpenAI groups retaining their current behavior

Run the relevant Go handler/service tests, frontend group tests, frontend typecheck, full backend tests, frontend build, and `git diff --check` before image construction.

## Deployment

This is an application-only change. Build the Linux `amd64` image locally, preserve the current server image as a rollback tag, load the new image on the server, and recreate only the `sub2api` service. PostgreSQL, Redis, and their volumes are unchanged.
