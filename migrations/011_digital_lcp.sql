-- Migration 011: add Readium LCP columns to digital_assets.

ALTER TABLE digital_assets
  ADD COLUMN IF NOT EXISTS lcp_content_id  VARCHAR(255),
  ADD COLUMN IF NOT EXISTS storage_backend VARCHAR(32) NOT NULL DEFAULT 'local',
  ADD COLUMN IF NOT EXISTS encrypted       BOOLEAN NOT NULL DEFAULT FALSE;

-- storage_backend values: 'local' (LCP-encrypted file), 'provider' (webhook-delivered)
