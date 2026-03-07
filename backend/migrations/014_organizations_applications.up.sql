-- Phase 8, Step 1: Organizations & Applications
-- Introduces multi-app structure: User → Org → App → Connection

-- Organizations
CREATE TABLE organizations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    slug       TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_organizations_slug ON organizations(slug);

-- Link users to orgs (nullable: users without an org haven't onboarded yet)
ALTER TABLE users ADD COLUMN org_id UUID REFERENCES organizations(id) ON DELETE SET NULL;
CREATE INDEX idx_users_org_id ON users(org_id);

-- Applications (belong to an org)
CREATE TABLE applications (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_applications_org_id ON applications(org_id);

-- Connections now belong to applications (not directly to users).
-- Clean break: wipe existing test data so NOT NULL is safe.
DELETE FROM log_buffer;
DELETE FROM connections;
ALTER TABLE connections ADD COLUMN app_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE;
CREATE INDEX idx_connections_app_id ON connections(app_id);
