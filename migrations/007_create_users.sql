CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    display_name  TEXT NOT NULL DEFAULT '',
    password_hash TEXT NOT NULL DEFAULT '',  -- empty for SSO-only accounts
    sso_provider  TEXT NOT NULL DEFAULT '',  -- e.g. 'google'
    sso_subject   TEXT NOT NULL DEFAULT '',  -- provider's unique user ID
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_idx ON users (email);
CREATE UNIQUE INDEX IF NOT EXISTS users_sso_idx ON users (sso_provider, sso_subject)
    WHERE sso_provider != '' AND sso_subject != '';
