-- Migration 010: replace available boolean with status enum on physical_copies,
-- and introduce the copy_transfers audit table.

-- ─── Step 1: alter physical_copies ───────────────────────────────────────────

ALTER TABLE physical_copies
  ADD COLUMN status VARCHAR(32) NOT NULL DEFAULT 'AVAILABLE';

UPDATE physical_copies
  SET status = CASE WHEN available THEN 'AVAILABLE' ELSE 'ON_LOAN' END;

ALTER TABLE physical_copies DROP COLUMN available;

DROP INDEX IF EXISTS copies_available_idx;

CREATE INDEX IF NOT EXISTS physical_copies_status ON physical_copies(status);

-- status values: AVAILABLE | ON_LOAN | REQUESTED | IN_TRANSIT

-- ─── Step 2: copy_transfers audit table ──────────────────────────────────────

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
