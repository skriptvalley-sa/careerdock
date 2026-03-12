-- Deferred FK: users.default_resume_id → resumes.id
-- Added after both tables exist to resolve circular reference.
ALTER TABLE users ADD CONSTRAINT fk_users_default_resume
    FOREIGN KEY (default_resume_id) REFERENCES resumes(id) ON DELETE SET NULL;
