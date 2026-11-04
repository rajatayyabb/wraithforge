# WraithForge

> **Observe. Orchestrate. Explain.**
>
> A focused, extensible control plane for running multi-agent workflows, streaming execution logs, tracking task trees, and producing auditable reports.

[![Status: Scaffold](https://img.shields.io/badge/status-scaffold-8b5cf6?style=for-the-badge)](https://github.com/rajatayyabb/wraithforge)
[![Go](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=for-the-badge&logo=react&logoColor=111827)](https://react.dev/)
[![Vite](https://img.shields.io/badge/Vite-5-646CFF?style=for-the-badge&logo=vite&logoColor=white)](https://vitejs.dev/)
[![License](https://img.shields.io/badge/license-not%20set-lightgrey?style=for-the-badge)](#project-status)

WraithForge is a scoped-down foundation for transparent, controllable agent orchestration. It combines a Go API, PostgreSQL-backed workflow data, a sandboxed job runner, live log streaming, a React dashboard, and Markdown reporting in one compact codebase.

## Why WraithForge?

Agent systems are powerful, but power without visibility is difficult to operate. WraithForge is designed around a simple idea: every workflow should be understandable while it is running and explainable after it finishes.

- **See the run** — Stream agent logs to the dashboard over Server-Sent Events.
- **Understand the work** — Browse flows, tasks, and nested subtasks as a live execution tree.
- **Contain the blast radius** — Run generic jobs with Docker-backed timeouts and resource limits.
- **Measure the cost** — Track model usage, token counts, and spend per flow.
- **Keep the paper trail** — Generate Markdown reports from results and cost data.
- **Extend deliberately** — Use the typed API, GraphQL schema, migrations, and taxonomy loader as stable seams for future features.

## Architecture

```mermaid
flowchart LR
    UI[React + Vite dashboard] -->|REST / SSE| API[Go Gin API]
    API --> DB[(PostgreSQL + pgvector)]
    API --> RUN[Sandbox runner]
    RUN --> JOB[Bounded Docker jobs]
    API --> REPORT[Markdown reports]
    API --> TAX[Asset taxonomy]
```

## What is included

### Backend

| Area | Purpose |
| --- | --- |
| `pkg/server` | Gin REST API, bearer/cookie authentication, and SSE log streaming |
| `pkg/database` | PostgreSQL data access with pgx |
| `pkg/sandbox` | Generic Docker job execution with timeouts and resource limits |
| `pkg/report` | Markdown report generation from flow results and cost data |
| `pkg/graph` | GraphQL schema foundation for flows, tasks, results, and costs |
| `pkg/taxonomy` | YAML-based asset taxonomy loader |
| `migrations` | Workflow, task, log, result, token, and cost-tracking tables |

### Frontend

- Flow picker and dashboard shell
- Expandable task/subtask tree
- Live flow terminal for streamed agent logs
- Typed API client
- Tailwind-powered dark interface

## Project layout

```text
wraithforge/
├── backend/
│   ├── cmd/
│   │   ├── orchestrator/    # API server
│   │   ├── seed/            # Sample-data seeder
│   │   ├── report/          # Flow report CLI
│   │   ├── sandboxtest/     # Sandbox smoke test
│   │   └── taxoncheck/      # Taxonomy validator
│   ├── migrations/          # PostgreSQL schema migrations
│   ├── pkg/                 # API, database, sandbox, reports, taxonomy
│   └── taxonomy/            # Taxonomy definitions
├── frontend/
│   ├── src/App.tsx          # Dashboard shell
│   ├── src/features/        # Flows and tasks UI
│   └── src/providers/       # API client
└── README.md
```

## Quick start

### Prerequisites

- Go **1.24+**
- Node.js **22+** and npm
- PostgreSQL with the `pgvector` extension
- Docker, for sandbox execution
- [`goose`](https://github.com/pressly/goose), if you want migration convenience

### 1. Start PostgreSQL

A local `pgvector` container is a convenient development option:

```bash
docker run --name wraithforge-db \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=wraithforge \
  -p 5432:5432 \
  -d pgvector/pgvector:pg16
```

Set the connection string:

```bash
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/wraithforge?sslmode=disable
```

Apply the migrations:

```bash
goose -dir backend/migrations postgres "$DATABASE_URL" up
```

### 2. Seed sample data

```bash
go run ./backend/cmd/seed
```

### 3. Start the API

```bash
export API_TOKEN=dev-token
go run ./backend/cmd/orchestrator
```

The API listens on `http://localhost:8080`.

### 4. Start the dashboard

```bash
cd frontend
npm ci
VITE_API_TOKEN=dev-token npm run dev
```

Open the Vite URL shown in your terminal.

## Useful commands

```bash
# Frontend production build
cd frontend && npm run build

# Run a sandbox smoke test
cd .. && go run ./backend/cmd/sandboxtest -cmd "echo hello from sandbox"

# Render a Markdown report for a flow
cd .. && go run ./backend/cmd/report -flow <flow-id> -out report.md

# Validate the taxonomy file
cd .. && go run ./backend/cmd/taxoncheck
```

## API shape

The dashboard currently uses the REST API for flow and task data and an SSE endpoint for live logs:

```text
GET /api/flows
GET /api/flows/:flow-id/tasks
GET /api/flows/:flow-id/agent-logs/stream?token=<token>
```

The repository also includes a GraphQL SDL foundation for flows, tasks, subtasks, results, and costs.

## Security and scope

WraithForge is intentionally a **generic orchestration scaffold**. It does not include offensive-security logic, exploit automation, credential harvesting, or target-specific attack workflows. The sandbox runner is intended for controlled jobs in environments you own or are authorized to test.

Before deploying beyond local development, replace the development token flow, review Docker isolation settings, lock down CORS and database access, and add production-grade secret management and audit controls.

## Project status

This repository is an early scaffold: the core plumbing and dashboard are in place, while production hardening, broader test coverage, richer workflow controls, and deployment automation remain open areas for development.

The frontend currently builds cleanly with `npm run build`. Backend dependency resolution and compilation should be verified in an environment with normal Go module network access.

## Contributing

Ideas, bug reports, and focused pull requests are welcome. Keep changes small, document behavior changes, and preserve the project's emphasis on visibility, containment, and auditability.

## License

No license has been selected yet. Until one is added, all rights are reserved by the repository owner.
