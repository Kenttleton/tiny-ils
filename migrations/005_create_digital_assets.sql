CREATE TABLE IF NOT EXISTS digital_assets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    curio_id        UUID NOT NULL REFERENCES curios(id) ON DELETE CASCADE,
    format          TEXT NOT NULL,
    file_ref        TEXT NOT NULL,
    checksum        TEXT NOT NULL DEFAULT '',
    max_concurrent  INT NOT NULL DEFAULT 0,  -- 0 = unlimited
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS digital_assets_curio_id_idx ON digital_assets (curio_id);
