# Instructions for coding agents

Read the canonical Galaxy architecture guide before making changes:

- Workspace: `../content_core/docs/SYSTEM_ARCHITECTURE.md`
- GitHub: https://github.com/trungndpc/content_core/blob/master/docs/SYSTEM_ARCHITECTURE.md

For this repository specifically:

- This service is the source of truth for tenants, authentication, users,
  identities, relationships, roles and permissions.
- A root/super-admin bearer token is not an implicit cross-tenant bypass.
- Enforce permissions in middleware and tenant scope in repositories.
- Keep `ADMIN_API_KEY`, `INTERNAL_API_KEY` and JWT secrets backend-only.
- Update `docs/API.md` when an API contract changes.
- Run `go test ./...` before committing or deploying.
