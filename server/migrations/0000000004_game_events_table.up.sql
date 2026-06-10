CREATE TABLE IF NOT EXISTS game_event (
    seq             BIGSERIAL       NOT NULL PRIMARY KEY,   -- deterministic total-order tiebreak
    key             TEXT            NOT NULL UNIQUE,        -- idempotency / dedup on retries
    type            TEXT            NOT NULL,
    process_after   TIMESTAMPTZ     NOT NULL DEFAULT now(),
    payload         JSONB           NOT NULL DEFAULT '{}'::jsonb,
    processed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT now()
);

-- efficient "due, unprocessed, in submission order" scan for the game loop
CREATE INDEX IF NOT EXISTS game_event_due_idx
    ON game_event (process_after, seq) WHERE processed_at IS NULL;
