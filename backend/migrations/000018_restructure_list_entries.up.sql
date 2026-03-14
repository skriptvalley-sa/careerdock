-- Add company_status to list_entries (company-level tracking, separate from application status).
ALTER TABLE list_entries ADD COLUMN company_status VARCHAR(20) NOT NULL DEFAULT 'marked'
  CHECK (company_status IN ('marked', 'researching', 'applied', 'interviewing', 'offered', 'accepted', 'rejected'));

-- Make role_title nullable — not required for company-level entries.
ALTER TABLE list_entries ALTER COLUMN role_title DROP NOT NULL;

-- Unique constraint: one company per list.
CREATE UNIQUE INDEX uq_list_entries_list_company ON list_entries (list_id, company_id);

-- Backfill: derive company_status from existing application status.
UPDATE list_entries SET company_status = CASE
  WHEN status IN ('offer', 'accepted') THEN 'offered'
  WHEN status IN ('interview', 'phone_screen') THEN 'interviewing'
  WHEN status = 'applied' THEN 'applied'
  WHEN status = 'rejected' THEN 'rejected'
  ELSE 'marked'
END;
