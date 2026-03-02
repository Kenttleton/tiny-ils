CREATE TABLE IF NOT EXISTS holds (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    curio_id    UUID NOT NULL REFERENCES curios(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL,
    user_node_id TEXT NOT NULL DEFAULT '',
    placed_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at  TIMESTAMPTZ,
    fulfilled   BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS holds_curio_id_idx ON holds (curio_id);
CREATE INDEX IF NOT EXISTS holds_user_id_idx ON holds (user_id);
CREATE INDEX IF NOT EXISTS holds_active_idx ON holds (curio_id, placed_at) WHERE fulfilled = false;
