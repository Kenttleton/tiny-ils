CREATE TABLE IF NOT EXISTS physical_loans (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    copy_id          UUID NOT NULL REFERENCES physical_copies(id) ON DELETE RESTRICT,
    user_id          UUID NOT NULL,
    user_node_id     TEXT NOT NULL DEFAULT '',
    checked_out      TIMESTAMPTZ NOT NULL DEFAULT now(),
    due_date         TIMESTAMPTZ NOT NULL,
    returned_at      TIMESTAMPTZ,
    requesting_node  TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS loans_copy_id_idx ON physical_loans (copy_id);
CREATE INDEX IF NOT EXISTS loans_user_id_idx ON physical_loans (user_id);
CREATE INDEX IF NOT EXISTS loans_returned_at_idx ON physical_loans (returned_at) WHERE returned_at IS NULL;
