CREATE TABLE IF NOT EXISTS curios (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    media_type  TEXT NOT NULL CHECK (media_type IN ('THING','BOOK','VIDEO','AUDIO','GAME')),
    format_type TEXT NOT NULL CHECK (format_type IN ('DIGITAL','PHYSICAL','BOTH')),
    tags        TEXT[] NOT NULL DEFAULT '{}',
    barcode     TEXT,
    qr_code     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS curios_media_type_idx ON curios (media_type);
CREATE INDEX IF NOT EXISTS curios_tags_idx ON curios USING GIN (tags);
