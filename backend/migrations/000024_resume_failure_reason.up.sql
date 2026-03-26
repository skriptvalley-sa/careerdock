-- Add failure_reason column to resumes for surfacing parse errors to the user.
ALTER TABLE resumes ADD COLUMN IF NOT EXISTS failure_reason TEXT;
