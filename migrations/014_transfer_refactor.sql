-- Transfer ID changes from auto-generated UUID to structured text key:
-- {source_node_fingerprint}/{dest_node_fingerprint}/{uuid_v7}
ALTER TABLE copy_transfers
  ALTER COLUMN id   TYPE TEXT,
  ALTER COLUMN id   DROP DEFAULT;

-- Drop the local copy FK: replaced by global_copy_id which is cross-node safe.
-- global_copy_id = "{home_node_fingerprint}/{copy_uuid}"
ALTER TABLE copy_transfers
  DROP COLUMN copy_id,
  ADD  COLUMN global_copy_id TEXT NOT NULL DEFAULT '';

-- Track the item's home (origin) node separately from its current location.
-- node_id = where the copy physically is now
-- home_node_id = the node that originally owns/cataloged this copy
ALTER TABLE physical_copies
  ADD COLUMN home_node_id TEXT NOT NULL DEFAULT '';

-- Backfill: existing copies belong to their current node
UPDATE physical_copies SET home_node_id = node_id WHERE home_node_id = '';
