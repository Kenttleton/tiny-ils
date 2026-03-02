CREATE TABLE IF NOT EXISTS node_claims (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    node_id     TEXT NOT NULL,  -- node public key fingerprint
    role        TEXT NOT NULL CHECK (role IN ('USER','MANAGER')),
    granted_by  UUID REFERENCES users(id) ON DELETE SET NULL,
    granted_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, node_id)
);

CREATE INDEX IF NOT EXISTS claims_node_id_idx ON node_claims (node_id);
