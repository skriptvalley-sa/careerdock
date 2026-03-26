-- Allow ATS resume checks without a stored resume slot (temp PDF upload path).

-- 1. Drop the NOT NULL constraint and FK on resume_id so temp-upload checks
--    can exist without a resume record.
ALTER TABLE ats_checks ALTER COLUMN resume_id DROP NOT NULL;
ALTER TABLE ats_checks DROP CONSTRAINT IF EXISTS ats_checks_resume_id_fkey;

-- 2. Add a column to store the temporary S3 key used for one-shot PDF uploads.
ALTER TABLE ats_checks ADD COLUMN IF NOT EXISTS temp_s3_key TEXT;

-- 3. Tighten the target constraint so every resume check has at least one
--    PDF source: either a slot resume or a temp upload.
ALTER TABLE ats_checks DROP CONSTRAINT IF EXISTS chk_ats_check_target;
ALTER TABLE ats_checks ADD CONSTRAINT chk_ats_check_target CHECK (
    (check_type = 'company' AND company_id IS NOT NULL) OR
    (check_type = 'job'     AND job_description IS NOT NULL) OR
    (check_type = 'resume'  AND (resume_id IS NOT NULL OR temp_s3_key IS NOT NULL))
);
