# Wraithforge — scaffold

This is a scoped-down scaffold of the "Wraithforge" spec. It implements the
generic multi-agent orchestration plumbing — database schema, API, sandboxed
job execution, dashboard, and reporting — without any offensive-security
logic. See `WHAT_WAS_NOT_BUILT.pdf` for what was intentionally left out and
why.

## Layout

```
backend/
  migrations/        goose SQL migrations (flows, tasks, subtasks,
                      agent_logs w/ pgvector, api_tokens, results,
                      cost_tracking)
  pkg/database/       pgx-based data access layer
  pkg/server/          Gin REST API (Bearer/cookie auth) + SSE log stream
  pkg/graph/           GraphQL SDL (flows, tasks, subtasks, results, costs)
  pkg/sandbox/          generic Docker job runner (timeout + resource limits)
  pkg/report/          Markdown report generator from Results + cost data
  pkg/taxonomy/        YAML asset-taxonomy loader
  cmd/orchestrator/    API server entrypoint
  cmd/seed/             sample-data seeder
  cmd/sandboxtest/      CLI to exercise the sandbox runner
  cmd/report/            CLI to render a flow's Markdown report
  cmd/taxoncheck/        CLI to validate taxonomy.yaml
frontend/
  src/App.tsx                    flow picker + dashboard layout
  src/features/flows/FlowTerminal.tsx   live SSE log pane
  src/features/tasks/TaskTree.tsx       task/subtask tree
  src/providers/api.ts                  typed API client
```

## Running it

Requires Go 1.24+, Node 22+, and a Postgres instance with the `pgvector`
extension available (e.g. `pgvector/pgvector:pg16` Docker image).

```bash
# 1. Database
createdb wraithforge
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/wraithforge?sslmode=disable
goose -dir backend/migrations postgres "$DATABASE_URL" up   # or run the .sql files manually
go run ./backend/cmd/seed

# 2. API
export API_TOKEN=dev-token
go run ./backend/cmd/orchestrator   # listens on :8080

# 3. Frontend
cd frontend
npm install
VITE_API_TOKEN=dev-token npm run dev
```

Sandbox smoke test (needs Docker):

```bash
go run ./backend/cmd/sandboxtest -cmd "echo hello from sandbox"
```

Markdown report for a flow:

```bash
go run ./backend/cmd/report -flow <flow-id> -out report.md
```

## Status

- Backend compiles clean (`gofmt -l .` empty); dependencies could not be
  fetched in this sandboxed session (Go module proxy returned 403 under the
  network policy here), so `go build` itself was not run — run `go mod tidy
  && go build ./...` in an environment with normal network access to confirm.
- Frontend installs and builds clean (`npm run build` passes, including
  `tsc`).
