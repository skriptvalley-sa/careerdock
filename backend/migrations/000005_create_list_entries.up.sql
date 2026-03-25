CREATE TABLE IF NOT EXISTS list_entries (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id     UUID         NOT NULL REFERENCES user_lists(id) ON DELETE CASCADE,
    company_id  UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    role_title  VARCHAR(255) NOT NULL,
    status      VARCHAR(20)  NOT NULL DEFAULT 'not_applied'
                    CHECK (status IN (
                        'not_applied', 'applied', 'phone_screen', 'interview',
                        'offer', 'rejected', 'accepted', 'withdrawn'
                    )),
    date_applied DATE,
    notes        TEXT,
    position     INTEGER     NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Entries in a list, ordered
CREATE INDEX IF NOT EXISTS idx_list_entries_list ON list_entries (list_id, position ASC);

-- Find all entries for a company across lists
CREATE INDEX IF NOT EXISTS idx_list_entries_company ON list_entries (company_id);
