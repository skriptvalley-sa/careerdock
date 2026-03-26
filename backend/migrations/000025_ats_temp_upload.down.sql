ALTER TABLE ats_checks DROP COLUMN IF EXISTS temp_s3_key;

ALTER TABLE ats_checks DROP CONSTRAINT IF EXISTS chk_ats_check_target;
ALTER TABLE ats_checks ADD CONSTRAINT chk_ats_check_target CHECK (
    (check_type = 'company' AND company_id IS NOT NULL) OR
    (check_type = 'job'     AND job_description IS NOT NULL) OR
    (check_type = 'resume')
);

ALTER TABLE ats_checks ALTER COLUMN resume_id SET NOT NULL;
ALTER TABLE ats_checks ADD CONSTRAINT ats_checks_resume_id_fkey
    FOREIGN KEY (resume_id) REFERENCES resumes(id) ON DELETE CASCADE;
