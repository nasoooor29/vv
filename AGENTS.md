# Visory Management Users - Agent Guidelines

## Stack
- **Frontend**: bun, React + React Router v7, oRPC, Zustand, Shadcn UI (max 250 lines/component)
- **Backend**: Go + Echo, SQLite, sqlc
- **Auth**: Session cookies, RBAC via comma-separated role strings

## API Communication
- oRPC via `OpenAPILink` at `http://localhost:9999/api`
- Routes: `@frontend/src/lib/routers/`, backend: `@internal/server/routes.go`
- Error handling: oRPC wraps as `HTTPError` schema with toast notification

## RBAC Policies
Available: `docker_read`, `docker_write`, `docker_update`, `docker_delete`, `qemu_read`, `qemu_write`, `qemu_update`, `qemu_delete`, `event_viewer`, `event_manager`, `user_admin` (bypasses all), `settings_manager`, `audit_log_viewer`, `health_checker`

Stored as comma-separated strings in `users.role` field.

## Frontend
- **Routes**: Redirect if user lacks permission (check frontend/src/lib/rbac.ts for some general utils, and frontend/src/components/protected-content.tsx for permission hooks)
- **Components**: Use Shadcn UI, max 250 lines
- **styling**: don't use arbtrary colors just use the shadcn colors and the ones on frontend/src/index.css
- **tanstack query**: auto invalidation is already implemented so no need for .refetch()
- **State**: Zustand stores in `@frontend/src/stores/`
- **Types**: Zod schemas in `@frontend/src/types/index.ts`

## Backend
- **Handlers**: `@internal/server/`, register in `RegisterRoutes()`
- **Middleware**: `Auth()` validates token, `RBAC(...policies)` checks permissions (403 if insufficient)
- **Database**: SQL in `@internal/database/queries/`, run `./scripts/sqlc.sh` after changes
- **Types Gen**: the converted types from go to ts only in only in `./internal/database` and `./internal/models` the others won't
- **Errors**: Use `echo.NewHTTPError(status, "message")`

## New Feature
1. Backend: SQL query → `sqlc.sh` → handler → register route with RBAC
2. Frontend: router definition → Zod schema (auto generated) → component → permission check
3. Test: `go test ./...` and verify 403 on missing permissions

## Key Files
- Frontend client: `@frontend/src/lib/orpc.ts`
- Auth handlers: `@internal/server/auth.go`
- User store: `@frontend/src/stores/user.ts`
- RBAC defs: `@internal/models/general.go`

## external docs
- orpc hub (link that route to other docs): https://orpc.dev/llms.txt
- orpc full: https://orpc.dev/llms-full.txt
- react router v7: https://reactrouter.com/tutorials/quickstart, https://reactrouter.com/how-to/file-route-conventions#setting-up

