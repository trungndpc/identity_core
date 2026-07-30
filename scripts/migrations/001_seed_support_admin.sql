-- Seed support tenant + admin user (username: admin, password: 123)
-- Idempotent — safe to run multiple times.
-- Password hash: bcrypt cost 10 for "123"

INSERT INTO tenants (code, name, status, metadata, created_at, updated_at)
SELECT
  'support',
  'Support',
  'active',
  '{}'::jsonb,
  NOW(),
  NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM tenants WHERE code = 'support' AND deleted_at IS NULL
);

INSERT INTO users (
  tenant_id,
  full_name,
  display_name,
  username,
  status,
  metadata,
  created_at,
  updated_at
)
SELECT
  t.id,
  'Support Admin',
  'Support Admin',
  'admin',
  'active',
  '{}'::jsonb,
  NOW(),
  NOW()
FROM tenants t
WHERE t.code = 'support'
  AND t.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM users u
    WHERE u.tenant_id = t.id
      AND u.username = 'admin'
      AND u.deleted_at IS NULL
  );

INSERT INTO user_identities (
  user_id,
  provider,
  identity,
  password_hash,
  metadata,
  created_at,
  updated_at
)
SELECT
  u.id,
  'local',
  'admin',
  '$2a$10$auDE9MMYvy5QyekiCCdbouJ1tV02VN/rDCwrp6vcnw6AJF2RVVOwm',
  '{}'::jsonb,
  NOW(),
  NOW()
FROM users u
INNER JOIN tenants t ON t.id = u.tenant_id
WHERE t.code = 'support'
  AND t.deleted_at IS NULL
  AND u.username = 'admin'
  AND u.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM user_identities ui
    WHERE ui.user_id = u.id
      AND ui.provider = 'local'
      AND ui.identity = 'admin'
      AND ui.deleted_at IS NULL
  );

-- Reset password to 123 if identity already exists (re-run updates hash)
UPDATE user_identities ui
SET
  password_hash = '$2a$10$auDE9MMYvy5QyekiCCdbouJ1tV02VN/rDCwrp6vcnw6AJF2RVVOwm',
  updated_at = NOW()
FROM users u
INNER JOIN tenants t ON t.id = u.tenant_id
WHERE ui.user_id = u.id
  AND ui.provider = 'local'
  AND ui.identity = 'admin'
  AND ui.deleted_at IS NULL
  AND u.username = 'admin'
  AND u.deleted_at IS NULL
  AND t.code = 'support'
  AND t.deleted_at IS NULL;
