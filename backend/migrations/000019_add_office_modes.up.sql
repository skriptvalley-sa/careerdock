-- Add office_modes column to companies table.
-- Values: "remote", "hybrid", "onsite" (stored as TEXT[] for companies with multiple modes).
ALTER TABLE companies ADD COLUMN office_modes TEXT[] NOT NULL DEFAULT '{}';

-- GIN index for array containment queries.
CREATE INDEX idx_companies_office_modes ON companies USING gin (office_modes);
