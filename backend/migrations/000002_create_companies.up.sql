CREATE TABLE companies (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug                VARCHAR(255) NOT NULL,
    name                VARCHAR(255) NOT NULL,
    logo_url            VARCHAR(512),
    description         TEXT,
    size                VARCHAR(20)
                            CHECK (size IN (
                                'startup', 'small', 'mid', 'large', 'enterprise'
                            )),
    headquarters        VARCHAR(255),
    founded_year        INTEGER
                            CHECK (founded_year >= 1900 AND founded_year <= 2100),
    careers_page_url    VARCHAR(512),
    glassdoor_url       VARCHAR(512),
    ambitionbox_url     VARCHAR(512),
    linkedin_url        VARCHAR(512),
    tech_stack          TEXT[]       NOT NULL DEFAULT '{}',
    domains             TEXT[]       NOT NULL DEFAULT '{}',
    hiring_status       VARCHAR(20)  NOT NULL DEFAULT 'unknown'
                            CHECK (hiring_status IN ('active', 'paused', 'unknown')),
    interview_patterns  JSONB,
    compensation_tier   VARCHAR(10)
                            CHECK (compensation_tier IN (
                                'tier_1', 'tier_2', 'tier_3', 'tier_4'
                            )),
    has_rsu             BOOLEAN      NOT NULL DEFAULT FALSE,
    has_rsu_refresher   BOOLEAN      NOT NULL DEFAULT FALSE,
    compensation_bands  JSONB,
    last_verified_at    TIMESTAMPTZ,

    -- Full-text search vector (maintained by trigger below)
    search_vector       TSVECTOR,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_companies_slug UNIQUE (slug)
);

-- Trigger function to maintain search_vector on INSERT/UPDATE.
-- to_tsvector is STABLE (not IMMUTABLE) so it cannot be used in a
-- GENERATED ALWAYS AS column; a trigger is the standard alternative.
CREATE OR REPLACE FUNCTION companies_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', coalesce(NEW.name, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(NEW.description, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(array_to_string(NEW.tech_stack, ' '), '')), 'B') ||
        setweight(to_tsvector('english', coalesce(array_to_string(NEW.domains, ' '), '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_companies_search_vector
    BEFORE INSERT OR UPDATE ON companies
    FOR EACH ROW
    EXECUTE FUNCTION companies_search_vector_update();

-- Full-text search
CREATE INDEX idx_companies_search ON companies USING GIN (search_vector);

-- Array containment filters
CREATE INDEX idx_companies_tech_stack ON companies USING GIN (tech_stack);
CREATE INDEX idx_companies_domains ON companies USING GIN (domains);

-- Common filter columns
CREATE INDEX idx_companies_hiring_status ON companies (hiring_status);
CREATE INDEX idx_companies_compensation_tier ON companies (compensation_tier);
CREATE INDEX idx_companies_size ON companies (size);

-- Sorting by update time
CREATE INDEX idx_companies_updated_at ON companies (updated_at DESC);
