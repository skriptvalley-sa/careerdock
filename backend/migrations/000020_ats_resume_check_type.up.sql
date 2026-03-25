-- Allow "resume" as a valid ATS check type (resume-only check, no company/JD needed).
ALTER TABLE ats_checks DROP CONSTRAINT IF EXISTS ats_checks_check_type_check;
ALTER TABLE ats_checks ADD CONSTRAINT ats_checks_check_type_check
    CHECK (check_type IN ('company', 'job', 'resume'));

-- Update the target constraint to allow resume checks with no company_id or job_description.
ALTER TABLE ats_checks DROP CONSTRAINT IF EXISTS chk_ats_check_target;
ALTER TABLE ats_checks ADD CONSTRAINT chk_ats_check_target CHECK (
    (check_type = 'company' AND company_id IS NOT NULL) OR
    (check_type = 'job' AND job_description IS NOT NULL) OR
    (check_type = 'resume')
);
