-- Deferred FK: users.default_resume_id → resumes.id
-- Added after both tables exist to resolve circular reference.
-- Wrapped in a DO block so re-running after a partial failure is safe.
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE constraint_name = 'fk_users_default_resume'
      AND table_name = 'users'
  ) THEN
    ALTER TABLE users ADD CONSTRAINT fk_users_default_resume
      FOREIGN KEY (default_resume_id) REFERENCES resumes(id) ON DELETE SET NULL;
  END IF;
END $$;
