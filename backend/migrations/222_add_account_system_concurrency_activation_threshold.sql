ALTER TABLE accounts
    ADD COLUMN IF NOT EXISTS system_concurrency_activation_threshold INTEGER;
