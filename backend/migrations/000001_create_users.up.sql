CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           VARCHAR(255) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    role            VARCHAR(20)  NOT NULL DEFAULT 'user'
                        CHECK (role IN ('user', 'moderator', 'admin')),
    premium_since   TIMESTAMPTZ,
    email_verified  BOOLEAN      NOT NULL DEFAULT FALSE,

    -- Profile fields (used for AI matching)
    current_title       VARCHAR(255),
    experience_level    VARCHAR(20)
                            CHECK (experience_level IN (
                                'fresher', 'junior', 'mid', 'senior', 'staff_plus'
                            )),
    preferred_tech_stacks TEXT[]   DEFAULT '{}',
    target_domains        TEXT[]   DEFAULT '{}',
    target_locations      TEXT[]   DEFAULT '{}',

    -- Default resume for AI-curated lists and ATS pre-selection
    default_resume_id UUID,

    -- Soft delete
    deleted_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_users_email UNIQUE (email)
);

-- Login lookup
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email);

-- Admin user management (filter by role, sort by creation)
CREATE INDEX IF NOT EXISTS idx_users_role_created ON users (role, created_at DESC);

-- Scheduled cleanup of soft-deleted users
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at)
    WHERE deleted_at IS NOT NULL;
