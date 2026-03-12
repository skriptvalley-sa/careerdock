CREATE TABLE admin_audit_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action      VARCHAR(255) NOT NULL,
    entity_type VARCHAR(50)  NOT NULL,
    entity_id   UUID,
    details     JSONB,
    ip_address  INET,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Audit log by admin
CREATE INDEX idx_audit_log_admin ON admin_audit_log (admin_id, created_at DESC);

-- Audit log by entity
CREATE INDEX idx_audit_log_entity ON admin_audit_log (entity_type, entity_id, created_at DESC);

-- Time-range queries for admin dashboard
CREATE INDEX idx_audit_log_created ON admin_audit_log (created_at DESC);
