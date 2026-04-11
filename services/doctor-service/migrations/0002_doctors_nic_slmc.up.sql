-- Upgrade path: add NIC / SLMC when upgrading from an older doctors table without these columns.
ALTER TABLE doctors ADD COLUMN IF NOT EXISTS nic VARCHAR(12);
ALTER TABLE doctors ADD COLUMN IF NOT EXISTS slmc_no VARCHAR(5);

UPDATE doctors SET
    nic = RIGHT(REPEAT('0', 12) || id::text, 12),
    slmc_no = RIGHT(REPEAT('0', 5) || id::text, 5)
WHERE nic IS NULL OR slmc_no IS NULL;

ALTER TABLE doctors ALTER COLUMN nic SET NOT NULL;
ALTER TABLE doctors ALTER COLUMN slmc_no SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_doctors_nic ON doctors (nic);
CREATE UNIQUE INDEX IF NOT EXISTS idx_doctors_slmc_no ON doctors (slmc_no);
