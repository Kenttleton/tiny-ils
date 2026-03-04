-- Digital leases are stubbed — access token delivery mechanism is pluggable (TODO).
CREATE TABLE IF NOT EXISTS digital_leases (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id     UUID NOT NULL REFERENCES digital_assets(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL,
    user_node_id TEXT NOT NULL DEFAULT '',
    access_token TEXT NOT NULL DEFAULT '',  -- TODO: pluggable delivery
    issued_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at   TIMESTAMPTZ NOT NULL,
    revoked      BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS leases_asset_id_idx ON digital_leases (asset_id);
CREATE INDEX IF NOT EXISTS leases_user_id_idx ON digital_leases (user_id);
CREATE INDEX IF NOT EXISTS leases_active_idx ON digital_leases (asset_id, expires_at) WHERE revoked = false;
