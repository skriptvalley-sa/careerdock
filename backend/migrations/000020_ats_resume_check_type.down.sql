-- Remove resume-only checks and restore original constraints.
DELETE FROM ats_checks WHERE check_type = 'resume';

ALTER TABLE ats_checks DROP CONSTRAINT IF EXISTS ats_checks_check_type_check;
ALTER TABLE ats_checks ADD CONSTRAINT ats_checks_check_type_check
    CHECK (check_type IN ('company', 'job'));

ALTER TABLE ats_checks DROP CONSTRAINT IF EXISTS chk_ats_check_target;
ALTER TABLE ats_checks ADD CONSTRAINT chk_ats_check_target CHECK (
    (check_type = 'company' AND company_id IS NOT NULL) OR
    (check_type = 'job' AND job_description IS NOT NULL)
);
