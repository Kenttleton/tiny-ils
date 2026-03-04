CREATE TABLE IF NOT EXISTS copy_transfers (
  id              TEXT PRIMARY KEY,
  global_copy_id  TEXT NOT NULL DEFAULT '',  -- "{home_node_fingerprint}/{copy_uuid}"
  transfer_type   VARCHAR(32) NOT NULL,   -- ILL | RETURN | PERMANENT
  source_node     VARCHAR(255) NOT NULL,  -- node sending the copy
  dest_node       VARCHAR(255) NOT NULL,  -- node receiving the copy
  initiated_by    UUID NOT NULL,          -- manager userId who opened the transfer
  approved_by     UUID,                   -- manager userId who confirmed receipt
  status          VARCHAR(32) NOT NULL DEFAULT 'PENDING',
  notes           TEXT,
  requested_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  approved_at     TIMESTAMPTZ,
  shipped_at      TIMESTAMPTZ,
  received_at     TIMESTAMPTZ
);

-- status values: PENDING | APPROVED | IN_TRANSIT | RECEIVED | REJECTED | CANCELLED

CREATE INDEX IF NOT EXISTS copy_transfers_global_copy_id ON copy_transfers(global_copy_id);

CREATE INDEX IF NOT EXISTS copy_transfers_open
  ON copy_transfers(status)
  WHERE status NOT IN ('RECEIVED', 'CANCELLED', 'REJECTED');
