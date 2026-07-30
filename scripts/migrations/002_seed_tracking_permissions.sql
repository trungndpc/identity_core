-- Tracking service permissions. Idempotent and safe to run repeatedly.
INSERT INTO permissions (code, name, module, description, created_at)
VALUES
  ('tracking.view', 'View tracking', 'tracking', 'View campaigns, links, QR codes and analytics', NOW()),
  ('tracking.manage', 'Manage tracking', 'tracking', 'Create, update and delete tracking campaigns and links', NOW())
ON CONFLICT (code) DO UPDATE SET
  name = EXCLUDED.name,
  module = EXCLUDED.module,
  description = EXCLUDED.description;
