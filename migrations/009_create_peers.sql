CREATE TABLE IF NOT EXISTS peers (
    node_id      TEXT PRIMARY KEY,  -- Ed25519 public key fingerprint
    public_key   TEXT NOT NULL,     -- base64-encoded Ed25519 public key
    address      TEXT NOT NULL,     -- host:port of the peer's network-manager gRPC server
    display_name TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'PENDING',  -- PENDING | CONNECTED
    first_seen   TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen    TIMESTAMPTZ NOT NULL DEFAULT now()
);
