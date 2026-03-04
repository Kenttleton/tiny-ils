CREATE TABLE IF NOT EXISTS physical_copies (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    curio_id     UUID NOT NULL REFERENCES curios(id) ON DELETE CASCADE,
    condition    TEXT NOT NULL CHECK (condition IN ('NEW','GOOD','FAIR','POOR','LOST')) DEFAULT 'GOOD',
    location     TEXT NOT NULL DEFAULT '',
    node_id      TEXT NOT NULL DEFAULT '',
    home_node_id TEXT NOT NULL DEFAULT '',
    status       VARCHAR(32) NOT NULL DEFAULT 'AVAILABLE',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- status values: AVAILABLE | ON_LOAN | REQUESTED | IN_TRANSIT

CREATE INDEX IF NOT EXISTS copies_curio_id_idx ON physical_copies (curio_id);
CREATE INDEX IF NOT EXISTS physical_copies_status ON physical_copies(status);
