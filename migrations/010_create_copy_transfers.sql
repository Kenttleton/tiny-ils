CREATE TABLE IF NOT EXISTS copy_transfers (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  copy_id         UUID NOT NULL REFERENCES physical_copies(id) ON DELETE RESTRICT,
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

CREATE INDEX IF NOT EXISTS copy_transfers_copy_id ON copy_transfers(copy_id);

CREATE INDEX IF NOT EXISTS copy_transfers_open
  ON copy_transfers(status)
  WHERE status NOT IN ('RECEIVED', 'CANCELLED', 'REJECTED');
