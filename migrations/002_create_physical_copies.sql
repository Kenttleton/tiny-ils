CREATE TABLE IF NOT EXISTS physical_copies (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    curio_id    UUID NOT NULL REFERENCES curios(id) ON DELETE CASCADE,
    condition   TEXT NOT NULL CHECK (condition IN ('NEW','GOOD','FAIR','POOR','LOST')) DEFAULT 'GOOD',
    location    TEXT NOT NULL DEFAULT '',
    node_id     TEXT NOT NULL DEFAULT '',
    available   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS copies_curio_id_idx ON physical_copies (curio_id);
CREATE INDEX IF NOT EXISTS copies_available_idx ON physical_copies (available);
