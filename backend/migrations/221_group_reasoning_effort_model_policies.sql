-- Add exact-model overrides for the existing group reasoning effort policy.
ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS reasoning_effort_model_policies JSONB NOT NULL DEFAULT '[]'::jsonb;
