ALTER TABLE doctors 
    ALTER COLUMN specialization DROP NOT NULL,
    ALTER COLUMN experience DROP NOT NULL,
    ALTER COLUMN hospital DROP NOT NULL,
    ALTER COLUMN nic DROP NOT NULL,
    ALTER COLUMN slmc_no DROP NOT NULL;

-- Ensure unique indexes apply only to meaningful values.
-- Skeleton profiles created from auth events may keep NIC/SLMC blank until completed.
DROP INDEX IF EXISTS idx_doctors_nic;
DROP INDEX IF EXISTS idx_doctors_slmc_no;
CREATE UNIQUE INDEX idx_doctors_nic
    ON doctors (nic)
    WHERE nic IS NOT NULL AND BTRIM(nic) <> '';

CREATE UNIQUE INDEX idx_doctors_slmc_no
    ON doctors (slmc_no)
    WHERE slmc_no IS NOT NULL AND BTRIM(slmc_no) <> '';
